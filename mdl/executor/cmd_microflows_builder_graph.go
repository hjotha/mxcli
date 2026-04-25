// SPDX-License-Identifier: Apache-2.0

// Package executor - Microflow flow graph: graph construction and statement dispatch
package executor

import (
	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/types"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/microflows"
)

// buildFlowGraph converts AST statements to a Microflow flow graph.
// Note: posY represents the CENTER LINE for element alignment, not the top position.
func (fb *flowBuilder) buildFlowGraph(stmts []ast.MicroflowStatement, returns *ast.MicroflowReturnType) *microflows.MicroflowObjectCollection {
	// Initialize maps if not set
	if fb.measurer == nil {
		fb.measurer = &layoutMeasurer{varTypes: fb.varTypes}
	}
	if fb.declaredVars == nil {
		fb.declaredVars = make(map[string]string)
	}
	if fb.callOutputRemaining == nil {
		fb.callOutputRemaining = countMicroflowCallOutputs(stmts)
	}
	if fb.listInputVariables == nil {
		fb.listInputVariables = collectListInputVariables(stmts)
	}
	if fb.objectInputVariables == nil {
		fb.objectInputVariables = collectObjectInputVariables(stmts)
	}
	// Set return value expression for error handler EndEvents
	if returns != nil && returns.Variable != "" {
		fb.returnValue = "$" + returns.Variable
	}
	fb.hasReturnValue = returns != nil && returns.Type.Kind != ast.TypeVoid
	// Set baseY for branch restoration (this is the center line)
	fb.baseY = fb.posY

	// Pre-scan: if the first statement carries an @position annotation, shift the
	// StartEvent to be one spacing unit to the left of that position so it doesn't
	// end up behind activities that use explicit coordinates.
	for _, stmt := range stmts {
		if ann := getStatementAnnotations(stmt); ann != nil && ann.Position != nil {
			fb.posX = ann.Position.X - fb.spacing
			fb.posY = ann.Position.Y
			fb.baseY = fb.posY
			break
		}
	}

	// Create StartEvent - Position is the CENTER point (RelativeMiddlePoint in Mendix)
	startEvent := &microflows.StartEvent{
		BaseMicroflowObject: microflows.BaseMicroflowObject{
			BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
			Position:    model.Point{X: fb.posX, Y: fb.posY},
			Size:        model.Size{Width: EventSize, Height: EventSize},
		},
	}
	fb.objects = append(fb.objects, startEvent)
	lastID := startEvent.ID

	fb.posX += fb.spacing

	// Process each statement
	// pendingCase holds the case value for the NEXT flow (set by merge-less splits)
	// pendingFlowAnchor carries branch anchors from a guard-pattern IF so the
	// deferred split→nextActivity flow honours @anchor(true: ..., false: ...).
	pendingCase := ""
	var pendingFlowAnchor *ast.FlowAnchors
	for i, stmt := range stmts {
		// Snapshot the current statement's anchor annotation before addStatement
		// can reset pendingAnnotations via recursive processing. The incoming
		// side (To) is applied when this statement is the destination of the
		// flow we're about to create; the outgoing side (From) is stashed in
		// previousStmtAnchor so the NEXT iteration can apply it.
		stmtAnchor := stmtOwnAnchor(stmt)

		activityID := fb.addStatement(stmt)
		if activityID != "" {
			// If there are pending annotations, apply them to this activity.
			fb.applyPendingAnnotations(activityID)
			// Connect to previous object with horizontal SequenceFlow
			var flow *microflows.SequenceFlow
			if pendingCase != "" {
				flow = newHorizontalFlowWithCase(lastID, activityID, pendingCase)
				pendingCase = ""
			} else {
				flow = newHorizontalFlow(lastID, activityID)
			}
			// Prefer the pendingFlowAnchor (carried from a guard-pattern IF's
			// branch) over the previous statement's own anchor — it encodes
			// exactly the @anchor(true/false: ...) the user asked for on the
			// deferred flow. When the pending anchor is present it applies to
			// both From (origin side on the split) and To (the side of the
			// continuing activity), unless the incoming statement explicitly
			// overrides its own To.
			originAnchor := fb.previousStmtAnchor
			destAnchor := stmtAnchor
			if pendingFlowAnchor != nil {
				originAnchor = pendingFlowAnchor
				if destAnchor == nil || destAnchor.To == ast.AnchorSideUnset {
					destAnchor = pendingFlowAnchor
				}
				pendingFlowAnchor = nil
			}
			applyUserAnchors(flow, originAnchor, destAnchor)
			fb.flows = append(fb.flows, flow)
			fb.addPendingErrorHandlerFlowForStatement(lastID, activityID, stmt, statementsReferenceVar(stmts[i+1:], fb.errorHandlerSkipVar))
			fb.previousStmtAnchor = stmtAnchor

			// For compound statements (IF, LOOP), the exit point differs from entry point
			if fb.nextConnectionPoint != "" {
				lastID = fb.nextConnectionPoint
				fb.nextConnectionPoint = ""
				// Save nextFlowCase / nextFlowAnchor for the NEXT iteration's flow creation
				pendingCase = fb.nextFlowCase
				fb.nextFlowCase = ""
				pendingFlowAnchor = fb.nextFlowAnchor
				fb.nextFlowAnchor = nil
				// Compound statements control their own internal anchors; don't
				// let the outer From leak into the flow leaving the merge.
				fb.previousStmtAnchor = nil
			} else {
				lastID = activityID
			}
		}
	}

	// Handle leftover pending annotations (free-floating annotation text)
	if fb.pendingAnnotations != nil {
		if fb.pendingAnnotations.AnnotationText != "" {
			fb.attachFreeAnnotation(fb.pendingAnnotations.AnnotationText)
		}
		fb.pendingAnnotations = nil
	}

	// Create EndEvent only if the flow doesn't already end with RETURN EndEvent(s)
	// (e.g., when both branches of an IF/ELSE end with RETURN, EndEvents are already created)
	if !fb.endsWithReturn {
		if fb.returnValue == "" {
			fb.returnValue = fb.inferReturnValueFromScope(returns)
		}
		fb.posX += fb.spacing / 2
		fb.posY = fb.baseY // Ensure end event is on the happy path center line
		endEvent := &microflows.EndEvent{
			BaseMicroflowObject: microflows.BaseMicroflowObject{
				BaseElement: model.BaseElement{ID: model.ID(types.GenerateID())},
				Position:    model.Point{X: fb.posX, Y: fb.posY},
				Size:        model.Size{Width: EventSize, Height: EventSize},
			},
			ReturnValue: fb.returnValue,
		}

		fb.objects = append(fb.objects, endEvent)

		// Connect last activity to end event
		var endFlow *microflows.SequenceFlow
		if pendingCase != "" {
			endFlow = newHorizontalFlowWithCase(lastID, endEvent.ID, pendingCase)
		} else {
			endFlow = newHorizontalFlow(lastID, endEvent.ID)
		}
		originAnchor := fb.previousStmtAnchor
		if pendingFlowAnchor != nil {
			originAnchor = pendingFlowAnchor
			pendingFlowAnchor = nil
		}
		applyUserAnchors(endFlow, originAnchor, nil)
		fb.flows = append(fb.flows, endFlow)
		fb.addPendingErrorHandlerFlowForStatement(lastID, endEvent.ID, nil)
		fb.previousStmtAnchor = nil
	}

	return &microflows.MicroflowObjectCollection{
		BaseElement:     model.BaseElement{ID: model.ID(types.GenerateID())},
		Objects:         fb.objects,
		Flows:           fb.flows,
		AnnotationFlows: fb.annotationFlows,
	}
}

func collectListInputVariables(stmts []ast.MicroflowStatement) map[string]bool {
	inputs := make(map[string]bool)
	var walk func([]ast.MicroflowStatement)
	walk = func(body []ast.MicroflowStatement) {
		for _, stmt := range body {
			switch s := stmt.(type) {
			case *ast.ListOperationStmt:
				if s.InputVariable != "" {
					inputs[s.InputVariable] = true
				}
			case *ast.AggregateListStmt:
				if s.InputVariable != "" {
					inputs[s.InputVariable] = true
				}
			case *ast.IfStmt:
				walk(s.ThenBody)
				walk(s.ElseBody)
			case *ast.LoopStmt:
				if s.ListVariable != "" {
					inputs[s.ListVariable] = true
				}
				walk(s.Body)
			case *ast.WhileStmt:
				walk(s.Body)
			case *ast.CallMicroflowStmt:
				if s.ErrorHandling != nil {
					walk(s.ErrorHandling.Body)
				}
			case *ast.CreateObjectStmt:
				if s.ErrorHandling != nil {
					walk(s.ErrorHandling.Body)
				}
			case *ast.MfCommitStmt:
				if s.ErrorHandling != nil {
					walk(s.ErrorHandling.Body)
				}
			case *ast.DeleteObjectStmt:
				if s.ErrorHandling != nil {
					walk(s.ErrorHandling.Body)
				}
			case *ast.RestCallStmt:
				if s.ErrorHandling != nil {
					walk(s.ErrorHandling.Body)
				}
			case *ast.CallJavaActionStmt:
				if s.ErrorHandling != nil {
					walk(s.ErrorHandling.Body)
				}
			case *ast.CallWebServiceStmt:
				if s.ErrorHandling != nil {
					walk(s.ErrorHandling.Body)
				}
			case *ast.CallExternalActionStmt:
				if s.ErrorHandling != nil {
					walk(s.ErrorHandling.Body)
				}
			case *ast.ImportFromMappingStmt:
				if s.ErrorHandling != nil {
					walk(s.ErrorHandling.Body)
				}
			}
		}
	}
	walk(stmts)
	return inputs
}

func collectObjectInputVariables(stmts []ast.MicroflowStatement) map[string]bool {
	inputs := make(map[string]bool)
	var walkExpr func(ast.Expression)
	walkExpr = func(expr ast.Expression) {
		switch e := expr.(type) {
		case *ast.AttributePathExpr:
			if e.Variable != "" {
				inputs[e.Variable] = true
			}
		case *ast.BinaryExpr:
			walkExpr(e.Left)
			walkExpr(e.Right)
		case *ast.UnaryExpr:
			walkExpr(e.Operand)
		case *ast.FunctionCallExpr:
			for _, arg := range e.Arguments {
				walkExpr(arg)
			}
		case *ast.ParenExpr:
			walkExpr(e.Inner)
		case *ast.IfThenElseExpr:
			walkExpr(e.Condition)
			walkExpr(e.ThenExpr)
			walkExpr(e.ElseExpr)
		case *ast.SourceExpr:
			walkExpr(e.Expression)
			for _, ref := range sourceAttributeVarRefs(e.Source) {
				inputs[ref] = true
			}
		}
	}
	var walk func([]ast.MicroflowStatement)
	walk = func(body []ast.MicroflowStatement) {
		for _, stmt := range body {
			switch s := stmt.(type) {
			case *ast.IfStmt:
				walkExpr(s.Condition)
				walk(s.ThenBody)
				walk(s.ElseBody)
			case *ast.WhileStmt:
				walkExpr(s.Condition)
				walk(s.Body)
			case *ast.LoopStmt:
				walk(s.Body)
			case *ast.ChangeObjectStmt:
				inputs[s.Variable] = true
				for _, change := range s.Changes {
					walkExpr(change.Value)
				}
			case *ast.CreateObjectStmt:
				for _, change := range s.Changes {
					walkExpr(change.Value)
				}
				if s.ErrorHandling != nil {
					walk(s.ErrorHandling.Body)
				}
			case *ast.LogStmt:
				walkExpr(s.Node)
				walkExpr(s.Message)
				for _, param := range s.Template {
					walkExpr(param.Value)
				}
			case *ast.CallMicroflowStmt:
				for _, arg := range s.Arguments {
					walkExpr(arg.Value)
				}
				if s.ErrorHandling != nil {
					walk(s.ErrorHandling.Body)
				}
			case *ast.CallJavaActionStmt:
				for _, arg := range s.Arguments {
					walkExpr(arg.Value)
				}
				if s.ErrorHandling != nil {
					walk(s.ErrorHandling.Body)
				}
			case *ast.CallWebServiceStmt:
				walkExpr(s.Timeout)
				if s.ErrorHandling != nil {
					walk(s.ErrorHandling.Body)
				}
			case *ast.RestCallStmt:
				walkExpr(s.URL)
				for _, param := range s.URLParams {
					walkExpr(param.Value)
				}
				for _, header := range s.Headers {
					walkExpr(header.Value)
				}
				if s.Body != nil {
					walkExpr(s.Body.Template)
					for _, param := range s.Body.TemplateParams {
						walkExpr(param.Value)
					}
				}
				walkExpr(s.Timeout)
				if s.ErrorHandling != nil {
					walk(s.ErrorHandling.Body)
				}
			case *ast.ReturnStmt:
				walkExpr(s.Value)
			case *ast.ImportFromMappingStmt:
				if s.ErrorHandling != nil {
					walk(s.ErrorHandling.Body)
				}
			}
		}
	}
	walk(stmts)
	return inputs
}

func sourceAttributeVarRefs(source string) []string {
	var refs []string
	for i := 0; i < len(source); i++ {
		if source[i] != '$' {
			continue
		}
		j := i + 1
		for j < len(source) {
			c := source[j]
			if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' {
				j++
				continue
			}
			break
		}
		if j > i+1 && j < len(source) && source[j] == '/' {
			refs = append(refs, source[i+1:j])
		}
		i = j
	}
	return refs
}

func countMicroflowCallOutputs(stmts []ast.MicroflowStatement) map[string]int {
	counts := make(map[string]int)
	var walk func([]ast.MicroflowStatement)
	walk = func(body []ast.MicroflowStatement) {
		for _, stmt := range body {
			switch s := stmt.(type) {
			case *ast.CallMicroflowStmt:
				if s.OutputVariable != "" {
					counts[s.OutputVariable]++
				}
				if s.ErrorHandling != nil {
					walk(s.ErrorHandling.Body)
				}
			case *ast.IfStmt:
				walk(s.ThenBody)
				walk(s.ElseBody)
			case *ast.LoopStmt:
				walk(s.Body)
			case *ast.WhileStmt:
				walk(s.Body)
			case *ast.CreateObjectStmt:
				if s.ErrorHandling != nil {
					walk(s.ErrorHandling.Body)
				}
			case *ast.MfCommitStmt:
				if s.ErrorHandling != nil {
					walk(s.ErrorHandling.Body)
				}
			case *ast.DeleteObjectStmt:
				if s.ErrorHandling != nil {
					walk(s.ErrorHandling.Body)
				}
			case *ast.RestCallStmt:
				if s.ErrorHandling != nil {
					walk(s.ErrorHandling.Body)
				}
			case *ast.CallJavaActionStmt:
				if s.ErrorHandling != nil {
					walk(s.ErrorHandling.Body)
				}
			case *ast.CallWebServiceStmt:
				if s.ErrorHandling != nil {
					walk(s.ErrorHandling.Body)
				}
			case *ast.CallExternalActionStmt:
				if s.ErrorHandling != nil {
					walk(s.ErrorHandling.Body)
				}
			case *ast.ImportFromMappingStmt:
				if s.ErrorHandling != nil {
					walk(s.ErrorHandling.Body)
				}
			}
		}
	}
	walk(stmts)
	return counts
}

func (fb *flowBuilder) inferReturnValueFromScope(returns *ast.MicroflowReturnType) string {
	if returns == nil || returns.Variable != "" {
		return ""
	}
	target := ""
	switch returns.Type.Kind {
	case ast.TypeEntity:
		if returns.Type.EntityRef != nil {
			target = returns.Type.EntityRef.Module + "." + returns.Type.EntityRef.Name
		}
	case ast.TypeListOf:
		if returns.Type.EntityRef != nil {
			target = "List of " + returns.Type.EntityRef.Module + "." + returns.Type.EntityRef.Name
		}
	default:
		return ""
	}
	if target == "" || fb.varTypes == nil {
		return ""
	}
	var candidate string
	for name, typ := range fb.varTypes {
		if typ != target {
			continue
		}
		if candidate != "" {
			return ""
		}
		candidate = name
	}
	if candidate == "" {
		return ""
	}
	return "$" + candidate
}

// addStatement converts an AST statement to a microflow activity and returns its ID.
func (fb *flowBuilder) addStatement(stmt ast.MicroflowStatement) model.ID {
	// Extract annotations from the statement and merge into pendingAnnotations
	fb.mergeStatementAnnotations(stmt)

	// Apply @position before creating the activity so it's placed at the right position
	if fb.pendingAnnotations != nil && fb.pendingAnnotations.Position != nil {
		fb.posX = fb.pendingAnnotations.Position.X
		fb.posY = fb.pendingAnnotations.Position.Y
	}

	switch s := stmt.(type) {
	case *ast.DeclareStmt:
		return fb.addCreateVariableAction(s)
	case *ast.InheritanceSplitStmt:
		return fb.addInheritanceSplit(s)
	case *ast.CastObjectStmt:
		return fb.addCastAction(s)
	case *ast.MfSetStmt:
		return fb.addChangeVariableAction(s)
	case *ast.ReturnStmt:
		return fb.addEndEventWithReturn(s)
	case *ast.RaiseErrorStmt:
		return fb.addErrorEvent()
	case *ast.LogStmt:
		return fb.addLogMessageAction(s)
	case *ast.CreateObjectStmt:
		return fb.addCreateObjectAction(s)
	case *ast.ChangeObjectStmt:
		return fb.addChangeObjectAction(s)
	case *ast.RetrieveStmt:
		return fb.addRetrieveAction(s)
	case *ast.MfCommitStmt:
		return fb.addCommitAction(s)
	case *ast.DeleteObjectStmt:
		return fb.addDeleteAction(s)
	case *ast.RollbackStmt:
		return fb.addRollbackAction(s)
	case *ast.IfStmt:
		return fb.addIfStatement(s)
	case *ast.LoopStmt:
		return fb.addLoopStatement(s)
	case *ast.WhileStmt:
		return fb.addWhileStatement(s)
	case *ast.ListOperationStmt:
		return fb.addListOperationAction(s)
	case *ast.AggregateListStmt:
		return fb.addAggregateListAction(s)
	case *ast.CreateListStmt:
		return fb.addCreateListAction(s)
	case *ast.AddToListStmt:
		return fb.addAddToListAction(s)
	case *ast.RemoveFromListStmt:
		return fb.addRemoveFromListAction(s)
	case *ast.CallMicroflowStmt:
		return fb.addCallMicroflowAction(s)
	case *ast.CallJavaActionStmt:
		return fb.addCallJavaActionAction(s)
	case *ast.CallWebServiceStmt:
		return fb.addCallWebServiceAction(s)
	case *ast.ExecuteDatabaseQueryStmt:
		return fb.addExecuteDatabaseQueryAction(s)
	case *ast.CallExternalActionStmt:
		return fb.addCallExternalActionAction(s)
	case *ast.ShowPageStmt:
		return fb.addShowPageAction(s)
	case *ast.ClosePageStmt:
		return fb.addClosePageAction(s)
	case *ast.ShowHomePageStmt:
		return fb.addShowHomePageAction(s)
	case *ast.ShowMessageStmt:
		return fb.addShowMessageAction(s)
	case *ast.DownloadFileStmt:
		return fb.addDownloadFileAction(s)
	case *ast.ValidationFeedbackStmt:
		return fb.addValidationFeedbackAction(s)
	case *ast.RestCallStmt:
		return fb.addRestCallAction(s)
	case *ast.SendRestRequestStmt:
		return fb.addSendRestRequestAction(s)
	case *ast.ImportFromMappingStmt:
		return fb.addImportFromMappingAction(s)
	case *ast.ExportToMappingStmt:
		return fb.addExportToMappingAction(s)
	case *ast.TransformJsonStmt:
		return fb.addTransformJsonAction(s)
	case *ast.BreakStmt:
		return fb.addBreakEvent()
	case *ast.ContinueStmt:
		return fb.addContinueEvent()
	// Workflow microflow actions
	case *ast.CallWorkflowStmt:
		return fb.addCallWorkflowAction(s)
	case *ast.GetWorkflowDataStmt:
		return fb.addGetWorkflowDataAction(s)
	case *ast.GetWorkflowsStmt:
		return fb.addGetWorkflowsAction(s)
	case *ast.GetWorkflowActivityRecordsStmt:
		return fb.addGetWorkflowActivityRecordsAction(s)
	case *ast.WorkflowOperationStmt:
		return fb.addWorkflowOperationAction(s)
	case *ast.SetTaskOutcomeStmt:
		return fb.addSetTaskOutcomeAction(s)
	case *ast.OpenUserTaskStmt:
		return fb.addOpenUserTaskAction(s)
	case *ast.NotifyWorkflowStmt:
		return fb.addNotifyWorkflowAction(s)
	case *ast.OpenWorkflowStmt:
		return fb.addOpenWorkflowAction(s)
	case *ast.LockWorkflowStmt:
		return fb.addLockWorkflowAction(s)
	case *ast.UnlockWorkflowStmt:
		return fb.addUnlockWorkflowAction(s)
	default:
		// For now, skip unknown statement types
		return ""
	}
}
