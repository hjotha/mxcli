// SPDX-License-Identifier: Apache-2.0

// Package executor - Widget commands (SHOW WIDGETS, UPDATE WIDGETS)
package executor

import (
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/mendixlabs/mxcli/mdl/ast"
	mdlerrors "github.com/mendixlabs/mxcli/mdl/errors"
	"github.com/mendixlabs/mxcli/model"
)

// execShowWidgets handles the SHOW WIDGETS statement.
func execShowWidgets(ctx *ExecContext, s *ast.ShowWidgetsStmt) error {

	if !ctx.Connected() {
		return mdlerrors.NewNotConnected()
	}

	// Ensure catalog is built (full mode for widgets)
	if err := ensureCatalog(ctx, true); err != nil {
		return mdlerrors.NewBackend("build catalog", err)
	}

	// Build SQL query from filters
	var query strings.Builder
	query.WriteString("SELECT Name, WidgetType, ContainerQualifiedName, ModuleName FROM widgets WHERE 1=1")
	args := []any{}

	for _, f := range s.Filters {
		col := mapWidgetFilterField(f.Field)
		if f.Operator == "LIKE" {
			query.WriteString(fmt.Sprintf(" AND %s LIKE ?", col))
		} else {
			query.WriteString(fmt.Sprintf(" AND %s = ?", col))
		}
		args = append(args, f.Value)
	}

	if s.InModule != "" {
		query.WriteString(" AND ModuleName = ?")
		args = append(args, s.InModule)
	}

	query.WriteString(" ORDER BY ModuleName, ContainerQualifiedName, Name")

	// Execute query using SQLite parameterization
	result, err := executeCatalogQueryWithArgs(ctx, query.String(), args...)
	if err != nil {
		return mdlerrors.NewBackend("query widgets", err)
	}

	// Output results as table
	if result.Count == 0 {
		fmt.Fprintln(ctx.Output, "No widgets found matching the criteria")
		return nil
	}

	// Print header
	fmt.Fprintf(ctx.Output, "\n%-30s %-40s %-40s %-20s\n",
		"NAME", "WIDGET TYPE", "CONTAINER", "MODULE")
	fmt.Fprintln(ctx.Output, strings.Repeat("-", 130))

	// Print rows
	for _, row := range result.Rows {
		name := formatCell(row[0], 30)
		widgetType := formatCell(row[1], 40)
		container := formatCell(row[2], 40)
		module := formatCell(row[3], 20)
		fmt.Fprintf(ctx.Output, "%-30s %-40s %-40s %-20s\n", name, widgetType, container, module)
	}

	fmt.Fprintf(ctx.Output, "\n%d widget(s) found\n", result.Count)
	return nil
}

// execUpdateWidgets handles the UPDATE WIDGETS statement.
func execUpdateWidgets(ctx *ExecContext, s *ast.UpdateWidgetsStmt) error {
	if !ctx.Connected() {
		return mdlerrors.NewNotConnected()
	}
	if !ctx.ConnectedForWrite() {
		return mdlerrors.NewNotConnectedWrite()
	}

	// Ensure catalog is built (full mode for widgets)
	if err := ensureCatalog(ctx, true); err != nil {
		return mdlerrors.NewBackend("build catalog", err)
	}

	// Find matching widgets
	widgets, err := findMatchingWidgets(ctx, s.Filters, s.InModule)
	if err != nil {
		return mdlerrors.NewBackend("find widgets", err)
	}

	if len(widgets) == 0 {
		fmt.Fprintln(ctx.Output, "No widgets found matching the criteria")
		return nil
	}

	// Group widgets by container
	containers := groupWidgetsByContainer(widgets)

	// Report what will be updated
	fmt.Fprintf(ctx.Output, "\nFound %d widget(s) in %d container(s) matching the criteria\n",
		len(widgets), len(containers))

	if s.DryRun {
		fmt.Fprintln(ctx.Output, "\n[DRY RUN] The following changes would be made:")
	}

	// Process each container
	totalUpdated := 0
	for containerID, widgetRefs := range containers {
		updated, err := updateWidgetsInContainer(ctx, containerID, widgetRefs, s.Assignments, s.DryRun)
		if err != nil {
			fmt.Fprintf(ctx.Output, "Warning: Failed to update widgets in %s: %v\n", containerID, err)
			continue
		}
		totalUpdated += updated
	}

	if s.DryRun {
		fmt.Fprintf(ctx.Output, "\n[DRY RUN] Would update %d widget(s)\n", totalUpdated)
		fmt.Fprintln(ctx.Output, "\nRun without DRY RUN to apply changes.")
	} else {
		fmt.Fprintf(ctx.Output, "\nUpdated %d widget(s)\n", totalUpdated)
		fmt.Fprintln(ctx.Output, "\nNote: Run 'REFRESH CATALOG FULL FORCE' to update the catalog with changes.")
	}

	return nil
}

// widgetRef holds information about a widget to be updated.
type widgetRef struct {
	ID            string
	Name          string
	WidgetType    string
	ContainerID   string
	ContainerName string
	ContainerType string // "page" or "snippet"
}

// findMatchingWidgets queries the catalog for widgets matching the filters.
func findMatchingWidgets(ctx *ExecContext, filters []ast.WidgetFilter, module string) ([]widgetRef, error) {
	var query strings.Builder
	query.WriteString(`SELECT Id, Name, WidgetType, ContainerId, ContainerQualifiedName, ContainerType
	          FROM widgets WHERE 1=1`)
	args := []any{}

	for _, f := range filters {
		col := mapWidgetFilterField(f.Field)
		if f.Operator == "LIKE" {
			query.WriteString(fmt.Sprintf(" AND %s LIKE ?", col))
		} else {
			query.WriteString(fmt.Sprintf(" AND %s = ?", col))
		}
		args = append(args, f.Value)
	}

	if module != "" {
		query.WriteString(" AND ModuleName = ?")
		args = append(args, module)
	}

	result, err := executeCatalogQueryWithArgs(ctx, query.String(), args...)
	if err != nil {
		return nil, err
	}

	widgets := make([]widgetRef, 0, result.Count)
	for _, row := range result.Rows {
		widgets = append(widgets, widgetRef{
			ID:            fmt.Sprintf("%v", row[0]),
			Name:          fmt.Sprintf("%v", row[1]),
			WidgetType:    fmt.Sprintf("%v", row[2]),
			ContainerID:   fmt.Sprintf("%v", row[3]),
			ContainerName: fmt.Sprintf("%v", row[4]),
			ContainerType: fmt.Sprintf("%v", row[5]),
		})
	}

	return widgets, nil
}

// groupWidgetsByContainer groups widgets by their container ID.
func groupWidgetsByContainer(widgets []widgetRef) map[string][]widgetRef {
	containers := make(map[string][]widgetRef)
	for _, w := range widgets {
		containers[w.ContainerID] = append(containers[w.ContainerID], w)
	}
	return containers
}

// updateWidgetsInContainer updates widgets within a single page or snippet.
func updateWidgetsInContainer(ctx *ExecContext, containerID string, widgetRefs []widgetRef, assignments []ast.WidgetPropertyAssignment, dryRun bool) (int, error) {
	if len(widgetRefs) == 0 {
		return 0, nil
	}

	containerType := widgetRefs[0].ContainerType
	containerName := widgetRefs[0].ContainerName

	// Load the page or snippet
	if strings.ToLower(containerType) == "page" {
		return updateWidgetsInPage(ctx, containerID, containerName, widgetRefs, assignments, dryRun)
	} else if strings.ToLower(containerType) == "snippet" {
		return updateWidgetsInSnippet(ctx, containerID, containerName, widgetRefs, assignments, dryRun)
	}

	return 0, mdlerrors.NewUnsupported(fmt.Sprintf("unsupported container type: %s", containerType))
}

// updateWidgetsInPage updates widgets in a page using raw BSON.
func updateWidgetsInPage(ctx *ExecContext, containerID, containerName string, widgetRefs []widgetRef, assignments []ast.WidgetPropertyAssignment, dryRun bool) (int, error) {

	// Load raw BSON as ordered document (preserves field ordering)
	rawBytes, err := ctx.Backend.GetRawUnitBytes(model.ID(containerID))
	if err != nil {
		return 0, mdlerrors.NewBackend(fmt.Sprintf("load page %s", containerName), err)
	}
	var rawData bson.D
	if err := bson.Unmarshal(rawBytes, &rawData); err != nil {
		return 0, mdlerrors.NewBackend(fmt.Sprintf("unmarshal page %s", containerName), err)
	}

	updated := 0
	for _, ref := range widgetRefs {
		result := findBsonWidget(rawData, ref.Name)
		if result == nil {
			fmt.Fprintf(ctx.Output, "  Warning: Widget %q not found in page %s\n", ref.Name, containerName)
			continue
		}
		for _, assignment := range assignments {
			if dryRun {
				fmt.Fprintf(ctx.Output, "  Would set '%s' = %v on %s (%s) in %s\n",
					assignment.PropertyPath, assignment.Value, ref.Name, ref.WidgetType, containerName)
			} else {
				if err := setRawWidgetProperty(result.widget, assignment.PropertyPath, assignment.Value); err != nil {
					fmt.Fprintf(ctx.Output, "  Warning: Failed to set '%s' on %s: %v\n",
						assignment.PropertyPath, ref.Name, err)
				}
			}
		}
		updated++
	}

	// Save back via raw BSON (bson.D preserves field ordering)
	if !dryRun && updated > 0 {
		outBytes, err := bson.Marshal(rawData)
		if err != nil {
			return updated, mdlerrors.NewBackend(fmt.Sprintf("marshal page %s", containerName), err)
		}
		if err := ctx.Backend.UpdateRawUnit(containerID, outBytes); err != nil {
			return updated, mdlerrors.NewBackend(fmt.Sprintf("save page %s", containerName), err)
		}
	}

	return updated, nil
}

// updateWidgetsInSnippet updates widgets in a snippet using raw BSON.
func updateWidgetsInSnippet(ctx *ExecContext, containerID, containerName string, widgetRefs []widgetRef, assignments []ast.WidgetPropertyAssignment, dryRun bool) (int, error) {

	// Load raw BSON as ordered document (preserves field ordering)
	rawBytes, err := ctx.Backend.GetRawUnitBytes(model.ID(containerID))
	if err != nil {
		return 0, mdlerrors.NewBackend(fmt.Sprintf("load snippet %s", containerName), err)
	}
	var rawData bson.D
	if err := bson.Unmarshal(rawBytes, &rawData); err != nil {
		return 0, mdlerrors.NewBackend(fmt.Sprintf("unmarshal snippet %s", containerName), err)
	}

	updated := 0
	for _, ref := range widgetRefs {
		result := findBsonWidgetInSnippet(rawData, ref.Name)
		if result == nil {
			fmt.Fprintf(ctx.Output, "  Warning: Widget %q not found in snippet %s\n", ref.Name, containerName)
			continue
		}
		for _, assignment := range assignments {
			if dryRun {
				fmt.Fprintf(ctx.Output, "  Would set '%s' = %v on %s (%s) in %s\n",
					assignment.PropertyPath, assignment.Value, ref.Name, ref.WidgetType, containerName)
			} else {
				if err := setRawWidgetProperty(result.widget, assignment.PropertyPath, assignment.Value); err != nil {
					fmt.Fprintf(ctx.Output, "  Warning: Failed to set '%s' on %s: %v\n",
						assignment.PropertyPath, ref.Name, err)
				}
			}
		}
		updated++
	}

	// Save back via raw BSON (bson.D preserves field ordering)
	if !dryRun && updated > 0 {
		outBytes, err := bson.Marshal(rawData)
		if err != nil {
			return updated, mdlerrors.NewBackend(fmt.Sprintf("marshal snippet %s", containerName), err)
		}
		if err := ctx.Backend.UpdateRawUnit(containerID, outBytes); err != nil {
			return updated, mdlerrors.NewBackend(fmt.Sprintf("save snippet %s", containerName), err)
		}
	}

	return updated, nil
}

// mapWidgetFilterField maps user-facing field names to catalog column names.
func mapWidgetFilterField(field string) string {
	switch strings.ToLower(field) {
	case "widgettype":
		return "WidgetType"
	case "name":
		return "Name"
	case "container":
		return "ContainerQualifiedName"
	case "module":
		return "ModuleName"
	default:
		return field
	}
}

// executeCatalogQueryWithArgs executes a parameterized SQL query against the catalog.
func executeCatalogQueryWithArgs(ctx *ExecContext, query string, args ...any) (*catalogQueryResult, error) {
	// Replace ? placeholders with values for SQLite
	// Note: This is a simplified implementation; production code should use prepared statements
	finalQuery := query
	for _, arg := range args {
		// Escape single quotes in string values
		strVal := fmt.Sprintf("%v", arg)
		strVal = strings.ReplaceAll(strVal, "'", "''")
		finalQuery = strings.Replace(finalQuery, "?", fmt.Sprintf("'%s'", strVal), 1)
	}

	result, err := ctx.Catalog.Query(finalQuery)
	if err != nil {
		return nil, err
	}

	return &catalogQueryResult{
		Columns: result.Columns,
		Rows:    result.Rows,
		Count:   result.Count,
	}, nil
}

// catalogQueryResult wraps catalog query results.
type catalogQueryResult struct {
	Columns []string
	Rows    [][]any
	Count   int
}

// formatCell formats a cell value for display, truncating if needed.
func formatCell(val any, maxLen int) string {
	s := fmt.Sprintf("%v", val)
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}
