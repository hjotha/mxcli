// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/backend/mock"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/javaactions"
	"github.com/mendixlabs/mxcli/sdk/microflows"
)

func TestBuildJavaAction_EmptyArgumentPreservesEmptyBasicValue(t *testing.T) {
	fb := &flowBuilder{posX: 100, posY: 100, spacing: HorizontalSpacing}
	stmt := &ast.CallJavaActionStmt{
		ActionName: ast.QualifiedName{Module: "SampleModule", Name: "Recalculate"},
		Arguments: []ast.CallArgument{
			{Name: "CompanyId", Value: &ast.LiteralExpr{Kind: ast.LiteralEmpty}},
			{Name: "RecalculateAll", Value: &ast.LiteralExpr{Kind: ast.LiteralBoolean, Value: true}},
			{Name: "ItemList", Value: &ast.LiteralExpr{Kind: ast.LiteralEmpty}},
		},
	}

	id := fb.addCallJavaActionAction(stmt)
	var activity *microflows.ActionActivity
	for _, obj := range fb.objects {
		if obj.GetID() == id {
			activity, _ = obj.(*microflows.ActionActivity)
			break
		}
	}
	if activity == nil {
		t.Fatal("expected Java action activity")
	}
	action, ok := activity.Action.(*microflows.JavaActionCallAction)
	if !ok {
		t.Fatalf("action = %T, want *JavaActionCallAction", activity.Action)
	}
	if len(action.ParameterMappings) != 3 {
		t.Fatalf("parameter mappings = %d, want 3", len(action.ParameterMappings))
	}

	for _, idx := range []int{0, 2} {
		value, ok := action.ParameterMappings[idx].Value.(*microflows.BasicCodeActionParameterValue)
		if !ok {
			t.Fatalf("mapping %d value = %T, want *BasicCodeActionParameterValue", idx, action.ParameterMappings[idx].Value)
		}
		if value.Argument != "" {
			t.Fatalf("mapping %d argument = %q, want empty string", idx, value.Argument)
		}
	}

	value, ok := action.ParameterMappings[1].Value.(*microflows.BasicCodeActionParameterValue)
	if !ok {
		t.Fatalf("boolean mapping value = %T, want *BasicCodeActionParameterValue", action.ParameterMappings[1].Value)
	}
	if value.Argument != "true" {
		t.Fatalf("boolean argument = %q, want true", value.Argument)
	}
}

// TestBuildJavaAction_EmptyListArgumentEmitsEmptyKeyword pins the BSON
// shape Studio Pro authors when a list-typed Java action parameter is
// bound to MDL `empty`: the BasicCodeActionParameterValue.Argument
// holds the literal string "empty" — distinguishable from primitive
// parameters whose unbound binding is the blank string. Emitting the
// blank string for a list parameter triggers `mx check` CE0126
// "Missing value for parameter X" because the model treats the
// parameter as missing rather than explicitly empty.
func TestBuildJavaAction_EmptyListArgumentEmitsEmptyKeyword(t *testing.T) {
	fb := &flowBuilder{
		posX:    100,
		posY:    100,
		spacing: HorizontalSpacing,
		backend: &mock.MockBackend{
			ReadJavaActionByNameFunc: func(qualifiedName string) (*javaactions.JavaAction, error) {
				if qualifiedName != "SampleModule.AddBatch" {
					t.Fatalf("java action lookup = %q", qualifiedName)
				}
				return &javaactions.JavaAction{
					Parameters: []*javaactions.JavaActionParameter{
						{
							Name:          "Tags",
							ParameterType: &javaactions.ListType{Entity: "SampleModule.Tag"},
						},
					},
				}, nil
			},
		},
	}
	stmt := &ast.CallJavaActionStmt{
		ActionName: ast.QualifiedName{Module: "SampleModule", Name: "AddBatch"},
		Arguments: []ast.CallArgument{
			{Name: "Tags", Value: &ast.LiteralExpr{Kind: ast.LiteralEmpty}},
		},
	}

	id := fb.addCallJavaActionAction(stmt)
	var activity *microflows.ActionActivity
	for _, obj := range fb.objects {
		if obj.GetID() == id {
			activity, _ = obj.(*microflows.ActionActivity)
			break
		}
	}
	if activity == nil {
		t.Fatal("expected Java action activity")
	}
	action := activity.Action.(*microflows.JavaActionCallAction)
	value, ok := action.ParameterMappings[0].Value.(*microflows.BasicCodeActionParameterValue)
	if !ok {
		t.Fatalf("mapping value = %T, want *BasicCodeActionParameterValue", action.ParameterMappings[0].Value)
	}
	if value.Argument != "empty" {
		t.Fatalf("list-typed empty argument = %q, want %q", value.Argument, "empty")
	}
}

func TestBuildJavaAction_EmptyMicroflowArgumentUsesMicroflowParameterValue(t *testing.T) {
	fb := &flowBuilder{
		posX:    100,
		posY:    100,
		spacing: HorizontalSpacing,
		backend: &mock.MockBackend{
			ReadJavaActionByNameFunc: func(qualifiedName string) (*javaactions.JavaAction, error) {
				if qualifiedName != "SampleModule.StartAsync" {
					t.Fatalf("java action lookup = %q", qualifiedName)
				}
				return &javaactions.JavaAction{
					Parameters: []*javaactions.JavaActionParameter{
						{
							Name: "Callback",
							ParameterType: &javaactions.MicroflowType{
								BaseElement: model.BaseElement{ID: "param-type"},
							},
						},
					},
				}, nil
			},
		},
	}
	stmt := &ast.CallJavaActionStmt{
		ActionName: ast.QualifiedName{Module: "SampleModule", Name: "StartAsync"},
		Arguments: []ast.CallArgument{
			{Name: "Callback", Value: &ast.LiteralExpr{Kind: ast.LiteralEmpty}},
		},
	}

	id := fb.addCallJavaActionAction(stmt)
	var activity *microflows.ActionActivity
	for _, obj := range fb.objects {
		if obj.GetID() == id {
			activity, _ = obj.(*microflows.ActionActivity)
			break
		}
	}
	if activity == nil {
		t.Fatal("expected Java action activity")
	}
	action, ok := activity.Action.(*microflows.JavaActionCallAction)
	if !ok {
		t.Fatalf("action = %T, want *JavaActionCallAction", activity.Action)
	}
	value, ok := action.ParameterMappings[0].Value.(*microflows.MicroflowParameterValue)
	if !ok {
		t.Fatalf("mapping value = %T, want *MicroflowParameterValue", action.ParameterMappings[0].Value)
	}
	if value.Microflow != "" {
		t.Fatalf("placeholder microflow = %q, want empty string", value.Microflow)
	}
}
