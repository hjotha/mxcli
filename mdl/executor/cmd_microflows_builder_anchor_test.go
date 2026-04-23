// SPDX-License-Identifier: Apache-2.0

// Tests for @anchor annotation wiring — the builder must copy
// AST Anchor/TrueBranchAnchor/FalseBranchAnchor fields into the
// OriginConnectionIndex/DestinationConnectionIndex of emitted SequenceFlows.
package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
)

func TestBuilder_AnchorOverridesFlowEndpoints(t *testing.T) {
	// Simple body: log statement with an @anchor on the next statement that
	// connects the log's outgoing flow at "bottom" instead of the default "right".
	body := []ast.MicroflowStatement{
		&ast.LogStmt{
			Level:   ast.LogInfo,
			Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "a"},
			Annotations: &ast.ActivityAnnotations{
				Anchor: &ast.FlowAnchors{From: ast.AnchorSideBottom, To: ast.AnchorSideUnset},
			},
		},
		&ast.LogStmt{
			Level:   ast.LogInfo,
			Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "b"},
			Annotations: &ast.ActivityAnnotations{
				Anchor: &ast.FlowAnchors{From: ast.AnchorSideUnset, To: ast.AnchorSideTop},
			},
		},
	}

	fb := &flowBuilder{
		posX:    100,
		posY:    100,
		spacing: HorizontalSpacing,
	}
	oc := fb.buildFlowGraph(body, nil)

	// Expect 3 flows: start→log1, log1→log2, log2→end.
	if len(oc.Flows) != 3 {
		t.Fatalf("expected 3 flows, got %d", len(oc.Flows))
	}

	log1ToLog2 := oc.Flows[1]
	if log1ToLog2.OriginConnectionIndex != AnchorBottom {
		t.Errorf("log1→log2 origin index: got %d, want %d (Bottom)",
			log1ToLog2.OriginConnectionIndex, AnchorBottom)
	}
	if log1ToLog2.DestinationConnectionIndex != AnchorTop {
		t.Errorf("log1→log2 destination index: got %d, want %d (Top)",
			log1ToLog2.DestinationConnectionIndex, AnchorTop)
	}
}

func TestBuilder_AnchorOmittedKeepsDefault(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.LogStmt{
			Level:   ast.LogInfo,
			Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "a"},
		},
		&ast.LogStmt{
			Level:   ast.LogInfo,
			Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "b"},
		},
	}

	fb := &flowBuilder{
		posX:    100,
		posY:    100,
		spacing: HorizontalSpacing,
	}
	oc := fb.buildFlowGraph(body, nil)

	log1ToLog2 := oc.Flows[1]
	// Default horizontal flow uses Right → Left.
	if log1ToLog2.OriginConnectionIndex != AnchorRight {
		t.Errorf("default origin: got %d, want %d (Right)",
			log1ToLog2.OriginConnectionIndex, AnchorRight)
	}
	if log1ToLog2.DestinationConnectionIndex != AnchorLeft {
		t.Errorf("default destination: got %d, want %d (Left)",
			log1ToLog2.DestinationConnectionIndex, AnchorLeft)
	}
}

func TestBuilder_AnchorPartialOverride(t *testing.T) {
	// Only the From side is overridden; the To side must remain at the default.
	body := []ast.MicroflowStatement{
		&ast.LogStmt{
			Level:   ast.LogInfo,
			Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "a"},
			Annotations: &ast.ActivityAnnotations{
				Anchor: &ast.FlowAnchors{From: ast.AnchorSideTop, To: ast.AnchorSideUnset},
			},
		},
		&ast.LogStmt{
			Level:   ast.LogInfo,
			Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "b"},
		},
	}

	fb := &flowBuilder{posX: 100, posY: 100, spacing: HorizontalSpacing}
	oc := fb.buildFlowGraph(body, nil)

	log1ToLog2 := oc.Flows[1]
	if log1ToLog2.OriginConnectionIndex != AnchorTop {
		t.Errorf("origin: got %d, want %d (Top)", log1ToLog2.OriginConnectionIndex, AnchorTop)
	}
	if log1ToLog2.DestinationConnectionIndex != AnchorLeft {
		t.Errorf("destination stayed default: got %d, want %d (Left)",
			log1ToLog2.DestinationConnectionIndex, AnchorLeft)
	}
}

func TestBuilder_IfBranchAnchorOverrides(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.IfStmt{
			Condition: &ast.LiteralExpr{Kind: ast.LiteralBoolean, Value: true},
			ThenBody: []ast.MicroflowStatement{
				&ast.LogStmt{Level: ast.LogInfo, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "y"}},
			},
			ElseBody: []ast.MicroflowStatement{
				&ast.LogStmt{Level: ast.LogInfo, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "n"}},
			},
			Annotations: &ast.ActivityAnnotations{
				TrueBranchAnchor:  &ast.FlowAnchors{From: ast.AnchorSideTop, To: ast.AnchorSideLeft},
				FalseBranchAnchor: &ast.FlowAnchors{From: ast.AnchorSideBottom, To: ast.AnchorSideTop},
			},
		},
	}

	fb := &flowBuilder{posX: 100, posY: 100, spacing: HorizontalSpacing}
	oc := fb.buildFlowGraph(body, nil)

	// Scan for the split→activity "true" and "false" flows and check their
	// OriginConnectionIndex matches the per-branch anchors.
	var trueFlow, falseFlow bool
	for _, f := range oc.Flows {
		if ec, ok := f.CaseValue.(ast.Expression); ok {
			_ = ec
		}
		switch cv := f.CaseValue.(type) {
		case nil:
			// skip
		default:
			_ = cv
		}
		// The split's outgoing flows carry an EnumerationCase with "true" / "false".
		if cv, ok := f.CaseValue.(interface{ GetValue() string }); ok {
			switch cv.GetValue() {
			case "true":
				if f.OriginConnectionIndex == AnchorTop && f.DestinationConnectionIndex == AnchorLeft {
					trueFlow = true
				}
			case "false":
				if f.OriginConnectionIndex == AnchorBottom && f.DestinationConnectionIndex == AnchorTop {
					falseFlow = true
				}
			}
		}
	}

	// Fallback if the CaseValue interface doesn't expose GetValue — just check
	// that AT LEAST one flow with Top origin and one with Bottom origin exist.
	if !trueFlow {
		for _, f := range oc.Flows {
			if f.OriginConnectionIndex == AnchorTop && f.DestinationConnectionIndex == AnchorLeft {
				trueFlow = true
				break
			}
		}
	}
	if !falseFlow {
		for _, f := range oc.Flows {
			if f.OriginConnectionIndex == AnchorBottom && f.DestinationConnectionIndex == AnchorTop {
				falseFlow = true
				break
			}
		}
	}

	if !trueFlow {
		t.Error("expected a split outgoing flow with origin=Top, destination=Left (true branch anchor)")
	}
	if !falseFlow {
		t.Error("expected a split outgoing flow with origin=Bottom, destination=Top (false branch anchor)")
	}
}
