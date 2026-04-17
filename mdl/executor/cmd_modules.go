// SPDX-License-Identifier: Apache-2.0

// Package executor - Module commands (SHOW/CREATE/DROP MODULES)
package executor

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
	mdlerrors "github.com/mendixlabs/mxcli/mdl/errors"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/domainmodel"
)

// execCreateModule handles CREATE MODULE statements.
func execCreateModule(ctx *ExecContext, s *ast.CreateModuleStmt) error {
	e := ctx.executor
	if e.reader == nil {
		return mdlerrors.NewNotConnected()
	}

	// Check if module already exists
	modules, err := e.reader.ListModules()
	if err != nil {
		return mdlerrors.NewBackend("list modules", err)
	}

	for _, m := range modules {
		if m.Name == s.Name {
			fmt.Fprintf(ctx.Output, "Module '%s' already exists\n", s.Name)
			return nil
		}
	}

	// Create the module
	module := &model.Module{
		Name: s.Name,
	}

	if err := e.writer.CreateModule(module); err != nil {
		return mdlerrors.NewBackend("create module", err)
	}

	// Invalidate cache so new module is visible
	invalidateModuleCache(ctx)

	fmt.Fprintf(ctx.Output, "Created module: %s\n", s.Name)
	return nil
}

// execDropModule handles DROP MODULE statements.
// This cascades to delete all objects inside the module:
// - Enumerations
// - Entities and Associations (in domain models)
// - Microflows
// - Nanoflows
// - Pages
// - Snippets
// - Constants
func execDropModule(ctx *ExecContext, s *ast.DropModuleStmt) error {
	e := ctx.executor
	if e.reader == nil {
		return mdlerrors.NewNotConnected()
	}

	// Find the module
	modules, err := e.reader.ListModules()
	if err != nil {
		return mdlerrors.NewBackend("list modules", err)
	}

	var targetModule *model.Module
	for _, m := range modules {
		if m.Name == s.Name {
			targetModule = m
			break
		}
	}

	if targetModule == nil {
		return mdlerrors.NewNotFound("module", s.Name)
	}

	// Build set of all container IDs belonging to this module (including nested folders)
	moduleContainers := getModuleContainers(ctx, targetModule.ID)

	// Counters for summary
	var nEnums, nEntities, nAssocs, nMicroflows, nNanoflows, nPages, nSnippets, nLayouts, nConstants, nJavaActions, nServices, nBizEvents, nDbConns int

	// Delete enumerations in this module
	if enums, err := e.reader.ListEnumerations(); err == nil {
		for _, enum := range enums {
			if moduleContainers[enum.ContainerID] {
				if err := e.writer.DeleteEnumeration(enum.ID); err != nil {
					fmt.Fprintf(ctx.Output, "Warning: failed to delete enumeration %s: %v\n", enum.Name, err)
				} else {
					nEnums++
				}
			}
		}
	}

	// Delete entities in domain models belonging to this module
	if dms, err := e.reader.ListDomainModels(); err == nil {
		for _, dm := range dms {
			if moduleContainers[dm.ContainerID] {
				// Delete all associations in this domain model first (they reference entities)
				for _, assoc := range dm.Associations {
					if err := e.writer.DeleteAssociation(dm.ID, assoc.ID); err != nil {
						fmt.Fprintf(ctx.Output, "Warning: failed to delete association %s: %v\n", assoc.Name, err)
					} else {
						nAssocs++
					}
				}
				// Delete all entities in this domain model
				for _, entity := range dm.Entities {
					if err := e.writer.DeleteEntity(dm.ID, entity.ID); err != nil {
						fmt.Fprintf(ctx.Output, "Warning: failed to delete entity %s: %v\n", entity.Name, err)
					} else {
						nEntities++
					}
				}
			}
		}
	}

	// Delete microflows in this module
	if mfs, err := e.reader.ListMicroflows(); err == nil {
		for _, mf := range mfs {
			if moduleContainers[mf.ContainerID] {
				if err := e.writer.DeleteMicroflow(mf.ID); err != nil {
					fmt.Fprintf(ctx.Output, "Warning: failed to delete microflow %s: %v\n", mf.Name, err)
				} else {
					nMicroflows++
				}
			}
		}
	}

	// Delete nanoflows in this module
	if nfs, err := e.reader.ListNanoflows(); err == nil {
		for _, nf := range nfs {
			if moduleContainers[nf.ContainerID] {
				if err := e.writer.DeleteNanoflow(nf.ID); err != nil {
					fmt.Fprintf(ctx.Output, "Warning: failed to delete nanoflow %s: %v\n", nf.Name, err)
				} else {
					nNanoflows++
				}
			}
		}
	}

	// Delete pages in this module
	if pages, err := e.reader.ListPages(); err == nil {
		for _, page := range pages {
			if moduleContainers[page.ContainerID] {
				if err := e.writer.DeletePage(page.ID); err != nil {
					fmt.Fprintf(ctx.Output, "Warning: failed to delete page %s: %v\n", page.Name, err)
				} else {
					nPages++
				}
			}
		}
	}

	// Delete snippets in this module
	if snippets, err := e.reader.ListSnippets(); err == nil {
		for _, snippet := range snippets {
			if moduleContainers[snippet.ContainerID] {
				if err := e.writer.DeleteSnippet(snippet.ID); err != nil {
					fmt.Fprintf(ctx.Output, "Warning: failed to delete snippet %s: %v\n", snippet.Name, err)
				} else {
					nSnippets++
				}
			}
		}
	}

	// Delete constants in this module
	if constants, err := e.reader.ListConstants(); err == nil {
		for _, c := range constants {
			if moduleContainers[c.ContainerID] {
				if err := e.writer.DeleteConstant(c.ID); err != nil {
					fmt.Fprintf(ctx.Output, "Warning: failed to delete constant %s: %v\n", c.Name, err)
				} else {
					nConstants++
				}
			}
		}
	}

	// Delete layouts in this module
	if layouts, err := e.reader.ListLayouts(); err == nil {
		for _, l := range layouts {
			if moduleContainers[l.ContainerID] {
				if err := e.writer.DeleteLayout(l.ID); err != nil {
					fmt.Fprintf(ctx.Output, "Warning: failed to delete layout %s: %v\n", l.Name, err)
				} else {
					nLayouts++
				}
			}
		}
	}

	// Delete Java actions in this module
	if jas, err := e.reader.ListJavaActions(); err == nil {
		for _, ja := range jas {
			if moduleContainers[ja.ContainerID] {
				if err := e.writer.DeleteJavaAction(ja.ID); err != nil {
					fmt.Fprintf(ctx.Output, "Warning: failed to delete Java action %s: %v\n", ja.Name, err)
				} else {
					nJavaActions++
				}
			}
		}
	}

	// Delete business event services in this module
	if services, err := e.reader.ListBusinessEventServices(); err == nil {
		for _, svc := range services {
			if moduleContainers[svc.ContainerID] {
				if err := e.writer.DeleteBusinessEventService(svc.ID); err != nil {
					fmt.Fprintf(ctx.Output, "Warning: failed to delete business event service %s: %v\n", svc.Name, err)
				} else {
					nBizEvents++
				}
			}
		}
	}

	// Delete database connections in this module
	if conns, err := e.reader.ListDatabaseConnections(); err == nil {
		for _, conn := range conns {
			if moduleContainers[conn.ContainerID] {
				if err := e.writer.DeleteDatabaseConnection(conn.ID); err != nil {
					fmt.Fprintf(ctx.Output, "Warning: failed to delete database connection %s: %v\n", conn.Name, err)
				} else {
					nDbConns++
				}
			}
		}
	}

	// Delete consumed OData services (clients) in this module
	if services, err := e.reader.ListConsumedODataServices(); err == nil {
		for _, svc := range services {
			if moduleContainers[svc.ContainerID] {
				if err := e.writer.DeleteConsumedODataService(svc.ID); err != nil {
					fmt.Fprintf(ctx.Output, "Warning: failed to delete OData client %s: %v\n", svc.Name, err)
				} else {
					nServices++
				}
			}
		}
	}

	// Delete published OData services in this module
	if services, err := e.reader.ListPublishedODataServices(); err == nil {
		for _, svc := range services {
			if moduleContainers[svc.ContainerID] {
				if err := e.writer.DeletePublishedODataService(svc.ID); err != nil {
					fmt.Fprintf(ctx.Output, "Warning: failed to delete OData service %s: %v\n", svc.Name, err)
				} else {
					nServices++
				}
			}
		}
	}

	// Remove module roles from user roles in ProjectSecurity
	if ms, err := e.reader.GetModuleSecurity(targetModule.ID); err == nil {
		if ps, err := e.reader.GetProjectSecurity(); err == nil {
			for _, mr := range ms.ModuleRoles {
				qualifiedRole := s.Name + "." + mr.Name
				if n, err := e.writer.RemoveModuleRoleFromAllUserRoles(ps.ID, qualifiedRole); err == nil && n > 0 {
					fmt.Fprintf(ctx.Output, "Removed %s from %d user role(s)\n", qualifiedRole, n)
				}
			}
		}
	}

	// Delete the module itself (and clean up themesource directory)
	if err := e.writer.DeleteModuleWithCleanup(targetModule.ID, s.Name); err != nil {
		return mdlerrors.NewBackend("delete module", err)
	}

	// Build summary of what was removed
	var parts []string
	if nEntities > 0 {
		parts = append(parts, fmt.Sprintf("%d entities", nEntities))
	}
	if nAssocs > 0 {
		parts = append(parts, fmt.Sprintf("%d associations", nAssocs))
	}
	if nEnums > 0 {
		parts = append(parts, fmt.Sprintf("%d enumerations", nEnums))
	}
	if nMicroflows > 0 {
		parts = append(parts, fmt.Sprintf("%d microflows", nMicroflows))
	}
	if nNanoflows > 0 {
		parts = append(parts, fmt.Sprintf("%d nanoflows", nNanoflows))
	}
	if nPages > 0 {
		parts = append(parts, fmt.Sprintf("%d pages", nPages))
	}
	if nSnippets > 0 {
		parts = append(parts, fmt.Sprintf("%d snippets", nSnippets))
	}
	if nLayouts > 0 {
		parts = append(parts, fmt.Sprintf("%d layouts", nLayouts))
	}
	if nConstants > 0 {
		parts = append(parts, fmt.Sprintf("%d constants", nConstants))
	}
	if nJavaActions > 0 {
		parts = append(parts, fmt.Sprintf("%d java actions", nJavaActions))
	}
	if nServices > 0 {
		parts = append(parts, fmt.Sprintf("%d OData services", nServices))
	}
	if nBizEvents > 0 {
		parts = append(parts, fmt.Sprintf("%d business event services", nBizEvents))
	}
	if nDbConns > 0 {
		parts = append(parts, fmt.Sprintf("%d database connections", nDbConns))
	}

	if len(parts) > 0 {
		fmt.Fprintf(ctx.Output, "Dropped module: %s (%s)\n", s.Name, strings.Join(parts, ", "))
	} else {
		fmt.Fprintf(ctx.Output, "Dropped module: %s (empty)\n", s.Name)
	}
	return nil
}

// Executor method wrapper — kept during migration for callers not yet

// getModuleContainers returns a set of all container IDs that belong to a module
// (including nested folders).
func getModuleContainers(ctx *ExecContext, moduleID model.ID) map[model.ID]bool {
	e := ctx.executor
	containers := make(map[model.ID]bool)
	containers[moduleID] = true

	// Build parent -> children map from units
	units, err := e.reader.ListUnits()
	if err != nil {
		return containers
	}

	childrenOf := make(map[model.ID][]model.ID)
	for _, u := range units {
		childrenOf[u.ContainerID] = append(childrenOf[u.ContainerID], u.ID)
	}

	// Also include folders
	folders, _ := e.reader.ListFolders()
	for _, f := range folders {
		childrenOf[f.ContainerID] = append(childrenOf[f.ContainerID], f.ID)
	}

	// BFS to find all containers under this module
	queue := []model.ID{moduleID}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, child := range childrenOf[current] {
			if !containers[child] {
				containers[child] = true
				queue = append(queue, child)
			}
		}
	}

	return containers
}

// showModules handles SHOW MODULES command.
func showModules(ctx *ExecContext) error {
	e := ctx.executor
	if e.reader == nil {
		return mdlerrors.NewNotConnected()
	}

	// Always get fresh module list and update cache
	invalidateModuleCache(ctx)
	modules, err := getModulesFromCache(ctx)
	if err != nil {
		return mdlerrors.NewBackend("list modules", err)
	}

	// Get hierarchy for module resolution
	h, err := getHierarchy(ctx)
	if err != nil {
		return mdlerrors.NewBackend("build hierarchy", err)
	}

	// Get units for type-based counting
	units, err := e.reader.ListUnits()
	if err != nil {
		return mdlerrors.NewBackend("list units", err)
	}

	// Count elements per module using unit types
	entityCounts := make(map[model.ID]int)
	enumCounts := make(map[model.ID]int)
	pageCounts := make(map[model.ID]int)
	snippetCounts := make(map[model.ID]int)
	microflowCounts := make(map[model.ID]int)
	nanoflowCounts := make(map[model.ID]int)
	constantCounts := make(map[model.ID]int)
	javaActionCounts := make(map[model.ID]int)
	workflowCounts := make(map[model.ID]int)
	pubRestCounts := make(map[model.ID]int)
	pubODataCounts := make(map[model.ID]int)
	conODataCounts := make(map[model.ID]int)
	bizEventCounts := make(map[model.ID]int)
	extDbCounts := make(map[model.ID]int)

	// Count from units by type prefix (efficient single pass)
	for _, u := range units {
		modID := h.FindModuleID(u.ContainerID)
		switch {
		case strings.HasPrefix(u.Type, "Rest$PublishedRestService"):
			pubRestCounts[modID]++
		case strings.HasPrefix(u.Type, "ODataPublish$PublishedODataService"):
			pubODataCounts[modID]++
		case strings.HasPrefix(u.Type, "Rest$ConsumedODataService"):
			conODataCounts[modID]++
		case strings.HasPrefix(u.Type, "BusinessEvents$"):
			bizEventCounts[modID]++
		case strings.HasPrefix(u.Type, "Workflows$Workflow"):
			workflowCounts[modID]++
		case strings.HasPrefix(u.Type, "DatabaseConnector$DatabaseConnection"):
			extDbCounts[modID]++
		}
	}

	// Count entities from domain models
	if dms, err := e.reader.ListDomainModels(); err == nil {
		for _, dm := range dms {
			modID := h.FindModuleID(dm.ContainerID)
			entityCounts[modID] += len(dm.Entities)
		}
	}

	// Count enumerations
	if enums, err := e.reader.ListEnumerations(); err == nil {
		for _, enum := range enums {
			modID := h.FindModuleID(enum.ContainerID)
			enumCounts[modID]++
		}
	}

	// Count pages
	if pages, err := e.reader.ListPages(); err == nil {
		for _, p := range pages {
			modID := h.FindModuleID(p.ContainerID)
			pageCounts[modID]++
		}
	}

	// Count snippets
	if snippets, err := e.reader.ListSnippets(); err == nil {
		for _, s := range snippets {
			modID := h.FindModuleID(s.ContainerID)
			snippetCounts[modID]++
		}
	}

	// Count microflows
	if mfs, err := e.reader.ListMicroflows(); err == nil {
		for _, mf := range mfs {
			modID := h.FindModuleID(mf.ContainerID)
			microflowCounts[modID]++
		}
	}

	// Count nanoflows
	if nfs, err := e.reader.ListNanoflows(); err == nil {
		for _, nf := range nfs {
			modID := h.FindModuleID(nf.ContainerID)
			nanoflowCounts[modID]++
		}
	}

	// Count constants
	if constants, err := e.reader.ListConstants(); err == nil {
		for _, c := range constants {
			modID := h.FindModuleID(c.ContainerID)
			constantCounts[modID]++
		}
	}

	// Count Java actions
	if jas, err := e.reader.ListJavaActions(); err == nil {
		for _, ja := range jas {
			modID := h.FindModuleID(ja.ContainerID)
			javaActionCounts[modID]++
		}
	}

	// Sort modules alphabetically by name
	sort.Slice(modules, func(i, j int) bool {
		return strings.ToLower(modules[i].Name) < strings.ToLower(modules[j].Name)
	})

	// Collect rows and calculate column widths
	type row struct {
		name        string
		source      string
		entities    int
		enums       int
		pages       int
		snippets    int
		microflows  int
		nanoflows   int
		workflows   int
		constants   int
		javaActions int
		pubRest     int
		pubOData    int
		conOData    int
		bizEvents   int
		extDb       int
	}
	var rows []row

	for _, m := range modules {
		// Determine source (AppStore with version or local)
		source := ""
		if m.FromAppStore {
			if m.AppStoreVersion != "" {
				source = "Marketplace v" + m.AppStoreVersion
			} else {
				source = "Marketplace"
			}
		}

		r := row{
			name:        m.Name,
			source:      source,
			entities:    entityCounts[m.ID],
			enums:       enumCounts[m.ID],
			pages:       pageCounts[m.ID],
			snippets:    snippetCounts[m.ID],
			microflows:  microflowCounts[m.ID],
			nanoflows:   nanoflowCounts[m.ID],
			workflows:   workflowCounts[m.ID],
			constants:   constantCounts[m.ID],
			javaActions: javaActionCounts[m.ID],
			pubRest:     pubRestCounts[m.ID],
			pubOData:    pubODataCounts[m.ID],
			conOData:    conODataCounts[m.ID],
			bizEvents:   bizEventCounts[m.ID],
			extDb:       extDbCounts[m.ID],
		}
		rows = append(rows, r)
	}

	// Build TableResult
	result := &TableResult{
		Columns: []string{"Module", "Source", "Entities", "Enums", "Pages", "Snippets", "Microflows", "Nanoflows", "Workflows", "Constants", "JavaActions", "PubREST", "PubOData", "ConOData", "BizEvents", "ExtDB"},
		Summary: fmt.Sprintf("(%d modules)", len(modules)),
	}
	for _, r := range rows {
		result.Rows = append(result.Rows, []any{r.name, r.source, r.entities, r.enums, r.pages, r.snippets, r.microflows, r.nanoflows, r.workflows, r.constants, r.javaActions, r.pubRest, r.pubOData, r.conOData, r.bizEvents, r.extDb})
	}
	return writeResult(ctx, result)
}

// describeModule handles DESCRIBE MODULE [WITH ALL] command.
func describeModule(ctx *ExecContext, moduleName string, withAll bool) error {
	e := ctx.executor
	if e.reader == nil {
		return mdlerrors.NewNotConnected()
	}

	// Find the module
	modules, err := e.reader.ListModules()
	if err != nil {
		return mdlerrors.NewBackend("list modules", err)
	}

	var targetModule *model.Module
	for _, m := range modules {
		if m.Name == moduleName {
			targetModule = m
			break
		}
	}

	if targetModule == nil {
		return mdlerrors.NewNotFound("module", moduleName)
	}

	// Output basic CREATE MODULE statement
	fmt.Fprintf(ctx.Output, "CREATE MODULE %s;\n", targetModule.Name)

	if !withAll {
		fmt.Fprintln(ctx.Output, "/")
		return nil
	}

	// Get all containers belonging to this module (including nested folders)
	moduleContainers := getModuleContainers(ctx, targetModule.ID)

	// Output separator
	fmt.Fprintln(ctx.Output)

	// Output enumerations in this module (no dependencies between enums)
	if enums, err := e.reader.ListEnumerations(); err == nil {
		for _, enum := range enums {
			if moduleContainers[enum.ContainerID] {
				if err := describeEnumeration(ctx, ast.QualifiedName{Module: moduleName, Name: enum.Name}); err == nil {
					fmt.Fprintln(ctx.Output)
				}
			}
		}
	}

	// Output constants (may reference enumerations)
	if constants, err := e.reader.ListConstants(); err == nil {
		for _, c := range constants {
			if moduleContainers[c.ContainerID] {
				if err := outputConstantMDL(ctx, c, moduleName); err == nil {
					fmt.Fprintln(ctx.Output)
				}
			}
		}
	}

	// Output entities in dependency order (base entities before derived entities)
	// and associations after all entities
	if dm, err := e.reader.GetDomainModel(targetModule.ID); err == nil {
		// Topologically sort entities by generalization (inheritance)
		sortedEntities := sortEntitiesByGeneralization(dm.Entities, moduleName)

		// Output entities in sorted order
		for _, entity := range sortedEntities {
			if err := describeEntity(ctx, ast.QualifiedName{Module: moduleName, Name: entity.Name}); err == nil {
				fmt.Fprintln(ctx.Output)
			}
		}

		// Output associations (after all entities are defined)
		for _, assoc := range dm.Associations {
			if err := describeAssociation(ctx, ast.QualifiedName{Module: moduleName, Name: assoc.Name}); err == nil {
				fmt.Fprintln(ctx.Output)
			}
		}
	}

	// Output microflows
	if mfs, err := e.reader.ListMicroflows(); err == nil {
		for _, mf := range mfs {
			if moduleContainers[mf.ContainerID] {
				if err := describeMicroflow(ctx, ast.QualifiedName{Module: moduleName, Name: mf.Name}); err == nil {
					fmt.Fprintln(ctx.Output)
				}
			}
		}
	}

	// Output java actions
	if jaList, err := e.reader.ListJavaActions(); err == nil {
		for _, ja := range jaList {
			if moduleContainers[ja.ContainerID] {
				if err := describeJavaAction(ctx, ast.QualifiedName{Module: moduleName, Name: ja.Name}); err == nil {
					fmt.Fprintln(ctx.Output)
				}
			}
		}
	}

	// Output pages
	if pageList, err := e.reader.ListPages(); err == nil {
		for _, p := range pageList {
			if moduleContainers[p.ContainerID] {
				if err := describePage(ctx, ast.QualifiedName{Module: moduleName, Name: p.Name}); err == nil {
					fmt.Fprintln(ctx.Output)
				}
			}
		}
	}

	// Output snippets
	if snippets, err := e.reader.ListSnippets(); err == nil {
		for _, s := range snippets {
			if moduleContainers[s.ContainerID] {
				if err := describeSnippet(ctx, ast.QualifiedName{Module: moduleName, Name: s.Name}); err == nil {
					fmt.Fprintln(ctx.Output)
				}
			}
		}
	}

	// Output layouts
	if layouts, err := e.reader.ListLayouts(); err == nil {
		for _, l := range layouts {
			if moduleContainers[l.ContainerID] {
				if err := describeLayout(ctx, ast.QualifiedName{Module: moduleName, Name: l.Name}); err == nil {
					fmt.Fprintln(ctx.Output)
				}
			}
		}
	}

	// Output database connections
	if conns, err := e.reader.ListDatabaseConnections(); err == nil {
		for _, conn := range conns {
			if moduleContainers[conn.ContainerID] {
				if err := outputDatabaseConnectionMDL(ctx, conn, moduleName); err == nil {
					fmt.Fprintln(ctx.Output)
				}
			}
		}
	}

	// Output business event services
	if services, err := e.reader.ListBusinessEventServices(); err == nil {
		for _, svc := range services {
			if moduleContainers[svc.ContainerID] {
				if err := describeBusinessEventService(ctx, ast.QualifiedName{Module: moduleName, Name: svc.Name}); err == nil {
					fmt.Fprintln(ctx.Output)
				}
			}
		}
	}

	// Get hierarchy for folder path resolution (used by OData sections below)
	h, _ := getHierarchy(ctx)

	// Output consumed OData services (clients)
	if services, err := e.reader.ListConsumedODataServices(); err == nil {
		for _, svc := range services {
			if moduleContainers[svc.ContainerID] {
				folderPath := ""
				if h != nil {
					folderPath = h.BuildFolderPath(svc.ContainerID)
				}
				if err := outputConsumedODataServiceMDL(ctx, svc, moduleName, folderPath); err == nil {
					fmt.Fprintln(ctx.Output)
				}
			}
		}
	}

	// Output published OData services
	if services, err := e.reader.ListPublishedODataServices(); err == nil {
		for _, svc := range services {
			if moduleContainers[svc.ContainerID] {
				folderPath := ""
				if h != nil {
					folderPath = h.BuildFolderPath(svc.ContainerID)
				}
				if err := outputPublishedODataServiceMDL(ctx, svc, moduleName, folderPath); err == nil {
					fmt.Fprintln(ctx.Output)
				}
			}
		}
	}

	// Output agent-editor documents. Order matches referential dependencies:
	// Model / KnowledgeBase / ConsumedMCPService first (leaves), Agent last
	// (may reference the others via its TOOL / KNOWLEDGE BASE / MCP SERVICE
	// blocks).
	if models, err := e.reader.ListAgentEditorModels(); err == nil {
		for _, m := range models {
			if moduleContainers[m.ContainerID] {
				if err := describeAgentEditorModel(ctx, ast.QualifiedName{Module: moduleName, Name: m.Name}); err == nil {
					fmt.Fprintln(ctx.Output)
				}
			}
		}
	}
	if kbs, err := e.reader.ListAgentEditorKnowledgeBases(); err == nil {
		for _, kb := range kbs {
			if moduleContainers[kb.ContainerID] {
				if err := describeAgentEditorKnowledgeBase(ctx, ast.QualifiedName{Module: moduleName, Name: kb.Name}); err == nil {
					fmt.Fprintln(ctx.Output)
				}
			}
		}
	}
	if svcs, err := e.reader.ListAgentEditorConsumedMCPServices(); err == nil {
		for _, svc := range svcs {
			if moduleContainers[svc.ContainerID] {
				if err := describeAgentEditorConsumedMCPService(ctx, ast.QualifiedName{Module: moduleName, Name: svc.Name}); err == nil {
					fmt.Fprintln(ctx.Output)
				}
			}
		}
	}
	if agents, err := e.reader.ListAgentEditorAgents(); err == nil {
		for _, a := range agents {
			if moduleContainers[a.ContainerID] {
				if err := describeAgentEditorAgent(ctx, ast.QualifiedName{Module: moduleName, Name: a.Name}); err == nil {
					fmt.Fprintln(ctx.Output)
				}
			}
		}
	}

	fmt.Fprintln(ctx.Output, "/")
	return nil
}

// sortEntitiesByGeneralization returns entities sorted so base entities come before derived entities.
// Uses topological sort based on GeneralizationRef (parent entity reference).
func sortEntitiesByGeneralization(entities []*domainmodel.Entity, moduleName string) []*domainmodel.Entity {
	if len(entities) <= 1 {
		return entities
	}

	// Build map of entity name to entity
	entityByName := make(map[string]*domainmodel.Entity)
	for _, ent := range entities {
		entityByName[ent.Name] = ent
		// Also index by qualified name
		entityByName[moduleName+"."+ent.Name] = ent
	}

	// Build adjacency: child -> parent (for entities within this module)
	// We need to output parent before child
	parentOf := make(map[string]string)
	for _, ent := range entities {
		if ent.GeneralizationRef != "" {
			// GeneralizationRef is like "ModuleName.EntityName"
			parentOf[ent.Name] = ent.GeneralizationRef
		}
	}

	// Kahn's algorithm for topological sort
	// Count incoming edges (how many entities extend this one)
	inDegree := make(map[string]int)
	for _, ent := range entities {
		inDegree[ent.Name] = 0
	}
	for child, parent := range parentOf {
		// Only count if parent is in the same module
		if _, inModule := entityByName[parent]; inModule {
			inDegree[child]++
		}
	}

	// Start with entities that have no parent in this module
	var queue []string
	for _, ent := range entities {
		if inDegree[ent.Name] == 0 {
			queue = append(queue, ent.Name)
		}
	}

	// Build children map for traversal
	childrenOf := make(map[string][]string)
	for child, parent := range parentOf {
		if _, inModule := entityByName[parent]; inModule {
			// Extract just the entity name if it's qualified
			parentName := parent
			if strings.Contains(parent, ".") {
				parts := strings.Split(parent, ".")
				parentName = parts[len(parts)-1]
			}
			childrenOf[parentName] = append(childrenOf[parentName], child)
		}
	}

	// Process queue
	var sorted []*domainmodel.Entity
	visited := make(map[string]bool)
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]

		if visited[name] {
			continue
		}
		visited[name] = true

		if ent, ok := entityByName[name]; ok {
			sorted = append(sorted, ent)
		}

		// Add children whose parents are now processed
		for _, child := range childrenOf[name] {
			inDegree[child]--
			if inDegree[child] == 0 {
				queue = append(queue, child)
			}
		}
	}

	// Add any remaining entities (handles cycles or external parents)
	for _, ent := range entities {
		if !visited[ent.Name] {
			sorted = append(sorted, ent)
		}
	}

	return sorted
}

// Executor method wrappers for callers in unmigrated files.
