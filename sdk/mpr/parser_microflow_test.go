// SPDX-License-Identifier: Apache-2.0

package mpr

import (
	"testing"

	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/javaactions"
	"github.com/mendixlabs/mxcli/sdk/microflows"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestParseSequenceFlow_NewCaseValueEnumerationCase(t *testing.T) {
	flow := parseSequenceFlow(map[string]any{
		"$ID":                        "flow-1",
		"OriginPointer":              "start-1",
		"DestinationPointer":         "dest-1",
		"OriginConnectionIndex":      int32(1),
		"DestinationConnectionIndex": int32(2),
		"NewCaseValue": primitive.D{
			{Key: "$ID", Value: "case-1"},
			{Key: "$Type", Value: "Microflows$EnumerationCase"},
			{Key: "Value", Value: "true"},
		},
	})

	got, ok := flow.CaseValue.(*microflows.EnumerationCase)
	if !ok {
		t.Fatalf("expected *EnumerationCase, got %T", flow.CaseValue)
	}
	if got.Value != "true" {
		t.Fatalf("expected true branch, got %q", got.Value)
	}
}

func TestParseSequenceFlow_NewCaseValueNoCase(t *testing.T) {
	flow := parseSequenceFlow(map[string]any{
		"$ID":                "flow-1",
		"OriginPointer":      "start-1",
		"DestinationPointer": "dest-1",
		"NewCaseValue": primitive.D{
			{Key: "$ID", Value: "case-1"},
			{Key: "$Type", Value: "Microflows$NoCase"},
		},
	})

	if _, ok := flow.CaseValue.(*microflows.NoCase); !ok {
		t.Fatalf("expected *NoCase, got %T", flow.CaseValue)
	}
}

func TestParseCommitAction_ErrorHandlingTypeExplicit(t *testing.T) {
	action := parseCommitAction(map[string]any{
		"$ID":                "commit-1",
		"CommitVariableName": "Order",
		"WithEvents":         true,
		"RefreshInClient":    false,
		"ErrorHandlingType":  "Continue",
	})

	if action.ErrorHandlingType != microflows.ErrorHandlingTypeContinue {
		t.Errorf("expected Continue, got %q", action.ErrorHandlingType)
	}
	if action.CommitVariable != "Order" {
		t.Errorf("expected CommitVariable Order, got %q", action.CommitVariable)
	}
}

func TestParseCommitAction_ErrorHandlingTypeDefaultsToRollback(t *testing.T) {
	// When ErrorHandlingType is absent from BSON, the describer must still
	// emit "on error rollback" — matching Mendix Studio Pro's default.
	// Without this default, describe → exec → describe drops the suffix
	// because the writer omits the field when it equals Rollback.
	action := parseCommitAction(map[string]any{
		"$ID":                "commit-1",
		"CommitVariableName": "Order",
		"WithEvents":         false,
		"RefreshInClient":    false,
	})

	if action.ErrorHandlingType != microflows.ErrorHandlingTypeRollback {
		t.Errorf("expected default Rollback, got %q", action.ErrorHandlingType)
	}
}

func TestParseCodeActionParameterValue_MicroflowParameterValue(t *testing.T) {
	value := parseCodeActionParameterValue(map[string]any{
		"$ID":       "pmv-1",
		"$Type":     "Microflows$MicroflowParameterValue",
		"Microflow": "MxDock.Example_OpenAdminPage",
	})

	got, ok := value.(*microflows.MicroflowParameterValue)
	if !ok {
		t.Fatalf("expected *MicroflowParameterValue, got %T", value)
	}
	if got.Microflow != "MxDock.Example_OpenAdminPage" {
		t.Fatalf("expected microflow name preserved, got %q", got.Microflow)
	}
}

func TestParseCodeActionParameterType_JavaActionMicroflowParameter(t *testing.T) {
	value := parseCodeActionParameterType(map[string]any{
		"$ID":   "type-1",
		"$Type": "JavaActions$MicroflowJavaActionParameterType",
	})

	if _, ok := value.(*javaactions.MicroflowType); !ok {
		t.Fatalf("expected *javaactions.MicroflowType, got %T", value)
	}
}

func TestParseResultHandlingMappingUsesRangeForSingleObject(t *testing.T) {
	got := parseResultHandling(map[string]any{
		"$ID":                "result-handling-1",
		"ResultVariableName": "CloudApp",
		"ImportMappingCall": map[string]any{
			"ReturnValueMapping":    "CloudIntegration.IMM_CloudApp",
			"ForceSingleOccurrence": false,
			"Range": map[string]any{
				"SingleObject": true,
			},
		},
		"VariableType": map[string]any{
			"$Type":  "DataTypes$ObjectType",
			"Entity": "CloudIntegration.CloudApp",
		},
	}, "Mapping")

	rh, ok := got.(*microflows.ResultHandlingMapping)
	if !ok {
		t.Fatalf("got %T, want *microflows.ResultHandlingMapping", got)
	}
	if !rh.SingleObject {
		t.Fatal("Range.SingleObject=true must make the result object-valued")
	}
	if rh.ForceSingleOccurrence == nil || *rh.ForceSingleOccurrence {
		t.Fatalf("ForceSingleOccurrence = %v, want explicit false", rh.ForceSingleOccurrence)
	}
}

func TestSerializeRestResultHandlingPreservesForceSingleOccurrenceSeparately(t *testing.T) {
	forceSingleOccurrence := false
	doc := serializeRestResultHandling(&microflows.ResultHandlingMapping{
		BaseElement:           model.BaseElement{ID: model.ID("result-handling-1")},
		MappingID:             model.ID("CloudIntegration.IMM_CloudApp"),
		ResultEntityID:        model.ID("CloudIntegration.CloudApp"),
		ResultVariable:        "CloudApp",
		SingleObject:          true,
		ForceSingleOccurrence: &forceSingleOccurrence,
	}, "CloudApp")

	importCall, ok := bsonDMap(doc)["ImportMappingCall"].(primitive.D)
	if !ok {
		t.Fatalf("ImportMappingCall missing or wrong type: %T", bsonDMap(doc)["ImportMappingCall"])
	}
	callFields := bsonDMap(importCall)
	if got := callFields["ForceSingleOccurrence"]; got != false {
		t.Fatalf("ForceSingleOccurrence = %v, want false", got)
	}
	rangeDoc, ok := callFields["Range"].(primitive.D)
	if !ok {
		t.Fatalf("Range missing or wrong type: %T", callFields["Range"])
	}
	if got := bsonDMap(rangeDoc)["SingleObject"]; got != true {
		t.Fatalf("Range.SingleObject = %v, want true", got)
	}
	varType, ok := bsonDMap(doc)["VariableType"].(primitive.D)
	if !ok {
		t.Fatalf("VariableType missing or wrong type: %T", bsonDMap(doc)["VariableType"])
	}
	if got := bsonDMap(varType)["$Type"]; got != "DataTypes$ObjectType" {
		t.Fatalf("VariableType.$Type = %v, want DataTypes$ObjectType", got)
	}
}

func TestSerializeRestResultHandlingHttpResponseUsesObjectType(t *testing.T) {
	doc := serializeRestResultHandling(&microflows.ResultHandlingHttpResponse{
		BaseElement:  model.BaseElement{ID: model.ID("result-handling-1")},
		VariableName: "Response",
	}, "Response")

	varType, ok := bsonDMap(doc)["VariableType"].(primitive.D)
	if !ok {
		t.Fatalf("VariableType missing or wrong type: %T", bsonDMap(doc)["VariableType"])
	}
	fields := bsonDMap(varType)
	if got := fields["$Type"]; got != "DataTypes$ObjectType" {
		t.Fatalf("VariableType.$Type = %v, want DataTypes$ObjectType", got)
	}
	if got := fields["Entity"]; got != "System.HttpResponse" {
		t.Fatalf("VariableType.Entity = %v, want System.HttpResponse", got)
	}
}

func TestSerializeSortItemPreservesIndirectEntityRef(t *testing.T) {
	doc := serializeSortItem(&microflows.SortItem{
		BaseElement:            model.BaseElement{ID: model.ID("sort-1")},
		AttributeQualifiedName: "AppsCombinedView.AppView.AppCreatedDate",
		EntityRefSteps: []microflows.EntityRefStep{
			{
				Association:       "AppsCombinedView.PrivateCloudEnvironment_AppView",
				DestinationEntity: "AppsCombinedView.AppView",
			},
		},
		Direction: microflows.SortDirectionDescending,
	})

	attrRef, ok := bsonDMap(doc)["AttributeRef"].(primitive.D)
	if !ok {
		t.Fatalf("AttributeRef missing or wrong type: %T", bsonDMap(doc)["AttributeRef"])
	}
	entityRef, ok := bsonDMap(attrRef)["EntityRef"].(primitive.D)
	if !ok {
		t.Fatalf("EntityRef missing or wrong type: %T", bsonDMap(attrRef)["EntityRef"])
	}
	if got := bsonDMap(entityRef)["$Type"]; got != "DomainModels$IndirectEntityRef" {
		t.Fatalf("EntityRef.$Type = %v, want DomainModels$IndirectEntityRef", got)
	}
	steps, ok := bsonDMap(entityRef)["Steps"].(primitive.A)
	if !ok || len(steps) != 2 {
		t.Fatalf("Steps = %#v, want marker plus one step", bsonDMap(entityRef)["Steps"])
	}
	step, ok := steps[1].(primitive.D)
	if !ok {
		t.Fatalf("step type = %T, want primitive.D", steps[1])
	}
	stepFields := bsonDMap(step)
	if got := stepFields["Association"]; got != "AppsCombinedView.PrivateCloudEnvironment_AppView" {
		t.Fatalf("Association = %v", got)
	}
	if got := stepFields["DestinationEntity"]; got != "AppsCombinedView.AppView" {
		t.Fatalf("DestinationEntity = %v", got)
	}
}

func TestParseSortItemsPreservesIndirectEntityRef(t *testing.T) {
	got := parseSortItems(map[string]any{
		"NewSortings": map[string]any{
			"Sortings": []any{
				int32(2),
				map[string]any{
					"$ID":       "sort-1",
					"$Type":     "Microflows$RetrieveSorting",
					"SortOrder": "Descending",
					"AttributeRef": map[string]any{
						"$Type":     "DomainModels$AttributeRef",
						"Attribute": "AppsCombinedView.AppView.AppCreatedDate",
						"EntityRef": map[string]any{
							"$Type": "DomainModels$IndirectEntityRef",
							"Steps": []any{
								int32(2),
								map[string]any{
									"$Type":             "DomainModels$EntityRefStep",
									"Association":       "AppsCombinedView.PrivateCloudEnvironment_AppView",
									"DestinationEntity": "AppsCombinedView.AppView",
								},
							},
						},
					},
				},
			},
		},
	})

	if len(got) != 1 {
		t.Fatalf("got %d sort items, want 1", len(got))
	}
	if steps := got[0].EntityRefSteps; len(steps) != 1 || steps[0].Association != "AppsCombinedView.PrivateCloudEnvironment_AppView" || steps[0].DestinationEntity != "AppsCombinedView.AppView" {
		t.Fatalf("EntityRefSteps = %#v", steps)
	}
}

func bsonDMap(doc primitive.D) map[string]any {
	out := make(map[string]any, len(doc))
	for _, elem := range doc {
		out[elem.Key] = elem.Value
	}
	return out
}
