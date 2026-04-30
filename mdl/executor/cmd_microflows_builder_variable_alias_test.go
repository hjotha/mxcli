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

func TestBuildFlowGraph_TerminalBranchDuplicateOutputAtSamePositionGetsAlias(t *testing.T) {
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
	if strings.Join(outputs, ",") != "CurrentItem,CurrentItem_2" {
		t.Fatalf("retrieve outputs = %#v, want duplicate same-position output aliased", outputs)
	}
}

func TestBuildFlowGraph_DuplicateJavaActionOutputAtSamePositionGetsAlias(t *testing.T) {
	entityRef := ast.QualifiedName{Module: "Sample", Name: "Item"}
	sharedJavaPosition := &ast.ActivityAnnotations{Position: &ast.Position{X: 400, Y: 100}}
	body := []ast.MicroflowStatement{
		&ast.IfStmt{
			Condition: &ast.LiteralExpr{Kind: ast.LiteralBoolean, Value: true},
			ThenBody: []ast.MicroflowStatement{
				&ast.CallJavaActionStmt{
					OutputVariable: "Token",
					ActionName:     ast.QualifiedName{Module: "Sample", Name: "GenerateToken"},
					Annotations:    sharedJavaPosition,
				},
				&ast.CreateObjectStmt{
					Variable:   "CreatedItem",
					EntityType: entityRef,
					Changes: []ast.ChangeItem{{
						Attribute: "Code",
						Value:     &ast.VariableExpr{Name: "Token"},
					}},
				},
				&ast.ReturnStmt{Value: &ast.VariableExpr{Name: "CreatedItem"}},
			},
		},
		&ast.CallJavaActionStmt{
			OutputVariable: "Token",
			ActionName:     ast.QualifiedName{Module: "Sample", Name: "GenerateToken"},
			Annotations:    sharedJavaPosition,
		},
		&ast.CreateObjectStmt{
			Variable:   "CreatedItem",
			EntityType: entityRef,
			Changes: []ast.ChangeItem{{
				Attribute: "Code",
				Value:     &ast.VariableExpr{Name: "Token"},
			}},
		},
		&ast.ReturnStmt{Value: &ast.VariableExpr{Name: "CreatedItem"}},
	}

	fb := &flowBuilder{
		posX: 100, posY: 100, baseY: 100, spacing: HorizontalSpacing,
		varTypes:     map[string]string{},
		declaredVars: map[string]string{},
	}
	oc := fb.buildFlowGraph(body, &ast.MicroflowReturnType{Type: ast.DataType{Kind: ast.TypeEntity, EntityRef: &entityRef}})

	var javaOutputs []string
	var createCodeValues []string
	for _, obj := range oc.Objects {
		activity, ok := obj.(*microflows.ActionActivity)
		if !ok {
			continue
		}
		switch action := activity.Action.(type) {
		case *microflows.JavaActionCallAction:
			javaOutputs = append(javaOutputs, action.ResultVariableName)
		case *microflows.CreateObjectAction:
			for _, change := range action.InitialMembers {
				if change.AttributeQualifiedName == "Sample.Item.Code" || change.AttributeQualifiedName == "Code" {
					createCodeValues = append(createCodeValues, change.Value)
				}
			}
		}
	}

	if strings.Join(javaOutputs, ",") != "Token,Token_2" {
		t.Fatalf("java outputs = %#v, want duplicate same-position output aliased", javaOutputs)
	}
	if len(createCodeValues) != 2 || createCodeValues[1] != "$Token_2" {
		t.Fatalf("create Code values = %#v, want second branch to reference $Token_2", createCodeValues)
	}
}

func TestExprToString_SourceExprAppliesVariableAliases(t *testing.T) {
	fb := &flowBuilder{
		variableAliases: map[string]string{
			"CurrentItem": "CurrentItem_2",
		},
	}

	got := fb.exprToString(&ast.SourceExpr{
		Source: "if $CurrentItem/Code = empty then '$CurrentItem' else $OtherItem/Code",
		Expression: &ast.IfThenElseExpr{
			Condition: &ast.BinaryExpr{
				Left:     &ast.AttributePathExpr{Variable: "CurrentItem", Path: []string{"Code"}},
				Operator: "=",
				Right:    &ast.LiteralExpr{Kind: ast.LiteralEmpty},
			},
			ThenExpr: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "$CurrentItem"},
			ElseExpr: &ast.AttributePathExpr{Variable: "OtherItem", Path: []string{"Code"}},
		},
	})

	want := "if $CurrentItem_2/Code = empty then '$CurrentItem' else $OtherItem/Code"
	if got != want {
		t.Fatalf("exprToString(SourceExpr) = %q, want %q", got, want)
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
