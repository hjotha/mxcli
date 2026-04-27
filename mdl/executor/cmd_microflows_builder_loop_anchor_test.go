// SPDX-License-Identifier: Apache-2.0

// Tests for LOOP/WHILE internal @anchor handling.
//
// @anchor(iterator: ..., tail: ...) is accepted by the grammar so authors can
// carry the intent forward for a future Mendix capability, but today the
// builder deliberately does NOT emit SequenceFlows between a LoopedActivity
// and its body statements: Studio Pro rejects those edges with CE0709
// "Sequence flow is not accepted by origin or destination." These tests pin
// that behaviour — the loop must round-trip without extra flows even when
// the annotation is present.
package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/microflows"
)

func TestBuilder_LoopIteratorAnchorIsParsedButNotSerialised(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.LogStmt{
			Level:   ast.LogInfo,
			Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "a"},
		},
	}
	stmts := []ast.MicroflowStatement{
		&ast.LoopStmt{
			ListVariable: "Items",
			LoopVariable: "Item",
			Body:         body,
			Annotations: &ast.ActivityAnnotations{
				IteratorAnchor: &ast.FlowAnchors{From: ast.AnchorSideBottom, To: ast.AnchorSideTop},
			},
		},
	}

	fb := &flowBuilder{
		posX: 100, posY: 100, spacing: HorizontalSpacing,
		varTypes:     map[string]string{"Items": "List of MfTest.Item"},
		declaredVars: map[string]string{"Items": "List of MfTest.Item"},
	}
	oc := fb.buildFlowGraph(stmts, nil)

	var loop *microflows.LoopedActivity
	for _, obj := range oc.Objects {
		if l, ok := obj.(*microflows.LoopedActivity); ok {
			loop = l
			break
		}
	}
	if loop == nil {
		t.Fatalf("expected a LoopedActivity in output objects")
	}
	firstID := loop.ObjectCollection.Objects[0].GetID()

	for _, f := range oc.Flows {
		if f.OriginID == loop.ID && f.DestinationID == firstID {
			t.Errorf("unexpected iterator flow loop→firstBody: CE0709 would reject it")
		}
		if f.OriginID == firstID && f.DestinationID == loop.ID {
			t.Errorf("unexpected tail flow firstBody→loop: CE0709 would reject it")
		}
	}
}

func TestBuilder_WhileIteratorAndTailAnchorIsParsedButNotSerialised(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.LogStmt{
			Level:   ast.LogInfo,
			Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "step"},
		},
	}
	stmts := []ast.MicroflowStatement{
		&ast.WhileStmt{
			Condition: &ast.LiteralExpr{Kind: ast.LiteralBoolean, Value: true},
			Body:      body,
			Annotations: &ast.ActivityAnnotations{
				IteratorAnchor: &ast.FlowAnchors{From: ast.AnchorSideTop, To: ast.AnchorSideLeft},
				BodyTailAnchor: &ast.FlowAnchors{From: ast.AnchorSideRight, To: ast.AnchorSideBottom},
			},
		},
	}

	fb := &flowBuilder{posX: 100, posY: 100, spacing: HorizontalSpacing}
	oc := fb.buildFlowGraph(stmts, nil)

	var loop *microflows.LoopedActivity
	for _, obj := range oc.Objects {
		if l, ok := obj.(*microflows.LoopedActivity); ok {
			loop = l
			break
		}
	}
	if loop == nil {
		t.Fatalf("expected LoopedActivity (from while)")
	}
	firstID := loop.ObjectCollection.Objects[0].GetID()

	for _, f := range oc.Flows {
		if f.OriginID == loop.ID && f.DestinationID == firstID {
			t.Errorf("unexpected iterator flow on while loop")
		}
		if f.OriginID == firstID && f.DestinationID == loop.ID {
			t.Errorf("unexpected tail flow on while loop")
		}
	}
}

func TestBuilder_LoopWithoutLoopAnchorEmitsNoIteratorOrTail(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.LogStmt{
			Level:   ast.LogInfo,
			Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "a"},
		},
	}
	stmts := []ast.MicroflowStatement{
		&ast.LoopStmt{
			ListVariable: "Items",
			LoopVariable: "Item",
			Body:         body,
		},
	}

	fb := &flowBuilder{
		posX: 100, posY: 100, spacing: HorizontalSpacing,
		varTypes:     map[string]string{"Items": "List of MfTest.Item"},
		declaredVars: map[string]string{"Items": "List of MfTest.Item"},
	}
	oc := fb.buildFlowGraph(stmts, nil)

	var loop *microflows.LoopedActivity
	for _, obj := range oc.Objects {
		if l, ok := obj.(*microflows.LoopedActivity); ok {
			loop = l
			break
		}
	}
	if loop == nil {
		t.Fatalf("expected LoopedActivity")
	}
	firstID := loop.ObjectCollection.Objects[0].GetID()

	// Baseline: no iterator/tail flows when no annotation is given.
	for _, f := range oc.Flows {
		if f.OriginID == loop.ID && f.DestinationID == firstID {
			t.Errorf("unexpected iterator flow emitted without @anchor(iterator: ...)")
		}
		if f.OriginID == firstID && f.DestinationID == loop.ID {
			t.Errorf("unexpected tail flow emitted without @anchor(tail: ...)")
		}
	}
}

func TestBuilder_WhileTrueWithNestedReturnUsesManualLoopBack(t *testing.T) {
	stmts := []ast.MicroflowStatement{
		&ast.WhileStmt{
			Condition: &ast.LiteralExpr{Kind: ast.LiteralBoolean, Value: true},
			Body: []ast.MicroflowStatement{
				&ast.LogStmt{Level: ast.LogInfo, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "batch"}},
				&ast.IfStmt{
					Condition: &ast.LiteralExpr{Kind: ast.LiteralBoolean, Value: true},
					ThenBody: []ast.MicroflowStatement{
						&ast.ReturnStmt{},
					},
					Annotations: &ast.ActivityAnnotations{
						FalseBranchAnchor: &ast.FlowAnchors{From: ast.AnchorSideTop, To: ast.AnchorSideTop},
					},
				},
			},
		},
	}

	fb := &flowBuilder{posX: 100, posY: 100, spacing: HorizontalSpacing}
	oc := fb.buildFlowGraph(stmts, nil)

	var mergeID model.ID
	endEvents := 0
	for _, obj := range oc.Objects {
		switch o := obj.(type) {
		case *microflows.LoopedActivity:
			t.Fatalf("while true with nested return must use a manual loop, got LoopedActivity %s", o.ID)
		case *microflows.ExclusiveMerge:
			if mergeID == "" {
				mergeID = o.ID
			}
		case *microflows.EndEvent:
			endEvents++
		}
	}
	if mergeID == "" {
		t.Fatal("expected manual loop header merge")
	}
	if endEvents != 1 {
		t.Fatalf("expected only the explicit return end event, got %d", endEvents)
	}
	for _, flow := range oc.Flows {
		cv, ok := flow.CaseValue.(microflows.EnumerationCase)
		if !ok {
			if ptr, ptrOK := flow.CaseValue.(*microflows.EnumerationCase); ptrOK {
				cv = *ptr
				ok = true
			}
		}
		if ok && cv.Value == "false" && flow.DestinationID == mergeID {
			if flow.OriginConnectionIndex != AnchorTop || flow.DestinationConnectionIndex != AnchorTop {
				t.Fatalf("loop-back anchor = %d→%d, want Top→Top", flow.OriginConnectionIndex, flow.DestinationConnectionIndex)
			}
			return
		}
	}
	t.Fatal("expected false branch loop-back flow")
}

func TestBuilder_WhileTrueEndingInReturnDoesNotLoopBackFromEndEvent(t *testing.T) {
	stmts := []ast.MicroflowStatement{
		&ast.WhileStmt{
			Condition: &ast.LiteralExpr{Kind: ast.LiteralBoolean, Value: true},
			Body: []ast.MicroflowStatement{
				&ast.LogStmt{Level: ast.LogInfo, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "batch"}},
				&ast.ReturnStmt{},
			},
		},
	}

	fb := &flowBuilder{posX: 100, posY: 100, spacing: HorizontalSpacing}
	oc := fb.buildFlowGraph(stmts, nil)

	endEvents := map[model.ID]bool{}
	for _, obj := range oc.Objects {
		if _, ok := obj.(*microflows.EndEvent); ok {
			endEvents[obj.GetID()] = true
		}
	}
	if len(endEvents) == 0 {
		t.Fatal("expected explicit return end event")
	}
	for _, flow := range oc.Flows {
		if endEvents[flow.OriginID] {
			t.Fatalf("EndEvent %s must not have outgoing loop-back flow to %s", flow.OriginID, flow.DestinationID)
		}
	}
}
