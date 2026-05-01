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
