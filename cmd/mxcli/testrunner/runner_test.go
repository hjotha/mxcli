// SPDX-License-Identifier: Apache-2.0

package testrunner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureDockerStack_CreatesComposeFiles(t *testing.T) {
	projectDir := t.TempDir()
	projectPath := filepath.Join(projectDir, "test-source.mpr")
	dockerDir := filepath.Join(projectDir, ".docker")

	var buf strings.Builder
	if err := ensureDockerStack(projectPath, dockerDir, &buf); err != nil {
		t.Fatalf("ensureDockerStack failed: %v", err)
	}

	for _, path := range []string{
		filepath.Join(dockerDir, "docker-compose.yml"),
		filepath.Join(dockerDir, ".env"),
		filepath.Join(dockerDir, ".env.example"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s to exist", path)
		}
	}

	if err := ensureDockerStack(projectPath, dockerDir, &buf); err != nil {
		t.Fatalf("ensureDockerStack should be idempotent: %v", err)
	}
}
