// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"strings"
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/microflows"
)

func TestValidateMicroflowBodyRejectsDuplicateImplicitOutputs(t *testing.T) {
	entityRef := ast.QualifiedName{Module: "Synthetic", Name: "Item"}
	stmt := &ast.CreateMicroflowStmt{
		Body: []ast.MicroflowStatement{
			&ast.CreateObjectStmt{
				Variable:   "Item",
				EntityType: entityRef,
			},
			&ast.RetrieveStmt{
				Variable: "Item",
				Source:   entityRef,
				Limit:    "1",
			},
		},
	}

	errs := ValidateMicroflowBody(stmt)
	if len(errs) == 0 {
		t.Fatalf("expected duplicate output variable validation error")
	}
	if !strings.Contains(errs[0], "duplicate variable name '$Item'") {
		t.Fatalf("validation error = %#v, want duplicate $Item", errs)
	}
}

func TestValidateMicroflowBodyRejectsDuplicateCallOutputs(t *testing.T) {
	stmt := &ast.CreateMicroflowStmt{
		Body: []ast.MicroflowStatement{
			&ast.CallMicroflowStmt{
				OutputVariable: "Result",
				MicroflowName:  ast.QualifiedName{Module: "Synthetic", Name: "Compute"},
			},
			&ast.CallJavaActionStmt{
				OutputVariable: "Result",
				ActionName:     ast.QualifiedName{Module: "Synthetic", Name: "ComputeInJava"},
			},
		},
	}

	errs := ValidateMicroflowBody(stmt)
	if len(errs) == 0 {
		t.Fatalf("expected duplicate call output validation error")
	}
	if !strings.Contains(errs[0], "duplicate variable name '$Result'") {
		t.Fatalf("validation error = %#v, want duplicate $Result", errs)
	}
}

func TestValidateMicroflowBodyAllowsDuplicateOutputsInExclusiveBranches(t *testing.T) {
	entityRef := ast.QualifiedName{Module: "Synthetic", Name: "Item"}
	stmt := &ast.CreateMicroflowStmt{
		Body: []ast.MicroflowStatement{
			&ast.IfStmt{
				Condition: &ast.VariableExpr{Name: "UsePrimaryPath"},
				ThenBody: []ast.MicroflowStatement{
					&ast.CreateObjectStmt{Variable: "Result", EntityType: entityRef},
					&ast.ReturnStmt{},
				},
				ElseBody: []ast.MicroflowStatement{
					&ast.RetrieveStmt{Variable: "Result", Source: entityRef, Limit: "1"},
					&ast.ReturnStmt{},
				},
			},
		},
	}

	errs := ValidateMicroflowBody(stmt)
	for _, err := range errs {
		if strings.Contains(err, "duplicate variable name '$Result'") {
			t.Fatalf("exclusive branches must not share duplicate-output scope: %#v", errs)
		}
	}
}

func TestValidateMicroflowBodyAllowsDuplicateOutputsInEnumCases(t *testing.T) {
	entityRef := ast.QualifiedName{Module: "Synthetic", Name: "Item"}
	stmt := &ast.CreateMicroflowStmt{
		Body: []ast.MicroflowStatement{
			&ast.EnumSplitStmt{
				Variable: "Route",
				Cases: []ast.EnumSplitCase{
					{
						Value: "First",
						Body: []ast.MicroflowStatement{
							&ast.CallJavaActionStmt{OutputVariable: "GeneratedID", ActionName: ast.QualifiedName{Module: "Synthetic", Name: "Generate"}},
							&ast.CreateObjectStmt{Variable: "Result", EntityType: entityRef},
							&ast.ReturnStmt{},
						},
					},
					{
						Value: "Second",
						Body: []ast.MicroflowStatement{
							&ast.CallJavaActionStmt{OutputVariable: "GeneratedID", ActionName: ast.QualifiedName{Module: "Synthetic", Name: "Generate"}},
							&ast.CreateObjectStmt{Variable: "Result", EntityType: entityRef},
							&ast.ReturnStmt{},
						},
					},
				},
			},
		},
	}

	errs := strings.Join(ValidateMicroflowBody(stmt), "\n")
	for _, name := range []string{"GeneratedID", "Result"} {
		if strings.Contains(errs, "duplicate variable name '$"+name+"'") {
			t.Fatalf("enum cases must not share duplicate-output scope: %s", errs)
		}
	}
}

func TestFormatMicroflowActivitiesWarnsAboutDuplicateModelOutputs(t *testing.T) {
	oc := &microflows.MicroflowObjectCollection{
		Objects: []microflows.MicroflowObject{
			&microflows.StartEvent{
				BaseMicroflowObject: microflows.BaseMicroflowObject{
					BaseElement: model.BaseElement{ID: "start"},
					Position:    model.Point{X: 0, Y: 100},
				},
			},
			&microflows.ActionActivity{
				BaseActivity: microflows.BaseActivity{
					BaseMicroflowObject: microflows.BaseMicroflowObject{
						BaseElement: model.BaseElement{ID: "first"},
						Position:    model.Point{X: 100, Y: 100},
					},
				},
				Action: &microflows.CreateObjectAction{OutputVariable: "Item", EntityQualifiedName: "Synthetic.Item"},
			},
			&microflows.ActionActivity{
				BaseActivity: microflows.BaseActivity{
					BaseMicroflowObject: microflows.BaseMicroflowObject{
						BaseElement: model.BaseElement{ID: "second"},
						Position:    model.Point{X: 200, Y: 100},
					},
				},
				Action: &microflows.CreateObjectAction{OutputVariable: "Item", EntityQualifiedName: "Synthetic.Item"},
			},
			&microflows.EndEvent{
				BaseMicroflowObject: microflows.BaseMicroflowObject{
					BaseElement: model.BaseElement{ID: "end"},
					Position:    model.Point{X: 300, Y: 100},
				},
			},
		},
		Flows: []*microflows.SequenceFlow{
			{OriginID: "start", DestinationID: "first"},
			{OriginID: "first", DestinationID: "second"},
			{OriginID: "second", DestinationID: "end"},
		},
	}
	lines := formatMicroflowActivities(&ExecContext{}, &microflows.Microflow{ObjectCollection: oc}, nil, nil)
	got := strings.Join(lines, "\n")

	if !strings.Contains(got, "-- WARNING: duplicate output variable $Item") {
		t.Fatalf("describe output missing duplicate warning:\n%s", got)
	}
	if strings.Contains(got, "$Item_2") {
		t.Fatalf("describe output must not invent aliases:\n%s", got)
	}
}

func TestFormatMicroflowActivitiesDoesNotWarnForExclusiveBranchOutputs(t *testing.T) {
	oc := &microflows.MicroflowObjectCollection{
		Objects: []microflows.MicroflowObject{
			&microflows.StartEvent{
				BaseMicroflowObject: microflows.BaseMicroflowObject{
					BaseElement: model.BaseElement{ID: "start"},
					Position:    model.Point{X: 0, Y: 100},
				},
			},
			&microflows.ExclusiveSplit{
				BaseMicroflowObject: microflows.BaseMicroflowObject{
					BaseElement: model.BaseElement{ID: "split"},
					Position:    model.Point{X: 100, Y: 100},
				},
				SplitCondition: &microflows.ExpressionSplitCondition{Expression: "$UsePrimaryPath"},
			},
			&microflows.ActionActivity{
				BaseActivity: microflows.BaseActivity{
					BaseMicroflowObject: microflows.BaseMicroflowObject{
						BaseElement: model.BaseElement{ID: "then_create"},
						Position:    model.Point{X: 200, Y: 100},
					},
				},
				Action: &microflows.CreateObjectAction{OutputVariable: "Result", EntityQualifiedName: "Synthetic.Item"},
			},
			&microflows.ActionActivity{
				BaseActivity: microflows.BaseActivity{
					BaseMicroflowObject: microflows.BaseMicroflowObject{
						BaseElement: model.BaseElement{ID: "else_retrieve"},
						Position:    model.Point{X: 200, Y: 200},
					},
				},
				Action: &microflows.RetrieveAction{
					OutputVariable: "Result",
					Source: &microflows.DatabaseRetrieveSource{
						EntityQualifiedName: "Synthetic.Item",
					},
				},
			},
			&microflows.EndEvent{
				BaseMicroflowObject: microflows.BaseMicroflowObject{
					BaseElement: model.BaseElement{ID: "then_end"},
					Position:    model.Point{X: 300, Y: 100},
				},
			},
			&microflows.EndEvent{
				BaseMicroflowObject: microflows.BaseMicroflowObject{
					BaseElement: model.BaseElement{ID: "else_end"},
					Position:    model.Point{X: 300, Y: 200},
				},
			},
		},
		Flows: []*microflows.SequenceFlow{
			{OriginID: "start", DestinationID: "split"},
			{OriginID: "split", DestinationID: "then_create", CaseValue: &microflows.ExpressionCase{Expression: "true"}},
			{OriginID: "split", DestinationID: "else_retrieve", CaseValue: &microflows.ExpressionCase{Expression: "false"}},
			{OriginID: "then_create", DestinationID: "then_end"},
			{OriginID: "else_retrieve", DestinationID: "else_end"},
		},
	}
	lines := formatMicroflowActivities(&ExecContext{}, &microflows.Microflow{ObjectCollection: oc}, nil, nil)
	got := strings.Join(lines, "\n")

	if strings.Contains(got, "-- WARNING: duplicate output variable $Result") {
		t.Fatalf("exclusive branch outputs must not be warned as linear duplicates:\n%s", got)
	}
}
