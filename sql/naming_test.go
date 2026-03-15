// SPDX-License-Identifier: Apache-2.0

package sql

import "testing"

func TestTableToEntityName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"employees", "Employee"},
		{"order_items", "OrderItem"},
		{"categories", "Category"},
		{"addresses", "Address"},
		{"status", "Status"},     // don't singularize words ending in "us"
		{"Employee", "Employee"}, // already singular PascalCase
		{"users", "User"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := TableToEntityName(tt.input)
			if got != tt.want {
				t.Errorf("TableToEntityName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestColumnToAttributeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"first_name", "FirstName"},
		{"employee_id", "EmployeeId"},
		{"id", "Id"},
		{"Name", "Name"},
		{"created_at", "CreatedAt"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ColumnToAttributeName(tt.input)
			if got != tt.want {
				t.Errorf("ColumnToAttributeName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
