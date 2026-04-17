// SPDX-License-Identifier: Apache-2.0

// Package executor - JSON structure commands (SHOW/DESCRIBE/CREATE/DROP JSON STRUCTURE)
package executor

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/mendixlabs/mxcli/mdl/ast"
	mdlerrors "github.com/mendixlabs/mxcli/mdl/errors"
	"github.com/mendixlabs/mxcli/sdk/mpr"
)

// showJsonStructures handles SHOW JSON STRUCTURES [IN module].
func showJsonStructures(ctx *ExecContext, moduleName string) error {
	e := ctx.executor
	structures, err := e.reader.ListJsonStructures()
	if err != nil {
		return mdlerrors.NewBackend("list JSON structures", err)
	}

	h, err := e.getHierarchy()
	if err != nil {
		return err
	}

	type row struct {
		qualifiedName string
		elemCount     int
		source        string
	}
	var rows []row

	for _, js := range structures {
		modID := h.FindModuleID(js.ContainerID)
		modName := h.GetModuleName(modID)
		if moduleName != "" && modName != moduleName {
			continue
		}

		qualifiedName := fmt.Sprintf("%s.%s", modName, js.Name)

		elemCount := 0
		if len(js.Elements) > 0 {
			elemCount = len(js.Elements[0].Children)
		}

		source := "Manual"
		if js.JsonSnippet != "" {
			source = "JSON Snippet"
		}

		rows = append(rows, row{qualifiedName: qualifiedName, elemCount: elemCount, source: source})
	}

	// Sort alphabetically
	sort.Slice(rows, func(i, j int) bool { return rows[i].qualifiedName < rows[j].qualifiedName })

	tr := &TableResult{
		Columns: []string{"JSON Structure", "Elements", "Source"},
		Summary: fmt.Sprintf("(%d JSON structure(s))", len(rows)),
	}
	for _, r := range rows {
		tr.Rows = append(tr.Rows, []any{r.qualifiedName, r.elemCount, r.source})
	}
	return e.writeResult(tr)
}

// showJsonStructures is a wrapper for callers that still use an Executor receiver.
func (e *Executor) showJsonStructures(moduleName string) error {
	return showJsonStructures(e.newExecContext(context.Background()), moduleName)
}

// describeJsonStructure handles DESCRIBE JSON STRUCTURE Module.Name.
// Output is re-executable CREATE OR REPLACE MDL followed by the element tree as comments.
func describeJsonStructure(ctx *ExecContext, name ast.QualifiedName) error {
	e := ctx.executor
	js := findJsonStructure(ctx, name.Module, name.Name)
	if js == nil {
		return mdlerrors.NewNotFound("JSON structure", name.String())
	}

	h, err := e.getHierarchy()
	if err != nil {
		return err
	}
	modID := h.FindModuleID(js.ContainerID)
	modName := h.GetModuleName(modID)

	qualifiedName := fmt.Sprintf("%s.%s", modName, js.Name)

	// Documentation as doc comment
	if js.Documentation != "" {
		fmt.Fprintf(ctx.Output, "/**\n * %s\n */\n", js.Documentation)
	}

	// Re-executable CREATE OR REPLACE statement
	fmt.Fprintf(ctx.Output, "CREATE OR REPLACE JSON STRUCTURE %s", qualifiedName)
	if folderPath := h.BuildFolderPath(js.ContainerID); folderPath != "" {
		fmt.Fprintf(ctx.Output, "\n  FOLDER '%s'", folderPath)
	}
	if js.Documentation != "" {
		fmt.Fprintf(ctx.Output, "\n  COMMENT '%s'", strings.ReplaceAll(js.Documentation, "'", "''"))
	}

	if js.JsonSnippet != "" {
		snippet := mpr.PrettyPrintJSON(js.JsonSnippet)
		if strings.Contains(snippet, "'") || strings.Contains(snippet, "\n") {
			fmt.Fprintf(ctx.Output, "\n  SNIPPET $$%s$$", snippet)
		} else {
			fmt.Fprintf(ctx.Output, "\n  SNIPPET '%s'", snippet)
		}
	}

	// Detect custom name mappings by comparing ExposedName to auto-generated names
	customMappings := collectCustomNameMappings(js.Elements)
	if len(customMappings) > 0 {
		// Sort keys for deterministic DESCRIBE output
		keys := make([]string, 0, len(customMappings))
		for k := range customMappings {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		fmt.Fprintf(ctx.Output, "\n  CUSTOM NAME MAP (\n")
		for i, jsonKey := range keys {
			sep := ","
			if i == len(keys)-1 {
				sep = ""
			}
			fmt.Fprintf(ctx.Output, "    '%s' AS '%s'%s\n", jsonKey, customMappings[jsonKey], sep)
		}
		fmt.Fprintf(ctx.Output, "  )")
	}

	fmt.Fprintln(ctx.Output, ";")
	return nil
}

// describeJsonStructure is a wrapper for callers that still use an Executor receiver.
func (e *Executor) describeJsonStructure(name ast.QualifiedName) error {
	return describeJsonStructure(e.newExecContext(context.Background()), name)
}

// collectCustomNameMappings walks the element tree and returns JSON key → ExposedName
// mappings where the ExposedName differs from the auto-generated default (capitalizeFirst).
func collectCustomNameMappings(elements []*mpr.JsonElement) map[string]string {
	mappings := make(map[string]string)
	for _, elem := range elements {
		collectCustomNames(elem, mappings)
	}
	return mappings
}

func collectCustomNames(elem *mpr.JsonElement, mappings map[string]string) {
	// Extract the JSON key from the last segment of the Path.
	// Path format: "(Object)|fieldName" or "(Object)|parent|(Object)|child"
	if parts := strings.Split(elem.Path, "|"); len(parts) > 1 {
		jsonKey := parts[len(parts)-1]
		// Skip structural markers like (Object), (Array)
		if jsonKey != "" && jsonKey[0] != '(' {
			expected := capitalizeFirstRune(jsonKey)
			if elem.ExposedName != expected && elem.ExposedName != "" {
				mappings[jsonKey] = elem.ExposedName
			}
		}
	}
	for _, child := range elem.Children {
		collectCustomNames(child, mappings)
	}
}

// capitalizeFirstRune capitalizes the first rune of s (for ExposedName comparison).
func capitalizeFirstRune(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// execCreateJsonStructure handles CREATE [OR REPLACE] JSON STRUCTURE statements.
func execCreateJsonStructure(ctx *ExecContext, s *ast.CreateJsonStructureStmt) error {
	e := ctx.executor
	if e.reader == nil {
		return mdlerrors.NewNotConnected()
	}

	// Find or auto-create module
	module, err := e.findOrCreateModule(s.Name.Module)
	if err != nil {
		return err
	}

	// Resolve folder if specified
	containerID := module.ID
	if s.Folder != "" {
		folderID, err := e.resolveFolder(module.ID, s.Folder)
		if err != nil {
			return mdlerrors.NewBackend("resolve folder "+s.Folder, err)
		}
		containerID = folderID
	}

	// Check if already exists
	existing := findJsonStructure(ctx, s.Name.Module, s.Name.Name)
	if existing != nil {
		if s.CreateOrReplace {
			// Delete existing before recreating
			if err := e.writer.DeleteJsonStructure(string(existing.ID)); err != nil {
				return mdlerrors.NewBackend("delete existing JSON structure", err)
			}
		} else {
			return mdlerrors.NewAlreadyExists("JSON structure", s.Name.Module+"."+s.Name.Name)
		}
	}

	// Build element tree from JSON snippet, applying custom name mappings
	elements, err := mpr.BuildJsonElementsFromSnippet(s.JsonSnippet, s.CustomNameMap)
	if err != nil {
		return mdlerrors.NewBackend("build element tree", err)
	}

	// For CREATE OR REPLACE, keep original folder unless a new one is specified
	if existing != nil && s.Folder == "" {
		containerID = existing.ContainerID
	}

	js := &mpr.JsonStructure{
		ContainerID:   containerID,
		Name:          s.Name.Name,
		Documentation: s.Documentation,
		JsonSnippet:   mpr.PrettyPrintJSON(s.JsonSnippet),
		Elements:      elements,
	}

	if err := e.writer.CreateJsonStructure(js); err != nil {
		return mdlerrors.NewBackend("create JSON structure", err)
	}

	// Invalidate hierarchy cache
	e.invalidateHierarchy()

	action := "Created"
	if existing != nil {
		action = "Replaced"
	}
	fmt.Fprintf(ctx.Output, "%s JSON structure: %s\n", action, s.Name)
	return nil
}

// execDropJsonStructure handles DROP JSON STRUCTURE statements.
func execDropJsonStructure(ctx *ExecContext, s *ast.DropJsonStructureStmt) error {
	e := ctx.executor
	if e.reader == nil {
		return mdlerrors.NewNotConnected()
	}

	js := findJsonStructure(ctx, s.Name.Module, s.Name.Name)
	if js == nil {
		return mdlerrors.NewNotFound("JSON structure", s.Name.String())
	}

	if err := e.writer.DeleteJsonStructure(string(js.ID)); err != nil {
		return mdlerrors.NewBackend("delete JSON structure", err)
	}

	fmt.Fprintf(ctx.Output, "Dropped JSON structure: %s\n", s.Name)
	return nil
}

// findJsonStructure finds a JSON structure by module and name.
func findJsonStructure(ctx *ExecContext, moduleName, structName string) *mpr.JsonStructure {
	e := ctx.executor
	structures, err := e.reader.ListJsonStructures()
	if err != nil {
		return nil
	}

	h, _ := e.getHierarchy()
	if h == nil {
		return nil
	}

	for _, js := range structures {
		modID := h.FindModuleID(js.ContainerID)
		modName := h.GetModuleName(modID)
		if modName == moduleName && js.Name == structName {
			return js
		}
	}
	return nil
}
