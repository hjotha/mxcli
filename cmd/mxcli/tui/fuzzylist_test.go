package tui

import "testing"

func TestFuzzyListFilterTypePrefix(t *testing.T) {
	items := []PickerItem{
		{QName: "MyModule.ProcessOrder", NodeType: "Microflow"},
		{QName: "MyModule.OnClick", NodeType: "Nanoflow"},
		{QName: "MyModule.ApprovalFlow", NodeType: "Workflow"},
		{QName: "MyModule.OrderPage", NodeType: "Page"},
		{QName: "MyModule.Customer", NodeType: "Entity"},
	}
	fl := NewFuzzyList(items, 10)

	tests := []struct {
		query         string
		expectedCount int
		expectType    string // if set, all matches must have this NodeType
	}{
		{"", 5, ""},
		{"mf:", 1, "Microflow"},
		{"nf:", 1, "Nanoflow"},
		{"wf:", 1, "Workflow"},
		{"pg:", 1, "Page"},
		{"en:", 1, "Entity"},
		{"mf:process", 1, "Microflow"},
		{"mf:nonexistent", 0, ""},
		{"microflow:", 1, "Microflow"},
		// No colon: plain fuzzy match across all types
		{"order", 2, ""}, // ProcessOrder + OrderPage
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			fl.Filter(tt.query)
			if len(fl.Matches) != tt.expectedCount {
				names := make([]string, len(fl.Matches))
				for i, m := range fl.Matches {
					names[i] = m.item.QName
				}
				t.Errorf("Filter(%q): got %d matches %v, want %d", tt.query, len(fl.Matches), names, tt.expectedCount)
			}
			if tt.expectType != "" {
				for _, m := range fl.Matches {
					if m.item.NodeType != tt.expectType {
						t.Errorf("Filter(%q): got NodeType %q, want %q", tt.query, m.item.NodeType, tt.expectType)
					}
				}
			}
		})
	}
}
