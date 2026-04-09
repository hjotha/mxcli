// SPDX-License-Identifier: Apache-2.0

// init_vscode.go - VS Code extension installation for Mendix projects
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// installVSCodeExtension extracts the embedded .vsix and installs it into VS Code.
func installVSCodeExtension(projectDir string) {
	// Skip if no embedded vsix data
	if len(vsixData) == 0 {
		return
	}

	// Write .vsix to the project directory
	vsixPath := filepath.Join(projectDir, ".claude", "vscode-mdl.vsix")
	if err := os.WriteFile(vsixPath, vsixData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "  Warning: could not write VS Code extension: %v\n", err)
		return
	}

	// Try to find the VS Code CLI
	codeCLI := findCodeCLI()
	if codeCLI == "" {
		fmt.Println()
		fmt.Println("  ┌─────────────────────────────────────────────────────────────┐")
		fmt.Println("  │  VS Code MDL extension extracted but could not auto-install │")
		fmt.Println("  │  ('code' command not found on PATH)                         │")
		fmt.Println("  │                                                             │")
		fmt.Println("  │  To install, either:                                        │")
		fmt.Println("  │  1. Run in terminal:                                        │")
		fmt.Printf("  │     code --install-extension %s\n", vsixPath)
		fmt.Println("  │  2. Or in VS Code: Ctrl+Shift+X → ··· → Install from VSIX  │")
		fmt.Printf("  │     Select: %s\n", vsixPath)
		fmt.Println("  └─────────────────────────────────────────────────────────────┘")
		fmt.Println()
		return
	}

	// Install the extension
	cmd := exec.Command(codeCLI, "--install-extension", vsixPath, "--force")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  Warning: VS Code extension install failed: %v\n", err)
		if len(output) > 0 {
			fmt.Fprintf(os.Stderr, "  %s\n", strings.TrimSpace(string(output)))
		}
		fmt.Printf("  Install manually: %s --install-extension %s\n", codeCLI, vsixPath)
		return
	}
	fmt.Println("  Installed VS Code MDL extension")

	// Clean up the extracted .vsix
	os.Remove(vsixPath)
}

// findCodeCLI looks for the VS Code CLI executable.
func findCodeCLI() string {
	// 1. Check PATH (works on all platforms when VS Code added to PATH)
	for _, name := range []string{"code", "code-insiders"} {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}

	// 2. Windows: check common install locations
	if runtime.GOOS == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		programFiles := os.Getenv("ProgramFiles")
		candidates := []string{}
		if localAppData != "" {
			candidates = append(candidates,
				filepath.Join(localAppData, "Programs", "Microsoft VS Code", "bin", "code.cmd"),
				filepath.Join(localAppData, "Programs", "Microsoft VS Code Insiders", "bin", "code-insiders.cmd"),
			)
		}
		if programFiles != "" {
			candidates = append(candidates,
				filepath.Join(programFiles, "Microsoft VS Code", "bin", "code.cmd"),
				filepath.Join(programFiles, "Microsoft VS Code Insiders", "bin", "code-insiders.cmd"),
			)
		}
		for _, path := range candidates {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}

	// 3. macOS: check standard application path
	if runtime.GOOS == "darwin" {
		candidates := []string{
			"/Applications/Visual Studio Code.app/Contents/Resources/app/bin/code",
			"/Applications/Visual Studio Code - Insiders.app/Contents/Resources/app/bin/code-insiders",
		}
		for _, path := range candidates {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}

	return ""
}
