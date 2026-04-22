// SPDX-License-Identifier: Apache-2.0

package catalog

import "testing"

func TestExtractAttrName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"three parts", "Module.Entity.Attribute", "Attribute"},
		{"four parts", "Module.Entity.Attribute.Sub", "Sub"},
		{"two parts", "Module.Entity", ""},
		{"single part", "Single", ""},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractAttrName(tt.input); got != tt.want {
				t.Errorf("extractAttrName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
