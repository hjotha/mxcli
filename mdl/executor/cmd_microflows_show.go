// SPDX-License-Identifier: Apache-2.0

// Package executor - Microflow SHOW/DESCRIBE commands
package executor

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/microflows"
)

// showMicroflows handles SHOW MICROFLOWS command.
func (e *Executor) showMicroflows(moduleName string) error {
	// Get hierarchy for module/folder resolution
	h, err := e.getHierarchy()
	if err != nil {
		return fmt.Errorf("failed to build hierarchy: %w", err)
	}

	// Get all microflows
	microflows, err := e.reader.ListMicroflows()
	if err != nil {
		return fmt.Errorf("failed to list microflows: %w", err)
	}

	// Collect rows and calculate column widths
	type row struct {
		qualifiedName string
		module        string
		name          string
		folderPath    string
		params        int
		activities    int
		complexity    int
		returnType    string
	}
	var rows []row
	qnWidth := len("Qualified Name")
	modWidth := len("Module")
	nameWidth := len("Name")
	pathWidth := len("Folder")
	paramsWidth := len("Params")
	actWidth := len("Actions")
	compWidth := len("McCabe")
	returnWidth := len("Returns")

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

			rows = append(rows, row{qualifiedName, modName, mf.Name, folderPath, len(mf.Parameters), activityCount, complexity, returnType})
			if len(qualifiedName) > qnWidth {
				qnWidth = len(qualifiedName)
			}
			if len(modName) > modWidth {
				modWidth = len(modName)
			}
			if len(mf.Name) > nameWidth {
				nameWidth = len(mf.Name)
			}
			if len(folderPath) > pathWidth {
				pathWidth = len(folderPath)
			}
			paramsStr := fmt.Sprintf("%d", len(mf.Parameters))
			if len(paramsStr) > paramsWidth {
				paramsWidth = len(paramsStr)
			}
			actStr := fmt.Sprintf("%d", activityCount)
			if len(actStr) > actWidth {
				actWidth = len(actStr)
			}
			compStr := fmt.Sprintf("%d", complexity)
			if len(compStr) > compWidth {
				compWidth = len(compStr)
			}
			if len(returnType) > returnWidth {
				returnWidth = len(returnType)
			}
		}
	}

	// Sort by qualified name
	sort.Slice(rows, func(i, j int) bool {
		return strings.ToLower(rows[i].qualifiedName) < strings.ToLower(rows[j].qualifiedName)
	})

	// Markdown table with aligned columns
	fmt.Fprintf(e.output, "| %-*s | %-*s | %-*s | %-*s | %-*s | %-*s | %-*s | %-*s |\n",
		qnWidth, "Qualified Name", modWidth, "Module", nameWidth, "Name",
		pathWidth, "Folder", paramsWidth, "Params", actWidth, "Actions", compWidth, "McCabe", returnWidth, "Returns")
	fmt.Fprintf(e.output, "|-%s-|-%s-|-%s-|-%s-|-%s-|-%s-|-%s-|-%s-|\n",
		strings.Repeat("-", qnWidth), strings.Repeat("-", modWidth), strings.Repeat("-", nameWidth),
		strings.Repeat("-", pathWidth), strings.Repeat("-", paramsWidth), strings.Repeat("-", actWidth),
		strings.Repeat("-", compWidth), strings.Repeat("-", returnWidth))
	for _, r := range rows {
		fmt.Fprintf(e.output, "| %-*s | %-*s | %-*s | %-*s | %-*d | %-*d | %-*d | %-*s |\n",
			qnWidth, r.qualifiedName, modWidth, r.module, nameWidth, r.name,
			pathWidth, r.folderPath, paramsWidth, r.params, actWidth, r.activities,
			compWidth, r.complexity, returnWidth, r.returnType)
	}
	fmt.Fprintf(e.output, "\n(%d microflows)\n", len(rows))
	return nil
}

// showNanoflows handles SHOW NANOFLOWS command.
func (e *Executor) showNanoflows(moduleName string) error {
	// Get hierarchy for module/folder resolution
	h, err := e.getHierarchy()
	if err != nil {
		return fmt.Errorf("failed to build hierarchy: %w", err)
	}

	// Get all nanoflows
	nanoflows, err := e.reader.ListNanoflows()
	if err != nil {
		return fmt.Errorf("failed to list nanoflows: %w", err)
	}

	// Collect rows and calculate column widths
	type row struct {
		qualifiedName string
		module        string
		name          string
		folderPath    string
		params        int
		activities    int
		complexity    int
		returnType    string
	}
	var rows []row
	qnWidth := len("Qualified Name")
	modWidth := len("Module")
	nameWidth := len("Name")
	pathWidth := len("Folder")
	paramsWidth := len("Params")
	actWidth := len("Actions")
	compWidth := len("McCabe")
	returnWidth := len("Returns")

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

			rows = append(rows, row{qualifiedName, modName, nf.Name, folderPath, len(nf.Parameters), activityCount, complexity, returnType})
			if len(qualifiedName) > qnWidth {
				qnWidth = len(qualifiedName)
			}
			if len(modName) > modWidth {
				modWidth = len(modName)
			}
			if len(nf.Name) > nameWidth {
				nameWidth = len(nf.Name)
			}
			if len(folderPath) > pathWidth {
				pathWidth = len(folderPath)
			}
			paramsStr := fmt.Sprintf("%d", len(nf.Parameters))
			if len(paramsStr) > paramsWidth {
				paramsWidth = len(paramsStr)
			}
			actStr := fmt.Sprintf("%d", activityCount)
			if len(actStr) > actWidth {
				actWidth = len(actStr)
			}
			compStr := fmt.Sprintf("%d", complexity)
			if len(compStr) > compWidth {
				compWidth = len(compStr)
			}
			if len(returnType) > returnWidth {
				returnWidth = len(returnType)
			}
		}
	}

	// Sort by qualified name
	sort.Slice(rows, func(i, j int) bool {
		return strings.ToLower(rows[i].qualifiedName) < strings.ToLower(rows[j].qualifiedName)
	})

	// Markdown table with aligned columns
	fmt.Fprintf(e.output, "| %-*s | %-*s | %-*s | %-*s | %-*s | %-*s | %-*s | %-*s |\n",
		qnWidth, "Qualified Name", modWidth, "Module", nameWidth, "Name",
		pathWidth, "Folder", paramsWidth, "Params", actWidth, "Actions", compWidth, "McCabe", returnWidth, "Returns")
	fmt.Fprintf(e.output, "|-%s-|-%s-|-%s-|-%s-|-%s-|-%s-|-%s-|-%s-|\n",
		strings.Repeat("-", qnWidth), strings.Repeat("-", modWidth), strings.Repeat("-", nameWidth),
		strings.Repeat("-", pathWidth), strings.Repeat("-", paramsWidth), strings.Repeat("-", actWidth),
		strings.Repeat("-", compWidth), strings.Repeat("-", returnWidth))
	for _, r := range rows {
		fmt.Fprintf(e.output, "| %-*s | %-*s | %-*s | %-*s | %-*d | %-*d | %-*d | %-*s |\n",
			qnWidth, r.qualifiedName, modWidth, r.module, nameWidth, r.name,
			pathWidth, r.folderPath, paramsWidth, r.params, actWidth, r.activities,
			compWidth, r.complexity, returnWidth, r.returnType)
	}
	fmt.Fprintf(e.output, "\n(%d nanoflows)\n", len(rows))
	return nil
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
func (e *Executor) describeMicroflow(name ast.QualifiedName) error {
	// Get hierarchy for module/folder resolution
	h, err := e.getHierarchy()
	if err != nil {
		return fmt.Errorf("failed to build hierarchy: %w", err)
	}

	// Build entity name lookup for resolving entity IDs
	entityNames := make(map[model.ID]string)
	domainModels, _ := e.reader.ListDomainModels()
	for _, dm := range domainModels {
		modName := h.GetModuleName(dm.ContainerID)
		for _, entity := range dm.Entities {
			entityNames[entity.ID] = modName + "." + entity.Name
		}
	}

	// Build microflow name lookup for resolving microflow IDs
	microflowNames := make(map[model.ID]string)

	// Build page name lookup for resolving page IDs
	pageNames := make(map[model.ID]string)
	allPages, _ := e.reader.ListPages()
	for _, p := range allPages {
		pageNames[p.ID] = h.GetQualifiedName(p.ContainerID, p.Name)
	}

	// Find the microflow
	allMicroflows, err := e.reader.ListMicroflows()
	if err != nil {
		return fmt.Errorf("failed to list microflows: %w", err)
	}

	// Build microflow name lookup
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
		return fmt.Errorf("microflow not found: %s", name)
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

	// CREATE MICROFLOW header
	qualifiedName := name.Module + "." + name.Name
	if len(targetMf.Parameters) > 0 {
		lines = append(lines, fmt.Sprintf("CREATE OR REPLACE MICROFLOW %s (", qualifiedName))
		for i, param := range targetMf.Parameters {
			paramType := "Object"
			if param.Type != nil {
				paramType = e.formatMicroflowDataType(param.Type, entityNames)
			}
			comma := ","
			if i == len(targetMf.Parameters)-1 {
				comma = ""
			}
			lines = append(lines, fmt.Sprintf("  $%s: %s%s", param.Name, paramType, comma))
		}
		lines = append(lines, ")")
	} else {
		lines = append(lines, fmt.Sprintf("CREATE OR REPLACE MICROFLOW %s ()", qualifiedName))
	}

	// Return type
	if targetMf.ReturnType != nil {
		returnType := e.formatMicroflowDataType(targetMf.ReturnType, entityNames)
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
		activityLines := e.formatMicroflowActivities(targetMf, entityNames, microflowNames)
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
	fmt.Fprintln(e.output, strings.Join(lines, "\n"))
	return nil
}

// describeMicroflowToString generates MDL source for a microflow and returns it as a string
// along with a source map mapping node IDs to line ranges.
func (e *Executor) describeMicroflowToString(name ast.QualifiedName) (string, map[string]elkSourceRange, error) {
	h, err := e.getHierarchy()
	if err != nil {
		return "", nil, fmt.Errorf("failed to build hierarchy: %w", err)
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
		return "", nil, fmt.Errorf("failed to list microflows: %w", err)
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
		return "", nil, fmt.Errorf("microflow not found: %s", name)
	}

	var lines []string

	if targetMf.Documentation != "" {
		lines = append(lines, "/**")
		for docLine := range strings.SplitSeq(targetMf.Documentation, "\n") {
			lines = append(lines, " * "+docLine)
		}
		lines = append(lines, " */")
	}

	qualifiedName := name.Module + "." + name.Name
	if len(targetMf.Parameters) > 0 {
		lines = append(lines, fmt.Sprintf("CREATE OR REPLACE MICROFLOW %s (", qualifiedName))
		for i, param := range targetMf.Parameters {
			paramType := "Object"
			if param.Type != nil {
				paramType = e.formatMicroflowDataType(param.Type, entityNames)
			}
			comma := ","
			if i == len(targetMf.Parameters)-1 {
				comma = ""
			}
			lines = append(lines, fmt.Sprintf("  $%s: %s%s", param.Name, paramType, comma))
		}
		lines = append(lines, ")")
	} else {
		lines = append(lines, fmt.Sprintf("CREATE OR REPLACE MICROFLOW %s ()", qualifiedName))
	}

	if targetMf.ReturnType != nil {
		returnType := e.formatMicroflowDataType(targetMf.ReturnType, entityNames)
		if returnType != "Void" && returnType != "" {
			returnLine := fmt.Sprintf("RETURNS %s", returnType)
			if targetMf.ReturnVariableName != "" && targetMf.ReturnVariableName != "Variable" {
				returnLine += fmt.Sprintf(" AS $%s", targetMf.ReturnVariableName)
			}
			lines = append(lines, returnLine)
		}
	}

	lines = append(lines, "BEGIN")
	headerLineCount := len(lines) // lines before activity body

	sourceMap := make(map[string]elkSourceRange)

	if targetMf.ObjectCollection != nil && len(targetMf.ObjectCollection.Objects) > 0 {
		activityLines := e.formatMicroflowActivitiesWithSourceMap(targetMf, entityNames, microflowNames, sourceMap, headerLineCount)
		for _, line := range activityLines {
			lines = append(lines, "  "+line)
		}
	} else {
		lines = append(lines, "  -- No activities")
	}

	lines = append(lines, "END;")

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

	return strings.Join(lines, "\n"), sourceMap, nil
}

// formatMicroflowDataType formats a microflow data type for MDL output.
func (e *Executor) formatMicroflowDataType(dt microflows.DataType, entityNames map[model.ID]string) string {
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
func (e *Executor) formatMicroflowActivities(
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
	splitMergeMap := e.findSplitMergePoints(mf.ObjectCollection, activityMap)

	// Traverse the flow graph recursively
	visited := make(map[model.ID]bool)

	// Build annotation map for @annotation emission
	annotationsByTarget := buildAnnotationsByTarget(mf.ObjectCollection)

	e.traverseFlow(startID, activityMap, flowsByOrigin, splitMergeMap, visited, entityNames, microflowNames, &lines, 0, nil, 0, annotationsByTarget)

	return lines
}

// formatMicroflowActivitiesWithSourceMap generates MDL statements and populates a source map
// mapping ELK node IDs ("node-<objectID>") to line ranges (0-indexed) in the full MDL output.
// headerLineCount is the number of lines before the BEGIN body (to compute absolute line numbers).
func (e *Executor) formatMicroflowActivitiesWithSourceMap(
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

	splitMergeMap := e.findSplitMergePoints(mf.ObjectCollection, activityMap)
	visited := make(map[model.ID]bool)

	// Build annotation map for @annotation emission
	annotationsByTarget := buildAnnotationsByTarget(mf.ObjectCollection)

	e.traverseFlow(startID, activityMap, flowsByOrigin, splitMergeMap, visited, entityNames, microflowNames, &lines, 0, sourceMap, headerLineCount, annotationsByTarget)

	return lines
}

// findSplitMergePoints finds the corresponding merge point for each exclusive split.
func (e *Executor) findSplitMergePoints(
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
			mergeID := e.findMergeForSplit(splitID, flowsByOrigin, activityMap)
			if mergeID != "" {
				result[splitID] = mergeID
			}
		}
	}

	return result
}

// findMergeForSplit finds the ExclusiveMerge where branches from a split converge.
func (e *Executor) findMergeForSplit(
	splitID model.ID,
	flowsByOrigin map[model.ID][]*microflows.SequenceFlow,
	activityMap map[model.ID]microflows.MicroflowObject,
) model.ID {
	flows := flowsByOrigin[splitID]
	if len(flows) < 2 {
		return ""
	}

	// Follow each branch and collect all reachable nodes
	branch0Nodes := e.collectReachableNodes(flows[0].DestinationID, flowsByOrigin, activityMap, make(map[model.ID]bool))
	branch1Nodes := e.collectReachableNodes(flows[1].DestinationID, flowsByOrigin, activityMap, make(map[model.ID]bool))

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
func (e *Executor) collectReachableNodes(
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
