// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/microflows"
)

func TestBuildFlowGraph_NoMergeIfElseContinuesFromBranchTail(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.IfStmt{
			Condition: &ast.VariableExpr{Name: "Done"},
			ThenBody: []ast.MicroflowStatement{
				&ast.ReturnStmt{Value: &ast.LiteralExpr{Kind: ast.LiteralBoolean, Value: true}},
			},
			ElseBody: []ast.MicroflowStatement{
				&ast.LogStmt{Level: ast.LogInfo, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "else branch"}},
			},
		},
		&ast.LogStmt{Level: ast.LogInfo, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "after branch"}},
	}

	fb := &flowBuilder{
		posX:         100,
		posY:         100,
		spacing:      HorizontalSpacing,
		declaredVars: map[string]string{"Done": "Boolean"},
		measurer:     &layoutMeasurer{},
	}
	oc := fb.buildFlowGraph(body, &ast.MicroflowReturnType{Type: ast.DataType{Kind: ast.TypeBoolean}})

	elseID := findLogActivityIDByMessage(t, oc, "else branch")
	afterID := findLogActivityIDByMessage(t, oc, "after branch")

	if !hasSequenceFlow(oc.Flows, elseID, afterID) {
		t.Fatal("continuing ELSE branch tail must connect to the following statement")
	}
	for _, flow := range oc.Flows {
		if flow.DestinationID == afterID && flow.OriginID != elseID {
			t.Fatalf("following statement must not be wired from stale split origin %q", flow.OriginID)
		}
	}
}

func findLogActivityIDByMessage(t *testing.T, oc *microflows.MicroflowObjectCollection, message string) model.ID {
	t.Helper()
	for _, obj := range oc.Objects {
		activity, ok := obj.(*microflows.ActionActivity)
		if !ok {
			continue
		}
		action, ok := activity.Action.(*microflows.LogMessageAction)
		if !ok || action.MessageTemplate == nil {
			continue
		}
		if action.MessageTemplate.GetTranslation("en_US") == message {
			return activity.ID
		}
	}
	t.Fatalf("missing log activity %q", message)
	return ""
}

func hasSequenceFlow(flows []*microflows.SequenceFlow, originID, destinationID model.ID) bool {
	for _, flow := range flows {
		if flow.OriginID == originID && flow.DestinationID == destinationID {
			return true
		}
	}
	return false
}
