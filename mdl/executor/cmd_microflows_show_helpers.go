// SPDX-License-Identifier: Apache-2.0

// Package executor - Microflow flow traversal and helper functions.
package executor

import (
	"context"
	"fmt"
	"strings"

	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/microflows"
)

// buildAnnotationsByTarget builds a map from activity ID to annotation captions.
// It joins AnnotationFlows (destination → activity) with Annotation objects (caption).
func buildAnnotationsByTarget(oc *microflows.MicroflowObjectCollection) map[model.ID][]string {
	result := make(map[model.ID][]string)
	if oc == nil {
		return result
	}

	// Build a map of annotation IDs to their captions
	annotCaptions := make(map[model.ID]string)
	for _, obj := range oc.Objects {
		if annot, ok := obj.(*microflows.Annotation); ok {
			annotCaptions[annot.ID] = annot.Caption
		}
	}

	// Map each annotation flow's destination (the activity) to the annotation's caption
	for _, af := range oc.AnnotationFlows {
		if caption, ok := annotCaptions[af.OriginID]; ok && caption != "" {
			result[af.DestinationID] = append(result[af.DestinationID], caption)
		}
	}

	return result
}

// collectFreeAnnotations returns captions for annotations not referenced by any AnnotationFlow.
func collectFreeAnnotations(oc *microflows.MicroflowObjectCollection) []string {
	if oc == nil {
		return nil
	}

	// Collect annotation IDs that are referenced by flows
	referencedAnnotations := make(map[model.ID]bool)
	for _, af := range oc.AnnotationFlows {
		referencedAnnotations[af.OriginID] = true
	}

	var result []string
	for _, obj := range oc.Objects {
		if annot, ok := obj.(*microflows.Annotation); ok {
			if !referencedAnnotations[annot.ID] && annot.Caption != "" {
				result = append(result, annot.Caption)
			}
		}
	}
	return result
}

// emitObjectAnnotations emits @position, @caption, @color, and @annotation lines
// for a microflow object before its statement.
func emitObjectAnnotations(obj microflows.MicroflowObject, lines *[]string, indentStr string, annotationsByTarget map[model.ID][]string) {
	currentID := obj.GetID()

	// @position (always emit)
	pos := obj.GetPosition()
	*lines = append(*lines, indentStr+fmt.Sprintf("@position(%d, %d)", pos.X, pos.Y))

	// @excluded, @caption, and @color (only for ActionActivity)
	if activity, ok := obj.(*microflows.ActionActivity); ok {
		if activity.Disabled {
			*lines = append(*lines, indentStr+"@excluded")
		}
		if !activity.AutoGenerateCaption && activity.Caption != "" {
			escapedCaption := strings.ReplaceAll(activity.Caption, "'", "''")
			*lines = append(*lines, indentStr+fmt.Sprintf("@caption '%s'", escapedCaption))
		}
		if activity.BackgroundColor != "" && activity.BackgroundColor != "Default" {
			*lines = append(*lines, indentStr+fmt.Sprintf("@color %s", activity.BackgroundColor))
		}
	}

	// @annotation (attached Annotation objects)
	if annotationsByTarget != nil {
		for _, caption := range annotationsByTarget[currentID] {
			escapedCaption := strings.ReplaceAll(caption, "'", "''")
			*lines = append(*lines, indentStr+fmt.Sprintf("@annotation '%s'", escapedCaption))
		}
	}
}

// emitActivityStatement appends the formatted activity statement (with error handling)
// to the lines slice. It handles ON ERROR CONTINUE/ROLLBACK suffixes and custom error
// handler blocks. This replaces the copy-pasted error handling logic in each traversal function.
func emitActivityStatement(
	ctx *ExecContext,
	obj microflows.MicroflowObject,
	stmt string,
	flowsByOrigin map[model.ID][]*microflows.SequenceFlow,
	activityMap map[model.ID]microflows.MicroflowObject,
	entityNames map[model.ID]string,
	microflowNames map[model.ID]string,
	lines *[]string,
	indentStr string,
	annotationsByTarget map[model.ID][]string,
) {
	if stmt == "" {
		return
	}

	// Emit @ annotations before the statement
	emitObjectAnnotations(obj, lines, indentStr, annotationsByTarget)

	currentID := obj.GetID()
	flows := flowsByOrigin[currentID]
	errorHandlerFlow := findErrorHandlerFlow(flows)

	activity, isAction := obj.(*microflows.ActionActivity)
	if !isAction {
		*lines = append(*lines, indentStr+stmt)
		return
	}

	errType := getActionErrorHandlingType(activity)
	suffix := formatErrorHandlingSuffix(errType)

	if errorHandlerFlow != nil && hasCustomErrorHandler(errType) {
		errStmts := collectErrorHandlerStatements(
			ctx,
			errorHandlerFlow.DestinationID,
			activityMap, flowsByOrigin, entityNames, microflowNames,
		)

		stmtWithoutSemi := strings.TrimSuffix(strings.TrimSpace(stmt), ";")

		errorSuffix := suffix
		if errorSuffix == "" {
			errorSuffix = " on error without rollback"
		}

		if len(errStmts) == 0 {
			*lines = append(*lines, indentStr+stmtWithoutSemi+errorSuffix+" { };")
		} else {
			*lines = append(*lines, indentStr+stmtWithoutSemi+errorSuffix+" {")
			for _, errStmt := range errStmts {
				*lines = append(*lines, indentStr+"  "+errStmt)
			}
			*lines = append(*lines, indentStr+"};")
		}
	} else if suffix != "" {
		stmtWithoutSemi := strings.TrimSuffix(strings.TrimSpace(stmt), ";")
		*lines = append(*lines, indentStr+stmtWithoutSemi+suffix+";")
	} else {
		*lines = append(*lines, indentStr+stmt)
	}
}

// recordSourceMap records the source map entry for a node if sourceMap is non-nil.
func recordSourceMap(sourceMap map[string]elkSourceRange, nodeID model.ID, startLine, endLine int) {
	if sourceMap != nil && endLine >= startLine {
		sourceMap["node-"+string(nodeID)] = elkSourceRange{StartLine: startLine, EndLine: endLine}
	}
}

// traverseFlow recursively traverses the microflow graph and generates MDL statements.
// When sourceMap is non-nil, it also records line ranges for each activity node.
func traverseFlow(
	ctx *ExecContext,
	currentID model.ID,
	activityMap map[model.ID]microflows.MicroflowObject,
	flowsByOrigin map[model.ID][]*microflows.SequenceFlow,
	splitMergeMap map[model.ID]model.ID,
	visited map[model.ID]bool,
	entityNames map[model.ID]string,
	microflowNames map[model.ID]string,
	lines *[]string,
	indent int,
	sourceMap map[string]elkSourceRange,
	headerLineCount int,
	annotationsByTarget map[model.ID][]string,
) {
	if currentID == "" || visited[currentID] {
		return
	}

	obj := activityMap[currentID]
	if obj == nil {
		return
	}

	// Check if this is a merge point - if so, don't process it here (it will be handled by the split)
	if _, isMerge := obj.(*microflows.ExclusiveMerge); isMerge {
		return
	}

	visited[currentID] = true

	stmt := formatActivity(ctx, obj, entityNames, microflowNames)
	indentStr := strings.Repeat("  ", indent)

	// Handle ExclusiveSplit specially - need to process both branches
	if _, isSplit := obj.(*microflows.ExclusiveSplit); isSplit {
		startLine := len(*lines) + headerLineCount
		if stmt != "" {
			emitObjectAnnotations(obj, lines, indentStr, annotationsByTarget)
			*lines = append(*lines, indentStr+stmt)
		}

		flows := flowsByOrigin[currentID]
		mergeID := splitMergeMap[currentID]

		trueFlow, falseFlow := findBranchFlows(flows)

		// Guard pattern: true branch is a single EndEvent (RETURN),
		// but only when the false branch does NOT also end directly.
		// If both branches return, use normal IF/ELSE/END IF.
		isGuard := false
		if trueFlow != nil {
			if _, isEnd := activityMap[trueFlow.DestinationID].(*microflows.EndEvent); isEnd {
				isGuard = true
				// Not a guard if both branches return directly
				if falseFlow != nil {
					if _, falseIsEnd := activityMap[falseFlow.DestinationID].(*microflows.EndEvent); falseIsEnd {
						isGuard = false
					}
				}
			}
		}

		if isGuard {
			traverseFlowUntilMerge(ctx, trueFlow.DestinationID, mergeID, activityMap, flowsByOrigin, splitMergeMap, visited, entityNames, microflowNames, lines, indent+1, sourceMap, headerLineCount, annotationsByTarget)
			*lines = append(*lines, indentStr+"end if;")
			recordSourceMap(sourceMap, currentID, startLine, len(*lines)+headerLineCount-1)

			// Continue from the false branch (skip through merge if present)
			if falseFlow != nil {
				contID := falseFlow.DestinationID
				if _, isMerge := activityMap[contID].(*microflows.ExclusiveMerge); isMerge {
					visited[contID] = true
					for _, flow := range flowsByOrigin[contID] {
						contID = flow.DestinationID
						break
					}
				}
				traverseFlow(ctx, contID, activityMap, flowsByOrigin, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
			}
		} else {
			if trueFlow != nil {
				traverseFlowUntilMerge(ctx, trueFlow.DestinationID, mergeID, activityMap, flowsByOrigin, splitMergeMap, visited, entityNames, microflowNames, lines, indent+1, sourceMap, headerLineCount, annotationsByTarget)
			}

			if falseFlow != nil {
				*lines = append(*lines, indentStr+"else")
				visitedFalseBranch := make(map[model.ID]bool)
				for id := range visited {
					visitedFalseBranch[id] = true
				}
				traverseFlowUntilMerge(ctx, falseFlow.DestinationID, mergeID, activityMap, flowsByOrigin, splitMergeMap, visitedFalseBranch, entityNames, microflowNames, lines, indent+1, sourceMap, headerLineCount, annotationsByTarget)
			}

			*lines = append(*lines, indentStr+"end if;")
			recordSourceMap(sourceMap, currentID, startLine, len(*lines)+headerLineCount-1)

			// Continue after the merge point
			if mergeID != "" {
				visited[mergeID] = true
				nextFlows := flowsByOrigin[mergeID]
				for _, flow := range nextFlows {
					traverseFlow(ctx, flow.DestinationID, activityMap, flowsByOrigin, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
				}
			}
		}
		return
	}

	// Handle LoopedActivity specially - need to process loop body
	if loop, isLoop := obj.(*microflows.LoopedActivity); isLoop {
		startLine := len(*lines) + headerLineCount
		if stmt != "" {
			emitObjectAnnotations(obj, lines, indentStr, annotationsByTarget)
			*lines = append(*lines, indentStr+stmt)
		}

		emitLoopBody(ctx, loop, flowsByOrigin, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)

		*lines = append(*lines, indentStr+loopEndKeyword(loop)+";")
		recordSourceMap(sourceMap, currentID, startLine, len(*lines)+headerLineCount-1)

		// Continue after the loop
		flows := flowsByOrigin[currentID]
		for _, flow := range flows {
			traverseFlow(ctx, flow.DestinationID, activityMap, flowsByOrigin, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
		}
		return
	}

	// Regular activity
	startLine := len(*lines) + headerLineCount
	normalFlows := findNormalFlows(flowsByOrigin[currentID])
	emitActivityStatement(ctx, obj, stmt, flowsByOrigin, activityMap, entityNames, microflowNames, lines, indentStr, annotationsByTarget)
	recordSourceMap(sourceMap, currentID, startLine, len(*lines)+headerLineCount-1)

	// Follow normal (non-error-handler) outgoing flows
	for _, flow := range normalFlows {
		traverseFlow(ctx, flow.DestinationID, activityMap, flowsByOrigin, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
	}
}

// traverseFlowUntilMerge traverses the flow until reaching a merge point.
// When sourceMap is non-nil, it also records line ranges for each activity node.
func traverseFlowUntilMerge(
	ctx *ExecContext,
	currentID model.ID,
	mergeID model.ID,
	activityMap map[model.ID]microflows.MicroflowObject,
	flowsByOrigin map[model.ID][]*microflows.SequenceFlow,
	splitMergeMap map[model.ID]model.ID,
	visited map[model.ID]bool,
	entityNames map[model.ID]string,
	microflowNames map[model.ID]string,
	lines *[]string,
	indent int,
	sourceMap map[string]elkSourceRange,
	headerLineCount int,
	annotationsByTarget map[model.ID][]string,
) {
	if currentID == "" || currentID == mergeID || visited[currentID] {
		return
	}

	obj := activityMap[currentID]
	if obj == nil {
		return
	}

	// Handle intermediate merge points - traverse through them without outputting anything
	if _, isMerge := obj.(*microflows.ExclusiveMerge); isMerge {
		flows := flowsByOrigin[currentID]
		for _, flow := range flows {
			traverseFlowUntilMerge(ctx, flow.DestinationID, mergeID, activityMap, flowsByOrigin, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
		}
		return
	}

	visited[currentID] = true

	stmt := formatActivity(ctx, obj, entityNames, microflowNames)
	indentStr := strings.Repeat("  ", indent)

	// Handle nested ExclusiveSplit
	if _, isSplit := obj.(*microflows.ExclusiveSplit); isSplit {
		startLine := len(*lines) + headerLineCount
		if stmt != "" {
			emitObjectAnnotations(obj, lines, indentStr, annotationsByTarget)
			*lines = append(*lines, indentStr+stmt)
		}

		flows := flowsByOrigin[currentID]
		nestedMergeID := splitMergeMap[currentID]

		trueFlow, falseFlow := findBranchFlows(flows)

		// Guard pattern: true branch is a single EndEvent (RETURN),
		// but only when the false branch does NOT also end directly.
		isGuard := false
		if trueFlow != nil {
			if _, isEnd := activityMap[trueFlow.DestinationID].(*microflows.EndEvent); isEnd {
				isGuard = true
				if falseFlow != nil {
					if _, falseIsEnd := activityMap[falseFlow.DestinationID].(*microflows.EndEvent); falseIsEnd {
						isGuard = false
					}
				}
			}
		}

		if isGuard {
			traverseFlowUntilMerge(ctx, trueFlow.DestinationID, nestedMergeID, activityMap, flowsByOrigin, splitMergeMap, visited, entityNames, microflowNames, lines, indent+1, sourceMap, headerLineCount, annotationsByTarget)
			*lines = append(*lines, indentStr+"end if;")
			recordSourceMap(sourceMap, currentID, startLine, len(*lines)+headerLineCount-1)

			// Continue from the false branch (skip through merge if present)
			if falseFlow != nil {
				contID := falseFlow.DestinationID
				if _, isMerge := activityMap[contID].(*microflows.ExclusiveMerge); isMerge {
					visited[contID] = true
					for _, flow := range flowsByOrigin[contID] {
						contID = flow.DestinationID
						break
					}
				}
				traverseFlowUntilMerge(ctx, contID, mergeID, activityMap, flowsByOrigin, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
			}
		} else {
			if trueFlow != nil {
				traverseFlowUntilMerge(ctx, trueFlow.DestinationID, nestedMergeID, activityMap, flowsByOrigin, splitMergeMap, visited, entityNames, microflowNames, lines, indent+1, sourceMap, headerLineCount, annotationsByTarget)
			}

			if falseFlow != nil {
				*lines = append(*lines, indentStr+"else")
				visitedFalseBranch := make(map[model.ID]bool)
				for id := range visited {
					visitedFalseBranch[id] = true
				}
				traverseFlowUntilMerge(ctx, falseFlow.DestinationID, nestedMergeID, activityMap, flowsByOrigin, splitMergeMap, visitedFalseBranch, entityNames, microflowNames, lines, indent+1, sourceMap, headerLineCount, annotationsByTarget)
			}

			*lines = append(*lines, indentStr+"end if;")
			recordSourceMap(sourceMap, currentID, startLine, len(*lines)+headerLineCount-1)

			// Continue after nested merge
			if nestedMergeID != "" && nestedMergeID != mergeID {
				visited[nestedMergeID] = true
				nextFlows := flowsByOrigin[nestedMergeID]
				for _, flow := range nextFlows {
					traverseFlowUntilMerge(ctx, flow.DestinationID, mergeID, activityMap, flowsByOrigin, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
				}
			}
		}
		return
	}

	// Handle LoopedActivity inside a branch
	if loop, isLoop := obj.(*microflows.LoopedActivity); isLoop {
		startLine := len(*lines) + headerLineCount
		if stmt != "" {
			emitObjectAnnotations(obj, lines, indentStr, annotationsByTarget)
			*lines = append(*lines, indentStr+stmt)
		}

		emitLoopBody(ctx, loop, flowsByOrigin, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)

		*lines = append(*lines, indentStr+loopEndKeyword(loop)+";")
		recordSourceMap(sourceMap, currentID, startLine, len(*lines)+headerLineCount-1)

		// Continue after the loop within the branch
		flows := flowsByOrigin[currentID]
		for _, flow := range flows {
			traverseFlowUntilMerge(ctx, flow.DestinationID, mergeID, activityMap, flowsByOrigin, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
		}
		return
	}

	// Regular activity
	startLine := len(*lines) + headerLineCount
	normalFlows := findNormalFlows(flowsByOrigin[currentID])
	emitActivityStatement(ctx, obj, stmt, flowsByOrigin, activityMap, entityNames, microflowNames, lines, indentStr, annotationsByTarget)
	recordSourceMap(sourceMap, currentID, startLine, len(*lines)+headerLineCount-1)

	// Follow normal (non-error-handler) outgoing flows until merge
	for _, flow := range normalFlows {
		traverseFlowUntilMerge(ctx, flow.DestinationID, mergeID, activityMap, flowsByOrigin, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
	}
}

// traverseLoopBody traverses activities inside a loop body.
// When sourceMap is non-nil, it also records line ranges for each activity node.
func traverseLoopBody(
	ctx *ExecContext,
	currentID model.ID,
	activityMap map[model.ID]microflows.MicroflowObject,
	flowsByOrigin map[model.ID][]*microflows.SequenceFlow,
	visited map[model.ID]bool,
	entityNames map[model.ID]string,
	microflowNames map[model.ID]string,
	lines *[]string,
	indent int,
	sourceMap map[string]elkSourceRange,
	headerLineCount int,
	annotationsByTarget map[model.ID][]string,
) {
	if currentID == "" || visited[currentID] {
		return
	}

	obj := activityMap[currentID]
	if obj == nil {
		return
	}

	visited[currentID] = true

	stmt := formatActivity(ctx, obj, entityNames, microflowNames)
	indentStr := strings.Repeat("  ", indent)

	// Handle nested LoopedActivity specially
	if loop, isLoop := obj.(*microflows.LoopedActivity); isLoop {
		startLine := len(*lines) + headerLineCount
		if stmt != "" {
			emitObjectAnnotations(obj, lines, indentStr, annotationsByTarget)
			*lines = append(*lines, indentStr+stmt)
		}

		emitLoopBody(ctx, loop, flowsByOrigin, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)

		*lines = append(*lines, indentStr+loopEndKeyword(loop)+";")
		recordSourceMap(sourceMap, currentID, startLine, len(*lines)+headerLineCount-1)

		// Continue after the nested loop within the parent loop body
		flows := flowsByOrigin[currentID]
		for _, flow := range flows {
			if _, inLoop := activityMap[flow.DestinationID]; inLoop {
				traverseLoopBody(ctx, flow.DestinationID, activityMap, flowsByOrigin, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
			}
		}
		return
	}

	// Regular activity
	startLine := len(*lines) + headerLineCount
	normalFlows := findNormalFlows(flowsByOrigin[currentID])
	emitActivityStatement(ctx, obj, stmt, flowsByOrigin, activityMap, entityNames, microflowNames, lines, indentStr, annotationsByTarget)
	recordSourceMap(sourceMap, currentID, startLine, len(*lines)+headerLineCount-1)

	// Follow normal (non-error-handler) outgoing flows within the loop body
	for _, flow := range normalFlows {
		if _, inLoop := activityMap[flow.DestinationID]; inLoop {
			traverseLoopBody(ctx, flow.DestinationID, activityMap, flowsByOrigin, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
		}
	}
}

// emitLoopBody processes the inner objects of a LoopedActivity.
// Shared by traverseFlow and traverseLoopBody for both top-level and nested loops.
func emitLoopBody(
	ctx *ExecContext,
	loop *microflows.LoopedActivity,
	flowsByOrigin map[model.ID][]*microflows.SequenceFlow,
	entityNames map[model.ID]string,
	microflowNames map[model.ID]string,
	lines *[]string,
	indent int,
	sourceMap map[string]elkSourceRange,
	headerLineCount int,
	annotationsByTarget map[model.ID][]string,
) {
	if loop.ObjectCollection == nil || len(loop.ObjectCollection.Objects) == 0 {
		return
	}

	// Build a map of objects in the loop body
	loopActivityMap := make(map[model.ID]microflows.MicroflowObject)
	for _, loopObj := range loop.ObjectCollection.Objects {
		loopActivityMap[loopObj.GetID()] = loopObj
	}

	// Build flow graph from the loop's own ObjectCollection flows
	loopFlowsByOrigin := make(map[model.ID][]*microflows.SequenceFlow)
	if loop.ObjectCollection != nil {
		for _, flow := range loop.ObjectCollection.Flows {
			loopFlowsByOrigin[flow.OriginID] = append(loopFlowsByOrigin[flow.OriginID], flow)
		}
	}
	// Also include parent flows that originate from loop body objects (for backward compatibility)
	for originID, flows := range flowsByOrigin {
		if _, inLoop := loopActivityMap[originID]; inLoop {
			if _, exists := loopFlowsByOrigin[originID]; !exists {
				loopFlowsByOrigin[originID] = flows
			}
		}
	}

	// Find the first activity in the loop body (the one with no incoming flow from within the loop)
	incomingCount := make(map[model.ID]int)
	for _, loopObj := range loop.ObjectCollection.Objects {
		incomingCount[loopObj.GetID()] = 0
	}
	for _, flows := range loopFlowsByOrigin {
		for _, flow := range flows {
			if _, inLoop := loopActivityMap[flow.DestinationID]; inLoop {
				incomingCount[flow.DestinationID]++
			}
		}
	}
	var firstID model.ID
	for id, count := range incomingCount {
		if count == 0 {
			firstID = id
			break
		}
	}

	// Traverse the loop body
	if firstID != "" {
		loopVisited := make(map[model.ID]bool)
		traverseLoopBody(ctx, firstID, loopActivityMap, loopFlowsByOrigin, loopVisited, entityNames, microflowNames, lines, indent+1, sourceMap, headerLineCount, annotationsByTarget)
	}
}

// findBranchFlows separates flows from a split into TRUE and FALSE branches based on CaseValue.
// Returns (trueFlow, falseFlow). Either may be nil if the branch doesn't exist.
func findBranchFlows(flows []*microflows.SequenceFlow) (trueFlow, falseFlow *microflows.SequenceFlow) {
	for _, flow := range flows {
		if flow.CaseValue == nil {
			continue
		}
		switch cv := flow.CaseValue.(type) {
		case *microflows.ExpressionCase:
			if cv.Expression == "true" {
				trueFlow = flow
			} else if cv.Expression == "false" {
				falseFlow = flow
			}
		case *microflows.EnumerationCase:
			if cv.Value == "true" {
				trueFlow = flow
			} else if cv.Value == "false" {
				falseFlow = flow
			}
		case microflows.EnumerationCase:
			if cv.Value == "true" {
				trueFlow = flow
			} else if cv.Value == "false" {
				falseFlow = flow
			}
		case *microflows.BooleanCase:
			if cv.Value {
				trueFlow = flow
			} else {
				falseFlow = flow
			}
		case microflows.BooleanCase:
			if cv.Value {
				trueFlow = flow
			} else {
				falseFlow = flow
			}
		}
	}
	return trueFlow, falseFlow
}

// findErrorHandlerFlow returns the error handler flow from an activity's outgoing flows.
func findErrorHandlerFlow(flows []*microflows.SequenceFlow) *microflows.SequenceFlow {
	for _, flow := range flows {
		if flow.IsErrorHandler {
			return flow
		}
	}
	return nil
}

// findNormalFlows returns all non-error-handler flows from an activity.
func findNormalFlows(flows []*microflows.SequenceFlow) []*microflows.SequenceFlow {
	var result []*microflows.SequenceFlow
	for _, flow := range flows {
		if !flow.IsErrorHandler {
			result = append(result, flow)
		}
	}
	return result
}

// formatErrorHandlingSuffix returns the ON ERROR suffix for an activity based on its ErrorHandlingType.
// Returns empty string if no special error handling.
func formatErrorHandlingSuffix(errType microflows.ErrorHandlingType) string {
	switch errType {
	case microflows.ErrorHandlingTypeContinue:
		return " on error continue"
	case microflows.ErrorHandlingTypeRollback:
		return " on error rollback"
	case microflows.ErrorHandlingTypeCustom:
		return " on error" // Will be followed by block
	case microflows.ErrorHandlingTypeCustomWithoutRollback:
		return " on error without rollback" // Will be followed by block
	default:
		return "" // Abort is the default, no suffix needed
	}
}

// hasCustomErrorHandler returns true if the error handling type requires a custom handler block.
func hasCustomErrorHandler(errType microflows.ErrorHandlingType) bool {
	return errType == microflows.ErrorHandlingTypeCustom || errType == microflows.ErrorHandlingTypeCustomWithoutRollback
}

// getActionErrorHandlingType extracts the ErrorHandlingType from the action inside an ActionActivity.
// Most action types store ErrorHandlingType at the action level, not the activity level.
func getActionErrorHandlingType(activity *microflows.ActionActivity) microflows.ErrorHandlingType {
	if activity == nil || activity.Action == nil {
		return ""
	}

	switch action := activity.Action.(type) {
	case *microflows.MicroflowCallAction:
		return action.ErrorHandlingType
	case *microflows.JavaActionCallAction:
		return action.ErrorHandlingType
	case *microflows.CallExternalAction:
		return action.ErrorHandlingType
	case *microflows.RestCallAction:
		return action.ErrorHandlingType
	case *microflows.RestOperationCallAction:
		return "" // RestOperationCallAction does not support custom error handling (CE6035)
	case *microflows.ExecuteDatabaseQueryAction:
		return action.ErrorHandlingType
	case *microflows.ImportXmlAction:
		return action.ErrorHandlingType
	case *microflows.ExportXmlAction:
		return action.ErrorHandlingType
	case *microflows.CommitObjectsAction:
		return action.ErrorHandlingType
	default:
		// Fall back to activity level for action types without ErrorHandlingType field
		return activity.ErrorHandlingType
	}
}

// collectErrorHandlerStatements traverses the error handler flow and collects statements.
// Returns a slice of MDL statements for the error handler block.
func collectErrorHandlerStatements(
	ctx *ExecContext,
	startID model.ID,
	activityMap map[model.ID]microflows.MicroflowObject,
	flowsByOrigin map[model.ID][]*microflows.SequenceFlow,
	entityNames map[model.ID]string,
	microflowNames map[model.ID]string,
) []string {
	var statements []string
	visited := make(map[model.ID]bool)

	var traverse func(id model.ID)
	traverse = func(id model.ID) {
		if id == "" || visited[id] {
			return
		}

		obj := activityMap[id]
		if obj == nil {
			return
		}

		// Stop at merge points (rejoin with main flow) or end events
		if _, isMerge := obj.(*microflows.ExclusiveMerge); isMerge {
			return
		}

		visited[id] = true

		stmt := formatActivity(ctx, obj, entityNames, microflowNames)
		if stmt != "" {
			statements = append(statements, stmt)
		}

		// Follow normal (non-error) flows
		flows := flowsByOrigin[id]
		normalFlows := findNormalFlows(flows)
		for _, flow := range normalFlows {
			traverse(flow.DestinationID)
		}
	}

	traverse(startID)
	return statements
}

// loopEndKeyword returns "END WHILE" for WHILE loops and "END LOOP" for FOR-EACH loops.
func loopEndKeyword(loop *microflows.LoopedActivity) string {
	if _, isWhile := loop.LoopSource.(*microflows.WhileLoopCondition); isWhile {
		return "end while"
	}
	return "end loop"
}

// --- Executor method wrappers for callers in unmigrated code and tests ---

func (e *Executor) traverseFlow(
	currentID model.ID,
	activityMap map[model.ID]microflows.MicroflowObject,
	flowsByOrigin map[model.ID][]*microflows.SequenceFlow,
	splitMergeMap map[model.ID]model.ID,
	visited map[model.ID]bool,
	entityNames map[model.ID]string,
	microflowNames map[model.ID]string,
	lines *[]string,
	indent int,
	sourceMap map[string]elkSourceRange,
	headerLineCount int,
	annotationsByTarget map[model.ID][]string,
) {
	traverseFlow(e.newExecContext(context.Background()), currentID, activityMap, flowsByOrigin, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
}

func (e *Executor) collectErrorHandlerStatements(
	startID model.ID,
	activityMap map[model.ID]microflows.MicroflowObject,
	flowsByOrigin map[model.ID][]*microflows.SequenceFlow,
	entityNames map[model.ID]string,
	microflowNames map[model.ID]string,
) []string {
	return collectErrorHandlerStatements(e.newExecContext(context.Background()), startID, activityMap, flowsByOrigin, entityNames, microflowNames)
}
