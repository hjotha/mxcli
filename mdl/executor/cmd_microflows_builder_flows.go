// SPDX-License-Identifier: Apache-2.0

// Package executor - Microflow flow graph: sequence flow constructors and error handler flows
package executor

import (
	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/types"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/microflows"
)

// convertErrorHandlingType converts AST error handling type to SDK error handling type.
func convertErrorHandlingType(eh *ast.ErrorHandlingClause) microflows.ErrorHandlingType {
	if eh == nil {
		return microflows.ErrorHandlingTypeRollback
	}
	switch eh.Type {
	case ast.ErrorHandlingContinue:
		return microflows.ErrorHandlingTypeContinue
	case ast.ErrorHandlingRollback:
		return microflows.ErrorHandlingTypeRollback
	case ast.ErrorHandlingCustom:
		return microflows.ErrorHandlingTypeCustom
	case ast.ErrorHandlingCustomWithoutRollback:
		return microflows.ErrorHandlingTypeCustomWithoutRollback
	default:
		return microflows.ErrorHandlingTypeRollback
	}
}

func isEmptyCustomErrorHandler(eh *ast.ErrorHandlingClause) bool {
	if eh == nil || len(eh.Body) != 0 {
		return false
	}
	return eh.Type == ast.ErrorHandlingCustom || eh.Type == ast.ErrorHandlingCustomWithoutRollback
}

func (fb *flowBuilder) registerEmptyCustomErrorHandler(activityID model.ID, eh *ast.ErrorHandlingClause) {
	if isEmptyCustomErrorHandler(eh) {
		fb.emptyErrorHandlerFrom = activityID
	}
}

func (fb *flowBuilder) registerEmptyCustomErrorHandlerWithSkip(activityID model.ID, eh *ast.ErrorHandlingClause, skipVar string) {
	if !isEmptyCustomErrorHandler(eh) {
		return
	}
	if skipVar == "" {
		fb.emptyErrorHandlerFrom = activityID
		return
	}
	fb.errorHandlerSource = activityID
	fb.errorHandlerTailFrom = activityID
	fb.errorHandlerSkipVar = skipVar
	fb.errorHandlerTailIsSource = true
}

type pendingErrorHandlerState struct {
	emptyFrom    model.ID
	tailFrom     model.ID
	source       model.ID
	skipVar      string
	tailIsSource bool
}

func (s pendingErrorHandlerState) isEmpty() bool {
	return s.emptyFrom == "" && s.tailFrom == "" && s.source == "" && s.skipVar == ""
}

func (fb *flowBuilder) capturePendingErrorHandler() pendingErrorHandlerState {
	return pendingErrorHandlerState{
		emptyFrom:    fb.emptyErrorHandlerFrom,
		tailFrom:     fb.errorHandlerTailFrom,
		source:       fb.errorHandlerSource,
		skipVar:      fb.errorHandlerSkipVar,
		tailIsSource: fb.errorHandlerTailIsSource,
	}
}

func (fb *flowBuilder) restorePendingErrorHandler(state pendingErrorHandlerState) {
	fb.emptyErrorHandlerFrom = state.emptyFrom
	fb.errorHandlerTailFrom = state.tailFrom
	fb.errorHandlerSource = state.source
	fb.errorHandlerSkipVar = state.skipVar
	fb.errorHandlerTailIsSource = state.tailIsSource
}

func (fb *flowBuilder) clearPendingErrorHandler() {
	fb.restorePendingErrorHandler(pendingErrorHandlerState{})
}

func (fb *flowBuilder) addPendingEmptyErrorHandlerFlow(originID, destinationID model.ID) {
	if fb.emptyErrorHandlerFrom != "" && fb.emptyErrorHandlerFrom == originID && destinationID != "" {
		fb.flows = append(fb.flows, newErrorHandlerFlow(originID, destinationID))
		fb.emptyErrorHandlerFrom = ""
	}
	if fb.errorHandlerSource != "" && fb.errorHandlerSource == originID && fb.errorHandlerTailFrom != "" && destinationID != "" {
		fb.flows = append(fb.flows, newHorizontalFlow(fb.errorHandlerTailFrom, destinationID))
		fb.errorHandlerSource = ""
		fb.errorHandlerTailFrom = ""
		fb.errorHandlerTailIsSource = false
	}
}

func (fb *flowBuilder) addPendingErrorHandlerFlowForStatement(originID, destinationID model.ID, stmt ast.MicroflowStatement, futureReferencesSkipVar ...bool) {
	if fb.emptyErrorHandlerFrom != "" && fb.emptyErrorHandlerFrom == originID && destinationID != "" {
		fb.flows = append(fb.flows, newErrorHandlerFlow(originID, destinationID))
		fb.emptyErrorHandlerFrom = ""
	}
	if fb.errorHandlerTailFrom == "" || destinationID == "" {
		return
	}
	if fb.errorHandlerSource != "" && destinationID == fb.errorHandlerSource {
		return
	}
	if fb.errorHandlerSkipVar != "" {
		if statementReferencesVar(stmt, fb.errorHandlerSkipVar) {
			return
		}
		if len(futureReferencesSkipVar) > 0 && futureReferencesSkipVar[0] {
			return
		}
		fb.addErrorHandlerRejoinFlow(originID, destinationID)
		fb.errorHandlerSource = ""
		fb.errorHandlerTailFrom = ""
		fb.errorHandlerSkipVar = ""
		fb.errorHandlerTailIsSource = false
		return
	}
	if fb.errorHandlerSource != "" && fb.errorHandlerSource == originID {
		fb.flows = append(fb.flows, newHorizontalFlow(fb.errorHandlerTailFrom, destinationID))
		fb.errorHandlerSource = ""
		fb.errorHandlerTailFrom = ""
		fb.errorHandlerTailIsSource = false
	}
}

func (fb *flowBuilder) addErrorHandlerRejoinFlow(originID, destinationID model.ID) {
	existingIdx := -1
	for i := len(fb.flows) - 1; i >= 0; i-- {
		flow := fb.flows[i]
		if !flow.IsErrorHandler && flow.OriginID == originID && flow.DestinationID == destinationID {
			existingIdx = i
			break
		}
	}
	if existingIdx == -1 {
		if fb.errorHandlerTailIsSource {
			fb.flows = append(fb.flows, newErrorHandlerFlow(fb.errorHandlerTailFrom, destinationID))
		} else {
			fb.flows = append(fb.flows, newHorizontalFlow(fb.errorHandlerTailFrom, destinationID))
		}
		return
	}

	existing := fb.flows[existingIdx]
	fb.flows = append(fb.flows[:existingIdx], fb.flows[existingIdx+1:]...)

	merge := &microflows.ExclusiveMerge{
		BaseMicroflowObject: microflows.BaseMicroflowObject{
			BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
			Position:    model.Point{X: fb.posX - HorizontalSpacing/2, Y: fb.baseY},
			Size:        model.Size{Width: MergeSize, Height: MergeSize},
		},
	}
	fb.objects = append(fb.objects, merge)

	normalFlow := newHorizontalFlow(originID, merge.ID)
	normalFlow.OriginConnectionIndex = existing.OriginConnectionIndex
	normalFlow.CaseValue = existing.CaseValue
	fb.flows = append(fb.flows, normalFlow)
	if fb.errorHandlerTailIsSource {
		fb.flows = append(fb.flows, newErrorHandlerFlow(fb.errorHandlerTailFrom, merge.ID))
	} else {
		fb.flows = append(fb.flows, newUpwardFlow(fb.errorHandlerTailFrom, merge.ID))
	}

	mergeFlow := newHorizontalFlow(merge.ID, destinationID)
	mergeFlow.DestinationConnectionIndex = existing.DestinationConnectionIndex
	fb.flows = append(fb.flows, mergeFlow)
}

func statementReferencesVar(stmt ast.MicroflowStatement, varName string) bool {
	if stmt == nil || varName == "" {
		return false
	}
	for _, ref := range statementVarRefs(stmt) {
		if ref == varName {
			return true
		}
	}
	return false
}

func statementsReferenceVar(stmts []ast.MicroflowStatement, varName string) bool {
	if varName == "" {
		return false
	}
	for _, stmt := range stmts {
		if statementReferencesVar(stmt, varName) {
			return true
		}
	}
	return false
}

func exprReferencesVar(expr ast.Expression, varName string) bool {
	if varName == "" {
		return false
	}
	for _, ref := range exprVarRefs(expr) {
		if ref == varName {
			return true
		}
	}
	return false
}

func statementVarRefs(stmt ast.MicroflowStatement) []string {
	var refs []string
	switch s := stmt.(type) {
	case *ast.ReturnStmt:
		refs = append(refs, exprVarRefs(s.Value)...)
	case *ast.LogStmt:
		refs = append(refs, exprVarRefs(s.Node)...)
		refs = append(refs, exprVarRefs(s.Message)...)
		for _, param := range s.Template {
			refs = append(refs, exprVarRefs(param.Value)...)
		}
	case *ast.IfStmt:
		refs = append(refs, exprVarRefs(s.Condition)...)
		refs = append(refs, statementsVarRefs(s.ThenBody)...)
		refs = append(refs, statementsVarRefs(s.ElseBody)...)
	case *ast.WhileStmt:
		refs = append(refs, exprVarRefs(s.Condition)...)
		refs = append(refs, statementsVarRefs(s.Body)...)
	case *ast.LoopStmt:
		refs = append(refs, s.ListVariable)
		refs = append(refs, statementsVarRefs(s.Body)...)
	case *ast.MfSetStmt:
		refs = append(refs, extractVarName(s.Target))
		refs = append(refs, exprVarRefs(s.Value)...)
	case *ast.ChangeObjectStmt:
		refs = append(refs, s.Variable)
		for _, change := range s.Changes {
			refs = append(refs, exprVarRefs(change.Value)...)
		}
	case *ast.CreateObjectStmt:
		for _, change := range s.Changes {
			refs = append(refs, exprVarRefs(change.Value)...)
		}
	case *ast.CallMicroflowStmt:
		for _, arg := range s.Arguments {
			refs = append(refs, exprVarRefs(arg.Value)...)
		}
	case *ast.CallJavaActionStmt:
		for _, arg := range s.Arguments {
			refs = append(refs, exprVarRefs(arg.Value)...)
		}
	case *ast.RestCallStmt:
		refs = append(refs, exprVarRefs(s.URL)...)
		for _, param := range s.URLParams {
			refs = append(refs, exprVarRefs(param.Value)...)
		}
		for _, header := range s.Headers {
			refs = append(refs, exprVarRefs(header.Value)...)
		}
		if s.Auth != nil {
			refs = append(refs, exprVarRefs(s.Auth.Username)...)
			refs = append(refs, exprVarRefs(s.Auth.Password)...)
		}
		if s.Body != nil {
			refs = append(refs, exprVarRefs(s.Body.Template)...)
			for _, param := range s.Body.TemplateParams {
				refs = append(refs, exprVarRefs(param.Value)...)
			}
			if s.Body.SourceVariable != "" {
				refs = append(refs, s.Body.SourceVariable)
			}
		}
		refs = append(refs, exprVarRefs(s.Timeout)...)
	case *ast.MfCommitStmt:
		refs = append(refs, s.Variable)
	case *ast.DeleteObjectStmt:
		refs = append(refs, s.Variable)
	case *ast.AddToListStmt:
		refs = append(refs, s.Item, s.List)
	case *ast.RemoveFromListStmt:
		refs = append(refs, s.Item, s.List)
	}
	return refs
}

func statementsVarRefs(stmts []ast.MicroflowStatement) []string {
	var refs []string
	for _, stmt := range stmts {
		refs = append(refs, statementVarRefs(stmt)...)
	}
	return refs
}

// newErrorHandlerFlow creates a SequenceFlow with IsErrorHandler=true,
// connecting from the bottom of the source activity to the left of the error handler.
func newErrorHandlerFlow(originID, destinationID model.ID) *microflows.SequenceFlow {
	return &microflows.SequenceFlow{
		BaseElement:                model.BaseElement{ID: model.ID(types.GenerateID())},
		OriginID:                   originID,
		DestinationID:              destinationID,
		OriginConnectionIndex:      AnchorBottom,
		DestinationConnectionIndex: AnchorLeft,
		IsErrorHandler:             true,
	}
}

// addErrorHandlerFlow builds error handler activities from the given body statements,
// positions them below the source activity, and connects them with an error handler flow.
// Returns the last activity ID if the error handler should merge back to the main flow.
// Returns empty model.ID if the error handler terminates (via RAISE ERROR or RETURN).
func (fb *flowBuilder) addErrorHandlerFlow(sourceActivityID model.ID, sourceX int, errorBody []ast.MicroflowStatement) model.ID {
	if len(errorBody) == 0 {
		return ""
	}

	// Position error handler below the main flow
	errorY := fb.posY + VerticalSpacing
	errorX := sourceX

	// Build error handler activities
	errBuilder := &flowBuilder{
		posX:         errorX,
		posY:         errorY,
		baseY:        errorY,
		spacing:      HorizontalSpacing,
		varTypes:     fb.varTypes,
		declaredVars: fb.declaredVars,
		measurer:     fb.measurer,
		backend:      fb.backend,
		hierarchy:    fb.hierarchy,
		restServices: fb.restServices,
	}

	var lastErrID model.ID
	for _, stmt := range errorBody {
		actID := errBuilder.addStatement(stmt)
		if actID != "" {
			if lastErrID == "" {
				// Connect source activity to first error handler activity
				fb.flows = append(fb.flows, newErrorHandlerFlow(sourceActivityID, actID))
			} else {
				errBuilder.flows = append(errBuilder.flows, newHorizontalFlow(lastErrID, actID))
			}
			if errBuilder.nextConnectionPoint != "" {
				lastErrID = errBuilder.nextConnectionPoint
				errBuilder.nextConnectionPoint = ""
			} else {
				lastErrID = actID
			}
		}
	}

	// Append error handler objects and flows to the main builder
	fb.objects = append(fb.objects, errBuilder.objects...)
	fb.flows = append(fb.flows, errBuilder.flows...)

	// If the error handler ends with RAISE ERROR or RETURN, it terminates there.
	// Otherwise, return the last activity ID so caller can create a merge.
	if errBuilder.endsWithReturn {
		return "" // Error handler terminates, no merge needed
	}
	return lastErrID // Error handler should merge back to main flow
}

// handleErrorHandlerMerge reconnects non-terminal custom error handlers to the
// same next activity as the main success path.
func (fb *flowBuilder) handleErrorHandlerMerge(lastErrID model.ID, activityID model.ID, errorY int) {
	fb.handleErrorHandlerMergeWithSkip(lastErrID, activityID, errorY, "")
}

func (fb *flowBuilder) handleErrorHandlerMergeWithSkip(lastErrID model.ID, activityID model.ID, errorY int, skipVar string) {
	if lastErrID == "" {
		return // No merge needed (error handler terminates with RETURN or RAISE ERROR)
	}
	_ = errorY
	if fb.manualLoopBackTarget != "" {
		fb.flows = append(fb.flows, newHorizontalFlow(lastErrID, fb.manualLoopBackTarget))
		return
	}
	fb.errorHandlerSource = activityID
	fb.errorHandlerTailFrom = lastErrID
	fb.errorHandlerSkipVar = skipVar
	fb.errorHandlerTailIsSource = false
}

// newHorizontalFlow creates a SequenceFlow with anchors for horizontal left-to-right connection
func newHorizontalFlow(originID, destinationID model.ID) *microflows.SequenceFlow {
	return &microflows.SequenceFlow{
		BaseElement:                model.BaseElement{ID: model.ID(types.GenerateID())},
		OriginID:                   originID,
		DestinationID:              destinationID,
		OriginConnectionIndex:      AnchorRight, // Connect from right side of origin
		DestinationConnectionIndex: AnchorLeft,  // Connect to left side of destination
	}
}

// newHorizontalFlowWithCase creates a horizontal SequenceFlow with a boolean case value (for splits)
func newHorizontalFlowWithCase(originID, destinationID model.ID, caseValue string) *microflows.SequenceFlow {
	flow := newHorizontalFlow(originID, destinationID)
	flow.CaseValue = microflows.EnumerationCase{
		BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
		Value:       caseValue, // "true" or "false" as string
	}
	return flow
}

func newHorizontalFlowWithInheritanceCase(originID, destinationID model.ID, caseValue string) *microflows.SequenceFlow {
	flow := newHorizontalFlow(originID, destinationID)
	flow.CaseValue = microflows.InheritanceCase{
		BaseElement:         model.BaseElement{ID: model.ID(types.GenerateID())},
		EntityQualifiedName: caseValue,
	}
	return flow
}

// newDownwardFlowWithCase creates a SequenceFlow going down from origin (Bottom) to destination (Left)
// Used when TRUE path goes below the main line
func newDownwardFlowWithCase(originID, destinationID model.ID, caseValue string) *microflows.SequenceFlow {
	return &microflows.SequenceFlow{
		BaseElement:                model.BaseElement{ID: model.ID(types.GenerateID())},
		OriginID:                   originID,
		DestinationID:              destinationID,
		OriginConnectionIndex:      AnchorBottom, // Connect from bottom of origin (going down)
		DestinationConnectionIndex: AnchorLeft,   // Connect to left side of destination
		CaseValue: microflows.EnumerationCase{
			BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
			Value:       caseValue, // "true" or "false" as string
		},
	}
}

func newDownwardFlowWithInheritanceCase(originID, destinationID model.ID, caseValue string) *microflows.SequenceFlow {
	return &microflows.SequenceFlow{
		BaseElement:                model.BaseElement{ID: model.ID(types.GenerateID())},
		OriginID:                   originID,
		DestinationID:              destinationID,
		OriginConnectionIndex:      AnchorBottom,
		DestinationConnectionIndex: AnchorLeft,
		CaseValue: microflows.InheritanceCase{
			BaseElement:         model.BaseElement{ID: model.ID(types.GenerateID())},
			EntityQualifiedName: caseValue,
		},
	}
}

// newUpwardFlow creates a SequenceFlow going up from origin (Right) to destination (Top)
// Used when returning from a lower branch to merge
func newUpwardFlow(originID, destinationID model.ID) *microflows.SequenceFlow {
	return &microflows.SequenceFlow{
		BaseElement:                model.BaseElement{ID: model.ID(types.GenerateID())},
		OriginID:                   originID,
		DestinationID:              destinationID,
		OriginConnectionIndex:      AnchorRight,  // Connect from right side of origin
		DestinationConnectionIndex: AnchorBottom, // Connect to bottom of destination (going up)
	}
}

// applyUserAnchors overrides a SequenceFlow's Origin/Destination connection
// indices with user-specified values from @anchor annotations. An AnchorSideUnset
// value leaves the builder-chosen default in place.
//
// Arguments are taken as *ast.FlowAnchors pointers for the origin and
// destination statements. Either can be nil (no annotation on that side).
// If both are non-nil, the origin's From and destination's To are applied.
func applyUserAnchors(flow *microflows.SequenceFlow, origin *ast.FlowAnchors, destination *ast.FlowAnchors) {
	if flow == nil {
		return
	}
	if origin != nil && origin.From != ast.AnchorSideUnset {
		flow.OriginConnectionIndex = int(origin.From)
	}
	if destination != nil && destination.To != ast.AnchorSideUnset {
		flow.DestinationConnectionIndex = int(destination.To)
	}
}

// lastStmtIsReturn reports whether execution of a body is guaranteed to terminate
// (via RETURN, RAISE ERROR, BREAK, or CONTINUE) on every path — i.e. control can
// never fall off the end of the body into the parent flow.
//
// Terminal statements: ReturnStmt, RaiseErrorStmt, BreakStmt, ContinueStmt. An
// IfStmt is terminal iff it has an ELSE and both branches are terminal
// (recursively). A LoopStmt is never terminal — BREAK can exit the loop even if
// the body returns.
//
// Naming kept for history; the predicate is really "last stmt is a guaranteed
// terminator". Missing this case causes the outer IF to emit a dangling
// continuation flow (duplicate "true" edge + orphan EndEvent), which Studio Pro
// rejects as "Sequence contains no matching element" when diffing.
func lastStmtIsReturn(stmts []ast.MicroflowStatement) bool {
	if len(stmts) == 0 {
		return false
	}
	return isTerminalStmt(stmts[len(stmts)-1])
}

func isTerminalStmt(stmt ast.MicroflowStatement) bool {
	switch s := stmt.(type) {
	case *ast.ReturnStmt:
		return true
	case *ast.RaiseErrorStmt:
		return true
	case *ast.BreakStmt:
		return true
	case *ast.ContinueStmt:
		return true
	case *ast.IfStmt:
		if len(s.ElseBody) == 0 {
			return false
		}
		return lastStmtIsReturn(s.ThenBody) && lastStmtIsReturn(s.ElseBody)
	case *ast.InheritanceSplitStmt:
		if len(s.ElseBody) == 0 {
			return false
		}
		for _, c := range s.Cases {
			if !lastStmtIsReturn(c.Body) {
				return false
			}
		}
		return lastStmtIsReturn(s.ElseBody)
	default:
		return false
	}
}
