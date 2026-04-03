// SPDX-License-Identifier: Apache-2.0

// tool_templates.go - Templates for multi-tool AI assistant support
package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ToolConfig defines configuration for an AI tool
type ToolConfig struct {
	Name        string
	Description string
	Files       []ToolFile
}

// ToolFile defines a configuration file to create
type ToolFile struct {
	Path     string
	Content  func(projectName, mprPath string) string
	Optional bool
}

// SupportedTools defines all AI tools that can be initialized
var SupportedTools = map[string]ToolConfig{
	"claude": {
		Name:        "Claude Code",
		Description: "Claude Code with skills and commands",
		Files: []ToolFile{
			{
				Path:    ".claude/settings.json",
				Content: generateClaudeSettings,
			},
			{
				Path:    "CLAUDE.md",
				Content: generateClaudeMD,
			},
		},
	},
	"cursor": {
		Name:        "Cursor",
		Description: "Cursor AI with MDL rules",
		Files: []ToolFile{
			{
				Path:    ".cursorrules",
				Content: generateCursorRules,
			},
		},
	},
	"continue": {
		Name:        "Continue.dev",
		Description: "Continue.dev with custom commands",
		Files: []ToolFile{
			{
				Path:    ".continue/config.json",
				Content: generateContinueConfig,
			},
		},
	},
	"windsurf": {
		Name:        "Windsurf",
		Description: "Windsurf (Codeium) with MDL rules",
		Files: []ToolFile{
			{
				Path:    ".windsurfrules",
				Content: generateWindsurfRules,
			},
		},
	},
	"aider": {
		Name:        "Aider",
		Description: "Aider with project configuration",
		Files: []ToolFile{
			{
				Path:    ".aider.conf.yml",
				Content: generateAiderConfig,
			},
		},
	},
	"opencode": {
		Name:        "OpenCode",
		Description: "OpenCode AI agent with MDL commands and skills",
		Files: []ToolFile{
			{
				Path:    "opencode.json",
				Content: generateOpenCodeConfig,
			},
		},
	},
	"vibe": {
		Name:        "Mistral Vibe",
		Description: "Mistral Vibe CLI agent with skills",
		Files: []ToolFile{
			{
				Path:    ".vibe/config.toml",
				Content: generateVibeConfig,
			},
			{
				Path:    ".vibe/prompts/mendix-mdl.md",
				Content: generateVibeSystemPrompt,
			},
		},
		// Skills are synced dynamically from the embedded skillsFS in init.go
	},
}

// Universal files created for all tools
var UniversalFiles = []ToolFile{
	{
		Path:    "AGENTS.md",
		Content: generateProjectAIMD,
	},
}

func generateClaudeSettings(projectName, mprPath string) string {
	return settingsJSON
}

func generateCursorRules(projectName, mprPath string) string {
	mprFile := filepath.Base(mprPath)
	return fmt.Sprintf(`# Mendix MDL Project: %s

You are working on a Mendix project with MDL (Mendix Definition Language) support via mxcli.

## Important: mxcli Location

The mxcli tool is in the PROJECT ROOT, not in system PATH. Always use:
- ./mxcli (correct)
- NOT mxcli (will fail)

## Quick Reference

### Project Connection
`+"```bash"+`
./mxcli -p %s -c "SHOW MODULES"
`+"```"+`

### Validate MDL Scripts
`+"```bash"+`
./mxcli check script.mdl                    # Syntax only
./mxcli check script.mdl -p %s --references  # With refs
`+"```"+`

### Execute MDL Scripts
`+"```bash"+`
./mxcli exec script.mdl -p %s
`+"```"+`

### Code Search (requires REFRESH CATALOG FULL)
`+"```bash"+`
./mxcli search -p %s "pattern"
./mxcli callers -p %s Module.Microflow
./mxcli refs -p %s Module.Entity
`+"```"+`

## MDL Syntax Quick Guide

### Microflows
- Variable: `+"`DECLARE $var Type = value;`"+`
- Entity: `+"`DECLARE $entity Module.Entity;`"+` (no AS, no = empty)
- Loop: `+"`LOOP $item IN $list BEGIN ... END LOOP;`"+`
- Change: `+"`CHANGE $obj (Attr = value);`"+`
- If: `+"`IF condition THEN ... END IF;`"+` (not END)
- Log: `+"`LOG WARNING NODE 'Name' 'Message';`"+`

### Pages
- Properties: `+"(Title: 'value', Layout: 'value')"+`
- Widget nesting: curly braces `+"`{ }`"+`
- Widget properties: `+"(Label: 'Name', Attribute: AttrName)"+`

## Documentation

See AGENTS.md for complete documentation and .ai-context/skills/ for patterns.

## Before Writing MDL

1. Read relevant skill file: .ai-context/skills/write-microflows.md or create-page.md
2. Validate: ./mxcli check script.mdl -p %s --references
3. Execute: ./mxcli exec script.mdl -p %s
`, projectName, mprFile, mprFile, mprFile, mprFile, mprFile, mprFile, mprFile, mprFile)
}

func generateWindsurfRules(projectName, mprPath string) string {
	// Windsurf uses same format as Cursor
	return generateCursorRules(projectName, mprPath)
}

func generateContinueConfig(projectName, mprPath string) string {
	mprFile := filepath.Base(mprPath)
	return fmt.Sprintf(`{
  "name": "%s - Mendix MDL",
  "systemMessage": "You are helping with Mendix development using MDL (Mendix Definition Language). The mxcli tool is located in the project root - always use './mxcli' not 'mxcli'.",
  "docs": [
    "AGENTS.md",
    ".ai-context/skills/"
  ],
  "customCommands": [
    {
      "name": "check-mdl",
      "description": "Check MDL script syntax",
      "prompt": "Run: ./mxcli check {filename}"
    },
    {
      "name": "check-mdl-refs",
      "description": "Check MDL with reference validation",
      "prompt": "Run: ./mxcli check {filename} -p %s --references"
    },
    {
      "name": "execute-mdl",
      "description": "Execute MDL script",
      "prompt": "Run: ./mxcli exec {filename} -p %s"
    },
    {
      "name": "show-entities",
      "description": "Show all entities in project",
      "prompt": "Run: ./mxcli -p %s -c \"SHOW ENTITIES\""
    },
    {
      "name": "search-project",
      "description": "Search project with catalog",
      "prompt": "Run: ./mxcli search -p %s \"{query}\""
    }
  ],
  "slashCommands": [
    {
      "name": "mdl-syntax",
      "description": "Show MDL syntax reference",
      "prompt": "Read and summarize: .ai-context/skills/write-microflows.md"
    },
    {
      "name": "page-syntax",
      "description": "Show page creation syntax",
      "prompt": "Read and summarize: .ai-context/skills/create-page.md"
    }
  ]
}
`, projectName, mprFile, mprFile, mprFile, mprFile)
}

func generateAiderConfig(projectName, mprPath string) string {
	mprFile := filepath.Base(mprPath)
	return fmt.Sprintf(`# Mendix MDL Project: %s
# Configuration for Aider AI coding assistant

# Files to read for context
read-files:
  - AGENTS.md
  - .ai-context/skills/*.md

# Project description
description: |
  Mendix project with MDL (Mendix Definition Language) support.
  Use ./mxcli for all project operations.

# Custom commands
commands:
  check: "./mxcli check {file}"
  check-refs: "./mxcli check {file} -p %s --references"
  execute: "./mxcli exec {file} -p %s"
  search: "./mxcli search -p %s {query}"

# Patterns to recognize
recognize:
  - "*.mdl files use MDL syntax (see .ai-context/skills/)"
  - "Always use ./mxcli (local binary) not mxcli"
  - "Microflows: LOOP BEGIN/END LOOP, CHANGE (attr=val)"
  - "Pages: { } blocks, (Prop: value)"
`, projectName, mprFile, mprFile, mprFile)
}

func generateDevcontainerJSON(projectName, mprPath, containerRuntime string) string {
	feature := `"ghcr.io/devcontainers/features/docker-in-docker:2": {}`
	containerEnv := `"PLAYWRIGHT_CLI_SESSION": "mendix-app"`
	if containerRuntime == "podman" {
		feature = `"ghcr.io/devcontainers/features/podman-in-podman:1": {}`
		containerEnv = `"PLAYWRIGHT_CLI_SESSION": "mendix-app",
    "MXCLI_CONTAINER_CLI": "podman"`
	}

	return fmt.Sprintf(`{
  "name": "%s",
  "build": {
    "dockerfile": "Dockerfile"
  },
  "features": {
    %s
  },
  "forwardPorts": [8080, 8090, 5432],
  "portsAttributes": {
    "8080-8099": { "onAutoForward": "silent" },
    "5432-5499": { "onAutoForward": "silent" }
  },
  "containerEnv": {
    %s
  },
  "postCreateCommand": "curl -fsSL https://claude.ai/install.sh | bash && if [ -f ./mxcli ] && ! file ./mxcli | grep -q Linux; then echo '⚠ ./mxcli is not a Linux binary. Replace it with the linux-amd64 or linux-arm64 build.'; fi",
  "customizations": {
    "vscode": {
      "extensions": [
        "anthropic.claude-code"
      ],
      "settings": {
        "mdl.mxcliPath": "./mxcli"
      }
    }
  },
  "remoteUser": "vscode"
}
`, projectName, feature, containerEnv)
}

func generateDockerfile(projectName, mprPath string) string {
	return `FROM mcr.microsoft.com/devcontainers/base:bookworm

# Install Adoptium JDK 21 (required by MxBuild), Node.js 22, and utility tools
RUN apt-get update && apt-get install -y --no-install-recommends wget apt-transport-https gpg ca-certificates curl && \
    wget -qO - https://packages.adoptium.net/artifactory/api/gpg/key/public | gpg --dearmor -o /etc/apt/keyrings/adoptium.gpg && \
    echo "deb [signed-by=/etc/apt/keyrings/adoptium.gpg] https://packages.adoptium.net/artifactory/deb bookworm main" > /etc/apt/sources.list.d/adoptium.list && \
    curl -fsSL https://deb.nodesource.com/setup_22.x | bash - && \
    apt-get update && \
    apt-get install -y --no-install-recommends \
       temurin-21-jdk \
       nodejs \
       postgresql-client \
       kafkacat \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install playwright-cli and Chromium with all system dependencies (must run as root)
RUN npm install -g @playwright/cli@latest && \
    npx playwright install --with-deps chromium
`
}

func generatePlaywrightConfig() string {
	return `{
  "browser": {
    "browserName": "chromium",
    "isolated": true,
    "launchOptions": {
      "headless": true
    }
  },
  "timeouts": {
    "action": 10000,
    "navigation": 30000
  },
  "network": {
    "allowedOrigins": [
      "http://localhost:8079",
      "http://localhost:8080",
      "http://localhost:8081",
      "http://localhost:8082",
      "http://localhost:8083",
      "http://localhost:8084",
      "http://localhost:8085"
    ]
  }
}
`
}

func generateProjectAIMD(projectName, mprPath string) string {
	return generateClaudeMD(projectName, mprPath)
}

func generateVibeConfig(projectName, mprPath string) string {
	return fmt.Sprintf(`# Mistral Vibe configuration for Mendix project: %s
# See: https://docs.mistral.ai/mistral-vibe/introduction/configuration

# Use the project AGENTS.md as system prompt context
system_prompt_id = "mendix-mdl"

# Skills from .vibe/skills/ are auto-discovered
# Additional context files
# skill_paths = [".ai-context/skills"]

# Tool permissions for MDL workflow
[tools.bash]
permission = "ask"

[tools.read_file]
permission = "always"

[tools.write_file]
permission = "ask"

[tools.search_replace]
permission = "ask"
`, projectName)
}

func generateVibeSystemPrompt(projectName, mprPath string) string {
	mprFile := filepath.Base(mprPath)
	return fmt.Sprintf(`You are helping with a Mendix project using MDL (Mendix Definition Language) via mxcli.

## Project: %s

MPR file: %s

## Key Rules

- The mxcli tool is in the project root. Always use ./mxcli, not mxcli.
- Read AGENTS.md for full project documentation.
- Read .ai-context/skills/ for MDL syntax patterns before writing scripts.
- Always validate MDL scripts: ./mxcli check script.mdl -p %s --references
- Microflow variables start with $. Entity declarations have no AS keyword.
- Page widgets nest with curly braces { }. Properties use (Key: value).
- Single quotes in expressions are escaped by doubling: 'it''s here'

## Quick Commands

- Explore: ./mxcli -p %s -c "SHOW STRUCTURE"
- Check: ./mxcli check script.mdl -p %s --references
- Execute: ./mxcli exec script.mdl -p %s
- Search: ./mxcli search -p %s "keyword"
`, projectName, mprFile, mprFile, mprFile, mprFile, mprFile, mprFile)
}

// wrapSkillForVibe wraps a plain markdown skill file with Vibe SKILL.md frontmatter.
func wrapSkillForVibe(skillName string, content []byte) []byte {
	// Derive a description from the first markdown heading
	description := "MDL skill for Mendix development"
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			description = strings.TrimPrefix(line, "# ")
			break
		}
	}

	frontmatter := fmt.Sprintf(`---
name: %s
description: %s
user-invocable: true
allowed-tools:
  - read_file
  - write_file
  - bash
  - grep
---

`, skillName, description)

	return []byte(frontmatter + string(content))
}

func generateOpenCodeConfig(projectName, mprPath string) string {
	return `{
  "$schema": "https://opencode.ai/config.json",
  "instructions": [
    "AGENTS.md",
    ".opencode/skills/**/SKILL.md",
    ".ai-context/skills/*.md"
  ]
}
`
}
