// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
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

	var hasBreak, hasContinue bool
	for _, obj := range loop.ObjectCollection.Objects {
		switch obj.(type) {
		case *microflows.BreakEvent:
			hasBreak = true
		case *microflows.ContinueEvent:
			hasContinue = true
		}
	}
	if !hasBreak || !hasContinue {
		t.Fatalf("expected break and continue events in loop body, got break=%v continue=%v", hasBreak, hasContinue)
	}
}

func TestBuildFlowGraph_ManualWhileTrueContinueUsesBackEdgeMerge(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.WhileStmt{
			Condition: &ast.LiteralExpr{Kind: ast.LiteralBoolean, Value: true},
			Body: []ast.MicroflowStatement{
				&ast.LogStmt{Level: ast.LogInfo, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "retry"}},
				&ast.ContinueStmt{},
			},
		},
	}

	fb := &flowBuilder{
		posX:     100,
		posY:     100,
		spacing:  HorizontalSpacing,
		measurer: &layoutMeasurer{},
	}
	oc := fb.buildFlowGraph(body, nil)

	var merge *microflows.ExclusiveMerge
	for _, obj := range oc.Objects {
		switch o := obj.(type) {
		case *microflows.LoopedActivity:
			t.Fatal("manual while true with continue must not be rebuilt as LoopedActivity")
		case *microflows.ContinueEvent:
			t.Fatal("manual while true back-edge must not emit ContinueEvent outside a LoopedActivity")
		case *microflows.ExclusiveMerge:
			merge = o
		}
	}
	if merge == nil {
		t.Fatal("expected manual loop header ExclusiveMerge")
	}

	var backEdges int
	for _, flow := range oc.Flows {
		if flow.DestinationID == merge.ID {
			backEdges++
		}
	}
	if backEdges == 0 {
		t.Fatal("expected continue branch to connect back to the manual-loop merge")
	}
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

func TestLastStmtIsReturn_EnumSplitAllBranchesReturn_Terminal(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.EnumSplitStmt{
			Cases: []ast.EnumSplitCase{
				{Values: []string{"Open"}, Body: []ast.MicroflowStatement{&ast.ReturnStmt{}}},
				{Values: []string{"Closed"}, Body: []ast.MicroflowStatement{&ast.RaiseErrorStmt{}}},
			},
			ElseBody: []ast.MicroflowStatement{&ast.ReturnStmt{}},
		},
	}
	if !lastStmtIsReturn(body) {
		t.Error("ENUM split where all cases and ELSE terminate must be terminal")
	}
}

func TestLastStmtIsReturn_EnumSplitWithoutElseAllBranchesReturn_Terminal(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.EnumSplitStmt{
			Cases: []ast.EnumSplitCase{
				{Values: []string{"Open"}, Body: []ast.MicroflowStatement{&ast.ReturnStmt{}}},
				{Values: []string{"Closed"}, Body: []ast.MicroflowStatement{&ast.ReturnStmt{}}},
			},
		},
	}
	if !lastStmtIsReturn(body) {
		t.Error("ENUM split without ELSE must be terminal when every emitted case terminates")
	}
}

func TestLastStmtIsReturn_EnumSplitWithoutElseNonTerminalCase_NotTerminal(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.EnumSplitStmt{
			Cases: []ast.EnumSplitCase{
				{Values: []string{"Open"}, Body: []ast.MicroflowStatement{&ast.ReturnStmt{}}},
				{Values: []string{"Closed"}, Body: []ast.MicroflowStatement{&ast.LogStmt{}}},
			},
		},
	}
	if lastStmtIsReturn(body) {
		t.Error("ENUM split without ELSE must not be terminal when any emitted case can continue")
	}
}
