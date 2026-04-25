// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"strings"
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/microflows"
)

func TestLastStmtIsReturn_EmptyBody(t *testing.T) {
	if lastStmtIsReturn(nil) {
		t.Error("empty body must not be terminal")
	}
}

func TestLastStmtIsReturn_PlainReturn(t *testing.T) {
	body := []ast.MicroflowStatement{&ast.ReturnStmt{}}
	if !lastStmtIsReturn(body) {
		t.Error("body ending in ReturnStmt must be terminal")
	}
}

func TestLastStmtIsReturn_RaiseError(t *testing.T) {
	body := []ast.MicroflowStatement{&ast.RaiseErrorStmt{}}
	if !lastStmtIsReturn(body) {
		t.Error("body ending in RaiseErrorStmt must be terminal")
	}
}

func TestLastStmtIsReturn_BreakAndContinue(t *testing.T) {
	for _, stmt := range []ast.MicroflowStatement{&ast.BreakStmt{}, &ast.ContinueStmt{}} {
		if !lastStmtIsReturn([]ast.MicroflowStatement{stmt}) {
			t.Errorf("body ending in %T must be terminal", stmt)
		}
	}
}

func TestLastStmtIsReturn_IfWithoutElse_NotTerminal(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.IfStmt{ThenBody: []ast.MicroflowStatement{&ast.ReturnStmt{}}},
	}
	if lastStmtIsReturn(body) {
		t.Error("IF without ELSE must not be terminal (false path falls through)")
	}
}

func TestLastStmtIsReturn_IfBothBranchesReturn_Terminal(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.IfStmt{
			ThenBody: []ast.MicroflowStatement{&ast.ReturnStmt{}},
			ElseBody: []ast.MicroflowStatement{&ast.ReturnStmt{}},
		},
	}
	if !lastStmtIsReturn(body) {
		t.Error("IF/ELSE where both branches return must be terminal")
	}
}

func TestLastStmtIsReturn_IfOnlyThenReturns_NotTerminal(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.IfStmt{
			ThenBody: []ast.MicroflowStatement{&ast.ReturnStmt{}},
			ElseBody: []ast.MicroflowStatement{&ast.LogStmt{}}, // non-terminal
		},
	}
	if lastStmtIsReturn(body) {
		t.Error("IF/ELSE where only THEN terminates must not be terminal")
	}
}

func TestLastStmtIsReturn_NestedIfChain_Terminal(t *testing.T) {
	// if { return } else if { return } else { return }
	inner := &ast.IfStmt{
		ThenBody: []ast.MicroflowStatement{&ast.ReturnStmt{}},
		ElseBody: []ast.MicroflowStatement{&ast.ReturnStmt{}},
	}
	outer := &ast.IfStmt{
		ThenBody: []ast.MicroflowStatement{&ast.ReturnStmt{}},
		ElseBody: []ast.MicroflowStatement{inner},
	}
	if !lastStmtIsReturn([]ast.MicroflowStatement{outer}) {
		t.Error("else-if chain where every terminal branch returns must be terminal")
	}
}

func TestLastStmtIsReturn_RaiseErrorMixed_Terminal(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.IfStmt{
			ThenBody: []ast.MicroflowStatement{&ast.ReturnStmt{}},
			ElseBody: []ast.MicroflowStatement{&ast.RaiseErrorStmt{}},
		},
	}
	if !lastStmtIsReturn(body) {
		t.Error("IF/ELSE with return on one side and raise error on the other must be terminal")
	}
}

func TestLastStmtIsReturn_IfBreakContinueBranches_Terminal(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.IfStmt{
			ThenBody: []ast.MicroflowStatement{&ast.ContinueStmt{}},
			ElseBody: []ast.MicroflowStatement{&ast.BreakStmt{}},
		},
	}
	if !lastStmtIsReturn(body) {
		t.Error("IF/ELSE with continue on one side and break on the other must be terminal")
	}
}

func TestLastStmtIsReturn_LoopNotTerminal(t *testing.T) {
	// A LOOP whose body only returns is still non-terminal — BREAK can exit.
	body := []ast.MicroflowStatement{
		&ast.LoopStmt{Body: []ast.MicroflowStatement{&ast.ReturnStmt{}}},
	}
	if lastStmtIsReturn(body) {
		t.Error("LOOP must never be terminal (BREAK path)")
	}
}

func TestBuildFlowGraph_LoopIfPreservesBreakAndContinue(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.LoopStmt{
			LoopVariable: "Item",
			ListVariable: "Items",
			Body: []ast.MicroflowStatement{
				&ast.IfStmt{
					Condition: &ast.VariableExpr{Name: "Changed"},
					ThenBody:  []ast.MicroflowStatement{&ast.ContinueStmt{}},
					ElseBody:  []ast.MicroflowStatement{&ast.BreakStmt{}},
				},
			},
		},
	}

	fb := &flowBuilder{
		posX:         100,
		posY:         100,
		spacing:      HorizontalSpacing,
		varTypes:     map[string]string{"Items": "List of MyModule.Item"},
		declaredVars: map[string]string{"Changed": "Boolean"},
		measurer:     &layoutMeasurer{},
	}
	oc := fb.buildFlowGraph(body, nil)

	var loop *microflows.LoopedActivity
	for _, obj := range oc.Objects {
		if l, ok := obj.(*microflows.LoopedActivity); ok {
			loop = l
			break
		}
	}
	if loop == nil || loop.ObjectCollection == nil {
		t.Fatal("expected loop with object collection")
	}

	var hasBreak, hasContinue, hasMerge bool
	for _, obj := range loop.ObjectCollection.Objects {
		switch obj.(type) {
		case *microflows.BreakEvent:
			hasBreak = true
		case *microflows.ContinueEvent:
			hasContinue = true
		case *microflows.ExclusiveMerge:
			hasMerge = true
		}
	}
	if !hasBreak || !hasContinue {
		t.Fatalf("expected break and continue events in loop body, got break=%v continue=%v", hasBreak, hasContinue)
	}
	if hasMerge {
		t.Fatal("break/continue branches must not be connected through an ExclusiveMerge")
	}
}

func TestBuildFlowGraph_NestedGuardIfPreservesFallthroughCase(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.IfStmt{
			Condition: &ast.VariableExpr{Name: "Outer"},
			ThenBody: []ast.MicroflowStatement{
				&ast.IfStmt{
					Condition: &ast.VariableExpr{Name: "Success"},
					ThenBody:  []ast.MicroflowStatement{&ast.ReturnStmt{Value: &ast.LiteralExpr{Kind: ast.LiteralBoolean, Value: true}}},
				},
				&ast.LogStmt{Level: ast.LogInfo, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "fallthrough"}},
			},
			ElseBody: []ast.MicroflowStatement{&ast.ReturnStmt{Value: &ast.LiteralExpr{Kind: ast.LiteralBoolean, Value: false}}},
		},
	}

	fb := &flowBuilder{
		posX:         100,
		posY:         100,
		spacing:      HorizontalSpacing,
		declaredVars: map[string]string{"Outer": "Boolean", "Success": "Boolean"},
		measurer:     &layoutMeasurer{},
	}
	oc := fb.buildFlowGraph(body, &ast.MicroflowReturnType{Type: ast.DataType{Kind: ast.TypeBoolean}})

	var found bool
	for _, flow := range oc.Flows {
		if flow.CaseValue == nil {
			continue
		}
		if enumCase, ok := flow.CaseValue.(microflows.EnumerationCase); ok && enumCase.Value == "false" {
			if _, ok := findMicroflowObjectByID(oc.Objects, flow.DestinationID).(*microflows.ActionActivity); ok {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatal("expected nested guard IF false branch to connect to the following action")
	}
}

func TestBuildFlowGraph_EmptyThenElseReturnNestedGuardKeepsTrueCase(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.IfStmt{
			Condition: &ast.VariableExpr{Name: "Support"},
			ElseBody: []ast.MicroflowStatement{
				&ast.IfStmt{
					Condition: &ast.VariableExpr{Name: "UserHasAdminRole"},
					ElseBody:  []ast.MicroflowStatement{&ast.ReturnStmt{}},
				},
			},
		},
		&ast.LogStmt{Level: ast.LogInfo, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "shared tail"}},
	}

	fb := &flowBuilder{
		posX:         100,
		posY:         100,
		spacing:      HorizontalSpacing,
		declaredVars: map[string]string{"Support": "Boolean", "UserHasAdminRole": "Boolean"},
		measurer:     &layoutMeasurer{},
	}
	oc := fb.buildFlowGraph(body, nil)

	var nestedSplitID model.ID
	for _, obj := range oc.Objects {
		split, ok := obj.(*microflows.ExclusiveSplit)
		if !ok {
			continue
		}
		cond, ok := split.SplitCondition.(*microflows.ExpressionSplitCondition)
		if ok && cond.Expression == "$UserHasAdminRole" {
			nestedSplitID = split.ID
			break
		}
	}
	if nestedSplitID == "" {
		t.Fatal("expected nested authorization split")
	}

	for _, flow := range oc.Flows {
		if flow.OriginID != nestedSplitID {
			continue
		}
		if enumCase, ok := flow.CaseValue.(microflows.EnumerationCase); ok && enumCase.Value == "true" {
			if _, ok := findMicroflowObjectByID(oc.Objects, flow.DestinationID).(*microflows.ExclusiveMerge); ok {
				return
			}
		}
	}
	t.Fatal("nested empty-then guard must connect its true branch to the parent merge with CaseValue=true")
}

func TestBuildFlowGraph_ManualWhileTrueContinueUsesBackEdgeMerge(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.WhileStmt{
			Condition: &ast.LiteralExpr{Kind: ast.LiteralBoolean, Value: true},
			Body: []ast.MicroflowStatement{
				&ast.IfStmt{
					Condition: &ast.VariableExpr{Name: "HasMore"},
					ThenBody: []ast.MicroflowStatement{
						&ast.LogStmt{Level: ast.LogInfo, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "retry"}},
						&ast.ContinueStmt{},
					},
					ElseBody: []ast.MicroflowStatement{
						&ast.LogStmt{Level: ast.LogInfo, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "done"}},
					},
				},
			},
		},
	}

	fb := &flowBuilder{
		posX:         100,
		posY:         100,
		spacing:      HorizontalSpacing,
		declaredVars: map[string]string{"HasMore": "Boolean"},
		measurer:     &layoutMeasurer{},
	}
	oc := fb.buildFlowGraph(body, nil)

	var merge *microflows.ExclusiveMerge
	var split *microflows.ExclusiveSplit
	for _, obj := range oc.Objects {
		switch o := obj.(type) {
		case *microflows.LoopedActivity:
			t.Fatal("manual while true with continue must not be rebuilt as LoopedActivity")
		case *microflows.ContinueEvent:
			t.Fatal("manual while true back-edge must not emit ContinueEvent outside a LoopedActivity")
		case *microflows.ExclusiveMerge:
			merge = o
		case *microflows.ExclusiveSplit:
			split = o
		}
	}
	if merge == nil {
		t.Fatal("expected manual loop header ExclusiveMerge")
	}
	if split == nil {
		t.Fatal("expected if split inside manual loop")
	}

	var falseFlows, backEdges int
	for _, flow := range oc.Flows {
		if flow.DestinationID == merge.ID {
			backEdges++
		}
		if flow.OriginID == split.ID {
			if enumCase, ok := flow.CaseValue.(microflows.EnumerationCase); ok && enumCase.Value == "false" {
				falseFlows++
			}
		}
	}
	if backEdges == 0 {
		t.Fatal("expected continue branch to connect back to the manual-loop merge")
	}
	if falseFlows != 1 {
		t.Fatalf("expected exactly one false flow out of split, got %d", falseFlows)
	}
}

func TestBuildFlowGraph_LoopTrailingGuardContinueKeepsFalseCase(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.LoopStmt{
			LoopVariable: "Issue",
			ListVariable: "Issues",
			Body: []ast.MicroflowStatement{
				&ast.IfStmt{
					Condition: &ast.VariableExpr{Name: "Retryable"},
					ThenBody: []ast.MicroflowStatement{
						&ast.ContinueStmt{},
					},
				},
			},
		},
	}

	fb := &flowBuilder{
		posX:     100,
		posY:     100,
		spacing:  HorizontalSpacing,
		varTypes: map[string]string{"Issues": "List of DataSyncIssues.Issue"},
		measurer: &layoutMeasurer{},
	}
	oc := fb.buildFlowGraph(body, nil)

	var split *microflows.ExclusiveSplit
	for _, obj := range oc.Objects {
		switch o := obj.(type) {
		case *microflows.ExclusiveSplit:
			split = o
		case *microflows.LoopedActivity:
			for _, nested := range o.ObjectCollection.Objects {
				if s, ok := nested.(*microflows.ExclusiveSplit); ok {
					split = s
					break
				}
			}
		}
	}
	if split == nil {
		t.Fatal("expected split")
	}
	for _, flow := range oc.Flows {
		if flow.OriginID != split.ID {
			continue
		}
		if enumCase, ok := flow.CaseValue.(microflows.EnumerationCase); ok && enumCase.Value == "false" {
			return
		}
	}
	t.Fatal("trailing guard IF inside loop must keep a configured false flow")
}

func TestBuildFlowGraph_ManualWhileTrueTerminalDoesNotAddFallthroughEnd(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.WhileStmt{
			Condition: &ast.LiteralExpr{Kind: ast.LiteralBoolean, Value: true},
			Body: []ast.MicroflowStatement{
				&ast.IfStmt{
					Condition: &ast.VariableExpr{Name: "Done"},
					ThenBody:  []ast.MicroflowStatement{&ast.ReturnStmt{Value: &ast.LiteralExpr{Kind: ast.LiteralBoolean, Value: true}}},
				},
				&ast.ContinueStmt{},
			},
		},
	}

	fb := &flowBuilder{
		posX:         100,
		posY:         100,
		spacing:      HorizontalSpacing,
		declaredVars: map[string]string{"Done": "Boolean"},
		measurer:     &layoutMeasurer{},
	}
	oc := fb.buildFlowGraph(body, &ast.MicroflowReturnType{Type: ast.DataType{Kind: ast.TypeBoolean}})

	for _, obj := range oc.Objects {
		end, ok := obj.(*microflows.EndEvent)
		if !ok {
			continue
		}
		if end.ReturnValue == "" {
			t.Fatal("manual while true ending in continue must not add a fallthrough EndEvent without return value")
		}
	}
}

func TestBuildFlowGraph_ManualWhileTrueReturnUsesBackEdgeMerge(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.WhileStmt{
			Condition: &ast.LiteralExpr{Kind: ast.LiteralBoolean, Value: true},
			Body: []ast.MicroflowStatement{
				&ast.IfStmt{
					Condition: &ast.VariableExpr{Name: "Retry"},
					ThenBody: []ast.MicroflowStatement{
						&ast.LogStmt{Level: ast.LogInfo, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "retry"}},
					},
					ElseBody: []ast.MicroflowStatement{
						&ast.ReturnStmt{Value: &ast.VariableExpr{Name: "Response"}},
					},
				},
				&ast.ReturnStmt{Value: &ast.VariableExpr{Name: "Response"}},
			},
		},
	}

	fb := &flowBuilder{
		posX:         100,
		posY:         100,
		spacing:      HorizontalSpacing,
		declaredVars: map[string]string{"Retry": "Boolean"},
		measurer:     &layoutMeasurer{},
	}
	oc := fb.buildFlowGraph(body, &ast.MicroflowReturnType{Type: ast.DataType{Kind: ast.TypeUnknown}})

	for _, obj := range oc.Objects {
		switch obj.(type) {
		case *microflows.LoopedActivity:
			t.Fatal("terminal while true must not be rebuilt as LoopedActivity")
		case *microflows.ExclusiveMerge:
			return
		}
	}
	t.Fatal("expected manual loop header ExclusiveMerge")
}

func TestBuildFlowGraph_ManualWhileTrueCustomErrorHandlerBacksToLoopMerge(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.WhileStmt{
			Condition: &ast.LiteralExpr{Kind: ast.LiteralBoolean, Value: true},
			Body: []ast.MicroflowStatement{
				&ast.CallMicroflowStmt{
					OutputVariable: "Response",
					MicroflowName:  ast.QualifiedName{Module: "SampleRuntime", Name: "REST_GetRuntimeChangeEvents"},
					ErrorHandling: &ast.ErrorHandlingClause{
						Type: ast.ErrorHandlingCustomWithoutRollback,
						Body: []ast.MicroflowStatement{
							&ast.IfStmt{
								Condition: &ast.VariableExpr{Name: "Retry"},
								ThenBody: []ast.MicroflowStatement{
									&ast.LogStmt{Level: ast.LogInfo, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "retry"}},
								},
								ElseBody: []ast.MicroflowStatement{
									&ast.RaiseErrorStmt{},
								},
							},
						},
					},
				},
				&ast.ReturnStmt{Value: &ast.VariableExpr{Name: "Response"}},
			},
		},
	}

	fb := &flowBuilder{
		posX:         100,
		posY:         100,
		spacing:      HorizontalSpacing,
		declaredVars: map[string]string{"Retry": "Boolean"},
		measurer:     &layoutMeasurer{},
	}
	oc := fb.buildFlowGraph(body, &ast.MicroflowReturnType{Type: ast.DataType{Kind: ast.TypeUnknown}})

	var merge *microflows.ExclusiveMerge
	for _, obj := range oc.Objects {
		if m, ok := obj.(*microflows.ExclusiveMerge); ok {
			merge = m
		}
		if end, ok := obj.(*microflows.EndEvent); ok && end.ReturnValue == "" {
			t.Fatal("non-terminal custom error handler inside manual while true must not create an empty EndEvent")
		}
	}
	if merge == nil {
		t.Fatal("expected manual loop header ExclusiveMerge")
	}

	for _, flow := range oc.Flows {
		if flow.DestinationID == merge.ID && !flow.IsErrorHandler {
			return
		}
	}
	t.Fatal("expected non-terminal custom error handler path to reconnect to the manual-loop merge")
}

func TestBuildFlowGraph_CustomErrorHandlerContinuesToNextStatement(t *testing.T) {
	entityRef := ast.QualifiedName{Module: "System", Name: "HttpResponse"}
	body := []ast.MicroflowStatement{
		&ast.CallMicroflowStmt{
			OutputVariable: "Response",
			MicroflowName:  ast.QualifiedName{Module: "SampleAudit", Name: "REST_Post"},
			ErrorHandling: &ast.ErrorHandlingClause{
				Type: ast.ErrorHandlingCustomWithoutRollback,
				Body: []ast.MicroflowStatement{
					&ast.LogStmt{Level: ast.LogError, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "failed"}},
				},
			},
		},
		&ast.LogStmt{Level: ast.LogDebug, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "after"}},
		&ast.ReturnStmt{Value: &ast.VariableExpr{Name: "Response"}},
	}

	fb := &flowBuilder{
		posX:     100,
		posY:     100,
		spacing:  HorizontalSpacing,
		varTypes: map[string]string{"Response": "System.HttpResponse"},
		measurer: &layoutMeasurer{},
	}
	oc := fb.buildFlowGraph(body, &ast.MicroflowReturnType{Type: ast.DataType{Kind: ast.TypeEntity, EntityRef: &entityRef}})

	for _, obj := range oc.Objects {
		end, ok := obj.(*microflows.EndEvent)
		if ok && end.ReturnValue == "" {
			var returnValues []string
			for _, candidate := range oc.Objects {
				if candidateEnd, ok := candidate.(*microflows.EndEvent); ok {
					returnValues = append(returnValues, candidateEnd.ReturnValue)
				}
			}
			t.Fatalf("non-terminal custom error handler must not create an empty EndEvent; returns=%v", returnValues)
		}
	}
}

func TestBuildFlowGraph_VoidCustomErrorHandlerTerminates(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.CallMicroflowStmt{
			MicroflowName: ast.QualifiedName{Module: "SampleApps", Name: "CommitApplicationViewDataChanges"},
			ErrorHandling: &ast.ErrorHandlingClause{
				Type: ast.ErrorHandlingCustomWithoutRollback,
				Body: []ast.MicroflowStatement{
					&ast.ChangeObjectStmt{
						Variable: "AppProcessingResult",
						Changes:  []ast.ChangeItem{{Attribute: "IsSuccessful", Value: &ast.LiteralExpr{Kind: ast.LiteralBoolean, Value: false}}},
					},
				},
			},
		},
		&ast.ChangeObjectStmt{
			Variable: "AppProcessingResult",
			Changes:  []ast.ChangeItem{{Attribute: "IsSuccessful", Value: &ast.LiteralExpr{Kind: ast.LiteralBoolean, Value: true}}},
		},
	}

	fb := &flowBuilder{posX: 100, posY: 100, spacing: HorizontalSpacing, measurer: &layoutMeasurer{}}
	oc := fb.buildFlowGraph(body, nil)

	var errorChangeID, successChangeID, errorEndID model.ID
	incomingError := map[model.ID]bool{}
	for _, flow := range oc.Flows {
		if flow.IsErrorHandler {
			incomingError[flow.DestinationID] = true
		}
	}
	for _, obj := range oc.Objects {
		activity, ok := obj.(*microflows.ActionActivity)
		if !ok {
			continue
		}
		change, ok := activity.Action.(*microflows.ChangeObjectAction)
		if !ok || change.ChangeVariable != "AppProcessingResult" {
			continue
		}
		if incomingError[activity.ID] {
			errorChangeID = activity.ID
		} else {
			successChangeID = activity.ID
		}
	}
	endIDs := map[model.ID]bool{}
	for _, obj := range oc.Objects {
		if _, ok := obj.(*microflows.EndEvent); ok {
			endIDs[obj.GetID()] = true
		}
	}
	for _, flow := range oc.Flows {
		if flow.OriginID == errorChangeID {
			if endIDs[flow.DestinationID] {
				errorEndID = flow.DestinationID
			}
			if flow.DestinationID == successChangeID {
				t.Fatal("void custom error handler must terminate instead of falling through to the success continuation")
			}
		}
	}
	if errorChangeID == "" || successChangeID == "" || errorEndID == "" {
		t.Fatalf("expected error change, success change, and terminal error end; error=%q success=%q end=%q", errorChangeID, successChangeID, errorEndID)
	}
}

func TestBuildFlowGraph_CustomErrorHandlerRejoinsBeforeOutputDependentContinuation(t *testing.T) {
	entityRef := ast.QualifiedName{Module: "System", Name: "HttpResponse"}
	body := []ast.MicroflowStatement{
		&ast.RestCallStmt{
			OutputVariable: "Response",
			Method:         ast.HttpMethodPost,
			URL:            &ast.LiteralExpr{Kind: ast.LiteralString, Value: "https://example.test"},
			Result:         ast.RestResult{Type: ast.RestResultResponse},
			ErrorHandling: &ast.ErrorHandlingClause{
				Type: ast.ErrorHandlingCustomWithoutRollback,
				Body: []ast.MicroflowStatement{
					&ast.LogStmt{Level: ast.LogError, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "failed"}},
				},
			},
		},
		&ast.LogStmt{
			Level:   ast.LogDebug,
			Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "response {1}"},
			Template: []ast.TemplateParam{
				{Index: 1, Value: &ast.SourceExpr{Source: "$Response/Content"}},
			},
		},
		&ast.ReturnStmt{Value: &ast.VariableExpr{Name: "latestHttpResponse"}},
	}

	fb := &flowBuilder{
		posX:     100,
		posY:     100,
		spacing:  HorizontalSpacing,
		varTypes: map[string]string{"latestHttpResponse": "System.HttpResponse"},
		measurer: &layoutMeasurer{},
	}
	oc := fb.buildFlowGraph(body, &ast.MicroflowReturnType{Type: ast.DataType{Kind: ast.TypeEntity, EntityRef: &entityRef}})

	var errorLogID, debugLogID, returnID model.ID
	for _, obj := range oc.Objects {
		switch o := obj.(type) {
		case *microflows.ActionActivity:
			logAction, ok := o.Action.(*microflows.LogMessageAction)
			if !ok {
				continue
			}
			switch logAction.LogLevel {
			case "Error":
				errorLogID = o.ID
			case "Debug":
				debugLogID = o.ID
			}
		case *microflows.EndEvent:
			if strings.TrimSpace(o.ReturnValue) == "$latestHttpResponse" {
				returnID = o.ID
			}
		}
	}
	if errorLogID == "" || debugLogID == "" || returnID == "" {
		t.Fatalf("expected error log, debug log, and return nodes; got error=%q debug=%q return=%q", errorLogID, debugLogID, returnID)
	}

	errorLogHandled := false
	for _, flow := range oc.Flows {
		if flow.OriginID == errorLogID && flow.DestinationID == debugLogID {
			t.Fatal("custom error handler must rejoin through a merge before the next continuation")
		}
		if flow.OriginID == errorLogID {
			errorLogHandled = true
		}
	}
	if errorLogHandled {
		return
	}
	var flowDescriptions []string
	for _, flow := range oc.Flows {
		flowDescriptions = append(flowDescriptions, string(flow.OriginID)+"->"+string(flow.DestinationID))
	}
	t.Fatalf("expected custom error handler to have a safe outgoing continuation; error=%q debug=%q return=%q flows=%v", errorLogID, debugLogID, returnID, flowDescriptions)
}

func TestBuildFlowGraph_CustomErrorHandlerTerminatesWhenFinalReturnUsesFailedOutput(t *testing.T) {
	entityRef := ast.QualifiedName{Module: "SampleApi", Name: "ResponseRoot"}
	body := []ast.MicroflowStatement{
		&ast.RestCallStmt{
			OutputVariable: "SuccessResponse",
			Method:         ast.HttpMethodGet,
			URL:            &ast.LiteralExpr{Kind: ast.LiteralString, Value: "https://example.test"},
			Result: ast.RestResult{
				Type:         ast.RestResultMapping,
				MappingName:  ast.QualifiedName{Module: "SampleApi", Name: "ImportResponse"},
				ResultEntity: entityRef,
			},
			ErrorHandling: &ast.ErrorHandlingClause{
				Type: ast.ErrorHandlingCustom,
				Body: []ast.MicroflowStatement{
					&ast.CreateObjectStmt{
						Variable:   "ErrorResponse",
						EntityType: entityRef,
					},
					&ast.ChangeObjectStmt{
						Variable: "ErrorResponse",
						Changes:  []ast.ChangeItem{{Attribute: "Message", Value: &ast.VariableExpr{Name: "latestError"}}},
					},
				},
			},
		},
		&ast.ReturnStmt{Value: &ast.VariableExpr{Name: "SuccessResponse"}},
	}

	fb := &flowBuilder{
		posX:     100,
		posY:     100,
		spacing:  HorizontalSpacing,
		varTypes: map[string]string{},
		measurer: &layoutMeasurer{},
	}
	oc := fb.buildFlowGraph(body, &ast.MicroflowReturnType{Type: ast.DataType{Kind: ast.TypeEntity, EntityRef: &entityRef}})

	returnValues := map[string]bool{}
	for _, obj := range oc.Objects {
		end, ok := obj.(*microflows.EndEvent)
		if ok {
			returnValues[strings.TrimSpace(end.ReturnValue)] = true
		}
	}
	if !returnValues["$SuccessResponse"] || !returnValues["$ErrorResponse"] {
		t.Fatalf("expected separate success and error return values, got %#v", returnValues)
	}
}

func TestBuildFlowGraph_NestedCustomErrorHandlerInsideErrorBodyKeepsTailFlow(t *testing.T) {
	entityRef := ast.QualifiedName{Module: "SampleApi", Name: "ResponseRoot"}
	body := []ast.MicroflowStatement{
		&ast.RestCallStmt{
			OutputVariable: "SuccessResponse",
			Method:         ast.HttpMethodGet,
			URL:            &ast.LiteralExpr{Kind: ast.LiteralString, Value: "https://example.test"},
			Result: ast.RestResult{
				Type:         ast.RestResultMapping,
				MappingName:  ast.QualifiedName{Module: "SampleApi", Name: "ImportSuccess"},
				ResultEntity: entityRef,
			},
			ErrorHandling: &ast.ErrorHandlingClause{
				Type: ast.ErrorHandlingCustom,
				Body: []ast.MicroflowStatement{
					&ast.CreateObjectStmt{
						Variable:   "ErrorResponse",
						EntityType: entityRef,
					},
					&ast.ImportFromMappingStmt{
						OutputVariable: "ParsedError",
						Mapping:        ast.QualifiedName{Module: "SampleApi", Name: "ImportError"},
						SourceVariable: "HttpResponseContent",
						ErrorHandling: &ast.ErrorHandlingClause{
							Type: ast.ErrorHandlingCustom,
							Body: []ast.MicroflowStatement{
								&ast.LogStmt{Level: ast.LogError, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "import failed"}},
								&ast.ChangeObjectStmt{
									Variable: "ErrorResponse",
									Changes:  []ast.ChangeItem{{Attribute: "Message", Value: &ast.VariableExpr{Name: "latestError"}}},
								},
							},
						},
					},
					&ast.LogStmt{Level: ast.LogError, Message: &ast.SourceExpr{Source: "$ParsedError/Message"}},
					&ast.ChangeObjectStmt{
						Variable: "ErrorResponse",
						Changes:  []ast.ChangeItem{{Attribute: "Message", Value: &ast.SourceExpr{Source: "$ParsedError/Message"}}},
					},
				},
			},
		},
		&ast.ReturnStmt{Value: &ast.VariableExpr{Name: "SuccessResponse"}},
	}

	fb := &flowBuilder{
		posX:     100,
		posY:     100,
		spacing:  HorizontalSpacing,
		varTypes: map[string]string{},
		measurer: &layoutMeasurer{},
	}
	oc := fb.buildFlowGraph(body, &ast.MicroflowReturnType{Type: ast.DataType{Kind: ast.TypeEntity, EntityRef: &entityRef}})

	outgoing := map[model.ID]bool{}
	returnValues := map[string]bool{}
	for _, flow := range oc.Flows {
		outgoing[flow.OriginID] = true
	}
	for _, obj := range oc.Objects {
		switch o := obj.(type) {
		case *microflows.ActionActivity:
			change, ok := o.Action.(*microflows.ChangeObjectAction)
			if ok && change.ChangeVariable == "ErrorResponse" && !outgoing[o.ID] {
				t.Fatalf("nested custom error handler change %q must not be left as a dangling tail", o.ID)
			}
		case *microflows.EndEvent:
			returnValues[strings.TrimSpace(o.ReturnValue)] = true
		}
	}
	if !returnValues["$ErrorResponse"] {
		t.Fatalf("nested handler should return the handler-local fallback object when its own output is unavailable, got %#v", returnValues)
	}
}

func TestBuildFlowGraph_EmptyCustomErrorHandlerSkipsOutputDependentContinuation(t *testing.T) {
	entityRef := ast.QualifiedName{Module: "System", Name: "HttpResponse"}
	body := []ast.MicroflowStatement{
		&ast.RestCallStmt{
			OutputVariable: "Response",
			Method:         ast.HttpMethodPost,
			URL:            &ast.LiteralExpr{Kind: ast.LiteralString, Value: "https://example.test"},
			Result:         ast.RestResult{Type: ast.RestResultResponse},
			ErrorHandling:  &ast.ErrorHandlingClause{Type: ast.ErrorHandlingCustomWithoutRollback},
		},
		&ast.LogStmt{
			Level:   ast.LogDebug,
			Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "response {1}"},
			Template: []ast.TemplateParam{
				{Index: 1, Value: &ast.SourceExpr{Source: "$Response/Content"}},
			},
		},
		&ast.ReturnStmt{Value: &ast.VariableExpr{Name: "latestHttpResponse"}},
	}

	fb := &flowBuilder{
		posX:     100,
		posY:     100,
		spacing:  HorizontalSpacing,
		varTypes: map[string]string{"latestHttpResponse": "System.HttpResponse"},
		measurer: &layoutMeasurer{},
	}
	oc := fb.buildFlowGraph(body, &ast.MicroflowReturnType{Type: ast.DataType{Kind: ast.TypeEntity, EntityRef: &entityRef}})

	var restID, debugLogID model.ID
	for _, obj := range oc.Objects {
		activity, ok := obj.(*microflows.ActionActivity)
		if !ok {
			continue
		}
		switch action := activity.Action.(type) {
		case *microflows.RestCallAction:
			restID = activity.ID
			if action.ErrorHandlingType != microflows.ErrorHandlingTypeCustomWithoutRollback {
				t.Fatalf("empty custom handler type = %q, want CustomWithoutRollBack", action.ErrorHandlingType)
			}
		case *microflows.LogMessageAction:
			if action.LogLevel == "Debug" {
				debugLogID = activity.ID
			}
		}
	}
	if restID == "" || debugLogID == "" {
		t.Fatalf("expected rest and debug log nodes; got rest=%q debug=%q", restID, debugLogID)
	}
	for _, flow := range oc.Flows {
		if flow.IsErrorHandler && flow.OriginID == restID && flow.DestinationID == debugLogID {
			t.Fatal("empty custom error handler must not flow into output-dependent debug statement")
		}
	}
}

func TestBuildFlowGraph_VoidEmptyOutputHandlerTerminatesBeforeOutputDependentTail(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.CallJavaActionStmt{
			OutputVariable: "ProcessedCount",
			ActionName:     ast.QualifiedName{Module: "SampleMigration", Name: "CountProcessedItems"},
			ErrorHandling:  &ast.ErrorHandlingClause{Type: ast.ErrorHandlingCustomWithoutRollback},
		},
		&ast.LogStmt{
			Level:   ast.LogInfo,
			Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "processed {1}"},
			Template: []ast.TemplateParam{{
				Index: 1,
				Value: &ast.VariableExpr{Name: "ProcessedCount"},
			}},
		},
	}

	fb := &flowBuilder{
		posX:     100,
		posY:     100,
		spacing:  HorizontalSpacing,
		measurer: &layoutMeasurer{},
	}
	oc := fb.buildFlowGraph(body, nil)

	var javaID, logID model.ID
	endIDs := map[model.ID]bool{}
	for _, obj := range oc.Objects {
		switch o := obj.(type) {
		case *microflows.ActionActivity:
			switch action := o.Action.(type) {
			case *microflows.JavaActionCallAction:
				if action.ResultVariableName == "ProcessedCount" {
					javaID = o.ID
				}
			case *microflows.LogMessageAction:
				if action.MessageTemplate != nil && action.MessageTemplate.Translations["en_US"] == "processed {1}" {
					logID = o.ID
				}
			}
		case *microflows.EndEvent:
			endIDs[o.ID] = true
		}
	}
	if javaID == "" || logID == "" || len(endIDs) == 0 {
		t.Fatalf("expected java action, output-dependent log, and end event; got java=%q log=%q ends=%v", javaID, logID, endIDs)
	}

	var errorFlowTerminates bool
	for _, flow := range oc.Flows {
		if !flow.IsErrorHandler || flow.OriginID != javaID {
			continue
		}
		if flowPathExists(oc.Flows, flow.DestinationID, logID) {
			t.Fatal("empty output handler in a void microflow must not rejoin before a statement that reads the missing output")
		}
		for endID := range endIDs {
			if flow.DestinationID == endID || flowPathExists(oc.Flows, flow.DestinationID, endID) {
				errorFlowTerminates = true
			}
		}
	}
	if !errorFlowTerminates {
		t.Fatal("empty output handler should terminate at an EndEvent before the output-dependent tail")
	}
}

func TestBuildFlowGraph_EmptyHandlerBeforeOutputHandlerRejoinsAtNextAction(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.CallMicroflowStmt{
			MicroflowName: ast.QualifiedName{Module: "SampleMigration", Name: "RefreshCache"},
			ErrorHandling: &ast.ErrorHandlingClause{Type: ast.ErrorHandlingCustomWithoutRollback},
		},
		&ast.CallJavaActionStmt{
			OutputVariable: "ProcessedCount",
			ActionName:     ast.QualifiedName{Module: "SampleMigration", Name: "CountProcessedItems"},
			ErrorHandling:  &ast.ErrorHandlingClause{Type: ast.ErrorHandlingCustomWithoutRollback},
		},
		&ast.LogStmt{
			Level:   ast.LogInfo,
			Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "processed {1}"},
			Template: []ast.TemplateParam{{
				Index: 1,
				Value: &ast.VariableExpr{Name: "ProcessedCount"},
			}},
		},
	}

	fb := &flowBuilder{
		posX:     100,
		posY:     100,
		spacing:  HorizontalSpacing,
		measurer: &layoutMeasurer{},
	}
	oc := fb.buildFlowGraph(body, nil)

	var callID, javaID model.ID
	for _, obj := range oc.Objects {
		activity, ok := obj.(*microflows.ActionActivity)
		if !ok {
			continue
		}
		switch action := activity.Action.(type) {
		case *microflows.MicroflowCallAction:
			if action.MicroflowCall != nil && action.MicroflowCall.Microflow == "SampleMigration.RefreshCache" {
				callID = activity.ID
			}
		case *microflows.JavaActionCallAction:
			if action.ResultVariableName == "ProcessedCount" {
				javaID = activity.ID
			}
		}
	}
	if callID == "" || javaID == "" {
		t.Fatalf("expected no-output call and output-producing java action; got call=%q java=%q", callID, javaID)
	}

	for _, flow := range oc.Flows {
		if flow.IsErrorHandler && flow.OriginID == callID && flowPathExists(oc.Flows, flow.DestinationID, javaID) {
			return
		}
	}
	t.Fatal("empty no-output handler should rejoin at the next action, even when that action has its own output error handler")
}

func TestBuildFlowGraph_EmptyCustomErrorHandlerRejoinsThroughMerge(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.CallMicroflowStmt{
			MicroflowName: ast.QualifiedName{Module: "SampleBranding", Name: "ResizeCropImageIfNecessary"},
			ErrorHandling: &ast.ErrorHandlingClause{
				Type: ast.ErrorHandlingCustomWithoutRollback,
			},
		},
		&ast.CallJavaActionStmt{
			OutputVariable: "Base64EncodedImage",
			ActionName:     ast.QualifiedName{Module: "FileHandling", Name: "Base64EncodeFile"},
		},
		&ast.CreateObjectStmt{
			Variable:   "NewBrand",
			EntityType: ast.QualifiedName{Module: "SampleBrandApi", Name: "UpdateBrandRequest"},
		},
	}

	fb := &flowBuilder{
		posX:     100,
		posY:     100,
		spacing:  HorizontalSpacing,
		measurer: &layoutMeasurer{},
	}
	oc := fb.buildFlowGraph(body, nil)

	var callID, javaID, mergeID model.ID
	for _, obj := range oc.Objects {
		switch o := obj.(type) {
		case *microflows.ActionActivity:
			switch o.Action.(type) {
			case *microflows.MicroflowCallAction:
				callID = o.ID
			case *microflows.JavaActionCallAction:
				javaID = o.ID
			}
		case *microflows.ExclusiveMerge:
			mergeID = o.ID
		}
	}
	if callID == "" || javaID == "" || mergeID == "" {
		t.Fatalf("expected call, java action, and merge; got call=%q java=%q merge=%q", callID, javaID, mergeID)
	}

	var normalToMerge, errorToMerge, mergeToJava bool
	for _, flow := range oc.Flows {
		if flow.OriginID == callID && flow.DestinationID == javaID {
			t.Fatal("empty custom handler must not create parallel normal/error flows directly to the next activity")
		}
		if flow.OriginID == callID && flow.DestinationID == mergeID && !flow.IsErrorHandler {
			normalToMerge = true
		}
		if flow.OriginID == callID && flow.DestinationID == mergeID && flow.IsErrorHandler {
			errorToMerge = true
		}
		if flow.OriginID == mergeID && flow.DestinationID == javaID {
			mergeToJava = true
		}
	}
	if !normalToMerge || !errorToMerge || !mergeToJava {
		t.Fatalf("expected normal/error call flows to merge and merge to java; normal=%v error=%v mergeToJava=%v", normalToMerge, errorToMerge, mergeToJava)
	}
}

func TestBuildFlowGraph_ConsecutiveEmptyCustomHandlersKeepErrorFlows(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.CallMicroflowStmt{
			MicroflowName: ast.QualifiedName{Module: "SampleMigration", Name: "DeleteAllData"},
			ErrorHandling: &ast.ErrorHandlingClause{
				Type: ast.ErrorHandlingCustomWithoutRollback,
			},
		},
		&ast.CallJavaActionStmt{
			OutputVariable: "ProcessedCount",
			ActionName:     ast.QualifiedName{Module: "SampleMigration", Name: "ProcessRows"},
			ErrorHandling: &ast.ErrorHandlingClause{
				Type: ast.ErrorHandlingCustomWithoutRollback,
			},
		},
		&ast.LogStmt{Level: ast.LogInfo, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "done"}},
	}

	fb := &flowBuilder{
		posX:     100,
		posY:     100,
		spacing:  HorizontalSpacing,
		measurer: &layoutMeasurer{},
	}
	oc := fb.buildFlowGraph(body, nil)

	var callID, javaID model.ID
	for _, obj := range oc.Objects {
		activity, ok := obj.(*microflows.ActionActivity)
		if !ok {
			continue
		}
		switch activity.Action.(type) {
		case *microflows.MicroflowCallAction:
			callID = activity.ID
		case *microflows.JavaActionCallAction:
			javaID = activity.ID
		}
	}
	if callID == "" || javaID == "" {
		t.Fatalf("expected microflow call and java action; got call=%q java=%q", callID, javaID)
	}

	hasErrorFlow := map[model.ID]bool{}
	for _, flow := range oc.Flows {
		if flow.IsErrorHandler {
			hasErrorFlow[flow.OriginID] = true
		}
	}
	if !hasErrorFlow[callID] || !hasErrorFlow[javaID] {
		t.Fatalf("consecutive empty custom handlers must both keep error flows; call=%v java=%v", hasErrorFlow[callID], hasErrorFlow[javaID])
	}
}

func TestBuildFlowGraph_EmptyCustomErrorHandlerTerminatesBeforeInheritanceSplitUsingOutput(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.CallMicroflowStmt{
			OutputVariable: "AccessToken",
			MicroflowName:  ast.QualifiedName{Module: "Synthetic", Name: "FetchToken"},
			ErrorHandling:  &ast.ErrorHandlingClause{Type: ast.ErrorHandlingCustomWithoutRollback},
		},
		&ast.InheritanceSplitStmt{
			Variable: "AccessToken",
			Cases: []ast.InheritanceSplitCase{
				{
					Entity: ast.QualifiedName{Module: "Synthetic", Name: "DecryptedToken"},
					Body: []ast.MicroflowStatement{
						&ast.MfSetStmt{
							Target: "TokenValue",
							Value:  &ast.SourceExpr{Source: "$AccessToken/Value"},
						},
					},
				},
			},
			ElseBody: []ast.MicroflowStatement{
				&ast.ReturnStmt{Value: &ast.LiteralExpr{Kind: ast.LiteralEmpty, Value: "empty"}},
			},
		},
		&ast.LogStmt{
			Level:   ast.LogInfo,
			Node:    &ast.LiteralExpr{Kind: ast.LiteralString, Value: "App"},
			Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "shared tail"},
		},
	}

	fb := &flowBuilder{
		posX:     100,
		posY:     100,
		spacing:  HorizontalSpacing,
		measurer: &layoutMeasurer{},
	}
	oc := fb.buildFlowGraph(body, nil)

	var callID, splitID model.ID
	endIDs := map[model.ID]bool{}
	for _, obj := range oc.Objects {
		switch o := obj.(type) {
		case *microflows.ActionActivity:
			if _, ok := o.Action.(*microflows.MicroflowCallAction); ok {
				callID = o.ID
			}
		case *microflows.InheritanceSplit:
			splitID = o.ID
		case *microflows.EndEvent:
			endIDs[o.ID] = true
		}
	}
	if callID == "" || splitID == "" || len(endIDs) == 0 {
		t.Fatalf("expected call, inheritance split, and end nodes; got call=%q split=%q ends=%v", callID, splitID, endIDs)
	}

	var errorFlowTerminates bool
	for _, flow := range oc.Flows {
		if !flow.IsErrorHandler || flow.OriginID != callID {
			continue
		}
		if flowPathExists(oc.Flows, flow.DestinationID, splitID) {
			t.Fatal("empty custom error handler must not reach an inheritance split that reads the missing output")
		}
		for endID := range endIDs {
			if flow.DestinationID == endID || flowPathExists(oc.Flows, flow.DestinationID, endID) {
				errorFlowTerminates = true
			}
		}
	}
	if !errorFlowTerminates {
		t.Fatal("empty custom error handler should terminate before the output-dependent inheritance split")
	}
}

func TestBuildFlowGraph_EmptyCustomErrorHandlerPreservesDecisionCases(t *testing.T) {
	entityRef := ast.QualifiedName{Module: "SampleAudit", Name: "ApiToken"}
	body := []ast.MicroflowStatement{
		&ast.RestCallStmt{
			OutputVariable: "NewApiToken",
			Method:         ast.HttpMethodPost,
			URL:            &ast.LiteralExpr{Kind: ast.LiteralString, Value: "https://example.test"},
			Result: ast.RestResult{
				Type:         ast.RestResultMapping,
				MappingName:  ast.QualifiedName{Module: "SampleAudit", Name: "IMP_ApiToken"},
				ResultEntity: entityRef,
			},
			ErrorHandling: &ast.ErrorHandlingClause{Type: ast.ErrorHandlingCustomWithoutRollback},
		},
		&ast.IfStmt{
			Condition: &ast.BinaryExpr{
				Left:     &ast.SourceExpr{Source: "$latestHttpResponse/StatusCode"},
				Operator: "=",
				Right:    &ast.LiteralExpr{Kind: ast.LiteralInteger, Value: "200"},
			},
			ThenBody: []ast.MicroflowStatement{
				&ast.ChangeObjectStmt{
					Variable: "NewApiToken",
					Changes:  []ast.ChangeItem{{Attribute: "StatusCode", Value: &ast.SourceExpr{Source: "$latestHttpResponse/StatusCode"}}},
				},
				&ast.ReturnStmt{Value: &ast.VariableExpr{Name: "NewApiToken"}},
			},
			ElseBody: []ast.MicroflowStatement{
				&ast.ReturnStmt{Value: &ast.LiteralExpr{Kind: ast.LiteralEmpty, Value: "empty"}},
			},
		},
	}

	fb := &flowBuilder{
		posX:     100,
		posY:     100,
		spacing:  HorizontalSpacing,
		varTypes: map[string]string{"latestHttpResponse": "System.HttpResponse"},
		measurer: &layoutMeasurer{},
	}
	oc := fb.buildFlowGraph(body, &ast.MicroflowReturnType{Type: ast.DataType{Kind: ast.TypeEntity, EntityRef: &entityRef}})

	var successSplitID model.ID
	for _, obj := range oc.Objects {
		split, ok := obj.(*microflows.ExclusiveSplit)
		if !ok {
			continue
		}
		if split.Caption == "$latestHttpResponse/StatusCode = 200" {
			successSplitID = split.ID
			break
		}
	}
	if successSplitID == "" {
		t.Fatal("expected success decision split")
	}

	cases := map[string]int{}
	for _, flow := range oc.Flows {
		if flow.OriginID != successSplitID {
			continue
		}
		if enumCase, ok := flow.CaseValue.(microflows.EnumerationCase); ok {
			cases[enumCase.Value]++
		}
	}
	if cases["true"] != 1 || cases["false"] != 1 {
		t.Fatalf("success decision cases = %#v, want one true and one false", cases)
	}
}

func TestBuildFlowGraph_EmptyCustomErrorHandlerRoutesOutputConditionToElse(t *testing.T) {
	entityRef := ast.QualifiedName{Module: "SampleAudit", Name: "ApiToken"}
	body := []ast.MicroflowStatement{
		&ast.CallMicroflowStmt{
			OutputVariable: "ApiToken",
			MicroflowName:  ast.QualifiedName{Module: "SampleAudit", Name: "GetValidAccessApiToken"},
			ErrorHandling:  &ast.ErrorHandlingClause{Type: ast.ErrorHandlingCustomWithoutRollback},
		},
		&ast.IfStmt{
			Condition: &ast.BinaryExpr{
				Left:     &ast.VariableExpr{Name: "ApiToken"},
				Operator: "!=",
				Right:    &ast.LiteralExpr{Kind: ast.LiteralEmpty, Value: "empty"},
			},
			ThenBody: []ast.MicroflowStatement{
				&ast.ReturnStmt{Value: &ast.VariableExpr{Name: "ApiToken"}},
			},
			ElseBody: []ast.MicroflowStatement{
				&ast.ReturnStmt{Value: &ast.LiteralExpr{Kind: ast.LiteralEmpty, Value: "empty"}},
			},
		},
	}

	fb := &flowBuilder{
		posX:     100,
		posY:     100,
		spacing:  HorizontalSpacing,
		varTypes: map[string]string{"ApiToken": "SampleAudit.ApiToken"},
		measurer: &layoutMeasurer{},
	}
	oc := fb.buildFlowGraph(body, &ast.MicroflowReturnType{Type: ast.DataType{Kind: ast.TypeEntity, EntityRef: &entityRef}})

	var callID model.ID
	for _, obj := range oc.Objects {
		activity, ok := obj.(*microflows.ActionActivity)
		if !ok {
			continue
		}
		if _, ok := activity.Action.(*microflows.MicroflowCallAction); ok {
			callID = activity.ID
			break
		}
	}
	if callID == "" {
		t.Fatal("expected microflow call")
	}
	for _, flow := range oc.Flows {
		if flow.OriginID == callID && flow.IsErrorHandler {
			return
		}
	}
	t.Fatal("empty custom handler on output-producing call must keep an outgoing error-handler flow")
}

func TestBuildFlowGraph_CustomErrorHandlerWaitsPastFutureOutputReferences(t *testing.T) {
	entityRef := ast.QualifiedName{Module: "SampleRuntime", Name: "Runtime"}
	body := []ast.MicroflowStatement{
		&ast.RestCallStmt{
			OutputVariable: "Runtime",
			Method:         ast.HttpMethodGet,
			URL:            &ast.LiteralExpr{Kind: ast.LiteralString, Value: "https://example.test"},
			Result: ast.RestResult{
				Type:         ast.RestResultMapping,
				MappingName:  ast.QualifiedName{Module: "SampleRuntime", Name: "IMM_RuntimeByUUID"},
				ResultEntity: entityRef,
			},
			ErrorHandling: &ast.ErrorHandlingClause{
				Type: ast.ErrorHandlingCustomWithoutRollback,
				Body: []ast.MicroflowStatement{
					&ast.LogStmt{Level: ast.LogError, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "failed"}},
				},
			},
		},
		&ast.LogStmt{Level: ast.LogInfo, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "success"}},
		&ast.ChangeObjectStmt{
			Variable: "Response",
			Changes:  []ast.ChangeItem{{Attribute: "RuntimeResponse_Runtime", Value: &ast.VariableExpr{Name: "Runtime"}}},
		},
		&ast.ReturnStmt{Value: &ast.VariableExpr{Name: "Response"}},
	}

	fb := &flowBuilder{
		posX:     100,
		posY:     100,
		spacing:  HorizontalSpacing,
		varTypes: map[string]string{"Response": "SampleRuntime.RuntimeResponse"},
		measurer: &layoutMeasurer{},
	}
	oc := fb.buildFlowGraph(body, &ast.MicroflowReturnType{Type: ast.DataType{Kind: ast.TypeEntity, EntityRef: &ast.QualifiedName{Module: "SampleRuntime", Name: "RuntimeResponse"}}})

	var infoLogID, changeID model.ID
	for _, obj := range oc.Objects {
		activity, ok := obj.(*microflows.ActionActivity)
		if !ok {
			continue
		}
		switch action := activity.Action.(type) {
		case *microflows.LogMessageAction:
			if action.LogLevel == "Info" {
				infoLogID = activity.ID
			}
		case *microflows.ChangeObjectAction:
			changeID = activity.ID
		}
	}
	if infoLogID == "" || changeID == "" {
		t.Fatalf("expected info log and change activities; got info=%q change=%q", infoLogID, changeID)
	}
	for _, flow := range oc.Flows {
		if flow.IsErrorHandler && (flow.DestinationID == infoLogID || flow.DestinationID == changeID) {
			t.Fatal("custom error handler must not rejoin before later output-dependent statements")
		}
	}
}

func TestBuildFlowGraph_EmptyOutputHandlerRoutesToElseBranchUsingOutput(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.CallMicroflowStmt{
			OutputVariable: "AccessToken",
			MicroflowName:  ast.QualifiedName{Module: "SampleAuth", Name: "GetToken"},
			ErrorHandling:  &ast.ErrorHandlingClause{Type: ast.ErrorHandlingCustomWithoutRollback},
		},
		&ast.IfStmt{
			Condition: &ast.SourceExpr{Source: "$AccessToken != empty"},
			ThenBody: []ast.MicroflowStatement{
				&ast.MfSetStmt{Target: "TokenValue", Value: &ast.AttributePathExpr{Variable: "AccessToken", Path: []string{"Value"}}},
			},
			ElseBody: []ast.MicroflowStatement{
				&ast.LogStmt{Level: ast.LogError, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "missing token"}},
				&ast.ReturnStmt{Value: &ast.LiteralExpr{Kind: ast.LiteralEmpty, Value: "empty"}},
			},
		},
		&ast.ReturnStmt{Value: &ast.VariableExpr{Name: "TokenValue"}},
	}

	fb := &flowBuilder{
		posX:         100,
		posY:         100,
		spacing:      HorizontalSpacing,
		declaredVars: map[string]string{"TokenValue": "String"},
		measurer:     &layoutMeasurer{},
	}
	oc := fb.buildFlowGraph(body, &ast.MicroflowReturnType{Type: ast.DataType{Kind: ast.TypeString}})

	var callID, splitID, elseLogID model.ID
	for _, obj := range oc.Objects {
		switch o := obj.(type) {
		case *microflows.ActionActivity:
			switch action := o.Action.(type) {
			case *microflows.MicroflowCallAction:
				if action.ResultVariableName == "AccessToken" {
					callID = o.ID
				}
			case *microflows.LogMessageAction:
				if action.MessageTemplate != nil && action.MessageTemplate.Translations["en_US"] == "missing token" {
					elseLogID = o.ID
				}
			}
		case *microflows.ExclusiveSplit:
			if cond, ok := o.SplitCondition.(*microflows.ExpressionSplitCondition); ok && cond.Expression == "$AccessToken != empty" {
				splitID = o.ID
			}
		}
	}
	if callID == "" || splitID == "" || elseLogID == "" {
		t.Fatalf("expected call, split, and else log; got call=%q split=%q else=%q", callID, splitID, elseLogID)
	}
	for _, flow := range oc.Flows {
		if flow.OriginID == callID && flow.DestinationID == splitID && flow.IsErrorHandler {
			t.Fatal("empty output handler must not route the error path into a decision that references the missing output")
		}
	}
	for _, flow := range oc.Flows {
		if flow.OriginID == callID && flow.IsErrorHandler {
			if merge, ok := findMicroflowObjectByID(oc.Objects, flow.DestinationID).(*microflows.ExclusiveMerge); ok {
				for _, mergeFlow := range oc.Flows {
					if mergeFlow.OriginID == merge.ID && mergeFlow.DestinationID == elseLogID {
						return
					}
				}
			}
		}
	}
	t.Fatal("empty output handler should join the else branch through a merge")
}

func TestBuildFlowGraph_CustomShowMessageHandlerUsesAcceptedAnchors(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.CallMicroflowStmt{
			OutputVariable: "Payload",
			MicroflowName:  ast.QualifiedName{Module: "SampleService", Name: "GetPayload"},
			ErrorHandling: &ast.ErrorHandlingClause{
				Type: ast.ErrorHandlingCustom,
				Body: []ast.MicroflowStatement{
					&ast.ShowMessageStmt{
						Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "failed"},
						Type:    "Error",
					},
					&ast.ReturnStmt{Value: &ast.LiteralExpr{Kind: ast.LiteralEmpty, Value: "empty"}},
				},
			},
			Annotations: &ast.ActivityAnnotations{
				Anchor: &ast.FlowAnchors{From: ast.AnchorSideTop, To: ast.AnchorSideLeft},
			},
		},
		&ast.ReturnStmt{Value: &ast.VariableExpr{Name: "Payload"}},
	}

	fb := &flowBuilder{
		posX:     100,
		posY:     100,
		spacing:  HorizontalSpacing,
		measurer: &layoutMeasurer{},
	}
	oc := fb.buildFlowGraph(body, &ast.MicroflowReturnType{Type: ast.DataType{Kind: ast.TypeEntity, EntityRef: &ast.QualifiedName{Module: "SampleService", Name: "Payload"}}})

	var callID, showID, emptyEndID model.ID
	for _, obj := range oc.Objects {
		switch o := obj.(type) {
		case *microflows.ActionActivity:
			switch o.Action.(type) {
			case *microflows.MicroflowCallAction:
				callID = o.ID
			case *microflows.ShowMessageAction:
				showID = o.ID
			}
		case *microflows.EndEvent:
			if strings.TrimSpace(o.ReturnValue) == "empty" {
				emptyEndID = o.ID
			}
		}
	}
	if callID == "" || showID == "" || emptyEndID == "" {
		t.Fatalf("expected call, show-message, and empty return; got call=%q show=%q end=%q", callID, showID, emptyEndID)
	}
	var foundErrorFlow, foundShowReturnFlow bool
	for _, flow := range oc.Flows {
		if flow.OriginID == callID && flow.DestinationID == showID && flow.IsErrorHandler {
			if flow.OriginConnectionIndex != AnchorTop || flow.DestinationConnectionIndex != AnchorBottom {
				t.Fatalf("show-message error flow anchors = from %d to %d, want top to bottom", flow.OriginConnectionIndex, flow.DestinationConnectionIndex)
			}
			foundErrorFlow = true
		}
		if flow.OriginID == showID && flow.DestinationID == emptyEndID {
			if flow.OriginConnectionIndex != AnchorTop || flow.DestinationConnectionIndex != AnchorBottom {
				t.Fatalf("show-message return flow anchors = from %d to %d, want top to bottom", flow.OriginConnectionIndex, flow.DestinationConnectionIndex)
			}
			foundShowReturnFlow = true
		}
	}
	if !foundErrorFlow || !foundShowReturnFlow {
		t.Fatalf("expected error-handler call->show and show->return flows; got error=%v showReturn=%v", foundErrorFlow, foundShowReturnFlow)
	}
}

func TestBuildFlowGraph_CustomLogHandlerIgnoresNormalRightAnchor(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.CallMicroflowStmt{
			MicroflowName: ast.QualifiedName{Module: "SampleService", Name: "ApplyValue"},
			ErrorHandling: &ast.ErrorHandlingClause{
				Type: ast.ErrorHandlingCustomWithoutRollback,
				Body: []ast.MicroflowStatement{
					&ast.LogStmt{Level: ast.LogError, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "failed"}},
				},
			},
			Annotations: &ast.ActivityAnnotations{
				Anchor: &ast.FlowAnchors{From: ast.AnchorSideRight, To: ast.AnchorSideLeft},
			},
		},
		&ast.ReturnStmt{},
	}

	fb := &flowBuilder{posX: 100, posY: 100, spacing: HorizontalSpacing, measurer: &layoutMeasurer{}}
	oc := fb.buildFlowGraph(body, nil)

	var callID, logID model.ID
	for _, obj := range oc.Objects {
		activity, ok := obj.(*microflows.ActionActivity)
		if !ok {
			continue
		}
		switch activity.Action.(type) {
		case *microflows.MicroflowCallAction:
			callID = activity.ID
		case *microflows.LogMessageAction:
			logID = activity.ID
		}
	}
	if callID == "" || logID == "" {
		t.Fatalf("expected call and log actions; got call=%q log=%q", callID, logID)
	}
	for _, flow := range oc.Flows {
		if flow.OriginID == callID && flow.DestinationID == logID && flow.IsErrorHandler {
			if flow.OriginConnectionIndex != AnchorBottom || flow.DestinationConnectionIndex != AnchorTop {
				t.Fatalf("log error flow anchors = from %d to %d, want bottom to top", flow.OriginConnectionIndex, flow.DestinationConnectionIndex)
			}
			return
		}
	}
	t.Fatal("expected error-handler flow from call to log")
}

func TestBuildFlowGraph_NonTerminalHandlerRejoinsBeforeReturnThroughMerge(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.CallMicroflowStmt{
			MicroflowName: ast.QualifiedName{Module: "SampleService", Name: "ApplyValue"},
			ErrorHandling: &ast.ErrorHandlingClause{
				Type: ast.ErrorHandlingCustomWithoutRollback,
				Body: []ast.MicroflowStatement{
					&ast.LogStmt{Level: ast.LogError, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "failed"}},
				},
			},
		},
		&ast.ReturnStmt{Value: &ast.VariableExpr{Name: "App"}},
	}

	fb := &flowBuilder{
		posX:     100,
		posY:     100,
		spacing:  HorizontalSpacing,
		varTypes: map[string]string{"App": "SampleService.App"},
		measurer: &layoutMeasurer{},
	}
	oc := fb.buildFlowGraph(body, &ast.MicroflowReturnType{Type: ast.DataType{Kind: ast.TypeEntity, EntityRef: &ast.QualifiedName{Module: "SampleService", Name: "App"}}})

	var callID, logID, endID, mergeID model.ID
	for _, obj := range oc.Objects {
		switch o := obj.(type) {
		case *microflows.ActionActivity:
			switch o.Action.(type) {
			case *microflows.MicroflowCallAction:
				callID = o.ID
			case *microflows.LogMessageAction:
				logID = o.ID
			}
		case *microflows.EndEvent:
			if strings.TrimSpace(o.ReturnValue) == "$App" {
				endID = o.ID
			}
		case *microflows.ExclusiveMerge:
			mergeID = o.ID
		}
	}
	if callID == "" || logID == "" || endID == "" || mergeID == "" {
		t.Fatalf("expected call, log, merge, and return; got call=%q log=%q merge=%q end=%q", callID, logID, mergeID, endID)
	}
	var normalToMerge, errorTailToMerge, mergeToEnd bool
	for _, flow := range oc.Flows {
		if flow.OriginID == callID && flow.DestinationID == endID {
			t.Fatal("normal path must not connect directly to an EndEvent shared with a non-terminal error handler")
		}
		if flow.OriginID == logID && flow.DestinationID == endID {
			t.Fatal("error-handler tail must not connect directly to the shared EndEvent")
		}
		if flow.OriginID == callID && flow.DestinationID == mergeID && !flow.IsErrorHandler {
			normalToMerge = true
		}
		if flow.OriginID == logID && flow.DestinationID == mergeID {
			errorTailToMerge = true
		}
		if flow.OriginID == mergeID && flow.DestinationID == endID {
			mergeToEnd = true
		}
	}
	if !normalToMerge || !errorTailToMerge || !mergeToEnd {
		t.Fatalf("expected normal and error paths to rejoin before return; normal=%v error=%v mergeToEnd=%v", normalToMerge, errorTailToMerge, mergeToEnd)
	}
}

func TestBuildFlowGraph_NestedImportErrorHandlerSkipsOutputDependentTail(t *testing.T) {
	resultEntity := ast.QualifiedName{Module: "SampleRepository", Name: "ResponseRoot"}
	body := []ast.MicroflowStatement{
		&ast.CreateObjectStmt{
			Variable:   "ResponseRoot",
			EntityType: resultEntity,
		},
		&ast.DeclareStmt{
			Variable:     "Payload",
			Type:         ast.DataType{Kind: ast.TypeString},
			InitialValue: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "{}"},
		},
		&ast.ImportFromMappingStmt{
			OutputVariable: "ErrorPayload",
			Mapping:        ast.QualifiedName{Module: "SampleRepository", Name: "ImportErrorPayload"},
			SourceVariable: "Payload",
			ErrorHandling: &ast.ErrorHandlingClause{
				Type: ast.ErrorHandlingCustomWithoutRollback,
				Body: []ast.MicroflowStatement{
					&ast.LogStmt{Level: ast.LogError, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "parse failed"}},
				},
			},
		},
		&ast.LogStmt{
			Level:   ast.LogError,
			Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "api failed"},
			Template: []ast.TemplateParam{{
				Index: 1,
				Value: &ast.AttributePathExpr{
					Variable: "ErrorPayload",
					Path:     []string{"errorMessage"},
				},
			}},
		},
		&ast.ChangeObjectStmt{
			Variable: "ResponseRoot",
			Changes: []ast.ChangeItem{{
				Attribute: "ErrorMessage",
				Value: &ast.AttributePathExpr{
					Variable: "ErrorPayload",
					Path:     []string{"errorMessage"},
				},
			}},
		},
		&ast.ReturnStmt{Value: &ast.VariableExpr{Name: "ResponseRoot"}},
	}

	fb := &flowBuilder{
		posX:     100,
		posY:     100,
		spacing:  HorizontalSpacing,
		varTypes: map[string]string{},
		measurer: &layoutMeasurer{},
	}
	oc := fb.buildFlowGraph(body, &ast.MicroflowReturnType{Type: ast.DataType{Kind: ast.TypeEntity, EntityRef: &resultEntity}})

	var innerLogID, outerLogID model.ID
	returnIDs := map[model.ID]bool{}
	for _, obj := range oc.Objects {
		switch o := obj.(type) {
		case *microflows.ActionActivity:
			logAction, ok := o.Action.(*microflows.LogMessageAction)
			if !ok || logAction.MessageTemplate == nil {
				continue
			}
			switch logAction.MessageTemplate.Translations["en_US"] {
			case "parse failed":
				innerLogID = o.ID
			case "api failed":
				outerLogID = o.ID
			}
		case *microflows.EndEvent:
			if strings.TrimSpace(o.ReturnValue) == "$ResponseRoot" {
				returnIDs[o.ID] = true
			}
		}
	}
	if innerLogID == "" || outerLogID == "" || len(returnIDs) == 0 {
		t.Fatalf("expected inner log, outer log, and return nodes; got inner=%q outer=%q returns=%v", innerLogID, outerLogID, returnIDs)
	}
	for _, flow := range oc.Flows {
		if flow.OriginID == innerLogID && flow.DestinationID == outerLogID {
			t.Fatal("nested import error handler must not continue into statements that reference the failed import output")
		}
		if flow.OriginID == innerLogID && returnIDs[flow.DestinationID] {
			return
		}
	}
	t.Fatal("nested import error handler should return the already-created response object")
}

func TestBuildFlowGraph_RepeatedMicroflowCallOutputDeclaresFirstUse(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.CallMicroflowStmt{
			OutputVariable: "UpdatedRecord",
			MicroflowName:  ast.QualifiedName{Module: "SampleRepositoryApi", Name: "GetFirstRecord"},
			ErrorHandling:  &ast.ErrorHandlingClause{Type: ast.ErrorHandlingRollback},
		},
		&ast.IfStmt{
			Condition: &ast.SourceExpr{Source: "$UpdatedRecord/IsActive = true"},
			ThenBody: []ast.MicroflowStatement{
				&ast.LogStmt{Level: ast.LogInfo, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "active"}},
			},
		},
		&ast.CallMicroflowStmt{
			OutputVariable: "UpdatedRecord",
			MicroflowName:  ast.QualifiedName{Module: "SampleRepositoryApi", Name: "RefreshRecord"},
			ErrorHandling:  &ast.ErrorHandlingClause{Type: ast.ErrorHandlingRollback},
		},
		&ast.ReturnStmt{Value: &ast.LiteralExpr{Kind: ast.LiteralBoolean, Value: true}},
	}

	fb := &flowBuilder{
		posX:     100,
		posY:     100,
		spacing:  HorizontalSpacing,
		varTypes: map[string]string{},
		measurer: &layoutMeasurer{},
	}
	oc := fb.buildFlowGraph(body, &ast.MicroflowReturnType{Type: ast.DataType{Kind: ast.TypeBoolean}})

	useByCall := map[string]bool{}
	for _, obj := range oc.Objects {
		activity, ok := obj.(*microflows.ActionActivity)
		if !ok {
			continue
		}
		action, ok := activity.Action.(*microflows.MicroflowCallAction)
		if !ok || action.MicroflowCall == nil {
			continue
		}
		useByCall[action.MicroflowCall.Microflow] = action.UseReturnVariable
	}
	if !useByCall["SampleRepositoryApi.GetFirstRecord"] {
		t.Fatal("first repeated output call must declare the result variable before downstream references")
	}
	if useByCall["SampleRepositoryApi.RefreshRecord"] {
		t.Fatal("later repeated output call must reassign the existing result variable")
	}
}

func TestBuildFlowGraph_EmptyIfIsNoOp(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.DeclareStmt{Variable: "Result", Type: ast.DataType{Kind: ast.TypeBoolean}, InitialValue: &ast.LiteralExpr{Kind: ast.LiteralBoolean, Value: true}},
		&ast.IfStmt{Condition: &ast.VariableExpr{Name: "EnumValue"}},
		&ast.ReturnStmt{Value: &ast.VariableExpr{Name: "Result"}},
	}

	fb := &flowBuilder{
		posX:     100,
		posY:     100,
		spacing:  HorizontalSpacing,
		measurer: &layoutMeasurer{},
	}
	oc := fb.buildFlowGraph(body, &ast.MicroflowReturnType{Type: ast.DataType{Kind: ast.TypeBoolean}})

	for _, obj := range oc.Objects {
		if _, ok := obj.(*microflows.ExclusiveSplit); ok {
			t.Fatal("empty IF statements should not emit a decision in the MPR")
		}
	}
}

func TestBuildFlowGraph_InfersUniqueEntityReturnAfterEmptyIfNoOp(t *testing.T) {
	entityRef := ast.QualifiedName{Module: "SampleRuntime", Name: "RuntimeProcessingResult"}
	body := []ast.MicroflowStatement{
		&ast.CreateObjectStmt{
			Variable:   "RuntimeProcessingResult",
			EntityType: entityRef,
		},
		&ast.IfStmt{Condition: &ast.VariableExpr{Name: "EventType"}},
	}

	fb := &flowBuilder{
		posX:     100,
		posY:     100,
		spacing:  HorizontalSpacing,
		varTypes: map[string]string{},
		measurer: &layoutMeasurer{},
	}
	oc := fb.buildFlowGraph(body, &ast.MicroflowReturnType{Type: ast.DataType{Kind: ast.TypeEntity, EntityRef: &entityRef}})

	for _, obj := range oc.Objects {
		end, ok := obj.(*microflows.EndEvent)
		if !ok {
			continue
		}
		if end.ReturnValue == "$RuntimeProcessingResult" {
			return
		}
		t.Fatalf("EndEvent return value = %q, want $RuntimeProcessingResult", end.ReturnValue)
	}
	t.Fatal("expected EndEvent")
}

func TestBuildFlowGraph_ReturnValueTrimsTrailingLineBreaks(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.ReturnStmt{Value: &ast.SourceExpr{Source: "empty\n"}},
	}

	entityRef := ast.QualifiedName{Module: "SampleModule", Name: "SampleEntity"}
	fb := &flowBuilder{
		posX:     100,
		posY:     100,
		spacing:  HorizontalSpacing,
		measurer: &layoutMeasurer{},
	}
	oc := fb.buildFlowGraph(body, &ast.MicroflowReturnType{Type: ast.DataType{Kind: ast.TypeListOf, EntityRef: &entityRef}})

	for _, obj := range oc.Objects {
		end, ok := obj.(*microflows.EndEvent)
		if !ok {
			continue
		}
		if end.ReturnValue != "empty" {
			t.Fatalf("EndEvent return value = %q, want %q", end.ReturnValue, "empty")
		}
		return
	}
	t.Fatal("expected EndEvent")
}

func TestConvertErrorHandlingType_EmptyCustomPreservesCustomType(t *testing.T) {
	got := convertErrorHandlingType(&ast.ErrorHandlingClause{Type: ast.ErrorHandlingCustomWithoutRollback})
	if got != microflows.ErrorHandlingTypeCustomWithoutRollback {
		t.Fatalf("empty custom error handler should preserve custom type, got %q", got)
	}
}

func TestBuildFlowGraph_ExplicitEmptyElseReceivesEmptyImportErrorHandler(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.IfStmt{
			Condition: &ast.SourceExpr{Source: "$Response != empty"},
			HasElse:   true,
			ThenBody: []ast.MicroflowStatement{
				&ast.ImportFromMappingStmt{
					OutputVariable: "ErrorPayload",
					Mapping:        ast.QualifiedName{Module: "SampleMapping", Name: "ImportErrorPayload"},
					SourceVariable: "Response",
					ErrorHandling:  &ast.ErrorHandlingClause{Type: ast.ErrorHandlingCustomWithoutRollback},
				},
				&ast.ReturnStmt{Value: &ast.VariableExpr{Name: "ErrorPayload"}},
			},
		},
	}

	entityRef := ast.QualifiedName{Module: "SampleMapping", Name: "ErrorPayload"}
	fb := &flowBuilder{
		posX:     100,
		posY:     100,
		spacing:  HorizontalSpacing,
		varTypes: map[string]string{"Response": "System.HttpResponse"},
		measurer: &layoutMeasurer{},
	}
	oc := fb.buildFlowGraph(body, &ast.MicroflowReturnType{Type: ast.DataType{Kind: ast.TypeEntity, EntityRef: &entityRef}})

	var splitID, mergeID, importID, returnID model.ID
	for _, obj := range oc.Objects {
		switch o := obj.(type) {
		case *microflows.ExclusiveSplit:
			splitID = o.ID
		case *microflows.ExclusiveMerge:
			mergeID = o.ID
		case *microflows.ActionActivity:
			if _, ok := o.Action.(*microflows.ImportXmlAction); ok {
				importID = o.ID
			}
		case *microflows.EndEvent:
			if strings.TrimSpace(o.ReturnValue) == "$ErrorPayload" {
				returnID = o.ID
			}
		}
	}
	if splitID == "" || mergeID == "" || importID == "" || returnID == "" {
		t.Fatalf("expected split, merge, import, and return nodes; got split=%q merge=%q import=%q return=%q", splitID, mergeID, importID, returnID)
	}

	var normalImportToReturn, errorImportToMerge, falseSplitToMerge bool
	for _, flow := range oc.Flows {
		if flow.OriginID == importID && flow.DestinationID == returnID && !flow.IsErrorHandler {
			normalImportToReturn = true
		}
		if flow.OriginID == importID && flow.DestinationID == mergeID && flow.IsErrorHandler {
			errorImportToMerge = true
		}
		if flow.OriginID == importID && flow.DestinationID == returnID && flow.IsErrorHandler {
			t.Fatal("empty import error handler must not flow into a return that depends on the failed import output")
		}
		if flow.OriginID == splitID && flow.DestinationID == mergeID {
			if enumCase, ok := flow.CaseValue.(microflows.EnumerationCase); ok && enumCase.Value == "false" {
				falseSplitToMerge = true
			}
		}
	}
	if !normalImportToReturn || !errorImportToMerge || !falseSplitToMerge {
		var flowDescriptions []string
		for _, flow := range oc.Flows {
			flowDescriptions = append(flowDescriptions, string(flow.OriginID)+"->"+string(flow.DestinationID))
		}
		t.Fatalf("expected import normal->return, import error->merge, and empty-else false->merge; normal=%v error=%v false=%v flows=%v", normalImportToReturn, errorImportToMerge, falseSplitToMerge, flowDescriptions)
	}
}

func findMicroflowObjectByID(objects []microflows.MicroflowObject, id model.ID) microflows.MicroflowObject {
	for _, obj := range objects {
		if obj.GetID() == id {
			return obj
		}
	}
	return nil
}

func flowPathExists(flows []*microflows.SequenceFlow, startID, targetID model.ID) bool {
	if startID == "" || targetID == "" {
		return false
	}
	seen := map[model.ID]bool{}
	queue := []model.ID{startID}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if id == targetID {
			return true
		}
		if seen[id] {
			continue
		}
		seen[id] = true
		for _, flow := range flows {
			if flow.OriginID == id && !seen[flow.DestinationID] {
				queue = append(queue, flow.DestinationID)
			}
		}
	}
	return false
}
