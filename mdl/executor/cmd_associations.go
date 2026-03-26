// SPDX-License-Identifier: Apache-2.0

// Package executor - Association commands (SHOW/DESCRIBE/CREATE/DROP ASSOCIATION)
package executor

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/domainmodel"
)

// execCreateAssociation handles CREATE ASSOCIATION statements.
func (e *Executor) execCreateAssociation(s *ast.CreateAssociationStmt) error {
	if e.reader == nil {
		return fmt.Errorf("not connected to a project")
	}

	// Find or auto-create module
	module, err := e.findOrCreateModule(s.Name.Module)
	if err != nil {
		return err
	}

	dm, err := e.reader.GetDomainModel(module.ID)
	if err != nil {
		return fmt.Errorf("failed to get domain model: %w", err)
	}

	// Find parent and child entities (supports cross-module associations)
	parentModule := s.Parent.Module
	if parentModule == "" {
		parentModule = s.Name.Module
	}
	parentEntity, err := e.findEntity(parentModule, s.Parent.Name)
	if err != nil {
		return fmt.Errorf("parent entity not found: %s", s.Parent)
	}
	parentID := parentEntity.ID

	childModule := s.Child.Module
	if childModule == "" {
		childModule = s.Name.Module
	}
	childEntity, err := e.findEntity(childModule, s.Child.Name)
	if err != nil {
		return fmt.Errorf("child entity not found: %s", s.Child)
	}
	childID := childEntity.ID

	// Convert types
	assocType := domainmodel.AssociationTypeReference
	if s.Type == ast.AssocReferenceSet {
		assocType = domainmodel.AssociationTypeReferenceSet
	}

	owner := domainmodel.AssociationOwnerDefault
	switch s.Owner {
	case ast.OwnerBoth:
		owner = domainmodel.AssociationOwnerBoth
	}

	// Convert delete behavior
	var deleteBehavior domainmodel.DeleteBehaviorType
	switch s.DeleteBehavior {
	case ast.DeleteKeepReferences:
		deleteBehavior = domainmodel.DeleteBehaviorTypeDeleteMeButKeepReferences
	case ast.DeleteCascade:
		deleteBehavior = domainmodel.DeleteBehaviorTypeDeleteMeAndReferences
	case ast.DeleteIfNoReferences:
		deleteBehavior = domainmodel.DeleteBehaviorTypeDeleteMeIfNoReferences
	default:
		deleteBehavior = domainmodel.DeleteBehaviorTypeDeleteMeButKeepReferences
	}

	// Convert storage type (default: Column = foreign key in parent table)
	storageFormat := domainmodel.StorageFormatColumn
	switch s.Storage {
	case ast.StorageColumn:
		storageFormat = domainmodel.StorageFormatColumn
	case ast.StorageTable:
		storageFormat = domainmodel.StorageFormatTable
	}

	// Create association
	// ParentID = FROM entity (the one with the FK)
	// ChildID = TO entity (the one being referenced)
	// Cross-module associations use BY_NAME for the child entity
	isCrossModule := parentModule != childModule

	if isCrossModule {
		childRef := childModule + "." + s.Child.Name
		ca := &domainmodel.CrossModuleAssociation{
			Name:          s.Name.Name,
			Type:          assocType,
			Owner:         owner,
			StorageFormat: storageFormat,
			ParentID:      parentID,
			ChildRef:      childRef,
			ChildDeleteBehavior: &domainmodel.DeleteBehavior{
				Type: deleteBehavior,
			},
		}
		if err := e.writer.CreateCrossAssociation(dm.ID, ca); err != nil {
			return fmt.Errorf("failed to create cross-module association: %w", err)
		}
	} else {
		assoc := &domainmodel.Association{
			Name:          s.Name.Name,
			Type:          assocType,
			Owner:         owner,
			StorageFormat: storageFormat,
			ParentID:      parentID,
			ChildID:       childID,
			ChildDeleteBehavior: &domainmodel.DeleteBehavior{
				Type: deleteBehavior,
			},
		}
		if err := e.writer.CreateAssociation(dm.ID, assoc); err != nil {
			return fmt.Errorf("failed to create association: %w", err)
		}
	}

	// Invalidate hierarchy cache so the new association's container is visible
	e.invalidateHierarchy()
	e.invalidateDomainModelsCache()

	// Reconcile MemberAccesses immediately — existing access rules on entities
	// in this DM need MemberAccess entries for the new association (CE0066).
	if freshDM, err := e.reader.GetDomainModel(module.ID); err == nil {
		if count, err := e.writer.ReconcileMemberAccesses(freshDM.ID, module.Name); err == nil && count > 0 {
			fmt.Fprintf(e.output, "Reconciled %d access rule(s) for new association\n", count)
		}
	}

	e.trackModifiedDomainModel(module.ID, module.Name)
	fmt.Fprintf(e.output, "Created association: %s\n", s.Name)
	return nil
}

// execAlterAssociation handles ALTER ASSOCIATION statements.
func (e *Executor) execAlterAssociation(s *ast.AlterAssociationStmt) error {
	if e.reader == nil {
		return fmt.Errorf("not connected to a project")
	}

	module, err := e.findModule(s.Name.Module)
	if err != nil {
		return err
	}

	dm, err := e.reader.GetDomainModel(module.ID)
	if err != nil {
		return fmt.Errorf("failed to get domain model: %w", err)
	}

	// Try intra-module associations first
	for _, assoc := range dm.Associations {
		if assoc.Name == s.Name.Name {
			switch s.Operation {
			case ast.AlterAssociationSetDeleteBehavior:
				assoc.ChildDeleteBehavior = &domainmodel.DeleteBehavior{
					Type: domainmodel.DeleteBehaviorType(s.DeleteBehavior.String()),
				}
			case ast.AlterAssociationSetOwner:
				assoc.Owner = domainmodel.AssociationOwner(s.Owner.String())
			case ast.AlterAssociationSetStorage:
				assoc.StorageFormat = domainmodel.AssociationStorageFormat(s.Storage.String())
			case ast.AlterAssociationSetComment:
				assoc.Documentation = s.Comment
			}
			if err := e.writer.UpdateDomainModel(dm); err != nil {
				return fmt.Errorf("failed to update association: %w", err)
			}
			fmt.Fprintf(e.output, "Altered association: %s\n", s.Name)
			return nil
		}
	}

	// Try cross-module associations
	for _, ca := range dm.CrossAssociations {
		if ca.Name == s.Name.Name {
			switch s.Operation {
			case ast.AlterAssociationSetDeleteBehavior:
				ca.ChildDeleteBehavior = &domainmodel.DeleteBehavior{
					Type: domainmodel.DeleteBehaviorType(s.DeleteBehavior.String()),
				}
			case ast.AlterAssociationSetOwner:
				ca.Owner = domainmodel.AssociationOwner(s.Owner.String())
			case ast.AlterAssociationSetStorage:
				ca.StorageFormat = domainmodel.AssociationStorageFormat(s.Storage.String())
			case ast.AlterAssociationSetComment:
				ca.Documentation = s.Comment
			}
			if err := e.writer.UpdateDomainModel(dm); err != nil {
				return fmt.Errorf("failed to update cross-module association: %w", err)
			}
			fmt.Fprintf(e.output, "Altered association: %s\n", s.Name)
			return nil
		}
	}

	return fmt.Errorf("association not found: %s", s.Name)
}

// execDropAssociation handles DROP ASSOCIATION statements.
func (e *Executor) execDropAssociation(s *ast.DropAssociationStmt) error {
	if e.reader == nil {
		return fmt.Errorf("not connected to a project")
	}

	// Find module
	module, err := e.findModule(s.Name.Module)
	if err != nil {
		return err
	}

	dm, err := e.reader.GetDomainModel(module.ID)
	if err != nil {
		return fmt.Errorf("failed to get domain model: %w", err)
	}

	for _, assoc := range dm.Associations {
		if assoc.Name == s.Name.Name {
			if err := e.writer.DeleteAssociation(dm.ID, assoc.ID); err != nil {
				return fmt.Errorf("failed to delete association: %w", err)
			}
			fmt.Fprintf(e.output, "Dropped association: %s\n", s.Name)
			return nil
		}
	}
	for _, ca := range dm.CrossAssociations {
		if ca.Name == s.Name.Name {
			if err := e.writer.DeleteCrossAssociation(dm.ID, ca.ID); err != nil {
				return fmt.Errorf("failed to delete cross-module association: %w", err)
			}
			fmt.Fprintf(e.output, "Dropped cross-module association: %s\n", s.Name)
			return nil
		}
	}

	return fmt.Errorf("association not found: %s", s.Name)
}

// showAssociations handles SHOW ASSOCIATIONS command.
func (e *Executor) showAssociations(moduleName string) error {
	// Build module ID -> name map (single query)
	modules, err := e.reader.ListModules()
	if err != nil {
		return fmt.Errorf("failed to list modules: %w", err)
	}
	moduleNames := make(map[model.ID]string)
	for _, m := range modules {
		moduleNames[m.ID] = m.Name
	}

	// Get all domain models in a single query (avoids O(n²) behavior)
	domainModels, err := e.reader.ListDomainModels()
	if err != nil {
		return fmt.Errorf("failed to list domain models: %w", err)
	}

	// Build entity ID -> qualified name map
	entityNames := make(map[model.ID]string)
	for _, dm := range domainModels {
		modName := moduleNames[dm.ContainerID]
		for _, entity := range dm.Entities {
			entityNames[entity.ID] = modName + "." + entity.Name
		}
	}

	// Collect rows and calculate column widths
	type row struct {
		qualifiedName string
		module        string
		name          string
		parent        string
		child         string
		assocType     string
		owner         string
		storage       string
	}
	var rows []row
	qnWidth := len("Qualified Name")
	modWidth := len("Module")
	nameWidth := len("Name")
	parentWidth := len("Parent")
	childWidth := len("Child")
	typeWidth := len("Type")
	ownerWidth := len("Owner")
	storageWidth := len("Storage")

	addRow := func(qualifiedName, modName, assocName, parent, child string, assocType, owner, storage string) {
		rows = append(rows, row{qualifiedName, modName, assocName, parent, child, assocType, owner, storage})
		if len(qualifiedName) > qnWidth {
			qnWidth = len(qualifiedName)
		}
		if len(modName) > modWidth {
			modWidth = len(modName)
		}
		if len(assocName) > nameWidth {
			nameWidth = len(assocName)
		}
		if len(parent) > parentWidth {
			parentWidth = len(parent)
		}
		if len(child) > childWidth {
			childWidth = len(child)
		}
		if len(assocType) > typeWidth {
			typeWidth = len(assocType)
		}
		if len(owner) > ownerWidth {
			ownerWidth = len(owner)
		}
		if len(storage) > storageWidth {
			storageWidth = len(storage)
		}
	}

	for _, dm := range domainModels {
		modName := moduleNames[dm.ContainerID]
		// Filter by module name if specified
		if moduleName != "" && modName != moduleName {
			continue
		}
		// Intra-module associations
		for _, assoc := range dm.Associations {
			qualifiedName := modName + "." + assoc.Name
			parent := entityNames[assoc.ParentID]
			child := entityNames[assoc.ChildID]
			if parent == "" {
				parent = string(assoc.ParentID)
			}
			if child == "" {
				child = string(assoc.ChildID)
			}
			addRow(qualifiedName, modName, assoc.Name, parent, child, string(assoc.Type), string(assoc.Owner), string(assoc.StorageFormat))
		}
		// Cross-module associations
		for _, ca := range dm.CrossAssociations {
			qualifiedName := modName + "." + ca.Name
			parent := entityNames[ca.ParentID]
			if parent == "" {
				parent = string(ca.ParentID)
			}
			addRow(qualifiedName, modName, ca.Name, parent, ca.ChildRef, string(ca.Type), string(ca.Owner), string(ca.StorageFormat))
		}
	}

	// Sort by qualified name
	sort.Slice(rows, func(i, j int) bool {
		return strings.ToLower(rows[i].qualifiedName) < strings.ToLower(rows[j].qualifiedName)
	})

	// Markdown table with aligned columns
	fmt.Fprintf(e.output, "| %-*s | %-*s | %-*s | %-*s | %-*s | %-*s | %-*s | %-*s |\n",
		qnWidth, "Qualified Name", modWidth, "Module", nameWidth, "Name",
		parentWidth, "Parent", childWidth, "Child", typeWidth, "Type", ownerWidth, "Owner", storageWidth, "Storage")
	fmt.Fprintf(e.output, "|-%s-|-%s-|-%s-|-%s-|-%s-|-%s-|-%s-|-%s-|\n",
		strings.Repeat("-", qnWidth), strings.Repeat("-", modWidth), strings.Repeat("-", nameWidth),
		strings.Repeat("-", parentWidth), strings.Repeat("-", childWidth),
		strings.Repeat("-", typeWidth), strings.Repeat("-", ownerWidth), strings.Repeat("-", storageWidth))
	for _, r := range rows {
		fmt.Fprintf(e.output, "| %-*s | %-*s | %-*s | %-*s | %-*s | %-*s | %-*s | %-*s |\n",
			qnWidth, r.qualifiedName, modWidth, r.module, nameWidth, r.name,
			parentWidth, r.parent, childWidth, r.child, typeWidth, r.assocType, ownerWidth, r.owner, storageWidth, r.storage)
	}
	fmt.Fprintf(e.output, "\n(%d associations)\n", len(rows))
	return nil
}

// showAssociation handles SHOW ASSOCIATION command.
func (e *Executor) showAssociation(name *ast.QualifiedName) error {
	if name == nil {
		return fmt.Errorf("association name required")
	}

	module, err := e.findModule(name.Module)
	if err != nil {
		return err
	}

	dm, err := e.reader.GetDomainModel(module.ID)
	if err != nil {
		return fmt.Errorf("failed to get domain model: %w", err)
	}

	for _, assoc := range dm.Associations {
		if assoc.Name == name.Name {
			fmt.Fprintf(e.output, "Association: %s.%s\n", module.Name, assoc.Name)
			fmt.Fprintf(e.output, "  Type: %s\n", assoc.Type)
			fmt.Fprintf(e.output, "  Owner: %s\n", assoc.Owner)
			fmt.Fprintf(e.output, "  Storage: %s\n", assoc.StorageFormat)
			return nil
		}
	}
	for _, ca := range dm.CrossAssociations {
		if ca.Name == name.Name {
			fmt.Fprintf(e.output, "Association: %s.%s (cross-module)\n", module.Name, ca.Name)
			fmt.Fprintf(e.output, "  Type: %s\n", ca.Type)
			fmt.Fprintf(e.output, "  Owner: %s\n", ca.Owner)
			fmt.Fprintf(e.output, "  Storage: %s\n", ca.StorageFormat)
			fmt.Fprintf(e.output, "  Child: %s\n", ca.ChildRef)
			return nil
		}
	}

	return fmt.Errorf("association not found: %s", name)
}

// describeAssociation handles DESCRIBE ASSOCIATION command.
func (e *Executor) describeAssociation(name ast.QualifiedName) error {
	module, err := e.findModule(name.Module)
	if err != nil {
		return err
	}

	dm, err := e.reader.GetDomainModel(module.ID)
	if err != nil {
		return fmt.Errorf("failed to get domain model: %w", err)
	}

	// Build entity ID -> qualified name map across all modules
	entityNames := make(map[model.ID]string)
	allDomainModels, err := e.reader.ListDomainModels()
	if err != nil {
		return fmt.Errorf("failed to list domain models: %w", err)
	}
	h, err := e.getHierarchy()
	if err != nil {
		return fmt.Errorf("failed to build hierarchy: %w", err)
	}
	for _, otherDM := range allDomainModels {
		modName := h.GetModuleName(otherDM.ContainerID)
		for _, entity := range otherDM.Entities {
			entityNames[entity.ID] = modName + "." + entity.Name
		}
	}

	// Helper to format association type, owner, storage, and delete behavior
	formatAssocDetails := func(assocType domainmodel.AssociationType, assocOwner domainmodel.AssociationOwner, storageFormat domainmodel.AssociationStorageFormat, childDeleteBehavior *domainmodel.DeleteBehavior) {
		typeName := "Reference"
		if assocType == domainmodel.AssociationTypeReferenceSet {
			typeName = "ReferenceSet"
		}
		fmt.Fprintf(e.output, "TYPE %s\n", typeName)

		owner := "Default"
		if assocOwner == domainmodel.AssociationOwnerBoth {
			owner = "Both"
		}
		fmt.Fprintf(e.output, "OWNER %s\n", owner)

		// Only output STORAGE when it's not the default (Table)
		if storageFormat == domainmodel.StorageFormatColumn {
			fmt.Fprintf(e.output, "STORAGE COLUMN\n")
		}

		deleteBehavior := "DELETE_BUT_KEEP_REFERENCES"
		if childDeleteBehavior != nil {
			switch childDeleteBehavior.Type {
			case domainmodel.DeleteBehaviorTypeDeleteMeAndReferences:
				deleteBehavior = "DELETE_CASCADE"
			case domainmodel.DeleteBehaviorTypeDeleteMeIfNoReferences:
				deleteBehavior = "DELETE_IF_NO_REFERENCES"
			case domainmodel.DeleteBehaviorTypeDeleteMeButKeepReferences:
				deleteBehavior = "DELETE_BUT_KEEP_REFERENCES"
			}
		}
		fmt.Fprintf(e.output, "DELETE_BEHAVIOR %s;\n", deleteBehavior)
	}

	for _, assoc := range dm.Associations {
		if assoc.Name == name.Name {
			fromEntity := entityNames[assoc.ParentID]
			toEntity := entityNames[assoc.ChildID]

			if assoc.Documentation != "" {
				fmt.Fprintf(e.output, "/**\n * %s\n */\n", assoc.Documentation)
			}

			fmt.Fprintf(e.output, "CREATE ASSOCIATION %s.%s\n", module.Name, assoc.Name)
			fmt.Fprintf(e.output, "FROM %s TO %s\n", fromEntity, toEntity)
			formatAssocDetails(assoc.Type, assoc.Owner, assoc.StorageFormat, assoc.ChildDeleteBehavior)
			fmt.Fprintln(e.output, "/")
			return nil
		}
	}
	for _, ca := range dm.CrossAssociations {
		if ca.Name == name.Name {
			fromEntity := entityNames[ca.ParentID]
			if fromEntity == "" {
				fromEntity = string(ca.ParentID)
			}

			if ca.Documentation != "" {
				fmt.Fprintf(e.output, "/**\n * %s\n */\n", ca.Documentation)
			}

			fmt.Fprintf(e.output, "CREATE ASSOCIATION %s.%s\n", module.Name, ca.Name)
			fmt.Fprintf(e.output, "FROM %s TO %s\n", fromEntity, ca.ChildRef)
			formatAssocDetails(ca.Type, ca.Owner, ca.StorageFormat, ca.ChildDeleteBehavior)
			fmt.Fprintln(e.output, "/")
			return nil
		}
	}

	return fmt.Errorf("association not found: %s", name)
}
