package tui

import (
	"testing"
)

func TestAgentParseChanges(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []agentExecChange
	}{
		{
			name:     "empty string returns nil",
			input:    "",
			expected: nil,
		},
		{
			name:     "no matching lines returns nil",
			input:    "no changes here",
			expected: nil,
		},
		{
			name:     "false positive: missing type keyword",
			input:    "Removed trailing whitespace",
			expected: nil,
		},
		{
			name:     "false positive: unknown type keyword",
			input:    "Created foobar Mod.E",
			expected: nil,
		},
		{
			name:  "single Created entity line",
			input: "Created entity Mod.E",
			expected: []agentExecChange{
				{Action: "created", Target: "entity Mod.E"},
			},
		},
		{
			name:  "case insensitive matching",
			input: "created entity Mod.E",
			expected: []agentExecChange{
				{Action: "created", Target: "entity Mod.E"},
			},
		},
		{
			name:  "multi-line with three verbs",
			input: "Created entity Mod.E\nModified attribute Mod.E.Name\nDropped association Mod.A_B",
			expected: []agentExecChange{
				{Action: "created", Target: "entity Mod.E"},
				{Action: "modified", Target: "attribute Mod.E.Name"},
				{Action: "dropped", Target: "association Mod.A_B"},
			},
		},
		{
			name:  "Deleted page",
			input: "Deleted page Mod.P",
			expected: []agentExecChange{
				{Action: "deleted", Target: "page Mod.P"},
			},
		},
		{
			name:  "Added microflow",
			input: "Added microflow Mod.MF",
			expected: []agentExecChange{
				{Action: "added", Target: "microflow Mod.MF"},
			},
		},
		{
			name:  "Removed constant",
			input: "Removed constant Mod.C",
			expected: []agentExecChange{
				{Action: "removed", Target: "constant Mod.C"},
			},
		},
		{
			name:  "image collection type",
			input: "Created image collection Mod.Icons",
			expected: []agentExecChange{
				{Action: "created", Target: "image collection Mod.Icons"},
			},
		},
		{
			name:  "java action type",
			input: "Added java action Mod.JA_Do",
			expected: []agentExecChange{
				{Action: "added", Target: "java action Mod.JA_Do"},
			},
		},
		{
			name:  "mixed content extracts only matching lines",
			input: "Some info\nCreated entity Mod.E\nRemoved trailing whitespace\nMore info",
			expected: []agentExecChange{
				{Action: "created", Target: "entity Mod.E"},
			},
		},
		{
			name:  "workflow type",
			input: "Modified workflow Mod.WF_Approval",
			expected: []agentExecChange{
				{Action: "modified", Target: "workflow Mod.WF_Approval"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := agentParseChanges(tc.input)
			if tc.expected == nil {
				if got != nil {
					t.Errorf("expected nil, got %v", got)
				}
				return
			}
			if len(got) != len(tc.expected) {
				t.Fatalf("expected %d changes, got %d: %v", len(tc.expected), len(got), got)
			}
			for i, want := range tc.expected {
				if got[i].Action != want.Action {
					t.Errorf("change[%d].Action = %q, want %q", i, got[i].Action, want.Action)
				}
				if got[i].Target != want.Target {
					t.Errorf("change[%d].Target = %q, want %q", i, got[i].Target, want.Target)
				}
			}
		})
	}
}

func TestBuildDropCmd(t *testing.T) {
	tests := []struct {
		nodeType string
		qname    string
		expected string
	}{
		{"entity", "Mod.E", "DROP ENTITY Mod.E"},
		{"association", "Mod.A_B", "DROP ASSOCIATION Mod.A_B"},
		{"enumeration", "Mod.Enum", "DROP ENUMERATION Mod.Enum"},
		{"constant", "Mod.C", "DROP CONSTANT Mod.C"},
		{"microflow", "Mod.MF", "DROP MICROFLOW Mod.MF"},
		{"page", "Mod.P", "DROP PAGE Mod.P"},
		{"snippet", "Mod.S", "DROP SNIPPET Mod.S"},
		{"module", "Mod.Sub", "DROP MODULE Mod"},
		{"module", "Mod", "DROP MODULE Mod"},
		{"workflow", "Mod.WF", "DROP WORKFLOW Mod.WF"},
		{"imagecollection", "Mod.IC", "DROP IMAGE COLLECTION Mod.IC"},
		{"javaaction", "Mod.JA", "DROP JAVA ACTION Mod.JA"},
		{"unknown", "Mod.X", ""},
	}

	for _, tc := range tests {
		t.Run(tc.nodeType+"_"+tc.qname, func(t *testing.T) {
			got := buildDropCmd(tc.nodeType, tc.qname)
			if got != tc.expected {
				t.Errorf("buildDropCmd(%q, %q) = %q, want %q", tc.nodeType, tc.qname, got, tc.expected)
			}
		})
	}
}
