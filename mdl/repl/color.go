// SPDX-License-Identifier: Apache-2.0

package repl

import (
	"os"

	"golang.org/x/term"
)

// ANSI escape codes.
const (
	ansiReset    = "\033[0m"
	ansiBold     = "\033[1m"
	ansiBoldCyan = "\033[1;36m"
	ansiRed      = "\033[31m"
	ansiGray     = "\033[90m"
)

// rlIgnoreStart / rlIgnoreEnd bracket invisible characters in readline prompts
// so readline counts the display width correctly.
const (
	rlIgnoreStart = "\001"
	rlIgnoreEnd   = "\002"
)

// colorPalette applies ANSI colors when stdout is a real TTY and NO_COLOR is unset.
type colorPalette struct {
	enabled bool
}

// newColorPalette returns a palette active only when stdout is a terminal and
// the NO_COLOR environment variable is absent.
func newColorPalette() colorPalette {
	if os.Getenv("NO_COLOR") != "" {
		return colorPalette{}
	}
	return colorPalette{enabled: term.IsTerminal(int(os.Stdout.Fd()))}
}

func (c colorPalette) wrap(code, s string) string {
	if !c.enabled {
		return s
	}
	return code + s + ansiReset
}

// Bold returns s in bold.
func (c colorPalette) Bold(s string) string { return c.wrap(ansiBold, s) }

// Red returns s in red (used for errors).
func (c colorPalette) Red(s string) string { return c.wrap(ansiRed, s) }

// Gray returns s in dark gray (used for secondary text).
func (c colorPalette) Gray(s string) string { return c.wrap(ansiGray, s) }

// PromptPrimary returns a readline-safe bold-cyan prompt string.
// The \001…\002 markers tell readline not to count the escape bytes in the
// visible line width, preventing cursor positioning bugs.
func (c colorPalette) PromptPrimary(s string) string {
	if !c.enabled {
		return s
	}
	return rlIgnoreStart + ansiBoldCyan + rlIgnoreEnd + s + rlIgnoreStart + ansiReset + rlIgnoreEnd
}

// PromptContinue returns a readline-safe gray continuation prompt string.
func (c colorPalette) PromptContinue(s string) string {
	if !c.enabled {
		return s
	}
	return rlIgnoreStart + ansiGray + rlIgnoreEnd + s + rlIgnoreStart + ansiReset + rlIgnoreEnd
}
