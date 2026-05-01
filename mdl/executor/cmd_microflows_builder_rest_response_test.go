// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/sdk/microflows"
)

// TestAddRestCallAction_ReturnsResponseUsesHttpResponseHandling pins the
// builder behavior for issue #377: when MDL authors `rest call ... returns
// response`, the builder must construct a ResultHandlingHttpResponse so the
// writer's matching branch (PR #376) emits a DataTypes$ObjectType bound to
// System.HttpResponse, instead of falling back to a string variable.
func TestAddRestCallAction_ReturnsResponseUsesHttpResponseHandling(t *testing.T) {
	fb := &flowBuilder{
		posX:         100,
		posY:         100,
		spacing:      HorizontalSpacing,
		varTypes:     map[string]string{},
		declaredVars: map[string]string{},
		measurer:     &layoutMeasurer{},
	}

	stmt := &ast.RestCallStmt{
		OutputVariable: "Response",
		Method:         ast.HttpMethodGet,
		URL:            &ast.LiteralExpr{Kind: ast.LiteralString, Value: "https://example.com"},
		Result:         ast.RestResult{Type: ast.RestResultResponse},
	}
	fb.addRestCallAction(stmt)

	if len(fb.objects) == 0 {
		t.Fatalf("expected one activity, got %d", len(fb.objects))
	}

	activity, ok := fb.objects[0].(*microflows.ActionActivity)
	if !ok {
		t.Fatalf("first object is %T, want *microflows.ActionActivity", fb.objects[0])
	}
	action, ok := activity.Action.(*microflows.RestCallAction)
	if !ok {
		t.Fatalf("activity.Action is %T, want *microflows.RestCallAction", activity.Action)
	}

	httpResponse, ok := action.ResultHandling.(*microflows.ResultHandlingHttpResponse)
	if !ok {
		t.Fatalf("ResultHandling is %T, want *microflows.ResultHandlingHttpResponse", action.ResultHandling)
	}
	if httpResponse.VariableName != "Response" {
		t.Errorf("VariableName = %q, want %q", httpResponse.VariableName, "Response")
	}
}
