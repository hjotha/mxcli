// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"strings"
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/model"
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
	if inbound != 3 {
		t.Fatalf("merge inbound flows: got %d, want 3", inbound)
	}
	if outbound != 1 {
		t.Fatalf("merge outbound flows: got %d, want 1", outbound)
	}
}

func TestTraverseFlow_InheritanceSplitPreservesEmptyCases(t *testing.T) {
	e := newTestExecutor()
	activityMap := map[model.ID]microflows.MicroflowObject{
		mkID("start"): &microflows.StartEvent{BaseMicroflowObject: mkObj("start")},
		mkID("split"): &microflows.InheritanceSplit{
			BaseMicroflowObject: mkObj("split"),
			VariableName:        "ListContext",
		},
		mkID("cast"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("cast")},
			Action:       &microflows.CastAction{OutputVariable: "EnvironmentListContext"},
		},
		mkID("merge"): &microflows.ExclusiveMerge{BaseMicroflowObject: mkObj("merge")},
		mkID("end"):   &microflows.EndEvent{BaseMicroflowObject: mkObj("end")},
	}
	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		mkID("start"): {mkFlow("start", "split")},
		mkID("split"): {
			mkBranchFlow("split", "cast", microflows.InheritanceCase{EntityQualifiedName: "Cloud.EnvironmentListContext"}),
			mkBranchFlow("split", "merge", microflows.InheritanceCase{EntityQualifiedName: "MultiSelectionListView.ListContext"}),
		},
		mkID("cast"):  {mkFlow("cast", "merge")},
		mkID("merge"): {mkFlow("merge", "end")},
	}
	splitMergeMap := map[model.ID]model.ID{mkID("split"): mkID("merge")}
	var lines []string
	e.traverseFlow(mkID("start"), activityMap, flowsByOrigin, splitMergeMap, map[model.ID]bool{}, nil, nil, &lines, 0, nil, 0, nil)
	got := strings.Join(lines, "\n")
	assertContains(t, got, "case Cloud.EnvironmentListContext")
	assertContains(t, got, "case MultiSelectionListView.ListContext")
	assertContains(t, got, "cast $EnvironmentListContext;")
}

func TestBuilder_InheritanceSplit_EmptyCaseCreatesConfiguredFlow(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.InheritanceSplitStmt{
			Variable: "ListContext",
			Cases: []ast.InheritanceSplitCase{
				{
					Entity: ast.QualifiedName{Module: "Cloud", Name: "EnvironmentListContext"},
					Body: []ast.MicroflowStatement{
						&ast.CastObjectStmt{OutputVariable: "EnvironmentListContext"},
					},
				},
				{Entity: ast.QualifiedName{Module: "MultiSelectionListView", Name: "ListContext"}},
			},
			ElseBody: []ast.MicroflowStatement{},
		},
		&ast.ReturnStmt{Value: &ast.LiteralExpr{Kind: ast.LiteralEmpty, Value: "empty"}},
	}
	fb := &flowBuilder{posX: 100, posY: 100, spacing: HorizontalSpacing, measurer: &layoutMeasurer{}}
	oc := fb.buildFlowGraph(body, nil)

	var splitID, mergeID model.ID
	for _, obj := range oc.Objects {
		switch obj.(type) {
		case *microflows.InheritanceSplit:
			splitID = obj.GetID()
		case *microflows.ExclusiveMerge:
			mergeID = obj.GetID()
		}
	}
	if splitID == "" || mergeID == "" {
		t.Fatalf("expected split and merge, got split=%q merge=%q", splitID, mergeID)
	}
	foundEmptyElse := false
	for _, flow := range oc.Flows {
		if flow.OriginID != splitID || flow.DestinationID != mergeID {
			continue
		}
		if inheritanceCaseValue(flow.CaseValue) == "MultiSelectionListView.ListContext" {
			foundEmptyElse = true
		}
	}
	if !foundEmptyElse {
		t.Fatal("expected empty inheritance case to produce a configured split-to-merge flow")
	}
	for _, flow := range oc.Flows {
		if flow.OriginID == splitID && flow.DestinationID == mergeID {
			if _, ok := flow.CaseValue.(microflows.InheritanceCase); ok && inheritanceCaseValue(flow.CaseValue) == "" {
				return
			}
			if _, ok := flow.CaseValue.(*microflows.InheritanceCase); ok && inheritanceCaseValue(flow.CaseValue) == "" {
				return
			}
		}
	}
	t.Fatal("expected empty else branch to use an InheritanceCase with empty value, not NoCase")
}

func TestBuilder_InheritanceSplitPreservesGuardFalseContinuation(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.InheritanceSplitStmt{
			Variable: "currentUser",
			Cases: []ast.InheritanceSplitCase{
				{
					Entity: ast.QualifiedName{Module: "ControlCenterCommons", Name: "MendixSSOUser"},
					Body: []ast.MicroflowStatement{
						&ast.IfStmt{
							Condition: &ast.BinaryExpr{
								Left:     &ast.VariableExpr{Name: "Member"},
								Operator: "!=",
								Right:    &ast.LiteralExpr{Kind: ast.LiteralEmpty, Value: "empty"},
							},
							ThenBody: []ast.MicroflowStatement{
								&ast.ReturnStmt{Value: &ast.VariableExpr{Name: "Member"}},
							},
						},
						&ast.LogStmt{Level: ast.LogError, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "missing member"}},
						&ast.ReturnStmt{Value: &ast.LiteralExpr{Kind: ast.LiteralEmpty, Value: "empty"}},
					},
				},
			},
			ElseBody: []ast.MicroflowStatement{
				&ast.ReturnStmt{Value: &ast.LiteralExpr{Kind: ast.LiteralEmpty, Value: "empty"}},
			},
		},
	}
	fb := &flowBuilder{
		posX:     100,
		posY:     100,
		spacing:  HorizontalSpacing,
		varTypes: map[string]string{"Member": "SprintrIntegration.Member"},
		measurer: &layoutMeasurer{},
	}
	oc := fb.buildFlowGraph(body, &ast.MicroflowReturnType{Type: ast.DataType{Kind: ast.TypeEntity, EntityRef: &ast.QualifiedName{Module: "SprintrIntegration", Name: "Member"}}})

	var guardSplitID, logID model.ID
	for _, obj := range oc.Objects {
		switch o := obj.(type) {
		case *microflows.ExclusiveSplit:
			if o.Caption == "$Member != empty" {
				guardSplitID = o.ID
			}
		case *microflows.ActionActivity:
			if action, ok := o.Action.(*microflows.LogMessageAction); ok && action.LogLevel == "Error" {
				logID = o.ID
			}
		}
	}
	if guardSplitID == "" || logID == "" {
		t.Fatalf("expected guard split and log; got split=%q log=%q", guardSplitID, logID)
	}
	for _, flow := range oc.Flows {
		if flow.OriginID != guardSplitID || flow.DestinationID != logID {
			continue
		}
		if enumCase, ok := flow.CaseValue.(microflows.EnumerationCase); ok && enumCase.Value == "false" {
			return
		}
		t.Fatalf("guard continuation flow has case %#v, want false", flow.CaseValue)
	}
	t.Fatal("expected false continuation from guard split to following log")
}

func TestBuilder_InheritanceSplitCastRegistersCaseType(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.InheritanceSplitStmt{
			Variable: "ListContext",
			Cases: []ast.InheritanceSplitCase{
				{
					Entity: ast.QualifiedName{Module: "Groups", Name: "AccessGroupMemberListContext"},
					Body: []ast.MicroflowStatement{
						&ast.CastObjectStmt{OutputVariable: "AccessGroupMemberListContext"},
						&ast.ChangeObjectStmt{
							Variable: "AccessGroupMemberListContext",
							Changes:  []ast.ChangeItem{{Attribute: "TotalListSize", Value: &ast.LiteralExpr{Kind: ast.LiteralInteger, Value: "0"}}},
						},
					},
				},
			},
		},
	}
	fb := &flowBuilder{
		posX:     100,
		posY:     100,
		spacing:  HorizontalSpacing,
		varTypes: map[string]string{},
		measurer: &layoutMeasurer{},
	}
	fb.buildFlowGraph(body, nil)

	if got := fb.varTypes["AccessGroupMemberListContext"]; got != "Groups.AccessGroupMemberListContext" {
		t.Fatalf("cast variable type = %q, want inheritance case type", got)
	}
	for _, obj := range fb.objects {
		activity, ok := obj.(*microflows.ActionActivity)
		if !ok {
			continue
		}
		action, ok := activity.Action.(*microflows.ChangeObjectAction)
		if !ok || len(action.Changes) == 0 {
			continue
		}
		if got := action.Changes[0].AttributeQualifiedName; got != "Groups.AccessGroupMemberListContext.TotalListSize" {
			t.Fatalf("change attribute = %q, want qualified cast case attribute", got)
		}
		return
	}
	t.Fatal("expected change action")
}
