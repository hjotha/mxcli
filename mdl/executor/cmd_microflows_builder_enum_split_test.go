// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/sdk/microflows"
)

func TestEnumSplitBuilderCreatesEnumerationCaseFlows(t *testing.T) {
	fb := &flowBuilder{
		spacing:  HorizontalSpacing,
		measurer: &layoutMeasurer{},
	}

	fb.addEnumSplit(&ast.EnumSplitStmt{
		Variable: "Status",
		Cases: []ast.EnumSplitCase{
			{
				Values: []string{"Open", "Pending"},
				Body: []ast.MicroflowStatement{
					&ast.LogStmt{Level: ast.LogInfo, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "open"}},
				},
			},
		},
		ElseBody: []ast.MicroflowStatement{
			&ast.LogStmt{Level: ast.LogInfo, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "other"}},
		},
	})

	var split *microflows.ExclusiveSplit
	for _, obj := range fb.objects {
		if candidate, ok := obj.(*microflows.ExclusiveSplit); ok {
			split = candidate
			break
		}
	}
	if split == nil {
		t.Fatal("Expected ExclusiveSplit")
	}
	cond, ok := split.SplitCondition.(*microflows.ExpressionSplitCondition)
	if !ok {
		t.Fatalf("SplitCondition = %T, want ExpressionSplitCondition", split.SplitCondition)
	}
	if cond.Expression != "$Status" {
		t.Fatalf("Expression = %q, want $Status", cond.Expression)
	}

	var cases []string
	for _, flow := range fb.flows {
		if flow.OriginID != split.ID {
			continue
		}
		if value, ok := enumCaseValue(flow); ok {
			cases = append(cases, value)
		}
	}
	if len(cases) != 2 || cases[0] != "Open" || cases[1] != "Pending" {
		t.Fatalf("enum case flows = %v, want [Open Pending]", cases)
	}
}
