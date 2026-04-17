// SPDX-License-Identifier: Apache-2.0

// Package executor — RENAME commands (entity, module)
package executor

import (
	"fmt"
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
	mdlerrors "github.com/mendixlabs/mxcli/mdl/errors"
	"github.com/mendixlabs/mxcli/sdk/mpr"
)

// execRename handles RENAME statements for all document types.
func execRename(ctx *ExecContext, s *ast.RenameStmt) error {
	e := ctx.executor
	if e.reader == nil {
		return mdlerrors.NewNotConnected()
	}

	switch s.ObjectType {
	case "ENTITY":
		return execRenameEntity(ctx, s)
	case "MICROFLOW":
		return execRenameDocument(ctx, s, "microflow")
	case "NANOFLOW":
		return execRenameDocument(ctx, s, "nanoflow")
	case "PAGE":
		return execRenameDocument(ctx, s, "page")
	case "ENUMERATION":
		return execRenameEnumeration(ctx, s)
	case "ASSOCIATION":
		return execRenameAssociation(ctx, s)
	case "CONSTANT":
		return execRenameDocument(ctx, s, "constant")
	case "MODULE":
		return execRenameModule(ctx, s)
	default:
		return mdlerrors.NewUnsupported(fmt.Sprintf("RENAME not supported for %s", s.ObjectType))
	}
}

// execRenameEntity renames an entity and updates all BY_NAME references.
func execRenameEntity(ctx *ExecContext, s *ast.RenameStmt) error {
	e := ctx.executor
	// Find the entity
	module, err := findModule(ctx, s.Name.Module)
	if err != nil {
		return err
	}

	dm, err := e.reader.GetDomainModel(module.ID)
	if err != nil {
		return mdlerrors.NewBackend("get domain model", err)
	}

	found := false
	for _, ent := range dm.Entities {
		if ent.Name == s.Name.Name {
			found = true
			break
		}
	}
	if !found {
		return mdlerrors.NewNotFound("entity", s.Name.String())
	}

	oldQualifiedName := s.Name.Module + "." + s.Name.Name
	newQualifiedName := s.Name.Module + "." + s.NewName

	// Scan for references
	hits, err := e.writer.RenameReferences(oldQualifiedName, newQualifiedName, s.DryRun)
	if err != nil {
		return mdlerrors.NewBackend("scan references", err)
	}

	if s.DryRun {
		printRenameReport(ctx, oldQualifiedName, newQualifiedName, hits)
		return nil
	}

	// Update the entity name in the domain model
	for _, ent := range dm.Entities {
		if ent.Name == s.Name.Name {
			ent.Name = s.NewName
			break
		}
	}
	if err := e.writer.UpdateDomainModel(dm); err != nil {
		return mdlerrors.NewBackend("update entity name", err)
	}

	invalidateHierarchy(ctx)
	invalidateDomainModelsCache(ctx)

	fmt.Fprintf(ctx.Output, "Renamed entity: %s → %s\n", oldQualifiedName, newQualifiedName)
	if len(hits) > 0 {
		fmt.Fprintf(ctx.Output, "Updated %d reference(s) in %d document(s)\n", totalRefCount(hits), len(hits))
	}
	return nil
}

// execRenameModule renames a module and updates all BY_NAME references with the module prefix.
func execRenameModule(ctx *ExecContext, s *ast.RenameStmt) error {
	e := ctx.executor
	oldModuleName := s.Name.Module
	newModuleName := s.NewName

	module, err := findModule(ctx, oldModuleName)
	if err != nil {
		return err
	}

	// Scan for all references with the old module prefix
	// Module rename replaces "OldModule." with "NewModule." in all qualified names
	hits, err := e.writer.RenameReferences(oldModuleName+".", newModuleName+".", s.DryRun)
	if err != nil {
		return mdlerrors.NewBackend("scan references", err)
	}

	// Also scan for exact module name matches (e.g., in navigation, security role refs)
	exactHits, err := e.writer.RenameReferences(oldModuleName, newModuleName, s.DryRun)
	if err != nil {
		return mdlerrors.NewBackend("scan exact module references", err)
	}

	// Merge hit lists (deduplicate by unit ID)
	allHits := mergeHits(hits, exactHits)

	if s.DryRun {
		printRenameReport(ctx, oldModuleName, newModuleName, allHits)
		return nil
	}

	// Update the module name
	module.Name = newModuleName
	if err := e.writer.UpdateModule(module); err != nil {
		return mdlerrors.NewBackend("update module name", err)
	}

	invalidateHierarchy(ctx)
	invalidateDomainModelsCache(ctx)

	fmt.Fprintf(ctx.Output, "Renamed module: %s → %s\n", oldModuleName, newModuleName)
	if len(allHits) > 0 {
		fmt.Fprintf(ctx.Output, "Updated %d reference(s) in %d document(s)\n", totalRefCount(allHits), len(allHits))
	}
	return nil
}

// execRenameDocument handles RENAME MICROFLOW/NANOFLOW/PAGE/CONSTANT.
// These are standalone documents where the Name field is in the document BSON itself.
// The reference scanner handles updating all BY_NAME references, and then we update
// the document's own Name field via a raw BSON rewrite.
func execRenameDocument(ctx *ExecContext, s *ast.RenameStmt, docType string) error {
	e := ctx.executor
	oldQualifiedName := s.Name.Module + "." + s.Name.Name
	newQualifiedName := s.Name.Module + "." + s.NewName

	// Verify the document exists
	h, err := getHierarchy(ctx)
	if err != nil {
		return err
	}

	found := false
	switch docType {
	case "microflow":
		mfs, _ := e.reader.ListMicroflows()
		for _, mf := range mfs {
			modID := h.FindModuleID(mf.ContainerID)
			if h.GetModuleName(modID) == s.Name.Module && mf.Name == s.Name.Name {
				found = true
				break
			}
		}
	case "nanoflow":
		nfs, _ := e.reader.ListNanoflows()
		for _, nf := range nfs {
			modID := h.FindModuleID(nf.ContainerID)
			if h.GetModuleName(modID) == s.Name.Module && nf.Name == s.Name.Name {
				found = true
				break
			}
		}
	case "page":
		pgs, _ := e.reader.ListPages()
		for _, pg := range pgs {
			modID := h.FindModuleID(pg.ContainerID)
			if h.GetModuleName(modID) == s.Name.Module && pg.Name == s.Name.Name {
				found = true
				break
			}
		}
	case "constant":
		cs, _ := e.reader.ListConstants()
		for _, c := range cs {
			modID := h.FindModuleID(c.ContainerID)
			if h.GetModuleName(modID) == s.Name.Module && c.Name == s.Name.Name {
				found = true
				break
			}
		}
	}

	if !found {
		return mdlerrors.NewNotFound(s.ObjectType, oldQualifiedName)
	}

	// The reference scanner will also update the document's own Name field
	// when it matches the old qualified name. But the Name field is just the
	// simple name (e.g., "OldName"), not the qualified name. So we need to
	// handle it separately — the scanner updates cross-references, and we
	// update the Name field directly.
	hits, err := e.writer.RenameReferences(oldQualifiedName, newQualifiedName, s.DryRun)
	if err != nil {
		return mdlerrors.NewBackend("scan references", err)
	}

	if s.DryRun {
		printRenameReport(ctx, oldQualifiedName, newQualifiedName, hits)
		return nil
	}

	// Update the document's own Name field via the raw BSON name updater
	if err := e.writer.RenameDocumentByName(s.Name.Module, s.Name.Name, s.NewName); err != nil {
		return mdlerrors.NewBackend(fmt.Sprintf("rename %s", docType), err)
	}

	invalidateHierarchy(ctx)

	fmt.Fprintf(ctx.Output, "Renamed %s: %s → %s\n", docType, oldQualifiedName, newQualifiedName)
	if len(hits) > 0 {
		fmt.Fprintf(ctx.Output, "Updated %d reference(s) in %d document(s)\n", totalRefCount(hits), len(hits))
	}
	return nil
}

// execRenameEnumeration renames an enumeration and updates all references.
func execRenameEnumeration(ctx *ExecContext, s *ast.RenameStmt) error {
	e := ctx.executor
	oldQualifiedName := s.Name.Module + "." + s.Name.Name
	newQualifiedName := s.Name.Module + "." + s.NewName

	// Verify it exists
	enums, err := e.reader.ListEnumerations()
	if err != nil {
		return mdlerrors.NewBackend("list enumerations", err)
	}
	h, err := getHierarchy(ctx)
	if err != nil {
		return err
	}
	found := false
	for _, en := range enums {
		modID := h.FindModuleID(en.ContainerID)
		if h.GetModuleName(modID) == s.Name.Module && en.Name == s.Name.Name {
			found = true
			break
		}
	}
	if !found {
		return mdlerrors.NewNotFound("enumeration", oldQualifiedName)
	}

	hits, err := e.writer.RenameReferences(oldQualifiedName, newQualifiedName, s.DryRun)
	if err != nil {
		return mdlerrors.NewBackend("scan references", err)
	}

	if s.DryRun {
		printRenameReport(ctx, oldQualifiedName, newQualifiedName, hits)
		return nil
	}

	// Update enumeration name via raw BSON
	if err := e.writer.RenameDocumentByName(s.Name.Module, s.Name.Name, s.NewName); err != nil {
		return mdlerrors.NewBackend("rename enumeration", err)
	}

	// Also update enumeration refs in domain models (attribute types store qualified enum names)
	e.writer.UpdateEnumerationRefsInAllDomainModels(oldQualifiedName, newQualifiedName)

	invalidateHierarchy(ctx)
	invalidateDomainModelsCache(ctx)

	fmt.Fprintf(ctx.Output, "Renamed enumeration: %s → %s\n", oldQualifiedName, newQualifiedName)
	if len(hits) > 0 {
		fmt.Fprintf(ctx.Output, "Updated %d reference(s) in %d document(s)\n", totalRefCount(hits), len(hits))
	}
	return nil
}

// execRenameAssociation renames an association and updates all references.
func execRenameAssociation(ctx *ExecContext, s *ast.RenameStmt) error {
	e := ctx.executor
	oldQualifiedName := s.Name.Module + "." + s.Name.Name
	newQualifiedName := s.Name.Module + "." + s.NewName

	module, err := findModule(ctx, s.Name.Module)
	if err != nil {
		return err
	}

	dm, err := e.reader.GetDomainModel(module.ID)
	if err != nil {
		return mdlerrors.NewBackend("get domain model", err)
	}

	found := false
	for _, assoc := range dm.Associations {
		if assoc.Name == s.Name.Name {
			found = true
			break
		}
	}
	if !found {
		return mdlerrors.NewNotFound("association", oldQualifiedName)
	}

	hits, err := e.writer.RenameReferences(oldQualifiedName, newQualifiedName, s.DryRun)
	if err != nil {
		return mdlerrors.NewBackend("scan references", err)
	}

	if s.DryRun {
		printRenameReport(ctx, oldQualifiedName, newQualifiedName, hits)
		return nil
	}

	// Update association name in domain model
	for _, assoc := range dm.Associations {
		if assoc.Name == s.Name.Name {
			assoc.Name = s.NewName
			break
		}
	}
	if err := e.writer.UpdateDomainModel(dm); err != nil {
		return mdlerrors.NewBackend("update association name", err)
	}

	invalidateHierarchy(ctx)
	invalidateDomainModelsCache(ctx)

	fmt.Fprintf(ctx.Output, "Renamed association: %s → %s\n", oldQualifiedName, newQualifiedName)
	if len(hits) > 0 {
		fmt.Fprintf(ctx.Output, "Updated %d reference(s) in %d document(s)\n", totalRefCount(hits), len(hits))
	}
	return nil
}

// printRenameReport outputs a dry-run report of what would change.
func printRenameReport(ctx *ExecContext, oldName, newName string, hits []mpr.RenameHit) {
	fmt.Fprintf(ctx.Output, "Would rename: %s → %s\n", oldName, newName)
	fmt.Fprintf(ctx.Output, "References found: %d in %d document(s)\n", totalRefCount(hits), len(hits))

	for _, h := range hits {
		label := h.Name
		if label == "" {
			label = h.UnitID
		}
		typeName := h.UnitType
		if idx := strings.Index(typeName, "$"); idx >= 0 {
			typeName = typeName[idx+1:]
		}
		fmt.Fprintf(ctx.Output, "  %s (%s) — %d reference(s)\n", label, typeName, h.Count)
	}
}

func totalRefCount(hits []mpr.RenameHit) int {
	total := 0
	for _, h := range hits {
		total += h.Count
	}
	return total
}

func mergeHits(a, b []mpr.RenameHit) []mpr.RenameHit {
	seen := make(map[string]int) // unitID → index in result
	result := make([]mpr.RenameHit, len(a))
	copy(result, a)
	for i := range result {
		seen[result[i].UnitID] = i
	}
	for _, h := range b {
		if idx, ok := seen[h.UnitID]; ok {
			result[idx].Count += h.Count
		} else {
			seen[h.UnitID] = len(result)
			result = append(result, h)
		}
	}
	return result
}
