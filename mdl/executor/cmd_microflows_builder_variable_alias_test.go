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
