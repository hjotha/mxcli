// SPDX-License-Identifier: Apache-2.0

package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const maxJumperHistory = 50

func jumperHistoryPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".mxcli", "tui-jumper-history.json")
}

// LoadJumperHistory returns recent jumper queries, most recent first.
func LoadJumperHistory() []string {
	data, err := os.ReadFile(jumperHistoryPath())
	if err != nil {
		return nil
	}
	var entries []string
	if err := json.Unmarshal(data, &entries); err != nil {
		Trace("jumperhistory: unmarshal error: %v", err)
		return nil
	}
	return entries
}

// AddJumperHistoryEntry adds a query to the front of jumper history (deduped, max 50).
func AddJumperHistoryEntry(query string) {
	query = strings.TrimSpace(query)
	if query == "" {
		return
	}
	entries := LoadJumperHistory()
	entries = slices.DeleteFunc(entries, func(e string) bool { return e == query })
	entries = append([]string{query}, entries...)
	if len(entries) > maxJumperHistory {
		entries = entries[:maxJumperHistory]
	}
	if err := os.MkdirAll(filepath.Dir(jumperHistoryPath()), 0755); err != nil {
		Trace("jumperhistory: mkdir error: %v", err)
		return
	}
	data, err := json.Marshal(entries)
	if err != nil {
		Trace("jumperhistory: marshal error: %v", err)
		return
	}
	if err := os.WriteFile(jumperHistoryPath(), data, 0644); err != nil {
		Trace("jumperhistory: write error: %v", err)
	}
}
