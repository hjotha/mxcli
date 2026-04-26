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

// anchorSideKeyword returns the MDL keyword (top/right/bottom/left) for a
// connection-index value. Returns "" for unknown values.
func anchorSideKeyword(idx int) string {
	switch idx {
	case AnchorTop:
		return "top"
	case AnchorRight:
		return "right"
	case AnchorBottom:
		return "bottom"
	case AnchorLeft:
		return "left"
	default:
		return ""
	}
}

// emitAnchorAnnotation emits an @anchor(...) line describing the incoming and
// outgoing SequenceFlow anchors for this object. Nothing is emitted when the
// object has no attached flows.
//
// For ExclusiveSplit / InheritanceSplit the split has up to two outgoing flows
// (true/false), so instead of the simple from/to form we emit the branch form:
//
//	@anchor(to: X, true: (from: Y, to: Z), false: (from: Y, to: Z))
//
// For LoopedActivity we emit the loop form when any iterator/tail flow exists:
//
//	@anchor(from: X, to: Y, iterator: (from: Y, to: Z), tail: (from: Y, to: Z))
//
// Any of the groups is omitted when its constituent values are both default
// (so a non-annotated split/loop produces no line). Non-split/non-loop objects
// use the simple @anchor(from: X, to: Y) form.
func emitAnchorAnnotation(
	obj microflows.MicroflowObject,
	flowsByOrigin map[model.ID][]*microflows.SequenceFlow,
	flowsByDest map[model.ID][]*microflows.SequenceFlow,
	lines *[]string,
	indentStr string,
) {
	id := obj.GetID()

	if _, isSplit := obj.(*microflows.ExclusiveSplit); isSplit {
		emitSplitAnchorAnnotation(id, flowsByOrigin, flowsByDest, lines, indentStr)
		return
	}
	if _, isSplit := obj.(*microflows.InheritanceSplit); isSplit {
		emitSplitAnchorAnnotation(id, flowsByOrigin, flowsByDest, lines, indentStr)
		return
	}
	if loop, isLoop := obj.(*microflows.LoopedActivity); isLoop {
		emitLoopAnchorAnnotation(loop, flowsByOrigin, flowsByDest, lines, indentStr)
		return
	}

	var from, to string
	if outgoing := flowsByOrigin[id]; len(outgoing) > 0 {
		from = anchorSideKeyword(outgoing[0].OriginConnectionIndex)
	}
	if incoming := flowsByDest[id]; len(incoming) > 0 {
		to = anchorSideKeyword(incoming[0].DestinationConnectionIndex)
	}

	if from == "" && to == "" {
		return
	}
	var parts []string
	if from != "" {
		parts = append(parts, "from: "+from)
	}
	if to != "" {
		parts = append(parts, "to: "+to)
	}
	*lines = append(*lines, indentStr+fmt.Sprintf("@anchor(%s)", strings.Join(parts, ", ")))
}

// emitSplitAnchorAnnotation emits the split form of @anchor — the incoming
// `to: X` plus per-branch `true: (...)` / `false: (...)` — whenever any of the
// three has a non-default value. The rendering matches the grammar accepted
// by parseAnchorAnnotation so describe → exec roundtrips bit-exactly.
//
// The TRUE/FALSE branches are identified by findBranchFlows, which already
// handles every CaseValue variant the parser produces (ExpressionCase,
// EnumerationCase, BooleanCase — both value and pointer forms). Sharing that
// helper keeps the anchor emission consistent with the rest of the describer.
func emitSplitAnchorAnnotation(
	id model.ID,
	flowsByOrigin map[model.ID][]*microflows.SequenceFlow,
	flowsByDest map[model.ID][]*microflows.SequenceFlow,
	lines *[]string,
	indentStr string,
) {
	// Incoming flow anchor (where the previous activity's flow lands on the split).
	var inTo string
	if incoming := flowsByDest[id]; len(incoming) > 0 {
		inTo = anchorSideKeyword(incoming[0].DestinationConnectionIndex)
	}

	trueFlow, falseFlow := findBranchFlows(flowsByOrigin[id])

	var trueFrom, trueTo, falseFrom, falseTo string
	if trueFlow != nil {
		trueFrom = anchorSideKeyword(trueFlow.OriginConnectionIndex)
		trueTo = anchorSideKeyword(trueFlow.DestinationConnectionIndex)
	}
	if falseFlow != nil {
		falseFrom = anchorSideKeyword(falseFlow.OriginConnectionIndex)
		falseTo = anchorSideKeyword(falseFlow.DestinationConnectionIndex)
	}

	if inTo == "" && trueFrom == "" && trueTo == "" && falseFrom == "" && falseTo == "" {
		return
	}

	var parts []string
	if inTo != "" {
		parts = append(parts, "to: "+inTo)
	}
	if p := branchAnchorFragment("true", trueFrom, trueTo); p != "" {
		parts = append(parts, p)
	}
	if p := branchAnchorFragment("false", falseFrom, falseTo); p != "" {
		parts = append(parts, p)
	}
	if len(parts) == 0 {
		return
	}
	*lines = append(*lines, indentStr+fmt.Sprintf("@anchor(%s)", strings.Join(parts, ", ")))
}

// branchAnchorFragment builds a `key: (from: X, to: Y)` fragment for a branch
// anchor. Returns "" when both sides are empty (default).
func branchAnchorFragment(label, from, to string) string {
	if from == "" && to == "" {
		return ""
	}
	var inner []string
	if from != "" {
		inner = append(inner, "from: "+from)
	}
	if to != "" {
		inner = append(inner, "to: "+to)
	}
	return fmt.Sprintf("%s: (%s)", label, strings.Join(inner, ", "))
}

// emitLoopAnchorAnnotation emits the loop form of @anchor for a LoopedActivity.
// A LoopedActivity has up to four flows worth describing:
//   - the incoming flow from the previous activity (normal `to:`)
//   - the outgoing flow to the next activity (normal `from:`)
//   - the iterator flow (loop boundary → first body statement), emitted as
//     `iterator: (from: X, to: Y)` when present
//   - the tail flow (last body statement → loop boundary), emitted as
//     `tail: (from: X, to: Y)` when present
//
// The iterator/tail edges are optional — Mendix normally renders them
// implicitly from the LoopedActivity geometry — so they appear only when the
// flows actually exist in the parent collection. This keeps plain loops
// free of `@anchor(iterator: ..., tail: ...)` noise on describe while still
// round-tripping projects that explicitly pinned them via @anchor annotations.
func emitLoopAnchorAnnotation(
	loop *microflows.LoopedActivity,
	flowsByOrigin map[model.ID][]*microflows.SequenceFlow,
	flowsByDest map[model.ID][]*microflows.SequenceFlow,
	lines *[]string,
	indentStr string,
) {
	id := loop.ID

	innerIDs := make(map[model.ID]bool)
	if loop.ObjectCollection != nil {
		for _, o := range loop.ObjectCollection.Objects {
			innerIDs[o.GetID()] = true
		}
	}

	// Outer incoming / outgoing (flows between this loop and its siblings).
	// Only the first non-iterator edge counts as the external connection.
	var outerFrom, outerTo string
	for _, f := range flowsByOrigin[id] {
		if innerIDs[f.DestinationID] {
			continue // iterator edge, handled below
		}
		outerFrom = anchorSideKeyword(f.OriginConnectionIndex)
		break
	}
	for _, f := range flowsByDest[id] {
		if innerIDs[f.OriginID] {
			continue // tail edge, handled below
		}
		outerTo = anchorSideKeyword(f.DestinationConnectionIndex)
		break
	}

	// Iterator edge: loop → inner. There can only be one in a well-formed
	// project; if the builder ever emits more we keep the first for now.
	var iterFrom, iterTo string
	for _, f := range flowsByOrigin[id] {
		if !innerIDs[f.DestinationID] {
			continue
		}
		iterFrom = anchorSideKeyword(f.OriginConnectionIndex)
		iterTo = anchorSideKeyword(f.DestinationConnectionIndex)
		break
	}

	// Tail edge: inner → loop.
	var tailFrom, tailTo string
	for _, f := range flowsByDest[id] {
		if !innerIDs[f.OriginID] {
			continue
		}
		tailFrom = anchorSideKeyword(f.OriginConnectionIndex)
		tailTo = anchorSideKeyword(f.DestinationConnectionIndex)
		break
	}

	if outerFrom == "" && outerTo == "" && iterFrom == "" && iterTo == "" && tailFrom == "" && tailTo == "" {
		return
	}

	var parts []string
	if outerFrom != "" {
		parts = append(parts, "from: "+outerFrom)
	}
	if outerTo != "" {
		parts = append(parts, "to: "+outerTo)
	}
	if p := branchAnchorFragment("iterator", iterFrom, iterTo); p != "" {
		parts = append(parts, p)
	}
	if p := branchAnchorFragment("tail", tailFrom, tailTo); p != "" {
		parts = append(parts, p)
	}
	if len(parts) == 0 {
		return
	}
	*lines = append(*lines, indentStr+fmt.Sprintf("@anchor(%s)", strings.Join(parts, ", ")))
}

// emitObjectAnnotations emits @position, @caption, @color, @annotation, and
// @anchor lines for a microflow object before its statement.
//
// flowsByOrigin / flowsByDest are optional: when both are non-nil, @anchor
// lines are emitted using them. Passing nil suppresses @anchor emission,
// which keeps small unit tests of this helper self-contained.
func emitObjectAnnotations(
	obj microflows.MicroflowObject,
	lines *[]string,
	indentStr string,
	annotationsByTarget map[model.ID][]string,
	flowsByOrigin map[model.ID][]*microflows.SequenceFlow,
	flowsByDest map[model.ID][]*microflows.SequenceFlow,
) {
	currentID := obj.GetID()

	pos := obj.GetPosition()
	*lines = append(*lines, indentStr+fmt.Sprintf("@position(%d, %d)", pos.X, pos.Y))

	if flowsByOrigin != nil && flowsByDest != nil {
		// @anchor — emit whenever attached flows exist, for roundtrip fidelity.
		// The emitter sorts out the right form (simple / split / loop) based on
		// the object type.
		emitAnchorAnnotation(obj, flowsByOrigin, flowsByDest, lines, indentStr)
	}
	if activity, ok := obj.(*microflows.ActionActivity); ok {
		if activity.Disabled {
			*lines = append(*lines, indentStr+"@excluded")
		}
		if !activity.AutoGenerateCaption && activity.Caption != "" {
			*lines = append(*lines, indentStr+fmt.Sprintf("@caption %s", mdlQuote(activity.Caption)))
		}
		if activity.BackgroundColor != "" && activity.BackgroundColor != "Default" {
			*lines = append(*lines, indentStr+fmt.Sprintf("@color %s", activity.BackgroundColor))
		}
	}

	if split, ok := obj.(*microflows.ExclusiveSplit); ok && split.Caption != "" {
		*lines = append(*lines, indentStr+fmt.Sprintf("@caption %s", mdlQuote(split.Caption)))
	}
	if split, ok := obj.(*microflows.InheritanceSplit); ok && split.Caption != "" {
		*lines = append(*lines, indentStr+fmt.Sprintf("@caption %s", mdlQuote(split.Caption)))
	}

	// @annotation (attached Annotation objects)
	if annotationsByTarget != nil {
		for _, caption := range annotationsByTarget[currentID] {
			*lines = append(*lines, indentStr+fmt.Sprintf("@annotation %s", mdlQuote(caption)))
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
	flowsByDest map[model.ID][]*microflows.SequenceFlow,
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

	// When the activity is unsupported by the describer (e.g. CallWebServiceAction,
	// CastAction, InheritanceSplit placeholder) we fall back to emitting just an
	// MDL line comment. Decorating that comment with @position/@anchor/@annotation
	// leaves the annotations orphaned — the grammar only accepts `annotation*`
	// as a prefix of a real microflowStatement, so line comments preceded by
	// annotations cause "no viable alternative at input '@position...end'" during
	// exec. Emit the comment on its own instead.
	if strings.HasPrefix(strings.TrimSpace(stmt), "--") {
		*lines = append(*lines, indentStr+stmt)
		return
	}

	// Emit @ annotations before the statement
	emitObjectAnnotations(obj, lines, indentStr, annotationsByTarget, flowsByOrigin, flowsByDest)
	appendFormattedStatement(ctx, obj, stmt, activityMap, flowsByOrigin, entityNames, microflowNames, lines, indentStr)
}

// appendFormattedStatement appends the formatted activity statement itself,
// including any ON ERROR suffix or custom error-handler block, but without
// emitting leading annotations. This is shared by the main describer and the
// error-handler collector, where annotations are intentionally suppressed.
func appendFormattedStatement(
	ctx *ExecContext,
	obj microflows.MicroflowObject,
	stmt string,
	activityMap map[model.ID]microflows.MicroflowObject,
	flowsByOrigin map[model.ID][]*microflows.SequenceFlow,
	entityNames map[model.ID]string,
	microflowNames map[model.ID]string,
	lines *[]string,
	indentStr string,
) {
	if stmt == "" {
		return
	}

	// Unsupported actions are rendered as MDL comments. They must stand alone;
	// callers that want annotations emit them separately before reaching here.
	if strings.HasPrefix(strings.TrimSpace(stmt), "--") {
		*lines = append(*lines, indentStr+stmt)
		return
	}

	currentID := obj.GetID()
	flows := flowsByOrigin[currentID]
	errorHandlerFlow := findErrorHandlerFlow(flows)

	activity, isAction := obj.(*microflows.ActionActivity)
	if !isAction {
		*lines = append(*lines, indentStr+stmt)
	} else {
		errType := getActionErrorHandlingType(activity)
		suffix := formatErrorHandlingSuffix(errType)

		if hasCustomErrorHandler(errType) && errorHandlerFlow == nil {
			stmtWithoutSemi := strings.TrimSuffix(strings.TrimSpace(stmt), ";")
			*lines = append(*lines, indentStr+stmtWithoutSemi+suffix+" { };")
		} else if errorHandlerFlow != nil && hasCustomErrorHandler(errType) {
			stmtWithoutSemi := strings.TrimSuffix(strings.TrimSpace(stmt), ";")

			errorSuffix := suffix
			if errorSuffix == "" {
				errorSuffix = " on error without rollback"
			}

			if hasNormalIncomingToDestination(flowsByOrigin, errorHandlerFlow.DestinationID) {
				*lines = append(*lines, indentStr+stmtWithoutSemi+errorSuffix+" { };")
				return
			}

			errStmts := collectErrorHandlerStatements(
				ctx,
				errorHandlerFlow.DestinationID,
				activityMap, flowsByOrigin, entityNames, microflowNames,
			)

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
	flowsByDest map[model.ID][]*microflows.SequenceFlow,
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
	traverseFlowLoopAware(ctx, currentID, "", activityMap, flowsByOrigin, flowsByDest, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
}

func traverseFlowLoopAware(
	ctx *ExecContext,
	currentID model.ID,
	loopHeaderID model.ID,
	activityMap map[model.ID]microflows.MicroflowObject,
	flowsByOrigin map[model.ID][]*microflows.SequenceFlow,
	flowsByDest map[model.ID][]*microflows.SequenceFlow,
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
	if currentID == "" {
		return
	}
	if loopHeaderID != "" && currentID == loopHeaderID {
		*lines = append(*lines, strings.Repeat("  ", indent)+"continue;")
		return
	}
	if visited[currentID] {
		return
	}

	obj := activityMap[currentID]
	if obj == nil {
		return
	}

	// Merge points are normally processed by the matching split's traversal —
	// returning here avoids emitting an `end if;` twice. But a merge can also
	// appear as a pure junction (multiple incoming flows converging outside
	// of an IF), e.g. a manual retry loop where a "Call REST → decide → loop
	// back" pattern routes the back-edge into a merge. Those merges are not
	// values in splitMergeMap, and stopping there drops every activity after
	// them (see issue #281). When that happens, walk through as a pass-
	// through — same as traverseFlowUntilMerge already does for intermediate
	// merges.
	if _, isMerge := obj.(*microflows.ExclusiveMerge); isMerge {
		if isMergePairedWithSplit(currentID, splitMergeMap) {
			return
		}
		if isManualLoopHeaderMerge(currentID, activityMap, flowsByOrigin, splitMergeMap) {
			startLine := len(*lines) + headerLineCount
			indentStr := strings.Repeat("  ", indent)
			visited[currentID] = true
			*lines = append(*lines, indentStr+"while true")
			*lines = append(*lines, indentStr+"begin")
			for _, flow := range findNormalFlows(flowsByOrigin[currentID]) {
				traverseFlowLoopAware(ctx, flow.DestinationID, currentID, activityMap, flowsByOrigin, flowsByDest, splitMergeMap, visited, entityNames, microflowNames, lines, indent+1, sourceMap, headerLineCount, annotationsByTarget)
			}
			*lines = append(*lines, indentStr+"end while;")
			recordSourceMap(sourceMap, currentID, startLine, len(*lines)+headerLineCount-1)
			return
		}
		visited[currentID] = true
		for _, flow := range findNormalFlows(flowsByOrigin[currentID]) {
			traverseFlowLoopAware(ctx, flow.DestinationID, loopHeaderID, activityMap, flowsByOrigin, flowsByDest, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
		}
		return
	}

	visited[currentID] = true

	stmt := formatActivity(ctx, obj, entityNames, microflowNames)
	indentStr := strings.Repeat("  ", indent)

	// Handle ExclusiveSplit specially - need to process both branches
	if _, isSplit := obj.(*microflows.ExclusiveSplit); isSplit {
		startLine := len(*lines) + headerLineCount
		if stmt != "" {
			emitObjectAnnotations(obj, lines, indentStr, annotationsByTarget, flowsByOrigin, flowsByDest)
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
			traverseFlowUntilMergeLoopAware(ctx, trueFlow.DestinationID, mergeID, loopHeaderID, activityMap, flowsByOrigin, flowsByDest, splitMergeMap, visited, entityNames, microflowNames, lines, indent+1, sourceMap, headerLineCount, annotationsByTarget)
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
				traverseFlowLoopAware(ctx, contID, loopHeaderID, activityMap, flowsByOrigin, flowsByDest, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
			}
		} else {
			if trueFlow != nil {
				traverseFlowUntilMergeLoopAware(ctx, trueFlow.DestinationID, mergeID, loopHeaderID, activityMap, flowsByOrigin, flowsByDest, splitMergeMap, visited, entityNames, microflowNames, lines, indent+1, sourceMap, headerLineCount, annotationsByTarget)
			}

			// Emit the ELSE branch only if it has statements. When the false
			// flow jumps straight to the merge (the MDL was `if X then ... end if`
			// with no else), emitting `else` with no body produces an empty
			// branch that normalizes away on re-parse.
			falseHasBody := falseFlow != nil && falseFlow.DestinationID != mergeID
			if falseHasBody {
				*lines = append(*lines, indentStr+"else")
				visitedFalseBranch := make(map[model.ID]bool)
				for id := range visited {
					visitedFalseBranch[id] = true
				}
				traverseFlowUntilMergeLoopAware(ctx, falseFlow.DestinationID, mergeID, loopHeaderID, activityMap, flowsByOrigin, flowsByDest, splitMergeMap, visitedFalseBranch, entityNames, microflowNames, lines, indent+1, sourceMap, headerLineCount, annotationsByTarget)
			}

			*lines = append(*lines, indentStr+"end if;")
			recordSourceMap(sourceMap, currentID, startLine, len(*lines)+headerLineCount-1)

			continueAfterSplitJoinLoopAware(ctx, mergeID, loopHeaderID, activityMap, flowsByOrigin, flowsByDest, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
		}
		return
	}

	if _, isSplit := obj.(*microflows.InheritanceSplit); isSplit && len(findNormalFlows(flowsByOrigin[currentID])) > 1 {
		startLine := len(*lines) + headerLineCount
		emitInheritanceSplitStatement(ctx, currentID, "", activityMap, flowsByOrigin, flowsByDest, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
		recordSourceMap(sourceMap, currentID, startLine, len(*lines)+headerLineCount-1)
		return
	}

	// Handle LoopedActivity specially - need to process loop body
	if loop, isLoop := obj.(*microflows.LoopedActivity); isLoop {
		startLine := len(*lines) + headerLineCount
		if stmt != "" {
			emitObjectAnnotations(obj, lines, indentStr, annotationsByTarget, flowsByOrigin, flowsByDest)
			*lines = append(*lines, indentStr+stmt)
		}

		*lines = append(*lines, indentStr+"begin")
		emitLoopBody(ctx, loop, flowsByOrigin, flowsByDest, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)

		*lines = append(*lines, indentStr+loopEndKeyword(loop)+";")
		recordSourceMap(sourceMap, currentID, startLine, len(*lines)+headerLineCount-1)

		// Continue after the loop
		flows := flowsByOrigin[currentID]
		for _, flow := range flows {
			traverseFlowLoopAware(ctx, flow.DestinationID, loopHeaderID, activityMap, flowsByOrigin, flowsByDest, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
		}
		return
	}

	// Regular activity
	startLine := len(*lines) + headerLineCount
	normalFlows := findNormalFlows(flowsByOrigin[currentID])
	emitActivityStatement(ctx, obj, stmt, flowsByOrigin, flowsByDest, activityMap, entityNames, microflowNames, lines, indentStr, annotationsByTarget)
	recordSourceMap(sourceMap, currentID, startLine, len(*lines)+headerLineCount-1)

	// Follow normal (non-error-handler) outgoing flows
	for _, flow := range normalFlows {
		traverseFlowLoopAware(ctx, flow.DestinationID, loopHeaderID, activityMap, flowsByOrigin, flowsByDest, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
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
	flowsByDest map[model.ID][]*microflows.SequenceFlow,
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
	traverseFlowUntilMergeLoopAware(ctx, currentID, mergeID, "", activityMap, flowsByOrigin, flowsByDest, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
}

func traverseFlowUntilMergeLoopAware(
	ctx *ExecContext,
	currentID model.ID,
	mergeID model.ID,
	loopHeaderID model.ID,
	activityMap map[model.ID]microflows.MicroflowObject,
	flowsByOrigin map[model.ID][]*microflows.SequenceFlow,
	flowsByDest map[model.ID][]*microflows.SequenceFlow,
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
	if currentID == "" || currentID == mergeID {
		return
	}
	if loopHeaderID != "" && currentID == loopHeaderID {
		*lines = append(*lines, strings.Repeat("  ", indent)+"continue;")
		return
	}
	if visited[currentID] {
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
			traverseFlowUntilMergeLoopAware(ctx, flow.DestinationID, mergeID, loopHeaderID, activityMap, flowsByOrigin, flowsByDest, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
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
			emitObjectAnnotations(obj, lines, indentStr, annotationsByTarget, flowsByOrigin, flowsByDest)
			*lines = append(*lines, indentStr+stmt)
		}

		flows := flowsByOrigin[currentID]
		nestedMergeID := splitMergeMap[currentID]

		trueFlow, falseFlow := findBranchFlows(flows)
		nestedMergeID = resolveNestedMergeID(nestedMergeID, mergeID, trueFlow, falseFlow, flowsByOrigin)

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
			guardMergeID := nestedMergeID
			if guardMergeID == "" {
				guardMergeID = mergeID
			}

			traverseFlowUntilMergeLoopAware(ctx, trueFlow.DestinationID, guardMergeID, loopHeaderID, activityMap, flowsByOrigin, flowsByDest, splitMergeMap, visited, entityNames, microflowNames, lines, indent+1, sourceMap, headerLineCount, annotationsByTarget)
			*lines = append(*lines, indentStr+"end if;")
			recordSourceMap(sourceMap, currentID, startLine, len(*lines)+headerLineCount-1)

			// Continue from the false branch (skip through merge if present)
			if falseFlow != nil {
				contID := falseFlow.DestinationID
				if contID != guardMergeID {
					if _, isMerge := activityMap[contID].(*microflows.ExclusiveMerge); isMerge {
						visited[contID] = true
						for _, flow := range flowsByOrigin[contID] {
							contID = flow.DestinationID
							break
						}
					}
					traverseFlowUntilMergeLoopAware(ctx, contID, guardMergeID, loopHeaderID, activityMap, flowsByOrigin, flowsByDest, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
				}
			}
			if guardMergeID != "" && guardMergeID != mergeID {
				continueAfterNestedSplitJoinLoopAware(ctx, guardMergeID, mergeID, loopHeaderID, activityMap, flowsByOrigin, flowsByDest, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
			}
		} else {
			if trueFlow != nil {
				traverseFlowUntilMergeLoopAware(ctx, trueFlow.DestinationID, nestedMergeID, loopHeaderID, activityMap, flowsByOrigin, flowsByDest, splitMergeMap, visited, entityNames, microflowNames, lines, indent+1, sourceMap, headerLineCount, annotationsByTarget)
			}

			falseHasBody := falseFlow != nil && falseFlow.DestinationID != nestedMergeID
			if falseHasBody {
				*lines = append(*lines, indentStr+"else")
				visitedFalseBranch := make(map[model.ID]bool)
				for id := range visited {
					visitedFalseBranch[id] = true
				}
				traverseFlowUntilMergeLoopAware(ctx, falseFlow.DestinationID, nestedMergeID, loopHeaderID, activityMap, flowsByOrigin, flowsByDest, splitMergeMap, visitedFalseBranch, entityNames, microflowNames, lines, indent+1, sourceMap, headerLineCount, annotationsByTarget)
			}

			*lines = append(*lines, indentStr+"end if;")
			recordSourceMap(sourceMap, currentID, startLine, len(*lines)+headerLineCount-1)

			continueAfterNestedSplitJoinLoopAware(ctx, nestedMergeID, mergeID, loopHeaderID, activityMap, flowsByOrigin, flowsByDest, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
		}
		return
	}

	if _, isSplit := obj.(*microflows.InheritanceSplit); isSplit && len(findNormalFlows(flowsByOrigin[currentID])) > 1 {
		startLine := len(*lines) + headerLineCount
		emitInheritanceSplitStatement(ctx, currentID, mergeID, activityMap, flowsByOrigin, flowsByDest, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
		recordSourceMap(sourceMap, currentID, startLine, len(*lines)+headerLineCount-1)
		return
	}

	// Handle LoopedActivity inside a branch
	if loop, isLoop := obj.(*microflows.LoopedActivity); isLoop {
		startLine := len(*lines) + headerLineCount
		if stmt != "" {
			emitObjectAnnotations(obj, lines, indentStr, annotationsByTarget, flowsByOrigin, flowsByDest)
			*lines = append(*lines, indentStr+stmt)
		}

		*lines = append(*lines, indentStr+"begin")
		emitLoopBody(ctx, loop, flowsByOrigin, flowsByDest, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)

		*lines = append(*lines, indentStr+loopEndKeyword(loop)+";")
		recordSourceMap(sourceMap, currentID, startLine, len(*lines)+headerLineCount-1)

		// Continue after the loop within the branch
		flows := flowsByOrigin[currentID]
		for _, flow := range flows {
			traverseFlowUntilMergeLoopAware(ctx, flow.DestinationID, mergeID, loopHeaderID, activityMap, flowsByOrigin, flowsByDest, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
		}
		return
	}

	// Regular activity
	startLine := len(*lines) + headerLineCount
	normalFlows := findNormalFlows(flowsByOrigin[currentID])
	emitActivityStatement(ctx, obj, stmt, flowsByOrigin, flowsByDest, activityMap, entityNames, microflowNames, lines, indentStr, annotationsByTarget)
	recordSourceMap(sourceMap, currentID, startLine, len(*lines)+headerLineCount-1)

	// Follow normal (non-error-handler) outgoing flows until merge
	for _, flow := range normalFlows {
		traverseFlowUntilMergeLoopAware(ctx, flow.DestinationID, mergeID, loopHeaderID, activityMap, flowsByOrigin, flowsByDest, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
	}
}

func continueAfterSplitJoinLoopAware(
	ctx *ExecContext,
	joinID model.ID,
	loopHeaderID model.ID,
	activityMap map[model.ID]microflows.MicroflowObject,
	flowsByOrigin map[model.ID][]*microflows.SequenceFlow,
	flowsByDest map[model.ID][]*microflows.SequenceFlow,
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
	if joinID == "" {
		return
	}
	if _, isMerge := activityMap[joinID].(*microflows.ExclusiveMerge); isMerge {
		visited[joinID] = true
		for _, flow := range findNormalFlows(flowsByOrigin[joinID]) {
			traverseFlowLoopAware(ctx, flow.DestinationID, loopHeaderID, activityMap, flowsByOrigin, flowsByDest, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
		}
		return
	}
	traverseFlowLoopAware(ctx, joinID, loopHeaderID, activityMap, flowsByOrigin, flowsByDest, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
}

func continueAfterNestedSplitJoinLoopAware(
	ctx *ExecContext,
	joinID model.ID,
	stopID model.ID,
	loopHeaderID model.ID,
	activityMap map[model.ID]microflows.MicroflowObject,
	flowsByOrigin map[model.ID][]*microflows.SequenceFlow,
	flowsByDest map[model.ID][]*microflows.SequenceFlow,
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
	if joinID == "" || joinID == stopID {
		return
	}
	if _, isMerge := activityMap[joinID].(*microflows.ExclusiveMerge); isMerge {
		visited[joinID] = true
		for _, flow := range findNormalFlows(flowsByOrigin[joinID]) {
			traverseFlowUntilMergeLoopAware(ctx, flow.DestinationID, stopID, loopHeaderID, activityMap, flowsByOrigin, flowsByDest, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
		}
		return
	}
	traverseFlowUntilMergeLoopAware(ctx, joinID, stopID, loopHeaderID, activityMap, flowsByOrigin, flowsByDest, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
}

func emitInheritanceSplitStatement(
	ctx *ExecContext,
	currentID model.ID,
	stopID model.ID,
	activityMap map[model.ID]microflows.MicroflowObject,
	flowsByOrigin map[model.ID][]*microflows.SequenceFlow,
	flowsByDest map[model.ID][]*microflows.SequenceFlow,
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
	obj := activityMap[currentID]
	stmt := strings.TrimSuffix(strings.TrimSpace(formatActivity(ctx, obj, entityNames, microflowNames)), ";")
	indentStr := strings.Repeat("  ", indent)
	if stmt != "" {
		emitObjectAnnotations(obj, lines, indentStr, annotationsByTarget, flowsByOrigin, flowsByDest)
		*lines = append(*lines, indentStr+stmt)
	}

	mergeID := splitMergeMap[currentID]
	branchStopID := mergeID
	if branchStopID == "" {
		branchStopID = stopID
	}
	var elseFlow *microflows.SequenceFlow
	for _, flow := range findNormalFlows(flowsByOrigin[currentID]) {
		caseValue := inheritanceCaseValue(flow.CaseValue)
		if caseValue == "" {
			elseFlow = flow
			continue
		}
		*lines = append(*lines, indentStr+"case "+caseValue)
		traverseFlowUntilMerge(ctx, flow.DestinationID, branchStopID, activityMap, flowsByOrigin, flowsByDest, splitMergeMap, cloneVisited(visited), entityNames, microflowNames, lines, indent+1, sourceMap, headerLineCount, annotationsByTarget)
	}
	if elseFlow != nil {
		*lines = append(*lines, indentStr+"else")
		traverseFlowUntilMerge(ctx, elseFlow.DestinationID, branchStopID, activityMap, flowsByOrigin, flowsByDest, splitMergeMap, cloneVisited(visited), entityNames, microflowNames, lines, indent+1, sourceMap, headerLineCount, annotationsByTarget)
	}
	*lines = append(*lines, indentStr+"end split;")

	if mergeID != "" && mergeID != stopID {
		if _, isMerge := activityMap[mergeID].(*microflows.ExclusiveMerge); isMerge {
			visited[mergeID] = true
			for _, flow := range findNormalFlows(flowsByOrigin[mergeID]) {
				if stopID != "" {
					traverseFlowUntilMerge(ctx, flow.DestinationID, stopID, activityMap, flowsByOrigin, flowsByDest, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
				} else {
					traverseFlow(ctx, flow.DestinationID, activityMap, flowsByOrigin, flowsByDest, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
				}
			}
		} else {
			if stopID != "" {
				traverseFlowUntilMerge(ctx, mergeID, stopID, activityMap, flowsByOrigin, flowsByDest, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
			} else {
				traverseFlow(ctx, mergeID, activityMap, flowsByOrigin, flowsByDest, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
			}
		}
	}
}

func inheritanceCaseValue(cv microflows.CaseValue) string {
	switch c := cv.(type) {
	case *microflows.InheritanceCase:
		return c.EntityQualifiedName
	case microflows.InheritanceCase:
		return c.EntityQualifiedName
	default:
		return ""
	}
}

// traverseLoopBody traverses activities inside a loop body.
// When sourceMap is non-nil, it also records line ranges for each activity node.
func traverseLoopBody(
	ctx *ExecContext,
	currentID model.ID,
	activityMap map[model.ID]microflows.MicroflowObject,
	flowsByOrigin map[model.ID][]*microflows.SequenceFlow,
	flowsByDest map[model.ID][]*microflows.SequenceFlow,
	visited map[model.ID]bool,
	entityNames map[model.ID]string,
	microflowNames map[model.ID]string,
	lines *[]string,
	indent int,
	sourceMap map[string]elkSourceRange,
	headerLineCount int,
	annotationsByTarget map[model.ID][]string,
) {
	// Loop bodies can contain the same structured control flow as top-level
	// microflows. Reuse the main traversal with a loop-local split/merge map so
	// nested IF/ELSE blocks emit `else` / `end if;` correctly.
	splitMergeMap := findSplitMergePointsForGraph(ctx, activityMap, flowsByOrigin)
	traverseFlow(ctx, currentID, activityMap, flowsByOrigin, flowsByDest, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
}

// emitLoopBody processes the inner objects of a LoopedActivity.
// Shared by traverseFlow and traverseLoopBody for both top-level and nested loops.
func emitLoopBody(
	ctx *ExecContext,
	loop *microflows.LoopedActivity,
	flowsByOrigin map[model.ID][]*microflows.SequenceFlow,
	flowsByDest map[model.ID][]*microflows.SequenceFlow,
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

	// flowsByDest for the loop body mirrors the top-level one: anchor sides
	// depend on where incoming flows land, which is the parent-level map for
	// flows entering the loop boundary and the loop's own for internal edges.
	loopFlowsByDest := make(map[model.ID][]*microflows.SequenceFlow)
	if loop.ObjectCollection != nil {
		for _, flow := range loop.ObjectCollection.Flows {
			loopFlowsByDest[flow.DestinationID] = append(loopFlowsByDest[flow.DestinationID], flow)
		}
	}
	for destID, flows := range flowsByDest {
		if _, inLoop := loopActivityMap[destID]; inLoop {
			if _, exists := loopFlowsByDest[destID]; !exists {
				loopFlowsByDest[destID] = flows
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
		traverseLoopBody(ctx, firstID, loopActivityMap, loopFlowsByOrigin, loopFlowsByDest, loopVisited, entityNames, microflowNames, lines, indent+1, sourceMap, headerLineCount, annotationsByTarget)
	}
}

// isMergePairedWithSplit reports whether an ExclusiveMerge appears as the
// matching end-of-branch point for some ExclusiveSplit recorded in
// splitMergeMap (i.e., it closes an IF/ELSE block). Merges that aren't paired
// — e.g. a junction used as the loop-back target of a manual retry pattern —
// must be traversed as pass-through, otherwise every activity after them is
// dropped from describe output (issue #281).
//
// splitMergeMap is split-ID → merge-ID, so the merge is paired iff it appears
// as a value in the map.
func isMergePairedWithSplit(mergeID model.ID, splitMergeMap map[model.ID]model.ID) bool {
	for _, v := range splitMergeMap {
		if v == mergeID {
			return true
		}
	}
	return false
}

func isManualLoopHeaderMerge(
	mergeID model.ID,
	activityMap map[model.ID]microflows.MicroflowObject,
	flowsByOrigin map[model.ID][]*microflows.SequenceFlow,
	splitMergeMap map[model.ID]model.ID,
) bool {
	if _, ok := activityMap[mergeID].(*microflows.ExclusiveMerge); !ok {
		return false
	}
	normalOutgoing := findNormalFlows(flowsByOrigin[mergeID])
	if isMergePairedWithSplit(mergeID, splitMergeMap) || len(normalOutgoing) == 0 {
		return false
	}
	for _, flow := range normalOutgoing {
		if canReachNode(flow.DestinationID, mergeID, flowsByOrigin, make(map[model.ID]bool)) {
			return true
		}
	}
	return false
}

func canReachNode(
	currentID model.ID,
	targetID model.ID,
	flowsByOrigin map[model.ID][]*microflows.SequenceFlow,
	visited map[model.ID]bool,
) bool {
	if currentID == "" {
		return false
	}
	if currentID == targetID {
		return true
	}
	if visited[currentID] {
		return false
	}
	visited[currentID] = true
	for _, flow := range findNormalFlows(flowsByOrigin[currentID]) {
		if canReachNode(flow.DestinationID, targetID, flowsByOrigin, visited) {
			return true
		}
	}
	return false
}

func resolveNestedMergeID(
	nestedMergeID model.ID,
	parentMergeID model.ID,
	trueFlow *microflows.SequenceFlow,
	falseFlow *microflows.SequenceFlow,
	flowsByOrigin map[model.ID][]*microflows.SequenceFlow,
) model.ID {
	if nestedMergeID != "" && parentMergeID != "" && nestedMergeID != parentMergeID &&
		canReachNode(parentMergeID, nestedMergeID, flowsByOrigin, make(map[model.ID]bool)) {
		for _, flow := range []*microflows.SequenceFlow{trueFlow, falseFlow} {
			if flow == nil {
				continue
			}
			if flow.DestinationID == parentMergeID ||
				canReachNode(flow.DestinationID, parentMergeID, flowsByOrigin, make(map[model.ID]bool)) {
				return parentMergeID
			}
		}
	}
	if nestedMergeID != "" || parentMergeID == "" {
		return nestedMergeID
	}
	for _, flow := range []*microflows.SequenceFlow{trueFlow, falseFlow} {
		if flow == nil {
			continue
		}
		if canReachNode(flow.DestinationID, parentMergeID, flowsByOrigin, make(map[model.ID]bool)) {
			return parentMergeID
		}
	}
	return ""
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
	case *microflows.WebServiceCallAction:
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
	case *microflows.DownloadFileAction:
		return action.ErrorHandlingType
	default:
		// Fall back to activity level for action types without ErrorHandlingType field
		return activity.ErrorHandlingType
	}
}

func cloneVisited(visited map[model.ID]bool) map[model.ID]bool {
	cloned := make(map[model.ID]bool, len(visited))
	for id := range visited {
		cloned[id] = true
	}
	return cloned
}

func hasNormalIncomingFromOtherOrigin(
	flowsByOrigin map[model.ID][]*microflows.SequenceFlow,
	destinationID model.ID,
	originID model.ID,
) bool {
	for candidateOriginID, flows := range flowsByOrigin {
		if candidateOriginID == originID {
			continue
		}
		for _, flow := range findNormalFlows(flows) {
			if flow.DestinationID == destinationID {
				return true
			}
		}
	}
	return false
}

func hasNormalIncomingToDestination(
	flowsByOrigin map[model.ID][]*microflows.SequenceFlow,
	destinationID model.ID,
) bool {
	for _, flows := range flowsByOrigin {
		for _, flow := range findNormalFlows(flows) {
			if flow.DestinationID == destinationID {
				return true
			}
		}
	}
	return false
}

func collectStructuredStatements(
	ctx *ExecContext,
	currentID model.ID,
	stopID model.ID,
	activityMap map[model.ID]microflows.MicroflowObject,
	flowsByOrigin map[model.ID][]*microflows.SequenceFlow,
	splitMergeMap map[model.ID]model.ID,
	visited map[model.ID]bool,
	entityNames map[model.ID]string,
	microflowNames map[model.ID]string,
	lines *[]string,
	indent int,
) {
	if currentID == "" || currentID == stopID || visited[currentID] {
		return
	}

	obj := activityMap[currentID]
	if obj == nil {
		return
	}

	if _, isMerge := obj.(*microflows.ExclusiveMerge); isMerge {
		// Rejoin points back to the outer graph are already handled by stopID.
		// Other merges can be local junctions inside the error handler itself,
		// for example an empty nested handler that rejoins before a decision.
		visited[currentID] = true
		for _, flow := range findNormalFlows(flowsByOrigin[currentID]) {
			collectStructuredStatements(ctx, flow.DestinationID, stopID, activityMap, flowsByOrigin, splitMergeMap, visited, entityNames, microflowNames, lines, indent)
		}
		return
	}

	visited[currentID] = true

	stmt := formatActivity(ctx, obj, entityNames, microflowNames)
	indentStr := strings.Repeat("  ", indent)

	if _, isSplit := obj.(*microflows.ExclusiveSplit); isSplit {
		if stmt != "" {
			*lines = append(*lines, indentStr+stmt)
		}

		flows := flowsByOrigin[currentID]
		nestedMergeID := splitMergeMap[currentID]
		trueFlow, falseFlow := findBranchFlows(flows)
		nestedMergeID = resolveNestedMergeID(nestedMergeID, stopID, trueFlow, falseFlow, flowsByOrigin)

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
			guardMergeID := nestedMergeID
			if guardMergeID == "" {
				guardMergeID = stopID
			}
			traverseToID := trueFlow.DestinationID
			collectStructuredStatements(ctx, traverseToID, guardMergeID, activityMap, flowsByOrigin, splitMergeMap, visited, entityNames, microflowNames, lines, indent+1)
			*lines = append(*lines, indentStr+"end if;")

			if falseFlow != nil {
				contID := falseFlow.DestinationID
				if contID != guardMergeID {
					if _, isMerge := activityMap[contID].(*microflows.ExclusiveMerge); isMerge {
						visited[contID] = true
						for _, flow := range findNormalFlows(flowsByOrigin[contID]) {
							contID = flow.DestinationID
							break
						}
					}
					collectStructuredStatements(ctx, contID, guardMergeID, activityMap, flowsByOrigin, splitMergeMap, visited, entityNames, microflowNames, lines, indent)
				}
			}
			if guardMergeID != "" && guardMergeID != stopID {
				visited[guardMergeID] = true
				for _, flow := range findNormalFlows(flowsByOrigin[guardMergeID]) {
					collectStructuredStatements(ctx, flow.DestinationID, stopID, activityMap, flowsByOrigin, splitMergeMap, visited, entityNames, microflowNames, lines, indent)
				}
			}
		} else {
			if trueFlow != nil {
				collectStructuredBranchStatements(ctx, trueFlow.DestinationID, nestedMergeID, activityMap, flowsByOrigin, splitMergeMap, visited, entityNames, microflowNames, lines, indent+1)
			}

			falseHasBody := falseFlow != nil && falseFlow.DestinationID != nestedMergeID
			if falseHasBody {
				*lines = append(*lines, indentStr+"else")
				collectStructuredBranchStatements(ctx, falseFlow.DestinationID, nestedMergeID, activityMap, flowsByOrigin, splitMergeMap, cloneVisited(visited), entityNames, microflowNames, lines, indent+1)
			}

			*lines = append(*lines, indentStr+"end if;")

			if nestedMergeID != "" && nestedMergeID != stopID {
				visited[nestedMergeID] = true
				for _, flow := range findNormalFlows(flowsByOrigin[nestedMergeID]) {
					collectStructuredStatements(ctx, flow.DestinationID, stopID, activityMap, flowsByOrigin, splitMergeMap, visited, entityNames, microflowNames, lines, indent)
				}
			}
		}
		return
	}

	if loop, isLoop := obj.(*microflows.LoopedActivity); isLoop {
		if stmt != "" {
			*lines = append(*lines, indentStr+stmt)
		}
		*lines = append(*lines, indentStr+"begin")
		collectLoopBodyStatements(ctx, loop, flowsByOrigin, entityNames, microflowNames, lines, indent)
		*lines = append(*lines, indentStr+loopEndKeyword(loop)+";")

		for _, flow := range findNormalFlows(flowsByOrigin[currentID]) {
			collectStructuredStatements(ctx, flow.DestinationID, stopID, activityMap, flowsByOrigin, splitMergeMap, visited, entityNames, microflowNames, lines, indent)
		}
		return
	}

	appendStructuredCollectedStatement(ctx, obj, stmt, activityMap, flowsByOrigin, visited, entityNames, microflowNames, lines, indentStr)

	for _, flow := range findNormalFlows(flowsByOrigin[currentID]) {
		collectStructuredStatements(ctx, flow.DestinationID, stopID, activityMap, flowsByOrigin, splitMergeMap, visited, entityNames, microflowNames, lines, indent)
	}
}

func appendStructuredCollectedStatement(
	ctx *ExecContext,
	obj microflows.MicroflowObject,
	stmt string,
	activityMap map[model.ID]microflows.MicroflowObject,
	flowsByOrigin map[model.ID][]*microflows.SequenceFlow,
	visited map[model.ID]bool,
	entityNames map[model.ID]string,
	microflowNames map[model.ID]string,
	lines *[]string,
	indentStr string,
) {
	activity, isAction := obj.(*microflows.ActionActivity)
	if isAction {
		errType := getActionErrorHandlingType(activity)
		errorHandlerFlow := findErrorHandlerFlow(flowsByOrigin[obj.GetID()])
		if errorHandlerFlow != nil &&
			hasCustomErrorHandler(errType) &&
			(visited[errorHandlerFlow.DestinationID] ||
				hasNormalIncomingToDestination(flowsByOrigin, errorHandlerFlow.DestinationID) ||
				hasNormalIncomingFromOtherOrigin(flowsByOrigin, errorHandlerFlow.DestinationID, obj.GetID())) {
			stmtWithoutSemi := strings.TrimSuffix(strings.TrimSpace(stmt), ";")
			errorSuffix := formatErrorHandlingSuffix(errType)
			if errorSuffix == "" {
				errorSuffix = " on error without rollback"
			}
			*lines = append(*lines, indentStr+stmtWithoutSemi+errorSuffix+" { };")
			return
		}
	}
	appendFormattedStatement(ctx, obj, stmt, activityMap, flowsByOrigin, entityNames, microflowNames, lines, indentStr)
}

func collectStructuredBranchStatements(
	ctx *ExecContext,
	currentID model.ID,
	stopID model.ID,
	activityMap map[model.ID]microflows.MicroflowObject,
	flowsByOrigin map[model.ID][]*microflows.SequenceFlow,
	splitMergeMap map[model.ID]model.ID,
	visited map[model.ID]bool,
	entityNames map[model.ID]string,
	microflowNames map[model.ID]string,
	lines *[]string,
	indent int,
) {
	if currentID == "" {
		return
	}
	if stopID == "" {
		if _, isMerge := activityMap[currentID].(*microflows.ExclusiveMerge); isMerge {
			visited[currentID] = true
			for _, flow := range findNormalFlows(flowsByOrigin[currentID]) {
				collectStructuredStatements(ctx, flow.DestinationID, stopID, activityMap, flowsByOrigin, splitMergeMap, visited, entityNames, microflowNames, lines, indent)
			}
			return
		}
	}
	collectStructuredStatements(ctx, currentID, stopID, activityMap, flowsByOrigin, splitMergeMap, visited, entityNames, microflowNames, lines, indent)
}

func collectLoopBodyStatements(
	ctx *ExecContext,
	loop *microflows.LoopedActivity,
	flowsByOrigin map[model.ID][]*microflows.SequenceFlow,
	entityNames map[model.ID]string,
	microflowNames map[model.ID]string,
	lines *[]string,
	indent int,
) {
	if loop == nil || loop.ObjectCollection == nil || len(loop.ObjectCollection.Objects) == 0 {
		return
	}

	loopActivityMap := make(map[model.ID]microflows.MicroflowObject)
	for _, loopObj := range loop.ObjectCollection.Objects {
		loopActivityMap[loopObj.GetID()] = loopObj
	}

	loopFlowsByOrigin := make(map[model.ID][]*microflows.SequenceFlow)
	for _, flow := range loop.ObjectCollection.Flows {
		loopFlowsByOrigin[flow.OriginID] = append(loopFlowsByOrigin[flow.OriginID], flow)
	}
	for originID, flows := range flowsByOrigin {
		if _, inLoop := loopActivityMap[originID]; inLoop {
			if _, exists := loopFlowsByOrigin[originID]; !exists {
				loopFlowsByOrigin[originID] = flows
			}
		}
	}

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
	if firstID == "" {
		return
	}

	loopVisited := make(map[model.ID]bool)
	loopSplitMergeMap := findSplitMergePointsForGraph(ctx, loopActivityMap, loopFlowsByOrigin)
	collectStructuredStatements(ctx, firstID, "", loopActivityMap, loopFlowsByOrigin, loopSplitMergeMap, loopVisited, entityNames, microflowNames, lines, indent+1)
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
	splitMergeMap := findSplitMergePointsForGraph(ctx, activityMap, flowsByOrigin)
	stopID := findErrorHandlerRejoinID(startID, activityMap, flowsByOrigin)
	collectStructuredStatements(ctx, startID, stopID, activityMap, flowsByOrigin, splitMergeMap, visited, entityNames, microflowNames, &statements, 0)
	return statements
}

func findErrorHandlerRejoinID(
	startID model.ID,
	activityMap map[model.ID]microflows.MicroflowObject,
	flowsByOrigin map[model.ID][]*microflows.SequenceFlow,
) model.ID {
	if startID == "" {
		return ""
	}

	reachable := make(map[model.ID]bool)
	var markReachable func(model.ID)
	markReachable = func(id model.ID) {
		if id == "" || reachable[id] {
			return
		}
		reachable[id] = true
		for _, flow := range findNormalFlows(flowsByOrigin[id]) {
			markReachable(flow.DestinationID)
		}
	}
	markReachable(startID)

	flowsByDest := make(map[model.ID][]*microflows.SequenceFlow)
	for _, flows := range flowsByOrigin {
		for _, flow := range flows {
			flowsByDest[flow.DestinationID] = append(flowsByDest[flow.DestinationID], flow)
		}
	}

	visited := make(map[model.ID]bool)
	queue := []model.ID{startID}
	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]
		if currentID == "" || visited[currentID] {
			continue
		}
		visited[currentID] = true

		if currentID != startID {
			if _, isEnd := activityMap[currentID].(*microflows.EndEvent); isEnd {
				continue
			}
			if _, isMerge := activityMap[currentID].(*microflows.ExclusiveMerge); isMerge {
				for _, incoming := range findNormalFlows(flowsByDest[currentID]) {
					if !reachable[incoming.OriginID] {
						return currentID
					}
				}
			} else {
				for _, incoming := range findNormalFlows(flowsByDest[currentID]) {
					if !reachable[incoming.OriginID] {
						return currentID
					}
				}
			}
		}

		for _, flow := range findNormalFlows(flowsByOrigin[currentID]) {
			queue = append(queue, flow.DestinationID)
		}
	}
	return ""
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
	// Legacy wrapper — preserved for tests and unmigrated callers that don't
	// supply flowsByDest. Passing nil suppresses @anchor emission, matching
	// the pre-refactor behaviour.
	traverseFlow(e.newExecContext(context.Background()), currentID, activityMap, flowsByOrigin, nil, splitMergeMap, visited, entityNames, microflowNames, lines, indent, sourceMap, headerLineCount, annotationsByTarget)
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
