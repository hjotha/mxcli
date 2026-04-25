// SPDX-License-Identifier: Apache-2.0

package mpr

import (
	"testing"

	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/pages"
	"go.mongodb.org/mongo-driver/bson"
)

// TestPageClientAction_Variable_NonNull verifies that Forms$PageParameterMapping
// always includes a non-null Forms$PageVariable in the Variable field (issue #295).
// Studio Pro's set_Variable property setter requires a non-null PageVariable and
// throws ArgumentNullException when it receives nil.
func TestPageClientAction_Variable_NonNull(t *testing.T) {
	action := &pages.PageClientAction{
		BaseElement: model.BaseElement{ID: "action-id"},
		PageName:    "AuditTrail.Log_View",
		ParameterMappings: []*pages.PageClientParameterMapping{
			{
				BaseElement:   model.BaseElement{ID: "mapping-id"},
				ParameterName: "Log",
				Variable:      "$currentObject",
			},
		},
	}

	doc := serializeClientAction(action)
	if doc == nil {
		t.Fatal("serializeClientAction returned nil")
	}

	// Navigate to FormSettings.ParameterMappings[1] (index 0 is the count int32)
	var formSettings bson.D
	for _, e := range doc {
		if e.Key == "FormSettings" {
			formSettings, _ = e.Value.(bson.D)
		}
	}
	if formSettings == nil {
		t.Fatal("FormSettings is nil")
	}

	var paramMappings bson.A
	for _, e := range formSettings {
		if e.Key == "ParameterMappings" {
			paramMappings, _ = e.Value.(bson.A)
		}
	}
	if len(paramMappings) < 2 {
		t.Fatalf("ParameterMappings: want at least 2 elements (count + mapping), got %d", len(paramMappings))
	}

	// Element 0 is the int32 count; element 1 is the first mapping
	mapping, ok := paramMappings[1].(bson.D)
	if !ok {
		t.Fatalf("ParameterMappings[1] is not bson.D, got %T", paramMappings[1])
	}

	var bsonType, argument string
	var variable any
	for _, e := range mapping {
		switch e.Key {
		case "$Type":
			bsonType, _ = e.Value.(string)
		case "Argument":
			argument, _ = e.Value.(string)
		case "Variable":
			variable = e.Value
		}
	}

	if bsonType != "Forms$PageParameterMapping" {
		t.Errorf("$Type = %q, want %q", bsonType, "Forms$PageParameterMapping")
	}
	if argument != "$currentObject" {
		t.Errorf("Argument = %q, want %q", argument, "$currentObject")
	}
	if variable == nil {
		t.Fatal("Variable is nil — Studio Pro requires a non-null Forms$PageVariable (issue #295)")
	}

	varDoc, ok := variable.(bson.D)
	if !ok {
		t.Fatalf("Variable is not bson.D, got %T", variable)
	}

	var varType string
	for _, e := range varDoc {
		if e.Key == "$Type" {
			varType, _ = e.Value.(string)
		}
	}
	if varType != "Forms$PageVariable" {
		t.Errorf("Variable.$Type = %q, want %q", varType, "Forms$PageVariable")
	}
}

// TestPageClientAction_NoParams_NoMappings verifies that a PageClientAction
// without parameter mappings serializes correctly (no regression).
func TestPageClientAction_NoParams_NoMappings(t *testing.T) {
	action := &pages.PageClientAction{
		BaseElement: model.BaseElement{ID: "action-id"},
		PageName:    "Sales.Customer_Overview",
	}

	doc := serializeClientAction(action)
	if doc == nil {
		t.Fatal("serializeClientAction returned nil")
	}

	var bsonType string
	for _, e := range doc {
		if e.Key == "$Type" {
			bsonType, _ = e.Value.(string)
		}
	}
	if bsonType != "Forms$FormAction" {
		t.Errorf("$Type = %q, want %q", bsonType, "Forms$FormAction")
	}
}

// TestPageClientAction_MultipleParams_AllHaveVariable verifies that each
// mapping in a multi-param action has a non-null Variable field.
func TestPageClientAction_MultipleParams_AllHaveVariable(t *testing.T) {
	action := &pages.PageClientAction{
		BaseElement: model.BaseElement{ID: "action-id"},
		PageName:    "Sales.Order_Detail",
		ParameterMappings: []*pages.PageClientParameterMapping{
			{
				BaseElement:   model.BaseElement{ID: "m1"},
				ParameterName: "Order",
				Variable:      "$Order",
			},
			{
				BaseElement:   model.BaseElement{ID: "m2"},
				ParameterName: "Customer",
				Variable:      "$Customer",
			},
		},
	}

	doc := serializeClientAction(action)

	var formSettings bson.D
	for _, e := range doc {
		if e.Key == "FormSettings" {
			formSettings, _ = e.Value.(bson.D)
		}
	}

	var paramMappings bson.A
	for _, e := range formSettings {
		if e.Key == "ParameterMappings" {
			paramMappings, _ = e.Value.(bson.A)
		}
	}

	// Index 0 is the int32 count; indices 1 and 2 are the mappings
	for i := 1; i <= 2; i++ {
		m, ok := paramMappings[i].(bson.D)
		if !ok {
			t.Fatalf("ParameterMappings[%d] is not bson.D", i)
		}
		var varVal any
		for _, e := range m {
			if e.Key == "Variable" {
				varVal = e.Value
			}
		}
		if varVal == nil {
			t.Errorf("ParameterMappings[%d].Variable is nil — must be non-null Forms$PageVariable", i)
		}
	}
}
