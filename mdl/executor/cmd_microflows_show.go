// SPDX-License-Identifier: Apache-2.0

// Package executor - Microflow SHOW/DESCRIBE commands
package executor

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
	mdlerrors "github.com/mendixlabs/mxcli/mdl/errors"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/microflows"
)

// showMicroflows handles SHOW MICROFLOWS command.
func showMicroflows(ctx *ExecContext, moduleName string) error {
	e := ctx.executor
	// Get hierarchy for module/folder resolution
	h, err := getHierarchy(ctx)
	if err != nil {
		return mdlerrors.NewBackend("build hierarchy", err)
	}

	// Get all microflows
	microflows, err := e.reader.ListMicroflows()
	if err != nil {
		return mdlerrors.NewBackend("list microflows", err)
	}

	// Collect rows and calculate column widths
	type row struct {
		qualifiedName string
		module        string
		name          string
		excluded      bool
		folderPath    string
		params        int
		activities    int
		complexity    int
		returnType    string
	}
	var rows []row

	for _, mf := range microflows {
		modID := h.FindModuleID(mf.ContainerID)
		modName := h.GetModuleName(modID)
		if moduleName == "" || modName == moduleName {
			qualifiedName := modName + "." + mf.Name
			folderPath := h.BuildFolderPath(mf.ContainerID)
			returnType := ""
			if mf.ReturnType != nil {
				returnType = mf.ReturnType.GetTypeName()
			}

			// Count activities (excluding structural elements like Start/End events)
			activityCount := countMicroflowActivities(mf)

			// Calculate McCabe cyclomatic complexity
			complexity := calculateMicroflowComplexity(mf)

			rows = append(rows, row{qualifiedName, modName, mf.Name, mf.Excluded, folderPath, len(mf.Parameters), activityCount, complexity, returnType})
		}
	}

	// Sort by qualified name
	sort.Slice(rows, func(i, j int) bool {
		return strings.ToLower(rows[i].qualifiedName) < strings.ToLower(rows[j].qualifiedName)
	})

	result := &TableResult{
		Columns: []string{"Qualified Name", "Module", "Name", "Excluded", "Folder", "Params", "Actions", "McCabe", "Returns"},
		Summary: fmt.Sprintf("(%d microflows)", len(rows)),
	}
	for _, r := range rows {
		result.Rows = append(result.Rows, []any{r.qualifiedName, r.module, r.name, r.excluded, r.folderPath, r.params, r.activities, r.complexity, r.returnType})
	}
	return writeResult(ctx, result)
}

// showNanoflows handles SHOW NANOFLOWS command.
func showNanoflows(ctx *ExecContext, moduleName string) error {
	e := ctx.executor
	// Get hierarchy for module/folder resolution
	h, err := getHierarchy(ctx)
	if err != nil {
		return mdlerrors.NewBackend("build hierarchy", err)
	}

	// Get all nanoflows
	nanoflows, err := e.reader.ListNanoflows()
	if err != nil {
		return mdlerrors.NewBackend("list nanoflows", err)
	}

	// Collect rows and calculate column widths
	type row struct {
		qualifiedName string
		module        string
		name          string
		excluded      bool
		folderPath    string
		params        int
		activities    int
		complexity    int
		returnType    string
	}
	var rows []row

	for _, nf := range nanoflows {
		modID := h.FindModuleID(nf.ContainerID)
		modName := h.GetModuleName(modID)
		if moduleName == "" || modName == moduleName {
			qualifiedName := modName + "." + nf.Name
			folderPath := h.BuildFolderPath(nf.ContainerID)
			returnType := ""
			if nf.ReturnType != nil {
				returnType = nf.ReturnType.GetTypeName()
			}

			// Count activities (excluding structural elements like Start/End events)
			activityCount := countNanoflowActivities(nf)

			// Calculate McCabe cyclomatic complexity
			complexity := calculateNanoflowComplexity(nf)

			rows = append(rows, row{qualifiedName, modName, nf.Name, nf.Excluded, folderPath, len(nf.Parameters), activityCount, complexity, returnType})
		}
	}

	// Sort by qualified name
	sort.Slice(rows, func(i, j int) bool {
		return strings.ToLower(rows[i].qualifiedName) < strings.ToLower(rows[j].qualifiedName)
	})

	result := &TableResult{
		Columns: []string{"Qualified Name", "Module", "Name", "Excluded", "Folder", "Params", "Actions", "McCabe", "Returns"},
		Summary: fmt.Sprintf("(%d nanoflows)", len(rows)),
	}
	for _, r := range rows {
		result.Rows = append(result.Rows, []any{r.qualifiedName, r.module, r.name, r.excluded, r.folderPath, r.params, r.activities, r.complexity, r.returnType})
	}
	return writeResult(ctx, result)
}

// countNanoflowActivities counts meaningful activities in a nanoflow.
func countNanoflowActivities(nf *microflows.Nanoflow) int {
	if nf.ObjectCollection == nil {
		return 0
	}
	count := 0
	for _, obj := range nf.ObjectCollection.Objects {
		switch obj.(type) {
		case *microflows.StartEvent, *microflows.EndEvent, *microflows.ExclusiveMerge:
			// Skip structural elements
		default:
			count++
		}
	}
	return count
}

// calculateNanoflowComplexity calculates McCabe cyclomatic complexity for a nanoflow.
func calculateNanoflowComplexity(nf *microflows.Nanoflow) int {
	if nf.ObjectCollection == nil {
		return 1
	}
	// McCabe = E - N + 2P where E = edges, N = nodes, P = connected components (1 for a single flow)
	// Simplified: 1 + number of decision points (ExclusiveSplit, InheritanceSplit, LoopedActivity)
	complexity := 1
	for _, obj := range nf.ObjectCollection.Objects {
		switch obj.(type) {
		case *microflows.ExclusiveSplit, *microflows.InheritanceSplit, *microflows.LoopedActivity:
			complexity++
		}
	}
	return complexity
}

// describeMicroflow handles DESCRIBE MICROFLOW command - outputs MDL source code.
func describeMicroflow(ctx *ExecContext, name ast.QualifiedName) error {
	e := ctx.executor
	// Get hierarchy for module/folder resolution
	h, err := getHierarchy(ctx)
	if err != nil {
		return mdlerrors.NewBackend("build hierarchy", err)
	}

	// Use pre-warmed cache if available (from PreWarmCache), otherwise build on demand
	entityNames := e.getEntityNames(h)
	microflowNames := e.getMicroflowNames(h)

	// Find the microflow
	allMicroflows, err := e.reader.ListMicroflows()
	if err != nil {
		return mdlerrors.NewBackend("list microflows", err)
	}

	// Supplement microflow name lookup if not pre-warmed
	if len(microflowNames) == 0 {
		for _, mf := range allMicroflows {
			microflowNames[mf.ID] = h.GetQualifiedName(mf.ContainerID, mf.Name)
		}
	}

	var targetMf *microflows.Microflow
	for _, mf := range allMicroflows {
		modID := h.FindModuleID(mf.ContainerID)
		modName := h.GetModuleName(modID)
		if modName == name.Module && mf.Name == name.Name {
			targetMf = mf
			break
		}
	}

	if targetMf == nil {
		return mdlerrors.NewNotFound("microflow", name.String())
	}

	// Generate MDL output
	var lines []string

	// Documentation
	if targetMf.Documentation != "" {
		lines = append(lines, "/**")
		for docLine := range strings.SplitSeq(targetMf.Documentation, "\n") {
			lines = append(lines, " * "+docLine)
		}
		lines = append(lines, " */")
	}

	// @excluded annotation
	if targetMf.Excluded {
		lines = append(lines, "@excluded")
	}

	// CREATE MICROFLOW header
	qualifiedName := name.Module + "." + name.Name
	if len(targetMf.Parameters) > 0 {
		lines = append(lines, fmt.Sprintf("CREATE OR MODIFY MICROFLOW %s (", qualifiedName))
		for i, param := range targetMf.Parameters {
			paramType := "Object"
			if param.Type != nil {
				paramType = formatMicroflowDataType(ctx, param.Type, entityNames)
			}
			comma := ","
			if i == len(targetMf.Parameters)-1 {
				comma = ""
			}
			lines = append(lines, fmt.Sprintf("  $%s: %s%s", param.Name, paramType, comma))
		}
		lines = append(lines, ")")
	} else {
		lines = append(lines, fmt.Sprintf("CREATE OR MODIFY MICROFLOW %s ()", qualifiedName))
	}

	// Return type
	if targetMf.ReturnType != nil {
		returnType := formatMicroflowDataType(ctx, targetMf.ReturnType, entityNames)
		if returnType != "Void" && returnType != "" {
			returnLine := fmt.Sprintf("RETURNS %s", returnType)
			// Add variable name if specified (AS $VarName)
			if targetMf.ReturnVariableName != "" && targetMf.ReturnVariableName != "Variable" {
				returnLine += fmt.Sprintf(" AS $%s", targetMf.ReturnVariableName)
			}
			lines = append(lines, returnLine)
		}
	}

	// Folder
	if folderPath := h.BuildFolderPath(targetMf.ContainerID); folderPath != "" {
		lines = append(lines, fmt.Sprintf("FOLDER '%s'", folderPath))
	}

	// BEGIN block
	lines = append(lines, "BEGIN")

	// Generate activities
	if targetMf.ObjectCollection != nil && len(targetMf.ObjectCollection.Objects) > 0 {
		activityLines := formatMicroflowActivities(ctx, targetMf, entityNames, microflowNames)
		for _, line := range activityLines {
			lines = append(lines, "  "+line)
		}
	} else {
		lines = append(lines, "  -- No activities")
	}

	lines = append(lines, "END;")

	// Add GRANT EXECUTE if roles are assigned
	if len(targetMf.AllowedModuleRoles) > 0 {
		roles := make([]string, len(targetMf.AllowedModuleRoles))
		for i, r := range targetMf.AllowedModuleRoles {
			roles[i] = string(r)
		}
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("GRANT EXECUTE ON MICROFLOW %s.%s TO %s;",
			name.Module, name.Name, strings.Join(roles, ", ")))
	}

	lines = append(lines, "/")

	// Output
	fmt.Fprintln(ctx.Output, strings.Join(lines, "\n"))
	return nil
}

// describeNanoflow generates re-executable CREATE OR MODIFY NANOFLOW MDL output
// with activities and control flows listed as comments.
func describeNanoflow(ctx *ExecContext, name ast.QualifiedName) error {
	e := ctx.executor
	h, err := getHierarchy(ctx)
	if err != nil {
		return mdlerrors.NewBackend("build hierarchy", err)
	}

	// Build entity name lookup
	entityNames := make(map[model.ID]string)
	domainModels, _ := e.reader.ListDomainModels()
	for _, dm := range domainModels {
		modName := h.GetModuleName(dm.ContainerID)
		for _, entity := range dm.Entities {
			entityNames[entity.ID] = modName + "." + entity.Name
		}
	}

	// Build microflow/nanoflow name lookup (used for call actions)
	microflowNames := make(map[model.ID]string)
	allMicroflows, _ := e.reader.ListMicroflows()
	for _, mf := range allMicroflows {
		microflowNames[mf.ID] = h.GetQualifiedName(mf.ContainerID, mf.Name)
	}

	// Find the nanoflow
	allNanoflows, err := e.reader.ListNanoflows()
	if err != nil {
		return mdlerrors.NewBackend("list nanoflows", err)
	}

	for _, nf := range allNanoflows {
		microflowNames[nf.ID] = h.GetQualifiedName(nf.ContainerID, nf.Name)
	}

	var targetNf *microflows.Nanoflow
	for _, nf := range allNanoflows {
		modID := h.FindModuleID(nf.ContainerID)
		modName := h.GetModuleName(modID)
		if modName == name.Module && nf.Name == name.Name {
			targetNf = nf
			break
		}
	}

	if targetNf == nil {
		return mdlerrors.NewNotFound("nanoflow", name.String())
	}

	var lines []string

	// Documentation
	if targetNf.Documentation != "" {
		lines = append(lines, "/**")
		for docLine := range strings.SplitSeq(targetNf.Documentation, "\n") {
			lines = append(lines, " * "+docLine)
		}
		lines = append(lines, " */")
	}

	// CREATE NANOFLOW header
	qualifiedName := name.Module + "." + name.Name
	if len(targetNf.Parameters) > 0 {
		lines = append(lines, fmt.Sprintf("CREATE OR MODIFY NANOFLOW %s (", qualifiedName))
		for i, param := range targetNf.Parameters {
			paramType := "Object"
			if param.Type != nil {
				paramType = formatMicroflowDataType(ctx, param.Type, entityNames)
			}
			comma := ","
			if i == len(targetNf.Parameters)-1 {
				comma = ""
			}
			lines = append(lines, fmt.Sprintf("  $%s: %s%s", param.Name, paramType, comma))
		}
		lines = append(lines, ")")
	} else {
		lines = append(lines, fmt.Sprintf("CREATE OR MODIFY NANOFLOW %s ()", qualifiedName))
	}

	// Return type
	if targetNf.ReturnType != nil {
		returnType := formatMicroflowDataType(ctx, targetNf.ReturnType, entityNames)
		if returnType != "Void" && returnType != "" {
			lines = append(lines, fmt.Sprintf("RETURNS %s", returnType))
		}
	}

	// Folder
	if folderPath := h.BuildFolderPath(targetNf.ContainerID); folderPath != "" {
		lines = append(lines, fmt.Sprintf("FOLDER '%s'", folderPath))
	}

	// BEGIN block with activities
	lines = append(lines, "BEGIN")

	// Wrap nanoflow in a Microflow to reuse formatMicroflowActivities
	if targetNf.ObjectCollection != nil && len(targetNf.ObjectCollection.Objects) > 0 {
		wrapperMf := &microflows.Microflow{
			ObjectCollection: targetNf.ObjectCollection,
		}
		activityLines := formatMicroflowActivities(ctx, wrapperMf, entityNames, microflowNames)
		for _, line := range activityLines {
			lines = append(lines, "  "+line)
		}
	} else {
		lines = append(lines, "  -- No activities")
	}

	lines = append(lines, "END;")
	lines = append(lines, "/")

	fmt.Fprintln(ctx.Output, strings.Join(lines, "\n"))
	return nil
}

// describeMicroflowToString generates MDL source for a microflow and returns it as a string
// along with a source map mapping node IDs to line ranges.
func describeMicroflowToString(ctx *ExecContext, name ast.QualifiedName) (string, map[string]elkSourceRange, error) {
	e := ctx.executor
	h, err := getHierarchy(ctx)
	if err != nil {
		return "", nil, mdlerrors.NewBackend("build hierarchy", err)
	}

	entityNames := make(map[model.ID]string)
	domainModels, _ := e.reader.ListDomainModels()
	for _, dm := range domainModels {
		modName := h.GetModuleName(dm.ContainerID)
		for _, entity := range dm.Entities {
			entityNames[entity.ID] = modName + "." + entity.Name
		}
	}

	microflowNames := make(map[model.ID]string)
	allMicroflows, err := e.reader.ListMicroflows()
	if err != nil {
		return "", nil, mdlerrors.NewBackend("list microflows", err)
	}
	for _, mf := range allMicroflows {
		microflowNames[mf.ID] = h.GetQualifiedName(mf.ContainerID, mf.Name)
	}

	var targetMf *microflows.Microflow
	for _, mf := range allMicroflows {
		modID := h.FindModuleID(mf.ContainerID)
		modName := h.GetModuleName(modID)
		if modName == name.Module && mf.Name == name.Name {
			targetMf = mf
			break
		}
	}

	if targetMf == nil {
		return "", nil, mdlerrors.NewNotFound("microflow", name.String())
	}

	sourceMap := make(map[string]elkSourceRange)
	mdl := renderMicroflowMDL(ctx, targetMf, name, entityNames, microflowNames, sourceMap)
	return mdl, sourceMap, nil
}

// renderMicroflowMDL formats a parsed Microflow as MDL text.
//
// Shared by DESCRIBE MICROFLOW and `diff-local`, so both paths produce the
// same output. entityNames/microflowNames provide ID → qualified-name
// resolution; pass empty maps if unavailable (types will fall back to
// "Object"/"List" stubs). If sourceMap is non-nil it will be populated with
// ELK node IDs → line ranges for visualization; pass nil when not needed.
func renderMicroflowMDL(
	ctx *ExecContext,
	mf *microflows.Microflow,
	name ast.QualifiedName,
	entityNames map[model.ID]string,
	microflowNames map[model.ID]string,
	sourceMap map[string]elkSourceRange,
) string {
	var lines []string

	if mf.Documentation != "" {
		lines = append(lines, "/**")
		for docLine := range strings.SplitSeq(mf.Documentation, "\n") {
			lines = append(lines, " * "+docLine)
		}
		lines = append(lines, " */")
	}

	if mf.Excluded {
		lines = append(lines, "@excluded")
	}

	qualifiedName := name.Module + "." + name.Name
	if len(mf.Parameters) > 0 {
		lines = append(lines, fmt.Sprintf("CREATE OR MODIFY MICROFLOW %s (", qualifiedName))
		for i, param := range mf.Parameters {
			paramType := "Object"
			if param.Type != nil {
				paramType = formatMicroflowDataType(ctx, param.Type, entityNames)
			}
			comma := ","
			if i == len(mf.Parameters)-1 {
				comma = ""
			}
			lines = append(lines, fmt.Sprintf("  $%s: %s%s", param.Name, paramType, comma))
		}
		lines = append(lines, ")")
	} else {
		lines = append(lines, fmt.Sprintf("CREATE OR MODIFY MICROFLOW %s ()", qualifiedName))
	}

	if mf.ReturnType != nil {
		returnType := formatMicroflowDataType(ctx, mf.ReturnType, entityNames)
		if returnType != "Void" && returnType != "" {
			returnLine := fmt.Sprintf("RETURNS %s", returnType)
			if mf.ReturnVariableName != "" && mf.ReturnVariableName != "Variable" {
				returnLine += fmt.Sprintf(" AS $%s", mf.ReturnVariableName)
			}
			lines = append(lines, returnLine)
		}
	}

	lines = append(lines, "BEGIN")
	headerLineCount := len(lines)

	if mf.ObjectCollection != nil && len(mf.ObjectCollection.Objects) > 0 {
		var activityLines []string
		if sourceMap != nil {
			activityLines = formatMicroflowActivitiesWithSourceMap(ctx, mf, entityNames, microflowNames, sourceMap, headerLineCount)
		} else {
			activityLines = formatMicroflowActivities(ctx, mf, entityNames, microflowNames)
		}
		for _, line := range activityLines {
			lines = append(lines, "  "+line)
		}
	} else {
		lines = append(lines, "  -- No activities")
	}

	lines = append(lines, "END;")

	if len(mf.AllowedModuleRoles) > 0 {
		roles := make([]string, len(mf.AllowedModuleRoles))
		for i, r := range mf.AllowedModuleRoles {
			roles[i] = string(r)
		}
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("GRANT EXECUTE ON MICROFLOW %s.%s TO %s;",
			name.Module, name.Name, strings.Join(roles, ", ")))
	}

	lines = append(lines, "/")

	return strings.Join(lines, "\n")
}

// formatMicroflowDataType formats a microflow data type for MDL output.
func formatMicroflowDataType(ctx *ExecContext, dt microflows.DataType, entityNames map[model.ID]string) string {
	if dt == nil {
		return "Unknown"
	}

	switch t := dt.(type) {
	case *microflows.BooleanType:
		return "Boolean"
	case *microflows.IntegerType:
		return "Integer"
	case *microflows.LongType:
		return "Long"
	case *microflows.DecimalType:
		return "Decimal"
	case *microflows.StringType:
		return "String"
	case *microflows.DateTimeType:
		return "DateTime"
	case *microflows.DateType:
		return "Date"
	case *microflows.BinaryType:
		return "Binary"
	case *microflows.VoidType:
		return "Void"
	case *microflows.ObjectType:
		// First try EntityQualifiedName (BY_NAME_REFERENCE), then fall back to EntityID lookup
		if t.EntityQualifiedName != "" {
			return t.EntityQualifiedName
		}
		if name, ok := entityNames[t.EntityID]; ok {
			return name
		}
		return "Object"
	case *microflows.ListType:
		// First try EntityQualifiedName (BY_NAME_REFERENCE), then fall back to EntityID lookup
		if t.EntityQualifiedName != "" {
			return "List of " + t.EntityQualifiedName
		}
		if name, ok := entityNames[t.EntityID]; ok {
			return "List of " + name
		}
		return "List"
	case *microflows.EnumerationType:
		if t.EnumerationQualifiedName != "" {
			return "ENUM " + t.EnumerationQualifiedName
		}
		return "Enumeration"
	default:
		return dt.GetTypeName()
	}
}

// formatMicroflowActivities generates MDL statements for microflow activities.
func formatMicroflowActivities(
	ctx *ExecContext,
	mf *microflows.Microflow,
	entityNames map[model.ID]string,
	microflowNames map[model.ID]string,
) []string {
	if mf.ObjectCollection == nil {
		return []string{"-- DEBUG: ObjectCollection is nil"}
	}

	// Build activity map by ID for flow traversal
	activityMap := make(map[model.ID]microflows.MicroflowObject)
	var startID model.ID

	for _, obj := range mf.ObjectCollection.Objects {
		activityMap[obj.GetID()] = obj
		if _, ok := obj.(*microflows.StartEvent); ok {
			startID = obj.GetID()
		}
	}

	// Build flow graph: map from origin ID to flows (sorted by OriginConnectionIndex)
	flowsByOrigin := make(map[model.ID][]*microflows.SequenceFlow)
	for _, flow := range mf.ObjectCollection.Flows {
		flowsByOrigin[flow.OriginID] = append(flowsByOrigin[flow.OriginID], flow)
	}

	var lines []string

	// Sort flows by OriginConnectionIndex for each origin
	for originID := range flowsByOrigin {
		flows := flowsByOrigin[originID]
		// Simple bubble sort since typically only 2 flows per split
		for i := 0; i < len(flows)-1; i++ {
			for j := i + 1; j < len(flows); j++ {
				if flows[i].OriginConnectionIndex > flows[j].OriginConnectionIndex {
					flows[i], flows[j] = flows[j], flows[i]
				}
			}
		}
	}

	// Find the merge point for each split (where branches converge)
	splitMergeMap := findSplitMergePoints(ctx, mf.ObjectCollection, activityMap)

	// Traverse the flow graph recursively
	visited := make(map[model.ID]bool)

	// Build annotation map for @annotation emission
	annotationsByTarget := buildAnnotationsByTarget(mf.ObjectCollection)

	traverseFlow(ctx, startID, activityMap, flowsByOrigin, splitMergeMap, visited, entityNames, microflowNames, &lines, 0, nil, 0, annotationsByTarget)

	return lines
}

// formatMicroflowActivitiesWithSourceMap generates MDL statements and populates a source map
// mapping ELK node IDs ("node-<objectID>") to line ranges (0-indexed) in the full MDL output.
// headerLineCount is the number of lines before the BEGIN body (to compute absolute line numbers).
func formatMicroflowActivitiesWithSourceMap(
	ctx *ExecContext,
	mf *microflows.Microflow,
	entityNames map[model.ID]string,
	microflowNames map[model.ID]string,
	sourceMap map[string]elkSourceRange,
	headerLineCount int,
) []string {
	if mf.ObjectCollection == nil {
		return []string{"-- DEBUG: ObjectCollection is nil"}
	}

	activityMap := make(map[model.ID]microflows.MicroflowObject)
	var startID model.ID

	for _, obj := range mf.ObjectCollection.Objects {
		activityMap[obj.GetID()] = obj
		if _, ok := obj.(*microflows.StartEvent); ok {
			startID = obj.GetID()
		}
	}

	flowsByOrigin := make(map[model.ID][]*microflows.SequenceFlow)
	for _, flow := range mf.ObjectCollection.Flows {
		flowsByOrigin[flow.OriginID] = append(flowsByOrigin[flow.OriginID], flow)
	}

	var lines []string

	for originID := range flowsByOrigin {
		flows := flowsByOrigin[originID]
		for i := 0; i < len(flows)-1; i++ {
			for j := i + 1; j < len(flows); j++ {
				if flows[i].OriginConnectionIndex > flows[j].OriginConnectionIndex {
					flows[i], flows[j] = flows[j], flows[i]
				}
			}
		}
	}

	splitMergeMap := findSplitMergePoints(ctx, mf.ObjectCollection, activityMap)
	visited := make(map[model.ID]bool)

	// Build annotation map for @annotation emission
	annotationsByTarget := buildAnnotationsByTarget(mf.ObjectCollection)

	traverseFlow(ctx, startID, activityMap, flowsByOrigin, splitMergeMap, visited, entityNames, microflowNames, &lines, 0, sourceMap, headerLineCount, annotationsByTarget)

	return lines
}

// findSplitMergePoints finds the corresponding merge point for each exclusive split.
func findSplitMergePoints(
	ctx *ExecContext,
	oc *microflows.MicroflowObjectCollection,
	activityMap map[model.ID]microflows.MicroflowObject,
) map[model.ID]model.ID {
	result := make(map[model.ID]model.ID)

	// Build flow graph for forward traversal
	flowsByOrigin := make(map[model.ID][]*microflows.SequenceFlow)
	for _, flow := range oc.Flows {
		flowsByOrigin[flow.OriginID] = append(flowsByOrigin[flow.OriginID], flow)
	}

	// For each ExclusiveSplit, find its merge point
	for _, obj := range oc.Objects {
		if _, ok := obj.(*microflows.ExclusiveSplit); ok {
			splitID := obj.GetID()
			// Find merge by following both branches until they converge
			mergeID := findMergeForSplit(ctx, splitID, flowsByOrigin, activityMap)
			if mergeID != "" {
				result[splitID] = mergeID
			}
		}
	}

	return result
}

// findMergeForSplit finds the ExclusiveMerge where branches from a split converge.
func findMergeForSplit(
	ctx *ExecContext,
	splitID model.ID,
	flowsByOrigin map[model.ID][]*microflows.SequenceFlow,
	activityMap map[model.ID]microflows.MicroflowObject,
) model.ID {
	flows := flowsByOrigin[splitID]
	if len(flows) < 2 {
		return ""
	}

	// Follow each branch and collect all reachable nodes
	branch0Nodes := collectReachableNodes(ctx, flows[0].DestinationID, flowsByOrigin, activityMap, make(map[model.ID]bool))
	branch1Nodes := collectReachableNodes(ctx, flows[1].DestinationID, flowsByOrigin, activityMap, make(map[model.ID]bool))

	// Find the first common node that is an ExclusiveMerge
	// This is a simplification - we look for the first merge point reachable from both branches
	for nodeID := range branch0Nodes {
		if branch1Nodes[nodeID] {
			if _, ok := activityMap[nodeID].(*microflows.ExclusiveMerge); ok {
				return nodeID
			}
		}
	}

	return ""
}

// collectReachableNodes collects all nodes reachable from a starting node.
func collectReachableNodes(
	ctx *ExecContext,
	startID model.ID,
	flowsByOrigin map[model.ID][]*microflows.SequenceFlow,
	activityMap map[model.ID]microflows.MicroflowObject,
	visited map[model.ID]bool,
) map[model.ID]bool {
	result := make(map[model.ID]bool)

	var traverse func(id model.ID)
	traverse = func(id model.ID) {
		if visited[id] {
			return
		}
		visited[id] = true
		result[id] = true

		for _, flow := range flowsByOrigin[id] {
			traverse(flow.DestinationID)
		}
	}

	traverse(startID)
	return result
}

// --- Executor method wrappers for callers in unmigrated code ---
