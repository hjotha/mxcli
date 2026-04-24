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

func TestBuilder_InheritanceSplit_NonReturningBranchesMerge(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.InheritanceSplitStmt{
			Variable: "currentUser",
			Cases: []ast.InheritanceSplitCase{
				{
					Entity: ast.QualifiedName{Module: "Administration", Name: "Account"},
					Body: []ast.MicroflowStatement{
						&ast.ShowMessageStmt{
							Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "account"},
							Type:    "Information",
						},
					},
				},
				{
					Entity: ast.QualifiedName{Module: "System", Name: "User"},
					Body: []ast.MicroflowStatement{
						&ast.ShowMessageStmt{
							Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "user"},
							Type:    "Information",
						},
					},
				},
			},
		},
	}
	fb := &flowBuilder{
		posX:    100,
		posY:    100,
		spacing: HorizontalSpacing,
	}

	oc := fb.buildFlowGraph(body, nil)
	var mergeID string
	for _, obj := range oc.Objects {
		if merge, ok := obj.(*microflows.ExclusiveMerge); ok {
			mergeID = string(merge.ID)
			break
		}
	}
	if mergeID == "" {
		t.Fatal("expected non-returning inheritance split branches to converge through an ExclusiveMerge")
	}

	inbound := 0
	outbound := 0
	for _, flow := range oc.Flows {
		if string(flow.DestinationID) == mergeID {
			inbound++
		}
		if string(flow.OriginID) == mergeID {
			outbound++
		}
	}
	if inbound != 2 {
		t.Fatalf("merge inbound flows: got %d, want 2", inbound)
	}
	if outbound != 1 {
		t.Fatalf("merge outbound flows: got %d, want 1", outbound)
	}
}
