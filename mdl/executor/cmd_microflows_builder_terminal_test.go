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
			t.Fatal("non-terminal custom error handler must not create an empty EndEvent")
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

func TestBuildFlowGraph_CustomErrorHandlerSkipsOutputDependentContinuation(t *testing.T) {
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

	mergeIncoming := map[model.ID]int{}
	mergeToReturn := map[model.ID]bool{}
	for _, flow := range oc.Flows {
		if flow.OriginID == errorLogID && flow.DestinationID == debugLogID {
			t.Fatal("custom error handler must skip statements that depend on the failed action output variable")
		}
		for _, obj := range oc.Objects {
			if _, ok := obj.(*microflows.ExclusiveMerge); ok && obj.GetID() == flow.DestinationID && (flow.OriginID == errorLogID || flow.OriginID == debugLogID) {
				mergeIncoming[flow.DestinationID]++
			}
			if _, ok := obj.(*microflows.ExclusiveMerge); ok && obj.GetID() == flow.OriginID && flow.DestinationID == returnID {
				mergeToReturn[flow.OriginID] = true
			}
		}
	}
	for mergeID, incoming := range mergeIncoming {
		if incoming == 2 && mergeToReturn[mergeID] {
			return
		}
	}
	var flowDescriptions []string
	for _, flow := range oc.Flows {
		flowDescriptions = append(flowDescriptions, string(flow.OriginID)+"->"+string(flow.DestinationID))
	}
	t.Fatalf("expected custom error handler to rejoin through a merge before the first output-independent return; error=%q debug=%q return=%q flows=%v", errorLogID, debugLogID, returnID, flowDescriptions)
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

func TestBuildFlowGraph_EmptyCustomErrorHandlerSkipsInheritanceSplitUsingOutput(t *testing.T) {
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
	for _, obj := range oc.Objects {
		switch o := obj.(type) {
		case *microflows.ActionActivity:
			if _, ok := o.Action.(*microflows.MicroflowCallAction); ok {
				callID = o.ID
			}
		case *microflows.InheritanceSplit:
			splitID = o.ID
		}
	}
	if callID == "" || splitID == "" {
		t.Fatalf("expected call and inheritance split nodes; got call=%q split=%q", callID, splitID)
	}

	hasDirectSuccessFlow := false
	for _, flow := range oc.Flows {
		if flow.OriginID == callID && flow.DestinationID == splitID && !flow.IsErrorHandler {
			hasDirectSuccessFlow = true
		}
		if flow.DestinationID == splitID && flow.IsErrorHandler {
			t.Fatal("empty custom error handler must not rejoin at output-dependent inheritance split")
		}
	}
	if !hasDirectSuccessFlow {
		t.Fatal("success flow should connect directly to the output-dependent inheritance split")
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

func TestBuildFlowGraph_RepeatedMicroflowCallOutputOnlyDeclaresLastUse(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.CallMicroflowStmt{
			OutputVariable: "UpdatedApp",
			MicroflowName:  ast.QualifiedName{Module: "SampleRepositoryApi", Name: "GetRepositoryTypeInfo"},
			ErrorHandling:  &ast.ErrorHandlingClause{Type: ast.ErrorHandlingRollback},
		},
		&ast.CallMicroflowStmt{
			OutputVariable: "UpdatedApp",
			MicroflowName:  ast.QualifiedName{Module: "SampleRepositoryApi", Name: "GetLatestRepositoryInfo"},
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
	if useByCall["SampleRepositoryApi.GetRepositoryTypeInfo"] {
		t.Fatal("first repeated output call must not redeclare the result variable")
	}
	if !useByCall["SampleRepositoryApi.GetLatestRepositoryInfo"] {
		t.Fatal("last repeated output call must declare the result variable")
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

func TestConvertErrorHandlingType_EmptyCustomPreservesCustomType(t *testing.T) {
	got := convertErrorHandlingType(&ast.ErrorHandlingClause{Type: ast.ErrorHandlingCustomWithoutRollback})
	if got != microflows.ErrorHandlingTypeCustomWithoutRollback {
		t.Fatalf("empty custom error handler should preserve custom type, got %q", got)
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
