package grammar

import (
	"bufio"
	"os"
	"regexp"
	"strings"
	"testing"
)

// TestKeywordRuleCoverage verifies that every lexer token (except structural
// ones like operators, punctuation, literals, and identifiers) is listed in
// the parser's `keyword` rule. This catches the common mistake of adding a
// new token to MDLLexer.g4 but forgetting to add it to the keyword rule,
// which would prevent it from being used as an identifier.
func TestKeywordRuleCoverage(t *testing.T) {
	allTokens := parseLexerTokens(t, "parser/MDLLexer.tokens")
	keywordTokens := parseKeywordRule(t, "MDLParser.g4")

	// Structural tokens that should NOT be in the keyword rule.
	excluded := map[string]bool{
		// Whitespace & comments
		"WS": true, "DOC_COMMENT": true, "BLOCK_COMMENT": true, "LINE_COMMENT": true,
		// Identifiers & variables
		"IDENTIFIER": true, "HYPHENATED_ID": true, "QUOTED_IDENTIFIER": true, "VARIABLE": true,
		// Literals
		"STRING_LITERAL": true, "DOLLAR_STRING": true, "NUMBER_LITERAL": true, "MENDIX_TOKEN": true,
		// Punctuation
		"SEMICOLON": true, "COMMA": true, "ELLIPSIS": true, "DOT": true,
		"LPAREN": true, "RPAREN": true,
		"LBRACE": true, "RBRACE": true,
		"LBRACKET": true, "RBRACKET": true,
		"COLON": true, "AT": true, "PIPE": true,
		"DOUBLE_COLON": true, "ARROW": true, "QUESTION": true, "HASH": true,
		// Operators
		"NOT_EQUALS": true, "LESS_THAN_OR_EQUAL": true, "GREATER_THAN_OR_EQUAL": true,
		"EQUALS": true, "LESS_THAN": true, "GREATER_THAN": true,
		"PLUS": true, "MINUS": true, "STAR": true, "SLASH": true, "PERCENT": true,
		// Version marker (not an identifier)
		"V3": true,
	}

	// Tokens missing from keyword rule (in lexer but not in keyword).
	var missing []string
	for _, tok := range allTokens {
		if excluded[tok] {
			continue
		}
		if !keywordTokens[tok] {
			missing = append(missing, tok)
		}
	}

	// Tokens in keyword rule but not in lexer (typos or stale entries).
	var extra []string
	allSet := make(map[string]bool, len(allTokens))
	for _, tok := range allTokens {
		allSet[tok] = true
	}
	for tok := range keywordTokens {
		if !allSet[tok] {
			extra = append(extra, tok)
		}
	}

	if len(missing) > 0 {
		t.Errorf("tokens in lexer but missing from keyword rule (%d):\n  %s\n"+
			"Add them to the keyword rule in MDLParser.g4 or to the excluded set in this test.",
			len(missing), strings.Join(missing, ", "))
	}
	if len(extra) > 0 {
		t.Errorf("tokens in keyword rule but not in lexer (%d):\n  %s",
			len(extra), strings.Join(extra, ", "))
	}
}

// parseLexerTokens reads MDLLexer.tokens and returns all symbolic token names
// (skipping literal aliases like '<='=515).
func parseLexerTokens(t *testing.T, path string) []string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()

	var tokens []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "'") {
			continue
		}
		if idx := strings.IndexByte(line, '='); idx > 0 {
			tokens = append(tokens, line[:idx])
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan %s: %v", path, err)
	}
	return tokens
}

// parseKeywordRule reads MDLParser.g4, extracts the `keyword` rule body, and
// returns the set of token names referenced in it.
func parseKeywordRule(t *testing.T, path string) map[string]bool {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	text := string(data)

	// Find the keyword rule: starts with "keyword" at the beginning of a line,
	// ends with ";".
	re := regexp.MustCompile(`(?m)^keyword\b([\s\S]*?);`)
	m := re.FindStringSubmatch(text)
	if m == nil {
		t.Fatal("keyword rule not found in MDLParser.g4")
	}

	// Strip single-line comments (// ...) to avoid matching words in comments.
	body := regexp.MustCompile(`//[^\n]*`).ReplaceAllString(m[1], "")

	// Extract all UPPERCASE token references (e.g., ADD, ALTER, STRING_TYPE).
	tokenRe := regexp.MustCompile(`\b([A-Z][A-Z0-9_]*)\b`)
	matches := tokenRe.FindAllStringSubmatch(body, -1)

	result := make(map[string]bool, len(matches))
	for _, match := range matches {
		result[match[1]] = true
	}
	return result
}
