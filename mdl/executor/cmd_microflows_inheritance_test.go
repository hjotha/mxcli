// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/sdk/microflows"
)

func TestFormatActivity_InheritanceSplit(t *testing.T) {
	e := newTestExecutor()
	obj := &microflows.InheritanceSplit{VariableName: "Input"}

	got := e.formatActivity(obj, nil, nil)
	want := "split type $Input;"
	if got != want {
		t.Fatalf("formatActivity: got %q, want %q", got, want)
	}
}

func TestFormatAction_CastAction(t *testing.T) {
	e := newTestExecutor()
	action := &microflows.CastAction{
		OutputVariable: "Specific",
	}

	got := e.formatAction(action, nil, nil)
	want := "cast $Specific;"
	if got != want {
		t.Fatalf("formatAction: got %q, want %q", got, want)
	}
}

func TestBuilder_InheritanceSplitAndCastAction(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.InheritanceSplitStmt{Variable: "Input"},
		&ast.CastObjectStmt{OutputVariable: "Specific"},
	}
	fb := &flowBuilder{
		posX:    100,
		posY:    100,
		spacing: HorizontalSpacing,
	}

	oc := fb.buildFlowGraph(body, nil)
	if len(oc.Objects) < 4 {
		t.Fatalf("objects: got %d, want at least 4", len(oc.Objects))
	}

	split, ok := oc.Objects[1].(*microflows.InheritanceSplit)
	if !ok {
		t.Fatalf("second object: got %T, want *microflows.InheritanceSplit", oc.Objects[1])
	}
	if split.VariableName != "Input" {
		t.Fatalf("split variable: got %q, want Input", split.VariableName)
	}

	activity, ok := oc.Objects[2].(*microflows.ActionActivity)
	if !ok {
		t.Fatalf("third object: got %T, want *microflows.ActionActivity", oc.Objects[2])
	}
	cast, ok := activity.Action.(*microflows.CastAction)
	if !ok {
		t.Fatalf("action: got %T, want *microflows.CastAction", activity.Action)
	}
	if cast.OutputVariable != "Specific" || cast.ObjectVariable != "" {
		t.Fatalf("cast vars: got output=%q object=%q", cast.OutputVariable, cast.ObjectVariable)
	}
}
