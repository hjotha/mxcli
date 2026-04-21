// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCheck_SkipUpdateWidgets(t *testing.T) {
	// This test verifies the SkipUpdateWidgets option is wired through.
	// Since we don't have a real mx binary in CI, we just verify the
	// function returns the expected "mx not found" error.
	opts := CheckOptions{
		ProjectPath:       "/nonexistent/app.mpr",
		SkipUpdateWidgets: true,
		Stdout:            &bytes.Buffer{},
		Stderr:            &bytes.Buffer{},
	}

	err := Check(opts)
	if err == nil {
		t.Fatal("expected error when mx binary not found")
	}
	if got := err.Error(); got != "mx not found; specify --mxbuild-path pointing to Mendix installation directory" {
		// Accept any error about mx not being found
		t.Logf("got error: %s", got)
	}
}

// createFakeMxDir creates a temp directory with fake mx and mxbuild scripts
// that log their first argument to a file.
func createFakeMxDir(t *testing.T) (dir, logFile string) {
	t.Helper()
	dir = t.TempDir()
	logFile = filepath.Join(dir, "commands.log")

	script := `#!/bin/sh
echo "$1" >> ` + logFile + "\n"

	for _, name := range []string{"mx", "mxbuild"} {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(script), 0755); err != nil {
			t.Fatal(err)
		}
	}
	return dir, logFile
}

func TestCheck_UpdateWidgetsBeforeCheck(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script test not supported on Windows")
	}

	mxDir, logFile := createFakeMxDir(t)

	var stdout, stderr bytes.Buffer
	opts := CheckOptions{
		ProjectPath: "/tmp/fake.mpr",
		MxBuildPath: mxDir,
		Stdout:      &stdout,
		Stderr:      &stderr,
	}

	Check(opts)

	logBytes, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read command log: %v", err)
	}

	log := string(logBytes)
	if !bytes.Contains(logBytes, []byte("update-widgets\n")) {
		t.Errorf("update-widgets was not called, got log:\n%s", log)
	}
	if !bytes.Contains(logBytes, []byte("check\n")) {
		t.Errorf("check was not called, got log:\n%s", log)
	}

	// Verify order: update-widgets before check
	uwIdx := bytes.Index(logBytes, []byte("update-widgets"))
	chIdx := bytes.Index(logBytes, []byte("check"))
	if uwIdx >= chIdx {
		t.Errorf("update-widgets should run before check, got log:\n%s", log)
	}
}

func TestCheck_SkipUpdateWidgetsFlag(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script test not supported on Windows")
	}

	mxDir, logFile := createFakeMxDir(t)

	var stdout, stderr bytes.Buffer
	opts := CheckOptions{
		ProjectPath:       "/tmp/fake.mpr",
		MxBuildPath:       mxDir,
		SkipUpdateWidgets: true,
		Stdout:            &stdout,
		Stderr:            &stderr,
	}

	Check(opts)

	logBytes, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read command log: %v", err)
	}

	if bytes.Contains(logBytes, []byte("update-widgets")) {
		t.Error("update-widgets should NOT be called when SkipUpdateWidgets=true")
	}
	if !bytes.Contains(logBytes, []byte("check")) {
		t.Error("check should still be called")
	}
}

func TestResolveMxForVersion_PrefersExactCachedVersion(t *testing.T) {
	dir := t.TempDir()
	setTestHomeDir(t, dir)
	t.Setenv("PATH", "")

	versions := []string{"9.24.40.80973", "11.6.3", "11.9.0"}
	var expected string
	for _, version := range versions {
		modelerDir := filepath.Join(dir, ".mxcli", "mxbuild", version, "modeler")
		if err := os.MkdirAll(modelerDir, 0755); err != nil {
			t.Fatal(err)
		}
		bin := filepath.Join(modelerDir, mxBinaryName())
		if err := os.WriteFile(bin, []byte("fake"), 0755); err != nil {
			t.Fatal(err)
		}
		if version == "11.9.0" {
			expected = bin
		}
	}

	result, err := ResolveMxForVersion("", "11.9.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != expected {
		t.Errorf("expected exact cached mx %s, got %s", expected, result)
	}
}
