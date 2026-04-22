// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// runCatalog executes the catalog subtree with the given args.
func runCatalog(t *testing.T, args ...string) (string, error) {
	t.Helper()
	for _, c := range []*cobra.Command{catalogSearchCmd, catalogShowCmd} {
		resetCmdFlags(c)
	}

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs(append([]string{"catalog"}, args...))
	err := rootCmd.ExecuteContext(context.Background())
	return out.String(), err
}

func TestCatalogSearch_NoAuth(t *testing.T) {
	withTestHome(t)

	_, err := runCatalog(t, "search", "test")
	if err == nil {
		t.Fatal("expected error when not authenticated")
	}
	// Error message should hint at auth login
	errMsg := err.Error()
	if !strings.Contains(errMsg, "auth login") && !strings.Contains(errMsg, "credential") && !strings.Contains(errMsg, "no credential") {
		t.Errorf("error should mention auth or credential: %v", err)
	}
}

func TestCatalogShow_NoAuth(t *testing.T) {
	withTestHome(t)

	_, err := runCatalog(t, "show", "test-uuid")
	if err == nil {
		t.Fatal("expected error when not authenticated")
	}
	// Error message should hint at auth
	errMsg := err.Error()
	if !strings.Contains(errMsg, "auth login") && !strings.Contains(errMsg, "credential") && !strings.Contains(errMsg, "no credential") {
		t.Errorf("error should mention auth or credential: %v", err)
	}
}

func TestCatalogSearch_RequiresQuery(t *testing.T) {
	withTestHome(t)

	_, err := runCatalog(t, "search")
	if err == nil {
		t.Fatal("expected error when query is missing")
	}
	// Should mention that query argument is required
	errMsg := err.Error()
	if !strings.Contains(errMsg, "requires") && !strings.Contains(errMsg, "arg") {
		t.Errorf("error should mention missing argument: %v", err)
	}
}

func TestCatalogShow_RequiresUUID(t *testing.T) {
	withTestHome(t)

	_, err := runCatalog(t, "show")
	if err == nil {
		t.Fatal("expected error when UUID is missing")
	}
	// Should mention that UUID argument is required
	errMsg := err.Error()
	if !strings.Contains(errMsg, "requires") && !strings.Contains(errMsg, "arg") {
		t.Errorf("error should mention missing argument: %v", err)
	}
}
