// SPDX-License-Identifier: Apache-2.0

package mpr

import (
	"testing"

	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/microflows"
	"go.mongodb.org/mongo-driver/bson"
)

func TestBuildSequenceFlowCase_InheritanceCase(t *testing.T) {
	doc := buildSequenceFlowCase(&microflows.InheritanceCase{
		BaseElement:         model.BaseElement{ID: "case-1"},
		EntityQualifiedName: "Sample.SpecializedInput",
	})

	if got := bsonGetKey(doc, "$Type"); got != "Microflows$InheritanceCase" {
		t.Fatalf("$Type = %v, want Microflows$InheritanceCase", got)
	}
	if got := bsonGetKey(doc, "Value"); got != "Sample.SpecializedInput" {
		t.Fatalf("Value = %v, want Sample.SpecializedInput", got)
	}
}

func TestSerializeMicroflowObject_InheritanceSplit(t *testing.T) {
	doc := serializeMicroflowObject(&microflows.InheritanceSplit{
		BaseMicroflowObject: microflows.BaseMicroflowObject{
			BaseElement: model.BaseElement{ID: "split-1"},
			Position:    model.Point{X: 100, Y: 200},
			Size:        model.Size{Width: 120, Height: 60},
		},
		VariableName:      "Input",
		ErrorHandlingType: microflows.ErrorHandlingTypeRollback,
	})

	if got := bsonGetKey(doc, "$Type"); got != "Microflows$InheritanceSplit" {
		t.Fatalf("$Type = %v, want Microflows$InheritanceSplit", got)
	}
	if got := bsonGetKey(doc, "SplitVariableName"); got != "Input" {
		t.Fatalf("SplitVariableName = %v, want Input", got)
	}
}

func TestCastAction_RoundtripVariableName(t *testing.T) {
	action := &microflows.CastAction{
		BaseElement:    model.BaseElement{ID: "cast-1"},
		OutputVariable: "SpecificInput",
	}
	doc := serializeMicroflowAction(action)
	data, err := bson.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal cast action: %v", err)
	}
	var raw map[string]any
	if err := bson.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal cast action: %v", err)
	}

	parsed := parseCastAction(raw)
	if parsed.OutputVariable != "SpecificInput" {
		t.Fatalf("OutputVariable = %q, want SpecificInput", parsed.OutputVariable)
	}
}
