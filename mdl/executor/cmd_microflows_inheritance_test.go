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
	stmt := formatActivity(nil, &microflows.InheritanceSplit{VariableName: "Input"}, nil, nil)
	if stmt != "split type $Input;" {
		t.Fatalf("formatActivity = %q, want split type $Input;", stmt)
	}
}

func TestFormatAction_CastAction(t *testing.T) {
	stmt := formatAction(nil, &microflows.CastAction{OutputVariable: "SpecificInput"}, nil, nil)
	if stmt != "cast $SpecificInput;" {
		t.Fatalf("formatAction = %q, want cast $SpecificInput;", stmt)
	}
}

func TestBuilder_InheritanceSplitAndCastAction(t *testing.T) {
	fb := &flowBuilder{spacing: HorizontalSpacing, measurer: &layoutMeasurer{}}
	oc := fb.buildFlowGraph([]ast.MicroflowStatement{
		&ast.InheritanceSplitStmt{
			Variable: "Input",
			Cases: []ast.InheritanceSplitCase{
				{
					Entity: ast.QualifiedName{Module: "Sample", Name: "SpecializedInput"},
					Body: []ast.MicroflowStatement{
						&ast.CastObjectStmt{OutputVariable: "SpecificInput"},
					},
				},
			},
			ElseBody: []ast.MicroflowStatement{&ast.ReturnStmt{}},
		},
	}, nil)

	var split *microflows.InheritanceSplit
	var cast *microflows.CastAction
	var caseFlow *microflows.SequenceFlow
	for _, obj := range oc.Objects {
		if candidate, ok := obj.(*microflows.InheritanceSplit); ok {
			split = candidate
		}
		if activity, ok := obj.(*microflows.ActionActivity); ok {
			if candidate, ok := activity.Action.(*microflows.CastAction); ok {
				cast = candidate
			}
		}
	}
	for _, flow := range oc.Flows {
		if split != nil && flow.OriginID == split.ID {
			if caseValue, ok := flow.CaseValue.(*microflows.InheritanceCase); ok && caseValue.EntityQualifiedName == "Sample.SpecializedInput" {
				caseFlow = flow
			}
		}
	}
	if split == nil {
		t.Fatal("expected InheritanceSplit object")
	}
	if split.VariableName != "Input" {
		t.Fatalf("split variable = %q, want Input", split.VariableName)
	}
	if cast == nil || cast.OutputVariable != "SpecificInput" {
		t.Fatalf("cast action = %#v, want output SpecificInput", cast)
	}
	if caseFlow == nil {
		t.Fatal("expected inheritance case flow")
	}
	caseValue := caseFlow.CaseValue.(*microflows.InheritanceCase)
	if caseValue.EntityQualifiedName != "Sample.SpecializedInput" {
		t.Fatalf("case entity = %q, want Sample.SpecializedInput", caseValue.EntityQualifiedName)
	}
}

func TestTraverseFlow_InheritanceSplit(t *testing.T) {
	e := newTestExecutor()
	entityID := mkID("entity-specialized")
	activityMap := map[model.ID]microflows.MicroflowObject{
		mkID("split"): &microflows.InheritanceSplit{
			BaseMicroflowObject: mkObj("split"),
			VariableName:        "Input",
		},
		mkID("cast"): &microflows.ActionActivity{
			BaseActivity: microflows.BaseActivity{BaseMicroflowObject: mkObj("cast")},
			Action:       &microflows.CastAction{OutputVariable: "SpecificInput"},
		},
		mkID("fallback"): &microflows.EndEvent{BaseMicroflowObject: mkObj("fallback")},
		mkID("merge"):    &microflows.ExclusiveMerge{BaseMicroflowObject: mkObj("merge")},
	}
	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		mkID("split"): {
			mkBranchFlow("split", "cast", &microflows.InheritanceCase{EntityID: entityID}),
			mkFlow("split", "fallback"),
		},
		mkID("cast"):     {mkFlow("cast", "merge")},
		mkID("fallback"): {mkFlow("fallback", "merge")},
	}
	splitMergeMap := map[model.ID]model.ID{mkID("split"): mkID("merge")}
	entityNames := map[model.ID]string{entityID: "Sample.SpecializedInput"}

	var lines []string
	visited := make(map[model.ID]bool)
	e.traverseFlow(mkID("split"), activityMap, flowsByOrigin, splitMergeMap, visited, entityNames, nil, &lines, 1, nil, 0, nil)

	assertLineContains(t, lines, "split type $Input")
	assertLineContains(t, lines, "case Sample.SpecializedInput")
	assertLineContains(t, lines, "cast $SpecificInput;")
	assertLineContains(t, lines, "else")
	assertLineContains(t, lines, "end split;")
}

func TestTraverseFlow_InheritanceSplitPreservesExplicitCaseOrder(t *testing.T) {
	e := newTestExecutor()
	activityMap := map[model.ID]microflows.MicroflowObject{
		mkID("split"): &microflows.InheritanceSplit{
			BaseMicroflowObject: mkObj("split"),
			VariableName:        "Input",
		},
		mkID("merge"): &microflows.ExclusiveMerge{BaseMicroflowObject: mkObj("merge")},
	}
	accountFlow := mkBranchFlow("split", "merge", &microflows.InheritanceCase{EntityQualifiedName: "Sample.Account"})
	userFlow := mkBranchFlow("split", "merge", &microflows.InheritanceCase{EntityQualifiedName: "Sample.User"})
	applyInheritanceSplitCaseOrder(accountFlow, 0)
	applyInheritanceSplitCaseOrder(userFlow, 1)
	flowsByOrigin := map[model.ID][]*microflows.SequenceFlow{
		mkID("split"): {userFlow, accountFlow},
	}
	splitMergeMap := map[model.ID]model.ID{mkID("split"): mkID("merge")}

	var lines []string
	visited := make(map[model.ID]bool)
	e.traverseFlow(mkID("split"), activityMap, flowsByOrigin, splitMergeMap, visited, nil, nil, &lines, 1, nil, 0, nil)

	out := strings.Join(lines, "\n")
	accountIdx := strings.Index(out, "case Sample.Account")
	userIdx := strings.Index(out, "case Sample.User")
	if accountIdx == -1 || userIdx == -1 {
		t.Fatalf("missing expected cases:\n%s", out)
	}
	if accountIdx > userIdx {
		t.Fatalf("case order was not preserved:\n%s", out)
	}
}

func TestLastStmtIsReturn_InheritanceSplitAllBranchesReturn(t *testing.T) {
	body := []ast.MicroflowStatement{
		&ast.InheritanceSplitStmt{
			Cases: []ast.InheritanceSplitCase{
				{Entity: ast.QualifiedName{Module: "Sample", Name: "SpecializedInput"}, Body: []ast.MicroflowStatement{&ast.ReturnStmt{}}},
			},
			ElseBody: []ast.MicroflowStatement{&ast.ReturnStmt{}},
		},
	}
	if !lastStmtIsReturn(body) {
		t.Fatal("inheritance split where all cases and ELSE return must be terminal")
	}
}

func TestBuilder_InheritanceSplitNestedEmptyThenBranchKeepsContinuationCase(t *testing.T) {
	fb := &flowBuilder{
		spacing:      HorizontalSpacing,
		declaredVars: map[string]string{"HasMember": "Boolean", "HasApp": "Boolean"},
		varTypes:     map[string]string{"Selection": "Sample.Selection"},
		measurer:     &layoutMeasurer{},
	}

	oc := fb.buildFlowGraph([]ast.MicroflowStatement{
		&ast.InheritanceSplitStmt{
			Variable: "Selection",
			Cases: []ast.InheritanceSplitCase{
				{
					Entity: ast.QualifiedName{Module: "Sample", Name: "MemberSelection"},
					Body: []ast.MicroflowStatement{
						&ast.IfStmt{
							Condition: &ast.VariableExpr{Name: "HasMember"},
							ElseBody:  []ast.MicroflowStatement{&ast.ReturnStmt{}},
						},
					},
				},
				{
					Entity: ast.QualifiedName{Module: "Sample", Name: "AppSelection"},
					Body: []ast.MicroflowStatement{
						&ast.IfStmt{
							Condition: &ast.VariableExpr{Name: "HasApp"},
							ElseBody:  []ast.MicroflowStatement{&ast.ReturnStmt{}},
						},
					},
				},
			},
			ElseBody: []ast.MicroflowStatement{&ast.ReturnStmt{}},
		},
		&ast.LogStmt{Level: ast.LogInfo, Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "shared tail"}},
	}, nil)

	objects := map[model.ID]microflows.MicroflowObject{}
	var nestedSplitID model.ID
	for _, obj := range oc.Objects {
		objects[obj.GetID()] = obj
		split, ok := obj.(*microflows.ExclusiveSplit)
		if !ok {
			continue
		}
		if condition, ok := split.SplitCondition.(*microflows.ExpressionSplitCondition); ok && condition.Expression == "$HasMember" {
			nestedSplitID = split.ID
		}
	}
	if nestedSplitID == "" {
		t.Fatal("expected nested decision split")
	}
	for _, flow := range oc.Flows {
		if flow.OriginID != nestedSplitID {
			continue
		}
		caseValue, ok := flow.CaseValue.(microflows.EnumerationCase)
		if !ok || caseValue.Value != "true" {
			continue
		}
		if _, ok := objects[flow.DestinationID].(*microflows.ExclusiveMerge); ok {
			return
		}
	}
	t.Fatal("nested empty-then inheritance branch must carry CaseValue=true to the inheritance merge")
}

func TestBuilder_InheritanceSplitBranchAnchorsApplyToBodyFlows(t *testing.T) {
	fb := &flowBuilder{spacing: HorizontalSpacing, measurer: &layoutMeasurer{}}
	message := &ast.ShowMessageStmt{
		Message: &ast.LiteralExpr{Kind: ast.LiteralString, Value: "No matching account"},
		Type:    "Information",
		Annotations: &ast.ActivityAnnotations{
			Anchor: &ast.FlowAnchors{From: ast.AnchorSideBottom, To: ast.AnchorSideTop},
		},
	}
	bodyReturn := &ast.ReturnStmt{
		Annotations: &ast.ActivityAnnotations{
			Anchor: &ast.FlowAnchors{From: ast.AnchorSideUnset, To: ast.AnchorSideTop},
		},
	}

	oc := fb.buildFlowGraph([]ast.MicroflowStatement{
		&ast.InheritanceSplitStmt{
			Variable: "Input",
			Cases: []ast.InheritanceSplitCase{
				{
					Entity: ast.QualifiedName{Module: "Sample", Name: "Primary"},
					Body:   []ast.MicroflowStatement{&ast.ReturnStmt{}},
				},
				{
					Entity: ast.QualifiedName{Module: "Sample", Name: "Secondary"},
					Body:   []ast.MicroflowStatement{message, bodyReturn},
				},
			},
			ElseBody: []ast.MicroflowStatement{&ast.ReturnStmt{}},
		},
	}, nil)

	var splitID, messageID model.ID
	for _, obj := range oc.Objects {
		switch obj := obj.(type) {
		case *microflows.InheritanceSplit:
			splitID = obj.ID
		case *microflows.ActionActivity:
			if _, ok := obj.Action.(*microflows.ShowMessageAction); ok {
				messageID = obj.ID
			}
		}
	}
	if splitID == "" || messageID == "" {
		t.Fatalf("expected split and show-message activity, got split=%q message=%q", splitID, messageID)
	}

	var splitToMessage, messageToReturn *microflows.SequenceFlow
	var elseCase *microflows.InheritanceCase
	for _, flow := range oc.Flows {
		if flow.OriginID == splitID && flow.DestinationID == messageID {
			splitToMessage = flow
		}
		if flow.OriginID == messageID {
			messageToReturn = flow
		}
		if flow.OriginID == splitID {
			if c, ok := flow.CaseValue.(*microflows.InheritanceCase); ok && c.EntityQualifiedName == "" {
				elseCase = c
			}
		}
	}
	if splitToMessage == nil {
		t.Fatal("expected inheritance split flow to annotated branch body")
	}
	if splitToMessage.OriginConnectionIndex != AnchorBottom || splitToMessage.DestinationConnectionIndex != AnchorTop {
		t.Fatalf("split branch anchors = (%d,%d), want (%d,%d)",
			splitToMessage.OriginConnectionIndex, splitToMessage.DestinationConnectionIndex,
			AnchorBottom, AnchorTop)
	}
	if messageToReturn == nil {
		t.Fatal("expected message to return flow")
	}
	if messageToReturn.OriginConnectionIndex != AnchorBottom || messageToReturn.DestinationConnectionIndex != AnchorTop {
		t.Fatalf("body flow anchors = (%d,%d), want (%d,%d)",
			messageToReturn.OriginConnectionIndex, messageToReturn.DestinationConnectionIndex,
			AnchorBottom, AnchorTop)
	}
	if elseCase == nil {
		t.Fatal("expected ELSE branch to keep an explicit empty inheritance case")
	}
}

func assertLineContains(t *testing.T, lines []string, want string) {
	t.Helper()
	for _, line := range lines {
		if contains(line, want) {
			return
		}
	}
	t.Fatalf("expected output to contain %q, got %v", want, lines)
}
