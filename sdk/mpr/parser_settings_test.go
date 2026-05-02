// SPDX-License-Identifier: Apache-2.0

package mpr

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestParseLanguageSettings_Languages verifies that Languages array items stored
// as primitive.D (the BSON decoded type) are correctly parsed via extractBsonMap.
// This is the fix for issue #480: bare .(map[string]any) assertions always fail
// on primitive.D values, so extractBsonMap must be used instead.
func TestParseLanguageSettings_Languages(t *testing.T) {
	raw := map[string]any{
		"$ID":                 "settings-lang-1",
		"$Type":               "Settings$LanguageSettings",
		"DefaultLanguageCode": "en_US",
		"Languages": primitive.A{
			int32(2),
			primitive.D{
				{Key: "$ID", Value: "lang-1"},
				{Key: "$Type", Value: "Texts$Language"},
				{Key: "Code", Value: "en_US"},
				{Key: "CheckCompleteness", Value: true},
				{Key: "CustomDateFormat", Value: "MM/dd/yyyy"},
				{Key: "CustomDateTimeFormat", Value: "MM/dd/yyyy HH:mm"},
				{Key: "CustomTimeFormat", Value: "HH:mm"},
			},
			primitive.D{
				{Key: "$ID", Value: "lang-2"},
				{Key: "$Type", Value: "Texts$Language"},
				{Key: "Code", Value: "fr_FR"},
				{Key: "CheckCompleteness", Value: false},
			},
		},
	}

	ls := parseLanguageSettings(raw)

	if ls.DefaultLanguageCode != "en_US" {
		t.Errorf("DefaultLanguageCode = %q, want %q", ls.DefaultLanguageCode, "en_US")
	}
	if len(ls.Languages) != 2 {
		t.Fatalf("len(Languages) = %d, want 2", len(ls.Languages))
	}

	en := ls.Languages[0]
	if en.Code != "en_US" {
		t.Errorf("Languages[0].Code = %q, want %q", en.Code, "en_US")
	}
	if !en.CheckCompleteness {
		t.Errorf("Languages[0].CheckCompleteness = false, want true")
	}
	if en.CustomDateFormat != "MM/dd/yyyy" {
		t.Errorf("Languages[0].CustomDateFormat = %q, want %q", en.CustomDateFormat, "MM/dd/yyyy")
	}
	if en.CustomDateTimeFormat != "MM/dd/yyyy HH:mm" {
		t.Errorf("Languages[0].CustomDateTimeFormat = %q, want %q", en.CustomDateTimeFormat, "MM/dd/yyyy HH:mm")
	}
	if en.CustomTimeFormat != "HH:mm" {
		t.Errorf("Languages[0].CustomTimeFormat = %q, want %q", en.CustomTimeFormat, "HH:mm")
	}

	fr := ls.Languages[1]
	if fr.Code != "fr_FR" {
		t.Errorf("Languages[1].Code = %q, want %q", fr.Code, "fr_FR")
	}
	if fr.CheckCompleteness {
		t.Errorf("Languages[1].CheckCompleteness = true, want false")
	}
}

// TestParseLanguageSettings_EmptyLanguages verifies that an absent or empty
// Languages array results in a nil/empty slice without panicking.
func TestParseLanguageSettings_EmptyLanguages(t *testing.T) {
	raw := map[string]any{
		"$ID":                 "settings-lang-2",
		"$Type":               "Settings$LanguageSettings",
		"DefaultLanguageCode": "en_US",
		"Languages":           primitive.A{int32(2)},
	}

	ls := parseLanguageSettings(raw)
	if len(ls.Languages) != 0 {
		t.Errorf("len(Languages) = %d, want 0", len(ls.Languages))
	}
}
