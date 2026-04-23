// SPDX-License-Identifier: Apache-2.0

// Regression test for the guard-pattern IF anchor leak. When an IF without
// ELSE has `thenReturns=true`, addIfStatement returns the splitID and sets
// nextFlowCase="false" so the OUTER loop in buildFlowGraph creates the
// splitID→nextActivity flow one iteration later. That flow needs the
// falseBranchAnchor from the IF's @anchor annotation — which addIfStatement
// now passes through fb.nextFlowAnchor.
package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/sdk/microflows"
)

func TestBuilder_GuardPatternPreservesFalseBranchAnchor(t *testing.T) {
	// Pattern from AcademyIntegration.GetOrCreateCertificate:
	//   retrieve; if cond then return X end if; create; return X
	//
	// The IF has no else, the then body returns — so the flow that runs
	// when the condition is FALSE connects split → create. That flow must
	// carry @anchor(from: bottom, to: top) (the continuation path drops
	// vertically to the next activity beneath the split).
	body := []ast.MicroflowStatement{
		&ast.LogStmt{
			Level:   ast.LogInfo,
			Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "start"},
		},
		&ast.IfStmt{
			Condition: &ast.LiteralExpr{Kind: ast.LiteralBoolean, Value: true},
			ThenBody: []ast.MicroflowStatement{
				&ast.ReturnStmt{Value: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "yes"}},
			},
			Annotations: &ast.ActivityAnnotations{
				FalseBranchAnchor: &ast.FlowAnchors{From: ast.AnchorSideBottom, To: ast.AnchorSideTop},
			},
		},
		&ast.LogStmt{
			Level:   ast.LogInfo,
			Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "tail"},
		},
	}

	fb := &flowBuilder{posX: 100, posY: 100, spacing: HorizontalSpacing}
	oc := fb.buildFlowGraph(body, nil)

	// Find the flow from the split to the tail log. It's the only one with
	// an EnumerationCase Value=="false" that doesn't target an EndEvent.
	var found *microflows.SequenceFlow
	for _, f := range oc.Flows {
		cv, ok := f.CaseValue.(microflows.EnumerationCase)
		if !ok {
			if p, okp := f.CaseValue.(*microflows.EnumerationCase); okp {
				cv = *p
				ok = true
			}
		}
		if !ok || cv.Value != "false" {
			continue
		}
		// Exclude flows pointing at an EndEvent.
		isEnd := false
		for _, obj := range oc.Objects {
			if obj.GetID() == f.DestinationID {
				if _, e := obj.(*microflows.EndEvent); e {
					isEnd = true
				}
				break
			}
		}
		if !isEnd {
			found = f
			break
		}
	}
	if found == nil {
		t.Fatal("expected a split→tail flow with false case, got none")
	}
	if found.OriginConnectionIndex != AnchorBottom {
		t.Errorf("origin: got %d, want %d (Bottom)", found.OriginConnectionIndex, AnchorBottom)
	}
	if found.DestinationConnectionIndex != AnchorTop {
		t.Errorf("destination: got %d, want %d (Top)", found.DestinationConnectionIndex, AnchorTop)
	}
}
