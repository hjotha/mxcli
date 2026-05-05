// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"fmt"
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/backend/mock"
	mdltypes "github.com/mendixlabs/mxcli/mdl/types"
	"github.com/mendixlabs/mxcli/model"
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

// REST call mappings backed by an XML schema or message definition (no
// JsonStructure set) must still infer single-vs-list from the import
// mapping's own root element kind. Otherwise the builder defaults to
// SingleObject=false and emits a ListType result, which mismatches the
// authored ObjectType return and triggers CE0117 / CE0019 / CE0136
// downstream when the microflow's return value references the result.
func TestAddRestCallAction_MappingFallsBackToImportMappingRootKindWhenJsonStructureMissing(t *testing.T) {
	fb := &flowBuilder{
		posX:         100,
		posY:         100,
		spacing:      HorizontalSpacing,
		varTypes:     map[string]string{},
		declaredVars: map[string]string{},
		measurer:     &layoutMeasurer{},
		backend: &mock.MockBackend{
			GetImportMappingByQualifiedNameFunc: func(moduleName, name string) (*model.ImportMapping, error) {
				if moduleName != "Synthetic" || name != "MsgDefMapping" {
					return nil, fmt.Errorf("unexpected import mapping %s.%s", moduleName, name)
				}
				return &model.ImportMapping{
					Name: "MsgDefMapping",
					// Empty JsonStructure simulates an XML-schema or message-
					// definition backed mapping.
					JsonStructure: "",
					Elements: []*model.ImportMappingElement{
						{Kind: "Object", Entity: "Synthetic.Item", MaxOccurs: 1, MinOccurs: 1},
					},
				}, nil
			},
		},
	}

	stmt := &ast.RestCallStmt{
		OutputVariable: "Item",
		Method:         ast.HttpMethodGet,
		URL:            &ast.LiteralExpr{Kind: ast.LiteralString, Value: "https://example.com"},
		Result: ast.RestResult{
			Type:         ast.RestResultMapping,
			MappingName:  ast.QualifiedName{Module: "Synthetic", Name: "MsgDefMapping"},
			ResultEntity: ast.QualifiedName{Module: "Synthetic", Name: "Item"},
		},
	}
	fb.addRestCallAction(stmt)

	activity := fb.objects[0].(*microflows.ActionActivity)
	action := activity.Action.(*microflows.RestCallAction)
	mapping := action.ResultHandling.(*microflows.ResultHandlingMapping)
	if !mapping.SingleObject {
		t.Errorf("SingleObject = false, want true (root mapping element Kind=Object)")
	}
}

// And the inverse: an Array root on the mapping element must yield a
// list-typed result handling.
// A repeating Object element (MaxOccurs > 1 or unbounded) is a list, even
// though the BSON Kind is "Object". Studio Pro models a list of objects
// this way for XML schema and message-definition mappings; treating it as
// a singleton triggers `mx check` CE0013/CE0100 ("Input variable must be
// of type 'List'") on downstream aggregate or loop activities.
func TestAddRestCallAction_MappingObjectKindWithUnboundedMaxOccursIsList(t *testing.T) {
	fb := &flowBuilder{
		posX:         100,
		posY:         100,
		spacing:      HorizontalSpacing,
		varTypes:     map[string]string{},
		declaredVars: map[string]string{},
		measurer:     &layoutMeasurer{},
		backend: &mock.MockBackend{
			GetImportMappingByQualifiedNameFunc: func(moduleName, name string) (*model.ImportMapping, error) {
				return &model.ImportMapping{
					Name:          "RepeatingObjectMapping",
					JsonStructure: "",
					Elements: []*model.ImportMappingElement{
						{Kind: "Object", Entity: "Synthetic.Item", MaxOccurs: -1, MinOccurs: 0},
					},
				}, nil
			},
		},
	}

	stmt := &ast.RestCallStmt{
		OutputVariable: "Items",
		Method:         ast.HttpMethodGet,
		URL:            &ast.LiteralExpr{Kind: ast.LiteralString, Value: "https://example.com"},
		Result: ast.RestResult{
			Type:         ast.RestResultMapping,
			MappingName:  ast.QualifiedName{Module: "Synthetic", Name: "RepeatingObjectMapping"},
			ResultEntity: ast.QualifiedName{Module: "Synthetic", Name: "Item"},
		},
	}
	fb.addRestCallAction(stmt)

	activity := fb.objects[0].(*microflows.ActionActivity)
	action := activity.Action.(*microflows.RestCallAction)
	mapping := action.ResultHandling.(*microflows.ResultHandlingMapping)
	if mapping.SingleObject {
		t.Errorf("SingleObject = true, want false (Kind=Object MaxOccurs=-1 should be a list)")
	}
}

// MaxOccurs > 1 (e.g. a fixed-bound repeating element) must also yield a
// list, not a singleton.
func TestAddRestCallAction_MappingObjectKindWithBoundedRepeatIsList(t *testing.T) {
	fb := &flowBuilder{
		posX:         100,
		posY:         100,
		spacing:      HorizontalSpacing,
		varTypes:     map[string]string{},
		declaredVars: map[string]string{},
		measurer:     &layoutMeasurer{},
		backend: &mock.MockBackend{
			GetImportMappingByQualifiedNameFunc: func(moduleName, name string) (*model.ImportMapping, error) {
				return &model.ImportMapping{
					Name:          "BoundedRepeatMapping",
					JsonStructure: "",
					Elements: []*model.ImportMappingElement{
						{Kind: "Object", Entity: "Synthetic.Item", MaxOccurs: 5, MinOccurs: 1},
					},
				}, nil
			},
		},
	}

	stmt := &ast.RestCallStmt{
		OutputVariable: "Items",
		Method:         ast.HttpMethodGet,
		URL:            &ast.LiteralExpr{Kind: ast.LiteralString, Value: "https://example.com"},
		Result: ast.RestResult{
			Type:         ast.RestResultMapping,
			MappingName:  ast.QualifiedName{Module: "Synthetic", Name: "BoundedRepeatMapping"},
			ResultEntity: ast.QualifiedName{Module: "Synthetic", Name: "Item"},
		},
	}
	fb.addRestCallAction(stmt)

	mapping := fb.objects[0].(*microflows.ActionActivity).Action.(*microflows.RestCallAction).ResultHandling.(*microflows.ResultHandlingMapping)
	if mapping.SingleObject {
		t.Errorf("SingleObject = true, want false (Kind=Object MaxOccurs=5 should be a list)")
	}
}

func TestAddRestCallAction_MappingResultPreservesExplicitOutputVariable(t *testing.T) {
	fb := &flowBuilder{
		posX:         100,
		posY:         100,
		spacing:      HorizontalSpacing,
		varTypes:     map[string]string{},
		declaredVars: map[string]string{},
		measurer:     &layoutMeasurer{},
		backend: &mock.MockBackend{
			GetImportMappingByQualifiedNameFunc: func(moduleName, name string) (*model.ImportMapping, error) {
				if moduleName != "Synthetic" || name != "ImportItems" {
					return nil, fmt.Errorf("unexpected import mapping %s.%s", moduleName, name)
				}
				return &model.ImportMapping{JsonStructure: "Synthetic.ItemsPayload"}, nil
			},
			GetJsonStructureByQualifiedNameFunc: func(moduleName, name string) (*mdltypes.JsonStructure, error) {
				if moduleName != "Synthetic" || name != "ItemsPayload" {
					return nil, fmt.Errorf("unexpected json structure %s.%s", moduleName, name)
				}
				return &mdltypes.JsonStructure{
					Elements: []*mdltypes.JsonElement{{ElementType: "Array"}},
				}, nil
			},
		},
	}

	stmt := &ast.RestCallStmt{
		OutputVariable: "Items",
		Method:         ast.HttpMethodGet,
		URL:            &ast.LiteralExpr{Kind: ast.LiteralString, Value: "https://example.com"},
		Result: ast.RestResult{
			Type:         ast.RestResultMapping,
			MappingName:  ast.QualifiedName{Module: "Synthetic", Name: "ImportItems"},
			ResultEntity: ast.QualifiedName{Module: "Synthetic", Name: "Item"},
		},
	}
	fb.addRestCallAction(stmt)

	activity, ok := fb.objects[0].(*microflows.ActionActivity)
	if !ok {
		t.Fatalf("first object is %T, want *microflows.ActionActivity", fb.objects[0])
	}
	action, ok := activity.Action.(*microflows.RestCallAction)
	if !ok {
		t.Fatalf("activity.Action is %T, want *microflows.RestCallAction", activity.Action)
	}
	mapping, ok := action.ResultHandling.(*microflows.ResultHandlingMapping)
	if !ok {
		t.Fatalf("ResultHandling is %T, want *microflows.ResultHandlingMapping", action.ResultHandling)
	}
	if action.OutputVariable != "Items" {
		t.Fatalf("OutputVariable = %q, want Items", action.OutputVariable)
	}
	if mapping.ResultVariable != "Items" {
		t.Fatalf("ResultVariable = %q, want Items", mapping.ResultVariable)
	}
}
