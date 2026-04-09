package tui

import (
	"fmt"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- View ---

func (a App) View() string {
	if a.width == 0 {
		return "mxcli tui — loading...\n\nPress q to quit"
	}

	if a.picker != nil {
		return a.picker.View()
	}

	active := a.views.Active()

	// For non-browser views, delegate rendering entirely
	if active.Mode() != ModeBrowser {
		contentH := a.height
		content := active.Render(a.width, contentH)

		if a.showHelp {
			helpView := renderHelp(a.width, a.height)
			content = lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, helpView,
				lipgloss.WithWhitespaceBackground(lipgloss.Color("0")))
		}

		return content
	}

	// Browser mode: App renders chrome (tab bar, hint bar, status bar)

	// Tab bar (line 1) with mode badge + context summary on the right
	tabLine := a.tabBar.View(a.width)
	tab := a.activeTabPtr()
	if tab != nil {
		modeBadge := AccentStyle.Render(active.Mode().String())
		summary := renderContextSummary(tab.AllNodes)
		rightSide := modeBadge
		if summary != "" {
			rightSide += BreadcrumbDimStyle.Render(" │ ") + BreadcrumbDimStyle.Render(summary)
		}
		rightWidth := lipgloss.Width(rightSide) + 1 // 1 char right padding
		tabWidth := lipgloss.Width(tabLine)
		if tabWidth+rightWidth <= a.width {
			// Replace trailing spaces with gap + right side
			trimmed := strings.TrimRight(tabLine, " ")
			trimmedWidth := lipgloss.Width(trimmed)
			gap := a.width - trimmedWidth - rightWidth
			if gap < 2 {
				gap = 2
			}
			tabLine = trimmed + strings.Repeat(" ", gap) + rightSide + " "
		}
	}

	// Content area (chromeHeight + 1 for the LLM anchor line)
	contentH := a.height - chromeHeight - 1
	if contentH < 5 {
		contentH = 5
	}
	content := active.Render(a.width, contentH)

	// Hint bar — declarative from active view
	a.hintBar.SetHints(active.Hints())
	hintLine := a.hintBar.View(a.width)

	// Status bar — declarative from active view
	info := active.StatusInfo()
	a.statusBar.SetBreadcrumb(info.Breadcrumb)
	a.statusBar.SetPosition(info.Position)
	a.statusBar.SetMode(info.Mode)
	if a.checkNavActive && len(a.checkNavLocations) > 0 {
		loc := a.checkNavLocations[a.checkNavIndex]
		navInfo := fmt.Sprintf("[%d/%d] %s: %s  ]e next  [e prev",
			a.checkNavIndex+1, len(a.checkNavLocations),
			loc.Code, docNameToQualifiedName(loc.ModuleName, loc.DocumentName))
		a.statusBar.SetCheckBadge(CheckWarnStyle.Render(navInfo))
	} else {
		a.statusBar.SetCheckBadge(formatCheckBadge(a.checkErrors, a.checkRunning))
	}
	// Agent activity badge
	if a.agentExecCtx != nil {
		a.statusBar.SetAgentBadge(AgentBadgeStyle.Render("⚡agent"))
	} else {
		a.statusBar.SetAgentBadge("")
	}
	viewModeNames := a.collectViewModeNames()
	a.statusBar.SetViewDepth(a.views.Depth(), viewModeNames)
	statusLine := StatusBarStyle.Width(a.width).Render(a.statusBar.View(a.width))

	// LLM anchor: machine-readable command list (Faint, not visible to users in practice)
	anchorStyle := lipgloss.NewStyle().Foreground(MutedColor).Faint(true)
	anchorLine := anchorStyle.Render("[mxcli:commands] h:back l:open Space:jump /:filter b:bson c:compare d:diagram z:zen Tab:toggle x:exec r:refresh y:copy !:check ]e:next-error [e:prev-error t:tab T:new-tab W:close-tab 1-9:switch ?:help ::palette")

	rendered := anchorLine + "\n" + tabLine + "\n" + content + "\n" + hintLine + "\n" + statusLine

	if a.showHelp {
		helpView := renderHelp(a.width, a.height)
		rendered = lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, helpView,
			lipgloss.WithWhitespaceBackground(lipgloss.Color("0")))
	}

	return rendered
}

// --- Load helpers ---

func buildDiagramHTML(elkJSON, nodeType, qualifiedName string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html><head><title>%s %s</title>
<script src="https://cdn.jsdelivr.net/npm/elkjs@0.9.3/lib/elk.bundled.js"></script>
<style>body{margin:0;background:#1e1e2e;color:#cdd6f4;font-family:monospace}svg{width:100vw;height:100vh}</style>
</head><body><div id="diagram"></div><script>
const elkData = %s;
const ELK = new ELKConstructor();
ELK.layout(elkData).then(graph=>{
  const svg=document.createElementNS("http://www.w3.org/2000/svg","svg");
  document.getElementById("diagram").appendChild(svg);
});
</script></body></html>`, nodeType, qualifiedName, elkJSON)
}

// CmdResultMsg carries output from any mxcli command.
type CmdResultMsg struct {
	Output string
	Err    error
}

// irregularPlurals maps singular type names to their correct plural forms
// for types where simply appending "s" produces incorrect English.
var irregularPlurals = map[string]string{
	"Index": "indexes",
}

// renderContextSummary counts top-level node types and returns a compact summary.
func renderContextSummary(nodes []*TreeNode) string {
	if len(nodes) == 0 {
		return ""
	}
	counts := map[string]int{}
	for _, n := range nodes {
		counts[n.Type]++
	}
	// Display in a predictable order
	order := []struct {
		key    string
		plural string
	}{
		{"Module", "modules"},
		{"Entity", "entities"},
		{"Microflow", "microflows"},
		{"Page", "pages"},
		{"Nanoflow", "nanoflows"},
		{"Enumeration", "enumerations"},
	}
	var parts []string
	used := map[string]bool{}
	for _, o := range order {
		if c, ok := counts[o.key]; ok {
			parts = append(parts, fmt.Sprintf("%d %s", c, o.plural))
			used[o.key] = true
		}
	}
	// Add remaining types not in the predefined order
	for k, c := range counts {
		if !used[k] {
			plural, ok := irregularPlurals[k]
			if !ok {
				plural = strings.ToLower(k) + "s"
			}
			parts = append(parts, fmt.Sprintf("%d %s", c, plural))
		}
	}
	if len(parts) > 3 {
		parts = parts[:3]
	}
	return strings.Join(parts, ", ")
}

// collectViewModeNames returns the mode names for all views in the stack.
func (a App) collectViewModeNames() []string {
	return a.views.ModeNames()
}

// inferBsonType maps tree node types to valid bson object types.
func inferBsonType(nodeType string) string {
	switch strings.ToLower(nodeType) {
	case "page", "microflow", "nanoflow", "workflow",
		"enumeration", "snippet", "layout", "entity", "association",
		"imagecollection", "javaaction", "javascriptaction", "constant":
		return strings.ToLower(nodeType)
	default:
		return ""
	}
}

// agentStateInfo is the structured state returned by the "state" action.
type agentStateInfo struct {
	Mode         string         `json:"mode"`
	Project      string         `json:"project"`
	SelectedNode *agentNodeInfo `json:"selectedNode,omitempty"`
	PreviewMode  string         `json:"previewMode,omitempty"`
	CheckErrors  int            `json:"checkErrors"`
	CheckRunning bool           `json:"checkRunning"`
}

// agentNodeInfo describes the currently selected tree node.
type agentNodeInfo struct {
	Type          string `json:"type"`
	QualifiedName string `json:"qualifiedName"`
}

// agentBuildState extracts rich TUI state for the agent.
func agentBuildState(a App) agentStateInfo {
	info := agentStateInfo{
		Mode:         a.views.Active().Mode().String(),
		Project:      a.activeTabProjectPath(),
		CheckErrors:  len(a.checkErrors),
		CheckRunning: a.checkRunning,
	}
	if bv, ok := a.views.Base().(BrowserView); ok {
		info.PreviewMode = "MDL"
		if bv.miller.preview.mode == PreviewNDSL {
			info.PreviewMode = "NDSL"
		}
		if node := bv.miller.SelectedNode(); node != nil {
			qname := node.QualifiedName
			if qname == "" {
				qname = node.Label
			}
			info.SelectedNode = &agentNodeInfo{
				Type:          node.Type,
				QualifiedName: qname,
			}
		}
	}
	return info
}

// agentExecChanges is a structured summary of exec output changes.
type agentExecChange struct {
	Action string `json:"action"` // "created", "modified", "dropped"
	Target string `json:"target"` // e.g. "entity Module.Entity"
}

// agentChangePattern matches exec output lines like "Created entity MyModule.Customer".
// Requires a known Mendix type keyword after the verb to avoid matching log noise
// such as "Removed trailing whitespace".
var agentChangePattern = regexp.MustCompile(`(?im)^(Created|Modified|Dropped|Deleted|Added|Removed)\s+(entity|association|attribute|enumeration|microflow|nanoflow|page|layout|snippet|module|folder|constant|workflow|image collection|java action|user role|module role|demo user|business event service)\s+(.+)$`)

// agentParseChanges extracts structured changes from exec output.
func agentParseChanges(output string) []agentExecChange {
	matches := agentChangePattern.FindAllStringSubmatch(output, -1)
	if len(matches) == 0 {
		return nil
	}
	changes := make([]agentExecChange, 0, len(matches))
	for _, m := range matches {
		changes = append(changes, agentExecChange{
			Action: strings.ToLower(m[1]),
			Target: strings.ToLower(m[2]) + " " + strings.TrimSpace(m[3]),
		})
	}
	return changes
}

// Shared by BrowserView and CompareView to avoid duplicate implementations.
func loadBsonNDSL(mxcliPath, projectPath, qname, nodeType string, side CompareFocus) tea.Cmd {
	return func() tea.Msg {
		bsonType := inferBsonType(nodeType)
		if bsonType == "" {
			return CompareLoadMsg{Side: side, Title: qname, NodeType: nodeType,
				Content: fmt.Sprintf("Error: type %q not supported for BSON dump", nodeType),
				Err:     fmt.Errorf("unsupported type")}
		}
		args := []string{"bson", "dump", "-p", projectPath, "--format", "ndsl",
			"--type", bsonType, "--object", qname}
		out, err := runMxcli(mxcliPath, args...)
		out = StripBanner(out)
		if err != nil {
			return CompareLoadMsg{Side: side, Title: qname, NodeType: nodeType, Content: "Error: " + out, Err: err}
		}
		return CompareLoadMsg{Side: side, Title: qname, NodeType: nodeType, Content: HighlightNDSL(out)}
	}
}
