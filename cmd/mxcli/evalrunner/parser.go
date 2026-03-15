// SPDX-License-Identifier: Apache-2.0

// Package evalrunner implements the eval framework for testing mxcli + Claude Code
// against structured evaluation test definitions.
package evalrunner

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// EvalTest represents a single evaluation test case parsed from a Markdown file.
type EvalTest struct {
	// Metadata from YAML frontmatter
	ID         string        `yaml:"id"`
	Category   string        `yaml:"category"`
	Tags       []string      `yaml:"tags"`
	Timeout    time.Duration `yaml:"-"`
	RawTimeout string        `yaml:"timeout"`

	// Parsed from Markdown sections
	Title    string
	Prompt   string
	Expected string
	Checks   []Check
	Criteria []string

	// Optional iteration scenario
	Iteration *EvalIteration

	// Source file
	SourceFile string
}

// EvalIteration represents a follow-up prompt and its validation criteria.
type EvalIteration struct {
	Prompt   string
	Checks   []Check
	Criteria []string
}

// Check represents an automated check to run against the project.
type Check struct {
	Type string `json:"type"` // entity_exists, entity_has_attribute, page_exists, etc.
	Args string `json:"args"` // "*.Book.Title String"
}

// String returns a human-readable representation of the check.
func (c Check) String() string {
	return fmt.Sprintf("%s %s", c.Type, c.Args)
}

// checkPattern matches "- check_type: \"args\"" or "- check_type: args"
var checkPattern = regexp.MustCompile(`^-\s+(\w+):\s*"?([^"]*)"?\s*$`)

// ParseEvalFile parses a single eval test definition from a Markdown file with YAML frontmatter.
func ParseEvalFile(path string) (*EvalTest, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading eval file: %w", err)
	}

	test, err := parseEvalContent(string(content))
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", filepath.Base(path), err)
	}

	test.SourceFile = path
	return test, nil
}

// ParseEvalDir parses all eval test files in a directory.
func ParseEvalDir(dir string) ([]*EvalTest, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading eval directory: %w", err)
	}

	var tests []*EvalTest
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !isEvalFile(e.Name()) {
			continue
		}
		test, err := ParseEvalFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		tests = append(tests, test)
	}

	return tests, nil
}

// isEvalFile returns true if the filename looks like an eval test definition.
func isEvalFile(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasPrefix(lower, "eval") && strings.HasSuffix(lower, ".md")
}

// parseEvalContent parses eval test content from a string.
func parseEvalContent(content string) (*EvalTest, error) {
	test := &EvalTest{}

	// Split YAML frontmatter from Markdown body
	frontmatter, body, err := splitFrontmatter(content)
	if err != nil {
		return nil, err
	}

	// Parse YAML frontmatter
	if frontmatter != "" {
		if err := yaml.Unmarshal([]byte(frontmatter), test); err != nil {
			return nil, fmt.Errorf("parsing YAML frontmatter: %w", err)
		}
	}

	// Parse timeout
	if test.RawTimeout != "" {
		d, err := time.ParseDuration(test.RawTimeout)
		if err != nil {
			return nil, fmt.Errorf("parsing timeout %q: %w", test.RawTimeout, err)
		}
		test.Timeout = d
	}
	if test.Timeout == 0 {
		test.Timeout = 10 * time.Minute
	}

	// Parse Markdown sections
	sections := parseSections(body)

	// Extract title from first heading
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			test.Title = strings.TrimPrefix(line, "# ")
			break
		}
	}

	// Extract main sections
	test.Prompt = strings.TrimSpace(sections["prompt"])
	test.Expected = strings.TrimSpace(sections["expected outcome"])
	test.Checks = parseChecks(sections["checks"])
	test.Criteria = parseCriteria(sections["acceptance criteria"])

	// Extract iteration section
	if iterContent, ok := sections["iteration"]; ok && strings.TrimSpace(iterContent) != "" {
		iter := &EvalIteration{}
		iterSections := parseSubSections(iterContent)

		iter.Prompt = strings.TrimSpace(iterSections["prompt"])
		iter.Checks = parseChecks(iterSections["checks"])
		iter.Criteria = parseCriteria(iterSections["acceptance criteria"])

		if iter.Prompt != "" {
			test.Iteration = iter
		}
	}

	// Validate
	if test.ID == "" {
		return nil, fmt.Errorf("eval test missing 'id' in YAML frontmatter")
	}
	if test.Prompt == "" {
		return nil, fmt.Errorf("eval test %s missing '## Prompt' section", test.ID)
	}

	return test, nil
}

// splitFrontmatter splits YAML frontmatter (between --- delimiters) from the body.
func splitFrontmatter(content string) (frontmatter, body string, err error) {
	content = strings.TrimSpace(content)

	if !strings.HasPrefix(content, "---") {
		return "", content, nil
	}

	// Find the closing ---
	rest := content[3:] // skip opening ---
	rest = strings.TrimLeft(rest, " \t")
	if len(rest) > 0 && rest[0] == '\n' {
		rest = rest[1:]
	} else if len(rest) > 1 && rest[0] == '\r' && rest[1] == '\n' {
		rest = rest[2:]
	}

	idx := strings.Index(rest, "\n---")
	if idx == -1 {
		return "", content, nil
	}

	frontmatter = strings.TrimSpace(rest[:idx])
	body = rest[idx+4:] // skip \n---
	// Skip any trailing whitespace on the --- line
	if len(body) > 0 && body[0] == '\n' {
		body = body[1:]
	} else if len(body) > 1 && body[0] == '\r' && body[1] == '\n' {
		body = body[2:]
	}

	return frontmatter, body, nil
}

// parseSections splits Markdown into sections by ## headings.
// Returns a map of lowercase heading → content.
func parseSections(body string) map[string]string {
	sections := make(map[string]string)
	var currentHeading string
	var currentContent strings.Builder

	scanner := bufio.NewScanner(strings.NewReader(body))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "## ") {
			// Save previous section
			if currentHeading != "" {
				sections[currentHeading] = currentContent.String()
			}
			currentHeading = strings.ToLower(strings.TrimPrefix(trimmed, "## "))
			currentContent.Reset()
		} else {
			if currentHeading != "" {
				currentContent.WriteString(line)
				currentContent.WriteString("\n")
			}
		}
	}

	// Save last section
	if currentHeading != "" {
		sections[currentHeading] = currentContent.String()
	}

	return sections
}

// parseSubSections splits content into sections by ### headings.
func parseSubSections(content string) map[string]string {
	sections := make(map[string]string)
	var currentHeading string
	var currentContent strings.Builder

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "### ") {
			// Save previous section
			if currentHeading != "" {
				sections[currentHeading] = currentContent.String()
			}
			currentHeading = strings.ToLower(strings.TrimPrefix(trimmed, "### "))
			currentContent.Reset()
		} else {
			if currentHeading != "" {
				currentContent.WriteString(line)
				currentContent.WriteString("\n")
			}
		}
	}

	// Save last section
	if currentHeading != "" {
		sections[currentHeading] = currentContent.String()
	}

	return sections
}

// parseChecks parses check lines from the Checks section.
func parseChecks(content string) []Check {
	var checks []Check
	if content == "" {
		return checks
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		m := checkPattern.FindStringSubmatch(line)
		if m == nil {
			continue
		}

		checks = append(checks, Check{
			Type: m[1],
			Args: strings.TrimSpace(m[2]),
		})
	}

	return checks
}

// parseCriteria parses acceptance criteria lines (bullet points).
func parseCriteria(content string) []string {
	var criteria []string
	if content == "" {
		return criteria
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Strip bullet prefix (- or *)
		if strings.HasPrefix(line, "- ") {
			line = strings.TrimPrefix(line, "- ")
		} else if strings.HasPrefix(line, "* ") {
			line = strings.TrimPrefix(line, "* ")
		} else {
			continue // skip non-bullet lines
		}
		criteria = append(criteria, strings.TrimSpace(line))
	}

	return criteria
}
