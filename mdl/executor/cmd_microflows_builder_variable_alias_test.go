// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"strings"
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/sdk/microflows"
)

func TestBuildFlowGraph_DuplicateImplicitOutputAtSamePositionGetsLocalAlias(t *testing.T) {
	entityRef := ast.QualifiedName{Module: "Sample", Name: "Item"}
	sharedPosition := &ast.ActivityAnnotations{Position: &ast.Position{X: 400, Y: 100}}
	body := []ast.MicroflowStatement{
		&ast.CreateObjectStmt{
			Variable:    "SelectedItem",
			EntityType:  entityRef,
			Annotations: sharedPosition,
		},
		&ast.CreateObjectStmt{
			Variable:    "SelectedItem",
			EntityType:  entityRef,
			Annotations: sharedPosition,
		},
		&ast.ChangeObjectStmt{
			Variable: "SelectedItem",
			Changes: []ast.ChangeItem{{
				Attribute: "Name",
				Value:     &ast.VariableExpr{Name: "SelectedItem"},
			}},
		},
		&ast.ReturnStmt{Value: &ast.VariableExpr{Name: "SelectedItem"}},
	}

	fb := &flowBuilder{
		posX: 100, posY: 100, baseY: 100, spacing: HorizontalSpacing,
		varTypes:     map[string]string{},
		declaredVars: map[string]string{},
	}
	oc := fb.buildFlowGraph(body, &ast.MicroflowReturnType{Type: ast.DataType{Kind: ast.TypeEntity, EntityRef: &entityRef}})

	createOutputs := map[string]bool{}
	var changeVariable, changeValue string
	var returnValue string
	for _, obj := range oc.Objects {
		switch o := obj.(type) {
		case *microflows.ActionActivity:
			switch action := o.Action.(type) {
			case *microflows.CreateObjectAction:
				createOutputs[action.OutputVariable] = true
			case *microflows.ChangeObjectAction:
				changeVariable = action.ChangeVariable
				if len(action.Changes) == 1 {
					changeValue = action.Changes[0].Value
				}
			}
		case *microflows.EndEvent:
			returnValue = strings.TrimSpace(o.ReturnValue)
		}
	}

	if !createOutputs["SelectedItem"] || !createOutputs["SelectedItem_2"] {
		t.Fatalf("duplicate implicit output should be aliased, got create outputs %#v", createOutputs)
	}
	if changeVariable != "SelectedItem_2" {
		t.Fatalf("change target = %q, want SelectedItem_2", changeVariable)
	}
	if changeValue != "$SelectedItem_2" {
		t.Fatalf("change value = %q, want $SelectedItem_2", changeValue)
	}
	if returnValue != "$SelectedItem_2" {
		t.Fatalf("return value = %q, want $SelectedItem_2", returnValue)
	}
}

func TestBuildFlowGraph_TerminalBranchDuplicateOutputDoesNotForceAlias(t *testing.T) {
	entityRef := ast.QualifiedName{Module: "Sample", Name: "Item"}
	sharedPosition := &ast.ActivityAnnotations{Position: &ast.Position{X: 400, Y: 100}}
	body := []ast.MicroflowStatement{
		&ast.IfStmt{
			Condition: &ast.LiteralExpr{Kind: ast.LiteralBoolean, Value: true},
			ThenBody: []ast.MicroflowStatement{
				&ast.RetrieveStmt{
					Variable:    "CurrentItem",
					Source:      entityRef,
					Limit:       "1",
					Annotations: sharedPosition,
				},
				&ast.ReturnStmt{Value: &ast.VariableExpr{Name: "CurrentItem"}},
			},
		},
		&ast.RetrieveStmt{
			Variable:    "CurrentItem",
			Source:      entityRef,
			Limit:       "1",
			Annotations: sharedPosition,
		},
		&ast.ReturnStmt{Value: &ast.VariableExpr{Name: "CurrentItem"}},
	}

	fb := &flowBuilder{
		posX: 100, posY: 100, baseY: 100, spacing: HorizontalSpacing,
		varTypes:     map[string]string{},
		declaredVars: map[string]string{},
	}
	oc := fb.buildFlowGraph(body, &ast.MicroflowReturnType{Type: ast.DataType{Kind: ast.TypeEntity, EntityRef: &entityRef}})

	var outputs []string
	for _, obj := range oc.Objects {
		activity, ok := obj.(*microflows.ActionActivity)
		if !ok {
			continue
		}
		if retrieve, ok := activity.Action.(*microflows.RetrieveAction); ok {
			outputs = append(outputs, retrieve.OutputVariable)
		}
	}
	if strings.Join(outputs, ",") != "CurrentItem,CurrentItem" {
		t.Fatalf("retrieve outputs = %#v, want duplicate name preserved without alias", outputs)
	}
}

func TestStatementVarRefsIncludesNonCallConsumers(t *testing.T) {
	refs := referencedVariableSet([]ast.MicroflowStatement{
		&ast.AggregateListStmt{
			InputVariable: "Items",
			Expression:    &ast.VariableExpr{Name: "CurrentItem"},
		},
		&ast.ListOperationStmt{
			InputVariable:  "PrimaryItems",
			SecondVariable: "SecondaryItems",
			Condition:      &ast.VariableExpr{Name: "Candidate"},
			OffsetExpr:     &ast.VariableExpr{Name: "Offset"},
			LimitExpr:      &ast.VariableExpr{Name: "Limit"},
		},
		&ast.CallExternalActionStmt{
			Arguments: []ast.CallArgument{{Value: &ast.VariableExpr{Name: "ExternalInput"}}},
		},
		&ast.RestCallStmt{
			Auth: &ast.RestAuth{
				Username: &ast.VariableExpr{Name: "Username"},
				Password: &ast.VariableExpr{Name: "Password"},
			},
			Body: &ast.RestBody{
				Template:       &ast.VariableExpr{Name: "BodyTemplate"},
				TemplateParams: []ast.TemplateParam{{Value: &ast.VariableExpr{Name: "TemplateArg"}}},
				SourceVariable: "BodyObject",
			},
		},
		&ast.SendRestRequestStmt{
			Parameters:   []ast.SendRestParamDef{{Expression: "$QueryValue + $OtherValue"}},
			BodyVariable: "RequestBody",
		},
		&ast.ImportFromMappingStmt{SourceVariable: "ImportSource"},
		&ast.ExportToMappingStmt{SourceVariable: "ExportSource"},
		&ast.TransformJsonStmt{InputVariable: "JsonInput"},
	})

	for _, name := range []string{
		"Items", "CurrentItem",
		"PrimaryItems", "SecondaryItems", "Candidate", "Offset", "Limit",
		"ExternalInput",
		"Username", "Password", "BodyTemplate", "TemplateArg", "BodyObject",
		"QueryValue", "OtherValue", "RequestBody",
		"ImportSource", "ExportSource", "JsonInput",
	} {
		if !refs[name] {
			t.Fatalf("referencedVariableSet missing %q in %#v", name, refs)
		}
	}
}

func TestAddCallMicroflowAction_DefaultsToReturnVariableWhenPlanMissing(t *testing.T) {
	fb := &flowBuilder{
		posX:                   100,
		posY:                   100,
		spacing:                HorizontalSpacing,
		callOutputDeclarations: map[*ast.CallMicroflowStmt]bool{},
	}

	fb.addCallMicroflowAction(&ast.CallMicroflowStmt{
		OutputVariable: "Result",
		MicroflowName:  ast.QualifiedName{Module: "Synthetic", Name: "Compute"},
	})

	activity, ok := fb.objects[0].(*microflows.ActionActivity)
	if !ok {
		t.Fatalf("object is %T, want *microflows.ActionActivity", fb.objects[0])
	}
	action, ok := activity.Action.(*microflows.MicroflowCallAction)
	if !ok {
		t.Fatalf("action is %T, want *microflows.MicroflowCallAction", activity.Action)
	}
	if !action.UseReturnVariable {
		t.Fatalf("unplanned output variable should default to UseReturnVariable=true")
	}
}
