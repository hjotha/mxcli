package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DiffViewMode determines the diff display layout.
type DiffViewMode int

const (
	DiffViewUnified    DiffViewMode = iota
	DiffViewSideBySide
)

// DiffOpenMsg requests opening a diff view.
type DiffOpenMsg struct {
	OldText  string
	NewText  string
	Language string // "sql", "go", "ndsl", "" (auto-detect)
	Title    string
}

// DiffCloseMsg signals the diff view should close.
type DiffCloseMsg struct{}

// DiffView is a Bubble Tea component for interactive diff viewing.
type DiffView struct {
	// Input
	oldText  string
	newText  string
	language string
	title    string

	// Computed
	result     *DiffResult
	unified    []DiffRenderedLine       // pre-rendered unified lines
	sideLeft   []SideBySideRenderedLine // pre-rendered side-by-side left
	sideRight  []SideBySideRenderedLine // pre-rendered side-by-side right
	hunkStarts []int                    // line indices where hunks begin

	// View state
	viewMode DiffViewMode
	yOffset  int
	xOffset  int // horizontal scroll offset (content only, line numbers stay fixed)
	width    int
	height   int

	// Side-by-side state
	syncScroll  bool
	focus       int // 0=left, 1=right
	leftOffset  int
	rightOffset int

	// Search
	searching   bool
	searchInput textinput.Model
	searchQuery string
	matchLines  []int
	matchIdx    int
}

// NewDiffView creates a DiffView from a DiffOpenMsg.
func NewDiffView(msg DiffOpenMsg, width, height int) DiffView {
	ti := textinput.New()
	ti.Prompt = "/"
	ti.CharLimit = 200

	dv := DiffView{
		oldText:     msg.OldText,
		newText:     msg.NewText,
		language:    msg.Language,
		title:       msg.Title,
		width:       width,
		height:      height,
		syncScroll:  true,
		searchInput: ti,
	}

	dv.result = ComputeDiff(msg.OldText, msg.NewText)
	dv.renderAll()
	dv.computeHunkStarts()

	return dv
}

func (dv *DiffView) renderAll() {
	dv.unified = RenderUnifiedDiff(dv.result, dv.language)
	dv.sideLeft, dv.sideRight = RenderSideBySideDiff(dv.result, dv.language)
}

func (dv *DiffView) computeHunkStarts() {
	dv.hunkStarts = nil
	if dv.result == nil {
		return
	}
	for i, dl := range dv.result.Lines {
		if dl.Type == DiffEqual {
			continue
		}
		if i == 0 || dv.result.Lines[i-1].Type == DiffEqual {
			dv.hunkStarts = append(dv.hunkStarts, i)
		}
	}
}

// SetSize updates dimensions.
func (dv *DiffView) SetSize(w, h int) {
	dv.width = w
	dv.height = h
}

// IsVisible returns true (always visible when DiffView exists).
func (dv DiffView) IsVisible() bool { return true }

func (dv DiffView) totalLines() int {
	if dv.viewMode == DiffViewSideBySide {
		return len(dv.sideLeft)
	}
	return len(dv.unified)
}

func (dv DiffView) contentHeight() int {
	h := dv.height - 2 // title bar + hint bar
	if dv.searching {
		h--
	}
	return max(5, h)
}

func (dv DiffView) maxOffset() int {
	return max(0, dv.totalLines()-dv.contentHeight())
}

// --- Update ---

func (dv DiffView) Update(msg tea.Msg) (DiffView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if dv.searching {
			return dv.updateSearch(msg)
		}
		return dv.updateNormal(msg)

	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress {
			switch msg.Button {
			case tea.MouseButtonWheelUp:
				dv.scroll(-3)
			case tea.MouseButtonWheelDown:
				dv.scroll(3)
			case tea.MouseButtonWheelLeft:
				dv.xOffset = max(0, dv.xOffset-8)
			case tea.MouseButtonWheelRight:
				dv.xOffset += 8
			}
		}

	case tea.WindowSizeMsg:
		dv.SetSize(msg.Width, msg.Height)
	}
	return dv, nil
}

func (dv DiffView) updateSearch(msg tea.KeyMsg) (DiffView, tea.Cmd) {
	switch msg.String() {
	case "esc":
		dv.searching = false
		dv.searchInput.Blur()
		return dv, nil
	case "enter":
		dv.commitSearch()
		return dv, nil
	default:
		var cmd tea.Cmd
		dv.searchInput, cmd = dv.searchInput.Update(msg)
		dv.searchQuery = strings.TrimSpace(dv.searchInput.Value())
		dv.buildMatchLines()
		if len(dv.matchLines) > 0 {
			dv.matchIdx = 0
			dv.scrollToMatch()
		}
		return dv, cmd
	}
}

func (dv DiffView) updateNormal(msg tea.KeyMsg) (DiffView, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		return dv, func() tea.Msg { return DiffCloseMsg{} }

	// Vertical scroll
	case "j", "down":
		dv.scroll(1)
	case "k", "up":
		dv.scroll(-1)
	case "d", "ctrl+d":
		dv.scroll(dv.contentHeight() / 2)
	case "u", "ctrl+u":
		dv.scroll(-dv.contentHeight() / 2)
	case "f", "pgdown":
		dv.scroll(dv.contentHeight())
	case "b", "pgup":
		dv.scroll(-dv.contentHeight())
	case "g", "home":
		dv.yOffset = 0
		dv.leftOffset = 0
		dv.rightOffset = 0
	case "G", "end":
		dv.yOffset = dv.maxOffset()
		dv.leftOffset = dv.maxOffset()
		dv.rightOffset = dv.maxOffset()

	// Horizontal scroll
	case "h", "left":
		dv.xOffset = max(0, dv.xOffset-8)
	case "l", "right":
		dv.xOffset += 8

	// View mode toggle
	case "tab":
		if dv.viewMode == DiffViewUnified {
			dv.viewMode = DiffViewSideBySide
		} else {
			dv.viewMode = DiffViewUnified
		}
		dv.yOffset = 0
		dv.xOffset = 0
		dv.leftOffset = 0
		dv.rightOffset = 0

	// Search
	case "/":
		dv.searching = true
		dv.searchInput.SetValue(dv.searchQuery)
		dv.searchInput.Focus()
	case "n":
		dv.nextMatch()
	case "N":
		dv.prevMatch()

	// Hunk navigation
	case "]":
		dv.nextHunk()
	case "[":
		dv.prevHunk()
	}

	return dv, nil
}

func (dv *DiffView) scroll(delta int) {
	if dv.viewMode == DiffViewSideBySide && !dv.syncScroll {
		if dv.focus == 0 {
			dv.leftOffset = clamp(dv.leftOffset+delta, 0, dv.maxOffset())
		} else {
			dv.rightOffset = clamp(dv.rightOffset+delta, 0, dv.maxOffset())
		}
	} else {
		dv.yOffset = clamp(dv.yOffset+delta, 0, dv.maxOffset())
		if dv.syncScroll {
			dv.leftOffset = dv.yOffset
			dv.rightOffset = dv.yOffset
		}
	}
}

// --- Hunk navigation ---

func (dv *DiffView) nextHunk() {
	if len(dv.hunkStarts) == 0 {
		return
	}
	for _, hs := range dv.hunkStarts {
		if hs > dv.yOffset {
			dv.yOffset = clamp(hs, 0, dv.maxOffset())
			return
		}
	}
	dv.yOffset = clamp(dv.hunkStarts[0], 0, dv.maxOffset())
}

func (dv *DiffView) prevHunk() {
	if len(dv.hunkStarts) == 0 {
		return
	}
	for i := len(dv.hunkStarts) - 1; i >= 0; i-- {
		if dv.hunkStarts[i] < dv.yOffset {
			dv.yOffset = clamp(dv.hunkStarts[i], 0, dv.maxOffset())
			return
		}
	}
	dv.yOffset = clamp(dv.hunkStarts[len(dv.hunkStarts)-1], 0, dv.maxOffset())
}

// --- Search ---

func (dv *DiffView) commitSearch() {
	dv.searching = false
	dv.searchInput.Blur()
	dv.searchQuery = strings.TrimSpace(dv.searchInput.Value())
	dv.buildMatchLines()
	if len(dv.matchLines) > 0 {
		dv.matchIdx = 0
		dv.scrollToMatch()
	}
}

func (dv *DiffView) buildMatchLines() {
	dv.matchLines = nil
	if dv.searchQuery == "" {
		return
	}
	q := strings.ToLower(dv.searchQuery)
	// Search in the raw content of DiffResult lines (not rendered)
	if dv.result != nil {
		for i, dl := range dv.result.Lines {
			if strings.Contains(strings.ToLower(dl.Content), q) {
				dv.matchLines = append(dv.matchLines, i)
			}
		}
	}
}

func (dv *DiffView) nextMatch() {
	if len(dv.matchLines) == 0 {
		return
	}
	dv.matchIdx = (dv.matchIdx + 1) % len(dv.matchLines)
	dv.scrollToMatch()
}

func (dv *DiffView) prevMatch() {
	if len(dv.matchLines) == 0 {
		return
	}
	dv.matchIdx--
	if dv.matchIdx < 0 {
		dv.matchIdx = len(dv.matchLines) - 1
	}
	dv.scrollToMatch()
}

func (dv *DiffView) scrollToMatch() {
	if dv.matchIdx >= len(dv.matchLines) {
		return
	}
	target := dv.matchLines[dv.matchIdx]
	dv.yOffset = clamp(target-dv.contentHeight()/2, 0, dv.maxOffset())
}

func (dv DiffView) searchInfo() string {
	if dv.searchQuery == "" || len(dv.matchLines) == 0 {
		return ""
	}
	return fmt.Sprintf("%d/%d", dv.matchIdx+1, len(dv.matchLines))
}

// --- View ---

func (dv DiffView) View() string {
	titleSt := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	dimSt := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	keySt := lipgloss.NewStyle().Foreground(lipgloss.Color("63")).Bold(true)
	activeSt := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	addSt := lipgloss.NewStyle().Foreground(diffAddedFg).Bold(true)
	delSt := lipgloss.NewStyle().Foreground(diffRemovedFg).Bold(true)

	// Title bar
	modeLabel := "Unified"
	if dv.viewMode == DiffViewSideBySide {
		modeLabel = "Side-by-Side"
	}
	stats := ""
	if dv.result != nil {
		stats = addSt.Render(fmt.Sprintf("+%d", dv.result.Stats.Additions)) + " " +
			delSt.Render(fmt.Sprintf("-%d", dv.result.Stats.Deletions))
	}
	pct := fmt.Sprintf("%d%%", dv.scrollPercent())
	xInfo := ""
	if dv.xOffset > 0 {
		xInfo = dimSt.Render(fmt.Sprintf(" col:%d", dv.xOffset))
	}
	titleBar := titleSt.Render(dv.title) + "  " + stats + "  " +
		dimSt.Render("["+modeLabel+"]") + "  " + dimSt.Render(pct) + xInfo

	// Content
	viewH := dv.contentHeight()
	var content string
	if dv.viewMode == DiffViewSideBySide {
		content = dv.renderSideBySide(viewH)
	} else {
		content = dv.renderUnified(viewH)
	}

	// Hint bar
	var hints []string
	hints = append(hints, keySt.Render("j/k")+" "+dimSt.Render("vert"))
	hints = append(hints, keySt.Render("h/l")+" "+dimSt.Render("horiz"))
	hints = append(hints, keySt.Render("Tab")+" "+dimSt.Render("mode"))
	hints = append(hints, keySt.Render("]/[")+" "+dimSt.Render("hunk"))
	hints = append(hints, keySt.Render("/")+" "+dimSt.Render("search"))
	if si := dv.searchInfo(); si != "" {
		hints = append(hints, keySt.Render("n/N")+" "+activeSt.Render(si))
	}
	hints = append(hints, keySt.Render("q")+" "+dimSt.Render("close"))
	hintLine := " " + strings.Join(hints, "  ")

	var sb strings.Builder
	sb.WriteString(titleBar)
	sb.WriteString("\n")
	sb.WriteString(content)
	sb.WriteString("\n")

	if dv.searching {
		matchInfo := ""
		if q := strings.TrimSpace(dv.searchInput.Value()); q != "" {
			matchInfo = fmt.Sprintf(" (%d matches)", len(dv.matchLines))
		}
		sb.WriteString(dv.searchInput.View() + dimSt.Render(matchInfo))
	} else {
		sb.WriteString(hintLine)
	}

	return sb.String()
}

func (dv DiffView) renderUnified(viewH int) string {
	lines := dv.unified
	total := len(lines)
	showScrollbar := total > viewH

	trackSt := lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	thumbSt := lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	var thumbStart, thumbEnd int
	if showScrollbar {
		thumbSize := max(1, viewH*viewH/total)
		if m := dv.maxOffset(); m > 0 {
			thumbStart = dv.yOffset * (viewH - thumbSize) / m
		}
		thumbEnd = thumbStart + thumbSize
	}

	scrollW := 0
	if showScrollbar {
		scrollW = 1
	}

	// Calculate content width after prefix
	// Prefix is fixed, content gets the remaining width
	prefixW := 0
	if len(lines) > 0 {
		prefixW = lipgloss.Width(lines[0].Prefix)
	}
	contentW := max(10, dv.width-prefixW-scrollW)

	var sb strings.Builder
	for vi := range viewH {
		lineIdx := dv.yOffset + vi
		var line string
		if lineIdx < total {
			rl := lines[lineIdx]
			// Prefix is sticky (always visible)
			content := hslice(rl.Content, dv.xOffset, contentW)
			// Pad content to fill width
			if pad := contentW - lipgloss.Width(content); pad > 0 {
				content += strings.Repeat(" ", pad)
			}
			line = rl.Prefix + content
		} else {
			line = strings.Repeat(" ", dv.width-scrollW)
		}

		if showScrollbar {
			if vi >= thumbStart && vi < thumbEnd {
				line += thumbSt.Render("█")
			} else {
				line += trackSt.Render("│")
			}
		}

		sb.WriteString(line)
		if vi < viewH-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

func (dv DiffView) renderSideBySide(viewH int) string {
	dividerSt := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	total := len(dv.sideLeft)
	showScrollbar := total > viewH

	trackSt := lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	thumbSt := lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	var thumbStart, thumbEnd int
	if showScrollbar {
		thumbSize := max(1, viewH*viewH/total)
		if m := dv.maxOffset(); m > 0 {
			thumbStart = dv.yOffset * (viewH - thumbSize) / m
		}
		thumbEnd = thumbStart + thumbSize
	}

	scrollW := 0
	if showScrollbar {
		scrollW = 1
	}
	dividerW := 3 // " │ "
	paneTotal := (dv.width - dividerW - scrollW) / 2

	// Calculate prefix width from rendered data
	prefixW := 0
	if len(dv.sideLeft) > 0 {
		prefixW = lipgloss.Width(dv.sideLeft[0].Prefix)
	}
	contentW := max(5, paneTotal-prefixW)

	var sb strings.Builder
	for vi := range viewH {
		lineIdx := dv.yOffset + vi

		var leftStr, rightStr string
		if lineIdx < total {
			ll := dv.sideLeft[lineIdx]
			leftContent := hslice(ll.Content, dv.xOffset, contentW)
			if pad := contentW - lipgloss.Width(leftContent); pad > 0 {
				leftContent += strings.Repeat(" ", pad)
			}
			leftStr = ll.Prefix + leftContent
		} else {
			leftStr = strings.Repeat(" ", paneTotal)
		}

		if lineIdx < len(dv.sideRight) {
			rl := dv.sideRight[lineIdx]
			rightContent := hslice(rl.Content, dv.xOffset, contentW)
			if pad := contentW - lipgloss.Width(rightContent); pad > 0 {
				rightContent += strings.Repeat(" ", pad)
			}
			rightStr = rl.Prefix + rightContent
		} else {
			rightStr = strings.Repeat(" ", paneTotal)
		}

		line := leftStr + dividerSt.Render(" │ ") + rightStr

		if showScrollbar {
			if vi >= thumbStart && vi < thumbEnd {
				line += thumbSt.Render("█")
			} else {
				line += trackSt.Render("│")
			}
		}

		sb.WriteString(line)
		if vi < viewH-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

func (dv DiffView) scrollPercent() int {
	m := dv.maxOffset()
	if m <= 0 {
		return 100
	}
	return int(float64(dv.yOffset) / float64(m) * 100)
}
