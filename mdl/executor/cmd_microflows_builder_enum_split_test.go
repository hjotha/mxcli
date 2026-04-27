// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/sdk/microflows"
)

func TestBuildFlowGraph_EnumSplitEmitsEnumerationCases(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.EnumSplitStmt{
			Variable: "EventType",
			Cases: []ast.EnumSplitCase{
				{Value: "CREATE"},
				{
					Value: "DELETE",
					Body: []ast.MicroflowStatement{
						&ast.LogStmt{Level: ast.LogInfo, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "delete"}},
					},
				},
				{
					Value: "(empty)",
					Body: []ast.MicroflowStatement{
						&ast.LogStmt{Level: ast.LogInfo, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "empty"}},
					},
				},
			},
		},
	}

	fb := &flowBuilder{posX: 100, posY: 100, spacing: HorizontalSpacing, measurer: &layoutMeasurer{}}
	oc := fb.buildFlowGraph(body, nil)

	var split *microflows.ExclusiveSplit
	for _, obj := range oc.Objects {
		if candidate, ok := obj.(*microflows.ExclusiveSplit); ok {
			split = candidate
			break
		}
	}
	if split == nil {
		t.Fatal("expected an ExclusiveSplit")
	}
	cond, ok := split.SplitCondition.(*microflows.ExpressionSplitCondition)
	if !ok {
		t.Fatalf("split condition: got %T, want *microflows.ExpressionSplitCondition", split.SplitCondition)
	}
	if cond.Expression != "$EventType" {
		t.Fatalf("split expression: got %q, want $EventType", cond.Expression)
	}

	values := map[string]bool{}
	for _, flow := range oc.Flows {
		if flow.OriginID != split.ID {
			continue
		}
		caseValue, ok := enumCaseValue(flow.CaseValue)
		if ok {
			values[caseValue] = true
		}
	}
	for _, want := range []string{"CREATE", "DELETE", "(empty)"} {
		if !values[want] {
			t.Fatalf("missing enum case %q in split flows; got %#v", want, values)
		}
	}
	if values["true"] || values["false"] {
		t.Fatalf("enum split must not encode branches as boolean cases: %#v", values)
	}
}

func TestBuildFlowGraph_EnumSplitGroupedCasesShareDestination(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.EnumSplitStmt{
			Variable: "Event/EventType",
			Cases: []ast.EnumSplitCase{
				{
					Value:  "CREATE",
					Values: []string{"CREATE", "UPDATE"},
					Body: []ast.MicroflowStatement{
						&ast.LogStmt{Level: ast.LogInfo, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "upsert"}},
					},
				},
				{
					Value: "DELETE",
					Body: []ast.MicroflowStatement{
						&ast.LogStmt{Level: ast.LogInfo, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "delete"}},
					},
				},
			},
		},
	}

	fb := &flowBuilder{posX: 100, posY: 100, spacing: HorizontalSpacing, measurer: &layoutMeasurer{}}
	oc := fb.buildFlowGraph(body, nil)

	var split *microflows.ExclusiveSplit
	for _, obj := range oc.Objects {
		if candidate, ok := obj.(*microflows.ExclusiveSplit); ok {
			split = candidate
			break
		}
	}
	if split == nil {
		t.Fatal("expected an ExclusiveSplit")
	}

	destinations := map[string]string{}
	destinationAnchors := map[string]int{}
	originAnchors := map[string]int{}
	originVectors := map[string]string{}
	destinationVectors := map[string]string{}
	for _, flow := range oc.Flows {
		if flow.OriginID != split.ID {
			continue
		}
		caseValue, ok := enumCaseValue(flow.CaseValue)
		if ok {
			destinations[caseValue] = string(flow.DestinationID)
			destinationAnchors[caseValue] = flow.DestinationConnectionIndex
			originAnchors[caseValue] = flow.OriginConnectionIndex
			originVectors[caseValue] = flow.OriginControlVector
			destinationVectors[caseValue] = flow.DestinationControlVector
		}
	}
	if destinations["CREATE"] == "" || destinations["UPDATE"] == "" {
		t.Fatalf("expected CREATE and UPDATE enum flows; got %#v", destinations)
	}
	if destinations["CREATE"] != destinations["UPDATE"] {
		t.Fatalf("grouped enum cases must share the same destination; got %#v", destinations)
	}
	var sharedDestination microflows.MicroflowObject
	for _, obj := range oc.Objects {
		if string(obj.GetID()) == destinations["CREATE"] {
			sharedDestination = obj
			break
		}
	}
	if _, ok := sharedDestination.(*microflows.ExclusiveMerge); !ok {
		t.Fatalf("grouped enum cases should join through an ExclusiveMerge before the shared body, got %T", sharedDestination)
	}
	if destinationAnchors["CREATE"] == destinationAnchors["UPDATE"] {
		t.Fatalf("grouped enum flows to the same activity must use distinct destination anchors; got %#v", destinationAnchors)
	}
	if originAnchors["CREATE"] != AnchorRight || originAnchors["UPDATE"] != AnchorRight {
		t.Fatalf("non-empty enum cases should leave the split from the right anchor; got %#v", originAnchors)
	}
	if originVectors["CREATE"] == "" || destinationVectors["CREATE"] == "" || originVectors["UPDATE"] == "" || destinationVectors["UPDATE"] == "" {
		t.Fatalf("grouped enum flows should carry non-default control vectors; origin=%#v destination=%#v", originVectors, destinationVectors)
	}
	if destinations["DELETE"] == "" {
		t.Fatalf("missing DELETE enum flow; got %#v", destinations)
	}
}

func TestBuildFlowGraph_EnumSplitAllTerminalCasesDoesNotAddDefaultFlow(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.EnumSplitStmt{
			Variable: "EventType",
			Cases: []ast.EnumSplitCase{
				{
					Value: "CREATE",
					Body:  []ast.MicroflowStatement{&ast.ReturnStmt{}},
				},
				{
					Value: "DELETE",
					Body:  []ast.MicroflowStatement{&ast.ReturnStmt{}},
				},
				{
					Value: "(empty)",
					Body:  []ast.MicroflowStatement{&ast.ReturnStmt{}},
				},
			},
		},
	}

	fb := &flowBuilder{posX: 100, posY: 100, spacing: HorizontalSpacing, measurer: &layoutMeasurer{}}
	oc := fb.buildFlowGraph(body, nil)

	var split *microflows.ExclusiveSplit
	for _, obj := range oc.Objects {
		if candidate, ok := obj.(*microflows.ExclusiveSplit); ok {
			split = candidate
			break
		}
	}
	if split == nil {
		t.Fatal("expected an ExclusiveSplit")
	}
	for _, flow := range oc.Flows {
		if flow.OriginID != split.ID {
			continue
		}
		if isNoCaseValue(flow.CaseValue) {
			t.Fatalf("terminal enum split must not synthesize a default NoCase flow: %#v", flow)
		}
	}
}

func TestBuildFlowGraph_EnumSplitSiblingBranchesDeclareRepeatedCallOutputs(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.EnumSplitStmt{
			Variable: "Mode",
			Cases: []ast.EnumSplitCase{
				{
					Value: "First",
					Body: []ast.MicroflowStatement{
						&ast.CallMicroflowStmt{OutputVariable: "GeneratedValue", MicroflowName: ast.QualifiedName{Module: "Sample", Name: "Generate"}},
						&ast.ChangeObjectStmt{Variable: "Target", Changes: []ast.ChangeItem{{Attribute: "Value", Value: &ast.VariableExpr{Name: "GeneratedValue"}}}},
					},
				},
				{
					Value: "Second",
					Body: []ast.MicroflowStatement{
						&ast.CallMicroflowStmt{OutputVariable: "GeneratedValue", MicroflowName: ast.QualifiedName{Module: "Sample", Name: "Generate"}},
						&ast.ChangeObjectStmt{Variable: "Target", Changes: []ast.ChangeItem{{Attribute: "Value", Value: &ast.VariableExpr{Name: "GeneratedValue"}}}},
					},
				},
			},
		},
	}

	fb := &flowBuilder{
		posX:         100,
		posY:         100,
		spacing:      HorizontalSpacing,
		measurer:     &layoutMeasurer{},
		declaredVars: map[string]string{"Target": "Sample.Target"},
	}
	oc := fb.buildFlowGraph(body, nil)

	declaringCalls := 0
	for _, obj := range oc.Objects {
		activity, ok := obj.(*microflows.ActionActivity)
		if !ok {
			continue
		}
		action, ok := activity.Action.(*microflows.MicroflowCallAction)
		if !ok || action.ResultVariableName != "GeneratedValue" {
			continue
		}
		if !action.UseReturnVariable {
			t.Fatalf("sibling enum branches must each declare repeated call output variables")
		}
		declaringCalls++
	}
	if declaringCalls != 2 {
		t.Fatalf("declaring calls: got %d, want 2", declaringCalls)
	}
}
