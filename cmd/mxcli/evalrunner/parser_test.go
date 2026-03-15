// SPDX-License-Identifier: Apache-2.0

package evalrunner

import (
	"testing"
	"time"
)

func TestSplitFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantFM   string
		wantBody string
	}{
		{
			name:     "with frontmatter",
			input:    "---\nid: APP-001\ncategory: App/Crud\n---\n# Title\nBody text",
			wantFM:   "id: APP-001\ncategory: App/Crud",
			wantBody: "# Title\nBody text",
		},
		{
			name:     "no frontmatter",
			input:    "# Title\nBody text",
			wantFM:   "",
			wantBody: "# Title\nBody text",
		},
		{
			name:     "empty frontmatter",
			input:    "---\n\n---\n# Title",
			wantFM:   "",
			wantBody: "# Title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, body, err := splitFrontmatter(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if fm != tt.wantFM {
				t.Errorf("frontmatter = %q, want %q", fm, tt.wantFM)
			}
			if body != tt.wantBody {
				t.Errorf("body = %q, want %q", body, tt.wantBody)
			}
		})
	}
}

func TestParseEvalContent(t *testing.T) {
	content := `---
id: APP-001
category: App/Crud
tags: [entity, crud, pages]
timeout: 5m
---

# APP-001: Bookstore Inventory

## Prompt
Create an app to manage my bookstore inventory.

## Expected Outcome
Domain model with Book entity, CRUD pages.

## Checks
- entity_exists: "*.Book"
- entity_has_attribute: "*.Book.Title String"
- page_exists: "*Overview*"
- navigation_has_item: true
- mx_check_passes: true

## Acceptance Criteria
- Book entity has all specified attributes
- Overview page with data grid

## Iteration

### Prompt
Add a category field to the books.

### Checks
- entity_has_attribute: "*.Book.Category"

### Acceptance Criteria
- Category attribute added
`

	test, err := parseEvalContent(content)
	if err != nil {
		t.Fatalf("parseEvalContent failed: %v", err)
	}

	// Check metadata
	if test.ID != "APP-001" {
		t.Errorf("ID = %q, want APP-001", test.ID)
	}
	if test.Category != "App/Crud" {
		t.Errorf("Category = %q, want App/Crud", test.Category)
	}
	if len(test.Tags) != 3 {
		t.Errorf("Tags = %v, want 3 tags", test.Tags)
	}
	if test.Timeout != 5*time.Minute {
		t.Errorf("Timeout = %v, want 5m", test.Timeout)
	}
	if test.Title != "APP-001: Bookstore Inventory" {
		t.Errorf("Title = %q, want 'APP-001: Bookstore Inventory'", test.Title)
	}

	// Check prompt
	if test.Prompt != "Create an app to manage my bookstore inventory." {
		t.Errorf("Prompt = %q", test.Prompt)
	}

	// Check checks
	if len(test.Checks) != 5 {
		t.Fatalf("len(Checks) = %d, want 5", len(test.Checks))
	}
	if test.Checks[0].Type != "entity_exists" || test.Checks[0].Args != "*.Book" {
		t.Errorf("Checks[0] = %+v", test.Checks[0])
	}
	if test.Checks[1].Type != "entity_has_attribute" || test.Checks[1].Args != "*.Book.Title String" {
		t.Errorf("Checks[1] = %+v", test.Checks[1])
	}
	if test.Checks[3].Type != "navigation_has_item" || test.Checks[3].Args != "true" {
		t.Errorf("Checks[3] = %+v", test.Checks[3])
	}

	// Check criteria
	if len(test.Criteria) != 2 {
		t.Errorf("len(Criteria) = %d, want 2", len(test.Criteria))
	}

	// Check iteration
	if test.Iteration == nil {
		t.Fatal("Iteration is nil")
	}
	if test.Iteration.Prompt != "Add a category field to the books." {
		t.Errorf("Iteration.Prompt = %q", test.Iteration.Prompt)
	}
	if len(test.Iteration.Checks) != 1 {
		t.Fatalf("len(Iteration.Checks) = %d, want 1", len(test.Iteration.Checks))
	}
	if test.Iteration.Checks[0].Type != "entity_has_attribute" {
		t.Errorf("Iteration.Checks[0].Type = %q", test.Iteration.Checks[0].Type)
	}
	if len(test.Iteration.Criteria) != 1 {
		t.Errorf("len(Iteration.Criteria) = %d, want 1", len(test.Iteration.Criteria))
	}
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		pattern string
		want    bool
	}{
		{"exact match", "MyModule.Book", "MyModule.Book", true},
		{"wildcard prefix", "MyModule.Book", "*.Book", true},
		{"wildcard both", "MyModule.Book_Overview", "*Overview*", true},
		{"wildcard suffix", "MyModule.Book_Overview", "*Overview", true},
		{"no match", "MyModule.Book", "*.Page", false},
		{"case insensitive", "MyModule.Book", "*.book", true},
		{"middle wildcard", "MyModule.Book_Overview", "*Book*Overview*", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchPattern(tt.input, tt.pattern)
			if got != tt.want {
				t.Errorf("matchPattern(%q, %q) = %v, want %v", tt.input, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestParseChecks(t *testing.T) {
	input := `- entity_exists: "*.Book"
- entity_has_attribute: "*.Book.Title String"
- page_exists: "*Overview*"
- navigation_has_item: true
`
	checks := parseChecks(input)
	if len(checks) != 4 {
		t.Fatalf("len(checks) = %d, want 4", len(checks))
	}

	if checks[0].Type != "entity_exists" || checks[0].Args != "*.Book" {
		t.Errorf("checks[0] = %+v", checks[0])
	}
	if checks[3].Type != "navigation_has_item" || checks[3].Args != "true" {
		t.Errorf("checks[3] = %+v", checks[3])
	}
}

func TestFindAttributeInDescribe(t *testing.T) {
	describe := `CREATE PERSISTENT ENTITY MyModule.Book (
  Title: String(200),
  Author: String(200),
  ISBN: String(50),
  Price: Decimal,
  StockQuantity: Integer
);
`
	tests := []struct {
		attr     string
		wantOK   bool
		wantType string
	}{
		{"Title", true, "String"},
		{"Price", true, "Decimal"},
		{"StockQuantity", true, "Integer"},
		{"Missing", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.attr, func(t *testing.T) {
			ok, typ := findAttributeInDescribe(describe, tt.attr)
			if ok != tt.wantOK {
				t.Errorf("found = %v, want %v", ok, tt.wantOK)
			}
			if typ != tt.wantType {
				t.Errorf("type = %q, want %q", typ, tt.wantType)
			}
		})
	}
}

func TestParseEvalFile(t *testing.T) {
	// Test parsing the actual eval-1.md file
	test, err := ParseEvalFile("../../../docs/14-eval/eval-1.md")
	if err != nil {
		t.Fatalf("ParseEvalFile failed: %v", err)
	}

	if test.ID != "APP-001" {
		t.Errorf("ID = %q, want APP-001", test.ID)
	}
	if test.Category != "App/Crud" {
		t.Errorf("Category = %q, want App/Crud", test.Category)
	}
	if len(test.Checks) == 0 {
		t.Error("no checks parsed")
	}
	if test.Iteration == nil {
		t.Error("no iteration parsed")
	}
}
