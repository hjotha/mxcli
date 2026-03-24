package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Diff color palette.
var (
	diffAddedFg        = lipgloss.Color("#00D787")
	diffAddedChangedFg = lipgloss.Color("#FFFFFF")
	diffAddedChangedBg = lipgloss.Color("#005F00")

	diffRemovedFg        = lipgloss.Color("#FF5F87")
	diffRemovedChangedFg = lipgloss.Color("#FFFFFF")
	diffRemovedChangedBg = lipgloss.Color("#5F0000")

	diffEqualGutter     = lipgloss.Color("#626262")
	diffGutterAddedFg   = lipgloss.Color("#00D787")
	diffGutterRemovedFg = lipgloss.Color("#FF5F87")
)

// DiffRenderedLine holds the sticky prefix (gutter + line numbers) and scrollable content separately.
type DiffRenderedLine struct {
	Prefix  string // gutter char + line numbers (sticky, never scrolled)
	Content string // actual code/text content (horizontally scrollable)
}

// RenderUnifiedDiff renders a DiffResult as unified diff lines with prefix/content split.
func RenderUnifiedDiff(result *DiffResult, lang string) []DiffRenderedLine {
	if result == nil || len(result.Lines) == 0 {
		return nil
	}

	gutterCharSt := lipgloss.NewStyle()
	lineNoSt := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	maxLineNo := 0
	for _, dl := range result.Lines {
		if dl.OldLineNo > maxLineNo {
			maxLineNo = dl.OldLineNo
		}
		if dl.NewLineNo > maxLineNo {
			maxLineNo = dl.NewLineNo
		}
	}
	lineNoW := max(3, len(fmt.Sprintf("%d", maxLineNo)))

	rendered := make([]DiffRenderedLine, 0, len(result.Lines))
	for _, dl := range result.Lines {
		var gutter, oldNo, newNo, content string

		switch dl.Type {
		case DiffEqual:
			gutter = gutterCharSt.Foreground(diffEqualGutter).Render("│")
			oldNo = lineNoSt.Render(fmt.Sprintf("%*d", lineNoW, dl.OldLineNo))
			newNo = lineNoSt.Render(fmt.Sprintf("%*d", lineNoW, dl.NewLineNo))
			content = highlightLine(dl.Content, lang)

		case DiffInsert:
			gutter = gutterCharSt.Foreground(diffGutterAddedFg).Render("+")
			oldNo = lineNoSt.Render(strings.Repeat(" ", lineNoW))
			newNo = lipgloss.NewStyle().Foreground(diffGutterAddedFg).Render(fmt.Sprintf("%*d", lineNoW, dl.NewLineNo))
			content = renderSegments(dl.Segments, DiffInsert)

		case DiffDelete:
			gutter = gutterCharSt.Foreground(diffGutterRemovedFg).Render("-")
			oldNo = lipgloss.NewStyle().Foreground(diffGutterRemovedFg).Render(fmt.Sprintf("%*d", lineNoW, dl.OldLineNo))
			newNo = lineNoSt.Render(strings.Repeat(" ", lineNoW))
			content = renderSegments(dl.Segments, DiffDelete)
		}

		prefix := gutter + " " + oldNo + " " + newNo + " "
		rendered = append(rendered, DiffRenderedLine{Prefix: prefix, Content: content})
	}
	return rendered
}

// SideBySideRenderedLine holds prefix and content for one pane in side-by-side view.
type SideBySideRenderedLine struct {
	Prefix  string // line number (sticky)
	Content string // code content (scrollable)
	Blank   bool   // true if this is a blank filler line
}

// RenderSideBySideDiff renders a DiffResult as two columns with prefix/content split.
func RenderSideBySideDiff(result *DiffResult, lang string) (left, right []SideBySideRenderedLine) {
	if result == nil || len(result.Lines) == 0 {
		return nil, nil
	}

	lineNoSt := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	maxLineNo := 0
	for _, dl := range result.Lines {
		if dl.OldLineNo > maxLineNo {
			maxLineNo = dl.OldLineNo
		}
		if dl.NewLineNo > maxLineNo {
			maxLineNo = dl.NewLineNo
		}
	}
	lineNoW := max(3, len(fmt.Sprintf("%d", maxLineNo)))
	blankPrefix := strings.Repeat(" ", lineNoW) + " "

	for _, dl := range result.Lines {
		switch dl.Type {
		case DiffEqual:
			highlighted := highlightLine(dl.Content, lang)
			oldNo := lineNoSt.Render(fmt.Sprintf("%*d", lineNoW, dl.OldLineNo)) + " "
			newNo := lineNoSt.Render(fmt.Sprintf("%*d", lineNoW, dl.NewLineNo)) + " "
			left = append(left, SideBySideRenderedLine{Prefix: oldNo, Content: highlighted})
			right = append(right, SideBySideRenderedLine{Prefix: newNo, Content: highlighted})

		case DiffDelete:
			content := renderSegments(dl.Segments, DiffDelete)
			oldNo := lipgloss.NewStyle().Foreground(diffGutterRemovedFg).Render(fmt.Sprintf("%*d", lineNoW, dl.OldLineNo)) + " "
			left = append(left, SideBySideRenderedLine{Prefix: oldNo, Content: content})
			right = append(right, SideBySideRenderedLine{Prefix: blankPrefix, Blank: true})

		case DiffInsert:
			content := renderSegments(dl.Segments, DiffInsert)
			newNo := lipgloss.NewStyle().Foreground(diffGutterAddedFg).Render(fmt.Sprintf("%*d", lineNoW, dl.NewLineNo)) + " "
			left = append(left, SideBySideRenderedLine{Prefix: blankPrefix, Blank: true})
			right = append(right, SideBySideRenderedLine{Prefix: newNo, Content: content})
		}
	}
	return left, right
}

// renderSegments renders word-level diff segments with appropriate styling.
func renderSegments(segments []DiffSegment, lineType DiffLineType) string {
	if len(segments) == 0 {
		return ""
	}

	var normalFg, changedFg, changedBg lipgloss.Color
	switch lineType {
	case DiffInsert:
		normalFg = diffAddedFg
		changedFg = diffAddedChangedFg
		changedBg = diffAddedChangedBg
	case DiffDelete:
		normalFg = diffRemovedFg
		changedFg = diffRemovedChangedFg
		changedBg = diffRemovedChangedBg
	default:
		var sb strings.Builder
		for _, seg := range segments {
			sb.WriteString(seg.Text)
		}
		return sb.String()
	}

	normalSt := lipgloss.NewStyle().Foreground(normalFg)
	changedSt := lipgloss.NewStyle().Foreground(changedFg).Background(changedBg)

	var sb strings.Builder
	for _, seg := range segments {
		if seg.Changed {
			sb.WriteString(changedSt.Render(seg.Text))
		} else {
			sb.WriteString(normalSt.Render(seg.Text))
		}
	}
	return sb.String()
}

// highlightLine applies syntax highlighting based on language.
func highlightLine(content, lang string) string {
	switch strings.ToLower(lang) {
	case "sql", "mdl":
		return HighlightMDL(content)
	case "ndsl":
		return HighlightNDSL(content)
	case "":
		return DetectAndHighlight(content)
	default:
		return content
	}
}

// hslice returns a horizontal slice of an ANSI-colored string,
// skipping the first `skip` visual columns and returning up to `take` visual columns.
func hslice(s string, skip, take int) string {
	if skip == 0 {
		return truncateToWidth(s, take)
	}

	var result strings.Builder
	visW := 0
	inEsc := false
	for _, r := range s {
		if r == '\x1b' {
			inEsc = true
			if visW >= skip {
				result.WriteRune(r)
			}
			continue
		}
		if inEsc {
			if visW >= skip {
				result.WriteRune(r)
			}
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEsc = false
			}
			continue
		}
		visW++
		if visW <= skip {
			continue
		}
		if visW-skip > take {
			break
		}
		result.WriteRune(r)
	}
	return result.String()
}

// truncateToWidth truncates a (possibly ANSI-colored) string to fit maxW visual columns.
func truncateToWidth(s string, maxW int) string {
	if lipgloss.Width(s) <= maxW {
		return s
	}

	var result strings.Builder
	visW := 0
	inEsc := false
	for _, r := range s {
		if r == '\x1b' {
			inEsc = true
			result.WriteRune(r)
			continue
		}
		if inEsc {
			result.WriteRune(r)
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEsc = false
			}
			continue
		}
		visW++
		if visW > maxW {
			break
		}
		result.WriteRune(r)
	}
	return result.String()
}
