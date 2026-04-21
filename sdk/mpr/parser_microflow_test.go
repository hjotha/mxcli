// SPDX-License-Identifier: Apache-2.0

package mpr

import (
	"testing"

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
