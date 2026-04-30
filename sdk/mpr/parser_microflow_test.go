// SPDX-License-Identifier: Apache-2.0

package mpr

import (
	"bytes"
	"testing"

	"github.com/mendixlabs/mxcli/sdk/microflows"
	"go.mongodb.org/mongo-driver/bson"
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

func TestParseActionActivityPreservesWebServiceActionRawBSONOrder(t *testing.T) {
	rawAction := primitive.D{
		{Key: "$ID", Value: "web-service-action-ordered"},
		{Key: "$Type", Value: "Microflows$CallWebServiceAction"},
		{Key: "ImportedService", Value: "SyntheticSOAP.OrderService"},
		{Key: "OperationName", Value: "FetchItemsByTenant"},
		{Key: "TimeOutExpression", Value: "30"},
		{Key: "NewResultHandling", Value: primitive.D{
			{Key: "$Type", Value: "Microflows$WebServiceOperationResultHandling"},
			{Key: "ResultVariableName", Value: "SampleResponse"},
		}},
	}
	expectedRaw, err := bson.Marshal(rawAction)
	if err != nil {
		t.Fatal(err)
	}

	activity := parseActionActivity(map[string]any{
		"$ID":    "activity-with-web-service-action",
		"$Type":  "Microflows$ActionActivity",
		"Action": rawAction,
	})
	action, ok := activity.Action.(*microflows.WebServiceCallAction)
	if !ok {
		t.Fatalf("Action = %T, want *WebServiceCallAction", activity.Action)
	}
	if !bytes.Equal(action.RawBSON, expectedRaw) {
		t.Fatalf("RawBSON was not preserved byte-for-byte")
	}

	serializedRaw, err := bson.Marshal(serializeWebServiceCallAction(action))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(serializedRaw, expectedRaw) {
		t.Fatalf("serialized raw BSON was not preserved byte-for-byte")
	}
}

func TestParseWebServiceActionFallsBackToRawBSONForUnsupportedFields(t *testing.T) {
	action := parseWebServiceCallAction(map[string]any{
		"$ID":             "soap-action-with-simple-request",
		"$Type":           "Microflows$CallWebServiceAction",
		"ImportedService": "SyntheticSOAP.OrderService",
		"OperationName":   "SubmitOrder",
		"RequestBodyHandling": map[string]any{
			"$Type": "Microflows$SimpleRequestHandling",
			"ParameterMappings": []any{
				int32(2),
				map[string]any{
					"$Type":    "Microflows$WebServiceOperationSimpleParameterMapping",
					"Argument": "$OrderID",
				},
			},
		},
	})

	if len(action.RawBSON) == 0 {
		t.Fatal("RawBSON was empty for unsupported SOAP request details")
	}
	serialized := serializeWebServiceCallAction(action)
	if got := bsonGetKey(serialized, "RequestBodyHandling"); got == nil {
		t.Fatalf("RequestBodyHandling was not preserved in raw fallback: %#v", serialized)
	}
}
