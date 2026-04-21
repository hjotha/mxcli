// SPDX-License-Identifier: Apache-2.0

package mpr

import "testing"

func TestParseImportObjectMappingElement_FindWithCreateBackupBecomesFindOrCreate(t *testing.T) {
	elem := parseImportObjectMappingElement(map[string]any{
		"$ID":                  "ignored",
		"$Type":                "ImportMappings$ObjectMappingElement",
		"Entity":               "MyModule.Pet",
		"ObjectHandling":       "Find",
		"ObjectHandlingBackup": "Create",
	})

	if elem == nil {
		t.Fatal("expected element, got nil")
	}
	if elem.ObjectHandling != "FindOrCreate" {
		t.Fatalf("ObjectHandling = %q, want %q", elem.ObjectHandling, "FindOrCreate")
	}
}
