// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/sdk/microflows"
)

func TestBuildJavaAction_PlaceholderArgumentPreservesEmptyBasicValue(t *testing.T) {
	fb := &flowBuilder{posX: 100, posY: 100, spacing: HorizontalSpacing}

	id := fb.addCallJavaActionAction(&ast.CallJavaActionStmt{
		OutputVariable: "Total",
		ActionName:     ast.QualifiedName{Module: "SampleModule", Name: "Recalculate"},
		Arguments: []ast.CallArgument{
			{Name: "CompanyId", Value: &ast.SourceExpr{Source: "..."}},
			{Name: "RecalculateAll", Value: &ast.LiteralExpr{Kind: ast.LiteralBoolean, Value: true}},
			{Name: "ItemList", Value: &ast.SourceExpr{Source: "..."}},
		},
	})

	if id == "" || len(fb.objects) != 1 {
		t.Fatalf("expected one java action activity, got id=%q objects=%d", id, len(fb.objects))
	}
	activity, ok := fb.objects[0].(*microflows.ActionActivity)
	if !ok {
		t.Fatalf("object type = %T, want *microflows.ActionActivity", fb.objects[0])
	}
	action, ok := activity.Action.(*microflows.JavaActionCallAction)
	if !ok {
		t.Fatalf("action type = %T, want *microflows.JavaActionCallAction", activity.Action)
	}
	if len(action.ParameterMappings) != 3 {
		t.Fatalf("got %d parameter mappings, want 3", len(action.ParameterMappings))
	}
	for _, idx := range []int{0, 2} {
		value, ok := action.ParameterMappings[idx].Value.(*microflows.BasicCodeActionParameterValue)
		if !ok {
			t.Fatalf("mapping %d value type = %T, want *BasicCodeActionParameterValue", idx, action.ParameterMappings[idx].Value)
		}
		if value.Argument != "" {
			t.Fatalf("mapping %d argument = %q, want empty string", idx, value.Argument)
		}
	}
	value, ok := action.ParameterMappings[1].Value.(*microflows.BasicCodeActionParameterValue)
	if !ok {
		t.Fatalf("boolean mapping value type = %T, want *BasicCodeActionParameterValue", action.ParameterMappings[1].Value)
	}
	if value.Argument != "true" {
		t.Fatalf("boolean argument = %q, want true", value.Argument)
	}
}
