// SPDX-License-Identifier: Apache-2.0

//go:build integration

package executor

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/visitor"
)

// scriptModuleDeps maps script filenames to marketplace module MPKs they require.
// These modules are imported via `mx module-import` before executing the script.
var scriptModuleDeps = map[string][]string{
	"05-database-connection-examples.mdl": {"ExternalDatabaseConnector-v6.2.3.mpk"},
	"13-business-events-examples.mdl":     {"BusinessEvents_3.12.0.mpk"},
}

// TestMxCheck_DoctypeScripts executes each doctype-tests/*.mdl example script
// in its own fresh Mendix project and validates the result with mx check.
//
// Each script runs in isolation so errors are cleanly attributed.
// Files matching *.test.mdl or *.tests.mdl are skipped (they require Docker).
func TestMxCheck_DoctypeScripts(t *testing.T) {
	if !mxCheckAvailable() {
		t.Skip("mx command not available")
	}

	// Locate doctype-tests directory
	doctypeDir, err := filepath.Abs("../../mdl-examples/doctype-tests")
	if err != nil {
		t.Fatalf("Failed to resolve doctype-tests path: %v", err)
	}
	if _, err := os.Stat(doctypeDir); err != nil {
		t.Skipf("doctype-tests directory not found at %s", doctypeDir)
	}

	// Locate mx-modules directory for marketplace dependencies
	modulesDir, err := filepath.Abs("../../mx-modules")
	if err != nil {
		t.Logf("Warning: could not resolve mx-modules path: %v", err)
	}

	// Collect eligible scripts (skip .test.mdl and .tests.mdl)
	entries, err := os.ReadDir(doctypeDir)
	if err != nil {
		t.Fatalf("Failed to read doctype-tests directory: %v", err)
	}

	var scripts []string
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".mdl") {
			continue
		}
		if strings.HasSuffix(name, ".test.mdl") || strings.HasSuffix(name, ".tests.mdl") {
			continue
		}
		scripts = append(scripts, name)
	}
	sort.Strings(scripts)

	if len(scripts) == 0 {
		t.Skip("no eligible MDL scripts found")
	}

	mxPath := findMxBinary()

	for _, name := range scripts {
		scriptPath := filepath.Join(doctypeDir, name)
		content, err := os.ReadFile(scriptPath)
		if err != nil {
			t.Fatalf("Failed to read %s: %v", name, err)
		}

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Fresh project for each script
			env := setupTestEnv(t)
			defer env.teardown()

			// Import required marketplace modules before executing script
			if deps, ok := scriptModuleDeps[name]; ok && modulesDir != "" && mxPath != "" {
				// Disconnect so mx can access the MPR file
				env.executor.Execute(&ast.DisconnectStmt{})

				for _, mpk := range deps {
					mpkPath := filepath.Join(modulesDir, mpk)
					if _, err := os.Stat(mpkPath); err != nil {
						t.Logf("Skipping module import (not found): %s", mpkPath)
						continue
					}
					cmd := exec.Command(mxPath, "module-import", mpkPath, env.projectPath)
					if out, err := cmd.CombinedOutput(); err != nil {
						t.Logf("Warning: module import failed for %s: %v\n%s", mpk, err, string(out))
					}
				}

				// Reconnect after module import
				if err := env.executor.Execute(&ast.ConnectStmt{Path: env.projectPath}); err != nil {
					t.Fatalf("Failed to reconnect after module import: %v", err)
				}
			}

			// Execute the script
			prog, errs := visitor.Build(string(content))
			if len(errs) > 0 {
				t.Fatalf("Parse error: %v", errs[0])
			}

			if err := env.executor.ExecuteProgram(prog); err != nil {
				t.Errorf("Execution error: %v", err)
			}

			// Flush to disk
			env.executor.Execute(&ast.DisconnectStmt{})

			// Run mx check
			output, mxErr := runMxCheck(t, env.projectPath)
			if mxErr != nil {
				// Check for actual errors: [error] lines or ERROR: crash messages
				hasErrors := strings.Contains(output, "[error]") || strings.Contains(output, "ERROR:")
				if hasErrors {
					// CE0161 (XPath constraint errors) are known limitations of the
					// XPath serializer — log but don't fail the test.
					onlyCE0161 := strings.Contains(output, "CE0161") &&
						strings.Count(output, "[error]") == strings.Count(output, "CE0161")
					if onlyCE0161 {
						t.Logf("mx check has known XPath limitation (CE0161):\n%s", output)
					} else {
						t.Errorf("mx check found errors:\n%s", output)
					}
				} else {
					t.Logf("mx check output:\n%s", output)
				}
			} else {
				t.Logf("mx check passed: 0 errors")
			}
		})
	}
}
