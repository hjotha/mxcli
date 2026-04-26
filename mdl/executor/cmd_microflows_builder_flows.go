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
		fb.queueActivePendingErrorHandler()
		fb.emptyErrorHandlerFrom = activityID
	}
}

func (fb *flowBuilder) registerEmptyCustomErrorHandlerWithSkip(activityID model.ID, eh *ast.ErrorHandlingClause, skipVar string) {
	if !isEmptyCustomErrorHandler(eh) {
		return
	}
	fb.queueActivePendingErrorHandler()
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
	returnValue  string
	queue        []pendingErrorHandlerState
}

func (s pendingErrorHandlerState) isEmpty() bool {
	return s.activeIsEmpty() && len(s.queue) == 0
}

func (s pendingErrorHandlerState) activeIsEmpty() bool {
	return s.emptyFrom == "" && s.tailFrom == "" && s.source == "" && s.skipVar == ""
}

func (s pendingErrorHandlerState) hasSkipVar() bool {
	if s.skipVar != "" {
		return true
	}
	for _, queued := range s.queue {
		if queued.skipVar != "" {
			return true
		}
	}
	return false
}

func (fb *flowBuilder) capturePendingErrorHandler() pendingErrorHandlerState {
	return pendingErrorHandlerState{
		emptyFrom:    fb.emptyErrorHandlerFrom,
		tailFrom:     fb.errorHandlerTailFrom,
		source:       fb.errorHandlerSource,
		skipVar:      fb.errorHandlerSkipVar,
		tailIsSource: fb.errorHandlerTailIsSource,
		returnValue:  fb.errorHandlerReturnValue,
		queue:        append([]pendingErrorHandlerState(nil), fb.pendingErrorHandlers...),
	}
}

func (fb *flowBuilder) restorePendingErrorHandler(state pendingErrorHandlerState) {
	fb.emptyErrorHandlerFrom = state.emptyFrom
	fb.errorHandlerTailFrom = state.tailFrom
	fb.errorHandlerSource = state.source
	fb.errorHandlerSkipVar = state.skipVar
	fb.errorHandlerTailIsSource = state.tailIsSource
	fb.errorHandlerReturnValue = state.returnValue
	fb.pendingErrorHandlers = append([]pendingErrorHandlerState(nil), state.queue...)
}

func (fb *flowBuilder) clearPendingErrorHandler() {
	fb.restorePendingErrorHandler(pendingErrorHandlerState{})
}

func (fb *flowBuilder) activePendingErrorHandler() pendingErrorHandlerState {
	return pendingErrorHandlerState{
		emptyFrom:    fb.emptyErrorHandlerFrom,
		tailFrom:     fb.errorHandlerTailFrom,
		source:       fb.errorHandlerSource,
		skipVar:      fb.errorHandlerSkipVar,
		tailIsSource: fb.errorHandlerTailIsSource,
		returnValue:  fb.errorHandlerReturnValue,
	}
}

func (fb *flowBuilder) setActivePendingErrorHandler(state pendingErrorHandlerState) {
	fb.emptyErrorHandlerFrom = state.emptyFrom
	fb.errorHandlerTailFrom = state.tailFrom
	fb.errorHandlerSource = state.source
	fb.errorHandlerSkipVar = state.skipVar
	fb.errorHandlerTailIsSource = state.tailIsSource
	fb.errorHandlerReturnValue = state.returnValue
}

func (fb *flowBuilder) queueActivePendingErrorHandler() {
	state := fb.activePendingErrorHandler()
	if state.activeIsEmpty() {
		return
	}
	fb.pendingErrorHandlers = append(fb.pendingErrorHandlers, state)
	fb.setActivePendingErrorHandler(pendingErrorHandlerState{})
}

func (fb *flowBuilder) addPendingEmptyErrorHandlerFlow(originID, destinationID model.ID) {
	fb.rewritePendingErrorHandlers(func(state pendingErrorHandlerState) pendingErrorHandlerState {
		return fb.addPendingEmptyErrorHandlerFlowForState(state, originID, destinationID)
	})
}

func (fb *flowBuilder) addPendingErrorHandlerFlowForStatement(originID, destinationID model.ID, stmt ast.MicroflowStatement, futureReferencesSkipVar ...bool) {
	futureReferences := len(futureReferencesSkipVar) > 0 && futureReferencesSkipVar[0]
	pendingSkipVars := fb.pendingErrorHandlerSkipVars()
	hasPendingOutput := len(pendingSkipVars) > 0
	referencesPendingOutput := statementReferencesAnyVar(stmt, pendingSkipVars)
	fb.rewritePendingErrorHandlers(func(state pendingErrorHandlerState) pendingErrorHandlerState {
		return fb.addPendingErrorHandlerFlowForState(state, originID, destinationID, stmt, futureReferences, hasPendingOutput, referencesPendingOutput)
	})
}

func (fb *flowBuilder) addPendingErrorHandlerFlowTo(destinationID model.ID) {
	if destinationID == "" {
		return
	}
	fb.rewritePendingErrorHandlers(func(state pendingErrorHandlerState) pendingErrorHandlerState {
		if state.emptyFrom != "" {
			fb.addEmptyErrorHandlerRejoinFlow(state.emptyFrom, destinationID)
			state.emptyFrom = ""
		}
		if state.source != "" && state.tailFrom != "" {
			fb.addErrorHandlerRejoinFlowForState(state, state.source, destinationID)
			state.source = ""
			state.tailFrom = ""
			state.skipVar = ""
			state.tailIsSource = false
			state.returnValue = ""
		}
		return state
	})
}

func (fb *flowBuilder) routePendingErrorHandlerToAlternative(originID, destinationID model.ID) {
	if originID == "" || destinationID == "" {
		return
	}
	fb.rewritePendingErrorHandlers(func(state pendingErrorHandlerState) pendingErrorHandlerState {
		if state.source != "" && state.tailFrom != "" {
			fb.addErrorHandlerRejoinFlowForState(state, originID, destinationID)
			state.source = ""
			state.tailFrom = ""
			state.skipVar = ""
			state.tailIsSource = false
			state.returnValue = ""
		}
		return state
	})
}

func (fb *flowBuilder) rewritePendingErrorHandlers(rewrite func(pendingErrorHandlerState) pendingErrorHandlerState) {
	queue := fb.pendingErrorHandlers[:0]
	for _, state := range fb.pendingErrorHandlers {
		state = rewrite(state)
		if !state.activeIsEmpty() {
			queue = append(queue, state)
		}
	}
	fb.pendingErrorHandlers = queue

	active := rewrite(fb.activePendingErrorHandler())
	fb.setActivePendingErrorHandler(active)
}

func (fb *flowBuilder) pendingErrorHandlerSkipVars() []string {
	var vars []string
	for _, state := range fb.pendingErrorHandlers {
		if state.skipVar != "" {
			vars = append(vars, state.skipVar)
		}
	}
	if fb.errorHandlerSkipVar != "" {
		vars = append(vars, fb.errorHandlerSkipVar)
	}
	return vars
}

func (fb *flowBuilder) addPendingEmptyErrorHandlerFlowForState(state pendingErrorHandlerState, originID, destinationID model.ID) pendingErrorHandlerState {
	if destinationID == "" {
		return state
	}
	if state.emptyFrom != "" && state.emptyFrom == originID {
		fb.addEmptyErrorHandlerRejoinFlow(originID, destinationID)
		state.emptyFrom = ""
	}
	if state.source != "" && state.source == originID && state.tailFrom != "" {
		fb.flows = append(fb.flows, newHorizontalFlow(state.tailFrom, destinationID))
		state.source = ""
		state.tailFrom = ""
		state.tailIsSource = false
		state.returnValue = ""
	}
	return state
}

func (fb *flowBuilder) addPendingErrorHandlerFlowForState(state pendingErrorHandlerState, originID, destinationID model.ID, stmt ast.MicroflowStatement, futureReferencesSkipVar bool, hasPendingOutput bool, referencesPendingOutput bool) pendingErrorHandlerState {
	if destinationID == "" {
		return state
	}
	if state.emptyFrom != "" {
		if state.emptyFrom != originID {
			return state
		}
		fb.addEmptyErrorHandlerRejoinFlowFrom(originID, state.emptyFrom, destinationID)
		state.emptyFrom = ""
	}
	if state.tailFrom == "" {
		return state
	}
	if state.source != "" && destinationID == state.source {
		return state
	}
	if state.skipVar != "" {
		referencesSkipVar := statementReferencesVar(stmt, state.skipVar)
		if referencesSkipVar {
			if state.returnValue != "" {
				endID := fb.addTerminalEndEventForPendingHandler(fb.returnType, state.returnValue)
				if state.tailIsSource {
					fb.flows = append(fb.flows, newErrorHandlerFlow(state.tailFrom, endID))
				} else {
					fb.flows = append(fb.flows, newHorizontalFlow(state.tailFrom, endID))
				}
				state.source = ""
				state.tailFrom = ""
				state.skipVar = ""
				state.tailIsSource = false
				state.returnValue = ""
				return state
			}
			if _, ok := stmt.(*ast.ReturnStmt); ok {
				return state
			}
			if !fb.hasReturnValue {
				endID := fb.addTerminalEndEventForPendingHandler(fb.returnType, "")
				if state.tailIsSource {
					fb.flows = append(fb.flows, newErrorHandlerFlow(state.tailFrom, endID))
				} else {
					fb.flows = append(fb.flows, newHorizontalFlow(state.tailFrom, endID))
				}
				state.source = ""
				state.tailFrom = ""
				state.skipVar = ""
				state.tailIsSource = false
				state.returnValue = ""
				return state
			}
			return state
		}

		if futureReferencesSkipVar {
			return state
		}

		fb.addErrorHandlerRejoinFlowForState(state, originID, destinationID)
		state.source = ""
		state.tailFrom = ""
		state.skipVar = ""
		state.tailIsSource = false
		state.returnValue = ""
		return state
	}
	if state.source != "" && state.source == originID {
		fb.addErrorHandlerRejoinFlowForState(state, originID, destinationID)
		state.source = ""
		state.tailFrom = ""
		state.tailIsSource = false
		state.returnValue = ""
	}
	return state
}

func (fb *flowBuilder) addEmptyErrorHandlerRejoinFlow(originID, destinationID model.ID) {
	fb.addEmptyErrorHandlerRejoinFlowFrom(originID, originID, destinationID)
}

func (fb *flowBuilder) addEmptyErrorHandlerRejoinFlowFrom(normalOriginID, errorOriginID, destinationID model.ID) {
	existingIdx := -1
	for i := len(fb.flows) - 1; i >= 0; i-- {
		flow := fb.flows[i]
		if !flow.IsErrorHandler && flow.OriginID == normalOriginID && flow.DestinationID == destinationID {
			existingIdx = i
			break
		}
	}
	if existingIdx == -1 {
		if mergeID := fb.findExistingRejoinMerge(normalOriginID, destinationID); mergeID != "" {
			fb.flows = append(fb.flows, newErrorHandlerFlow(errorOriginID, mergeID))
			return
		}
		fb.flows = append(fb.flows, newErrorHandlerFlow(errorOriginID, destinationID))
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

	normalFlow := newHorizontalFlow(normalOriginID, merge.ID)
	normalFlow.OriginConnectionIndex = existing.OriginConnectionIndex
	normalFlow.CaseValue = existing.CaseValue
	fb.flows = append(fb.flows, normalFlow)
	fb.flows = append(fb.flows, newErrorHandlerFlow(errorOriginID, merge.ID))

	mergeFlow := newHorizontalFlow(merge.ID, destinationID)
	mergeFlow.DestinationConnectionIndex = existing.DestinationConnectionIndex
	fb.flows = append(fb.flows, mergeFlow)
}

func (fb *flowBuilder) addErrorHandlerRejoinFlowForState(state pendingErrorHandlerState, originID, destinationID model.ID) {
	existingIdx := -1
	for i := len(fb.flows) - 1; i >= 0; i-- {
		flow := fb.flows[i]
		if !flow.IsErrorHandler && flow.OriginID == originID && flow.DestinationID == destinationID {
			existingIdx = i
			break
		}
	}
	if existingIdx == -1 {
		if mergeID := fb.findExistingRejoinMerge(originID, destinationID); mergeID != "" {
			if state.tailIsSource {
				fb.flows = append(fb.flows, newErrorHandlerFlow(state.tailFrom, mergeID))
			} else {
				fb.flows = append(fb.flows, newUpwardFlow(state.tailFrom, mergeID))
			}
			return
		}
		if state.tailIsSource {
			fb.flows = append(fb.flows, newErrorHandlerFlow(state.tailFrom, destinationID))
		} else {
			fb.flows = append(fb.flows, newHorizontalFlow(state.tailFrom, destinationID))
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
	if state.tailIsSource {
		fb.flows = append(fb.flows, newErrorHandlerFlow(state.tailFrom, merge.ID))
	} else {
		fb.flows = append(fb.flows, newUpwardFlow(state.tailFrom, merge.ID))
	}

	mergeFlow := newHorizontalFlow(merge.ID, destinationID)
	mergeFlow.DestinationConnectionIndex = existing.DestinationConnectionIndex
	fb.flows = append(fb.flows, mergeFlow)
}

func (fb *flowBuilder) findExistingRejoinMerge(originID, destinationID model.ID) model.ID {
	for _, flow := range fb.flows {
		if flow.OriginID != originID || flow.IsErrorHandler {
			continue
		}
		if !fb.isExclusiveMerge(flow.DestinationID) {
			continue
		}
		for _, mergeFlow := range fb.flows {
			if mergeFlow.OriginID == flow.DestinationID && mergeFlow.DestinationID == destinationID && !mergeFlow.IsErrorHandler {
				return flow.DestinationID
			}
		}
	}
	return ""
}

func (fb *flowBuilder) isExclusiveMerge(id model.ID) bool {
	for _, obj := range fb.objects {
		if obj.GetID() != id {
			continue
		}
		_, ok := obj.(*microflows.ExclusiveMerge)
		return ok
	}
	return false
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

func statementReferencesAnyVar(stmt ast.MicroflowStatement, varNames []string) bool {
	if len(varNames) == 0 {
		return false
	}
	lookup := map[string]bool{}
	for _, name := range varNames {
		if name != "" {
			lookup[name] = true
		}
	}
	for _, ref := range statementVarRefs(stmt) {
		if lookup[ref] {
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
	case *ast.InheritanceSplitStmt:
		refs = append(refs, s.Variable)
		for _, c := range s.Cases {
			refs = append(refs, statementsVarRefs(c.Body)...)
		}
		refs = append(refs, statementsVarRefs(s.ElseBody)...)
	case *ast.CastObjectStmt:
		refs = append(refs, s.ObjectVariable)
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
	case *ast.CallWebServiceStmt:
		refs = append(refs, exprVarRefs(s.Timeout)...)
	case *ast.DownloadFileStmt:
		refs = append(refs, s.FileDocument)
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
		refs = append(refs, exprVarRefs(s.Value)...)
		if s.Item != "" {
			refs = append(refs, s.Item)
		}
		refs = append(refs, s.List)
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

// newErrorHandlerFlow creates a SequenceFlow with IsErrorHandler=true.
// Studio Pro connects error-handler flows bottom-to-top; using a horizontal
// destination can produce invalid sequence flows after roundtrip.
func newErrorHandlerFlow(originID, destinationID model.ID) *microflows.SequenceFlow {
	return &microflows.SequenceFlow{
		BaseElement:                model.BaseElement{ID: model.ID(types.GenerateID())},
		OriginID:                   originID,
		DestinationID:              destinationID,
		OriginConnectionIndex:      AnchorBottom,
		DestinationConnectionIndex: AnchorTop,
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
	var sourceAnchor *ast.FlowAnchors
	if fb.pendingAnnotations != nil && fb.pendingAnnotations.Anchor != nil &&
		fb.pendingAnnotations.Anchor.From == ast.AnchorSideTop {
		sourceAnchor = fb.pendingAnnotations.Anchor
	}

	// Position error handler below the main flow
	errorY := fb.posY + VerticalSpacing
	errorX := sourceX

	// Build error handler activities
	errBuilder := &flowBuilder{
		posX:                   errorX,
		posY:                   errorY,
		baseY:                  errorY,
		spacing:                HorizontalSpacing,
		returnType:             fb.returnType,
		hasReturnValue:         fb.hasReturnValue,
		varTypes:               fb.varTypes,
		declaredVars:           fb.declaredVars,
		measurer:               fb.measurer,
		backend:                fb.backend,
		hierarchy:              fb.hierarchy,
		restServices:           fb.restServices,
		returnScopeBaseline:    copyVarTypes(fb.varTypes),
		callOutputDeclarations: fb.callOutputDeclarations,
	}

	var lastErrID model.ID
	for i, stmt := range errorBody {
		thisAnchor := stmtOwnAnchor(stmt)
		actID := errBuilder.addStatement(stmt)
		if actID != "" {
			if lastErrID == "" {
				// Connect source activity to first error handler activity
				flow := newErrorHandlerFlow(sourceActivityID, actID)
				applyUserAnchors(flow, sourceAnchor, thisAnchor)
				if thisAnchor == nil && errBuilder.isShowMessageActivity(actID) {
					flow.DestinationConnectionIndex = AnchorBottom
				}
				fb.flows = append(fb.flows, flow)
			} else {
				flow := newHorizontalFlow(lastErrID, actID)
				if errBuilder.isShowMessageActivity(lastErrID) && errBuilder.isEndEvent(actID) {
					flow.OriginConnectionIndex = AnchorTop
					flow.DestinationConnectionIndex = AnchorBottom
				}
				errBuilder.flows = append(errBuilder.flows, flow)
				errBuilder.addPendingErrorHandlerFlowForStatement(lastErrID, actID, stmt, statementsReferenceVar(errorBody[i+1:], errBuilder.errorHandlerSkipVar))
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

func (fb *flowBuilder) isShowMessageActivity(activityID model.ID) bool {
	for _, obj := range fb.objects {
		activity, ok := obj.(*microflows.ActionActivity)
		if !ok || activity.ID != activityID {
			continue
		}
		_, ok = activity.Action.(*microflows.ShowMessageAction)
		return ok
	}
	return false
}

func (fb *flowBuilder) isEndEvent(activityID model.ID) bool {
	for _, obj := range fb.objects {
		end, ok := obj.(*microflows.EndEvent)
		if ok && end.ID == activityID {
			return true
		}
	}
	return false
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
	if fb.manualLoopBackTarget != "" {
		fb.flows = append(fb.flows, newHorizontalFlow(lastErrID, fb.manualLoopBackTarget))
		return
	}
	if !fb.hasReturnValue {
		fb.terminateErrorHandlerFlow(lastErrID, errorY)
		return
	}
	_ = errorY
	fb.errorHandlerSource = activityID
	fb.errorHandlerTailFrom = lastErrID
	fb.errorHandlerSkipVar = skipVar
	fb.errorHandlerTailIsSource = false
	fb.errorHandlerReturnValue = fb.inferReturnValueFromScopeExcluding(fb.returnType, skipVar)
}

func (fb *flowBuilder) terminateErrorHandlerFlow(lastErrID model.ID, errorY int) {
	end := &microflows.EndEvent{
		BaseMicroflowObject: microflows.BaseMicroflowObject{
			BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
			Position:    model.Point{X: fb.posX + HorizontalSpacing/2, Y: errorY},
			Size:        model.Size{Width: EventSize, Height: EventSize},
		},
	}
	fb.objects = append(fb.objects, end)
	fb.flows = append(fb.flows, newHorizontalFlow(lastErrID, end.ID))
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

func branchDestinationAnchor(branchAnchor, stmtAnchor *ast.FlowAnchors) *ast.FlowAnchors {
	if branchAnchor != nil && branchAnchor.To != ast.AnchorSideUnset {
		return branchAnchor
	}
	return stmtAnchor
}

func pendingFlowAnchors(previousAnchor, pendingAnchor, stmtAnchor *ast.FlowAnchors) (*ast.FlowAnchors, *ast.FlowAnchors) {
	if pendingAnchor == nil {
		return previousAnchor, stmtAnchor
	}
	return pendingAnchor, branchDestinationAnchor(pendingAnchor, stmtAnchor)
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
	case *ast.WhileStmt:
		return isManualWhileTrueCandidate(s)
	default:
		return false
	}
}

func containsTerminalStmt(stmts []ast.MicroflowStatement) bool {
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.ReturnStmt, *ast.RaiseErrorStmt:
			return true
		case *ast.IfStmt:
			if containsTerminalStmt(s.ThenBody) || containsTerminalStmt(s.ElseBody) {
				return true
			}
		case *ast.InheritanceSplitStmt:
			if containsTerminalStmt(s.ElseBody) {
				return true
			}
			for _, c := range s.Cases {
				if containsTerminalStmt(c.Body) {
					return true
				}
			}
		case *ast.LoopStmt:
			if containsTerminalStmt(s.Body) {
				return true
			}
		case *ast.WhileStmt:
			if containsTerminalStmt(s.Body) {
				return true
			}
		}
	}
	return false
}

func containsBreakStmt(stmts []ast.MicroflowStatement) bool {
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.BreakStmt:
			return true
		case *ast.IfStmt:
			if containsBreakStmt(s.ThenBody) || containsBreakStmt(s.ElseBody) {
				return true
			}
		case *ast.InheritanceSplitStmt:
			if containsBreakStmt(s.ElseBody) {
				return true
			}
			for _, c := range s.Cases {
				if containsBreakStmt(c.Body) {
					return true
				}
			}
		case *ast.LoopStmt:
			if containsBreakStmt(s.Body) {
				return true
			}
		case *ast.WhileStmt:
			if containsBreakStmt(s.Body) {
				return true
			}
		}
	}
	return false
}
