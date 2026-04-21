# TUI MDL Execution Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add MDL script execution capability to the TUI — both from pasted text and from file selection.

**Architecture:** New `ExecView` implementing the `view` interface, with a `textarea` for MDL input and a file picker fallback. Execution delegates to `runMxcli("exec", ...)` subprocess (consistent with existing TUI patterns). Results display in an OverlayView; project tree refreshes after successful execution.

**Tech Stack:** `github.com/charmbracelet/bubbles/textarea`, existing TUI View/ViewStack infrastructure

---

### Task 1: Add ModeExec and ExecResultMsg

**Files:**
- Modify: `cmd/mxcli/tui/view.go` (add `ModeExec` constant and String case)
- Modify: `cmd/mxcli/tui/hintbar.go` (add `ExecViewHints`)

**Step 1: Add ModeExec to ViewMode**

In `cmd/mxcli/tui/view.go`, add `ModeExec` after `ModeJumper`:

```go
const (
    ModeBrowser ViewMode = iota
    ModeOverlay
    ModeCompare
    ModeDiff
    ModePicker
    ModeJumper
    ModeExec
)
```

And in the `string()` method, add:
```go
case ModeExec:
    return "exec"
```

**Step 2: Add ExecViewHints to hintbar.go**

```go
ExecViewHints = []Hint{
    {key: "Ctrl+E", label: "execute"},
    {key: "Ctrl+O", label: "open file"},
    {key: "Esc", label: "close"},
}
```

**Step 3: Run build to verify**

Run: `cd /mnt/data_sdd/gh/mxcli-wt-01 && go build ./cmd/mxcli/tui/...`
Expected: PASS (no new references yet)

**Step 4: Commit**

```bash
git add cmd/mxcli/tui/view.go cmd/mxcli/tui/hintbar.go
git commit -m "feat(tui): add ModeExec view mode and ExecViewHints"
```

---

### Task 2: Create ExecView with textarea

**Files:**
- Create: `cmd/mxcli/tui/execview.go`
- Create: `cmd/mxcli/tui/execview_test.go`

**Step 1: Write test for ExecView construction and Mode**

Create `cmd/mxcli/tui/execview_test.go`:

```go
package tui

import "testing"

func TestExecView_Mode(t *testing.T) {
    ev := NewExecView("mxcli", "/tmp/test.mpr", 80, 24)
    if ev.Mode() != ModeExec {
        t.Errorf("expected ModeExec, got %v", ev.Mode())
    }
}

func TestExecView_StatusInfo(t *testing.T) {
    ev := NewExecView("mxcli", "/tmp/test.mpr", 80, 24)
    info := ev.StatusInfo()
    if info.Mode != "exec" {
        t.Errorf("expected mode 'Exec', got %q", info.Mode)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd /mnt/data_sdd/gh/mxcli-wt-01 && go test ./cmd/mxcli/tui/ -run TestExecView -v`
Expected: FAIL (NewExecView undefined)

**Step 3: Implement ExecView**

Create `cmd/mxcli/tui/execview.go`:

```go
package tui

import (
    "fmt"
    "os"
    "os/exec"
    "strings"

    "github.com/charmbracelet/bubbles/textarea"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

// ExecDoneMsg carries the result of MDL execution.
type ExecDoneMsg struct {
    Output string
    Err    error
}

// ExecView provides a textarea for entering/pasting MDL scripts and executing them.
type ExecView struct {
    textarea    textarea.Model
    mxcliPath   string
    projectPath string
    width       int
    height      int
    executing   bool
    flash       string // status message
}

// NewExecView creates an ExecView with a textarea for MDL input.
func NewExecView(mxcliPath, projectPath string, width, height int) ExecView {
    ta := textarea.New()
    ta.Placeholder = "Paste or type MDL script here...\n\nCtrl+E to execute, Ctrl+O to open file, Esc to close"
    ta.ShowLineNumbers = true
    ta.Focus()
    ta.SetWidth(width - 4)
    ta.SetHeight(height - 6) // room for title and status line

    return ExecView{
        textarea:    ta,
        mxcliPath:   mxcliPath,
        projectPath: projectPath,
        width:       width,
        height:      height,
    }
}

func (ev ExecView) Mode() ViewMode {
    return ModeExec
}

func (ev ExecView) Hints() []Hint {
    return ExecViewHints
}

func (ev ExecView) StatusInfo() StatusInfo {
    lines := strings.Count(ev.textarea.Value(), "\n") + 1
    return StatusInfo{
        Breadcrumb: []string{"execute MDL"},
        position:   fmt.Sprintf("L%d", lines),
        Mode:       "exec",
    }
}

func (ev ExecView) Render(width, height int) string {
    ev.textarea.SetWidth(width - 4)
    ev.textarea.SetHeight(height - 6)

    titleStyle := lipgloss.NewStyle().Bold(true).Foreground(AccentColor).Padding(0, 1)
    title := titleStyle.Render("execute MDL")

    statusLine := ""
    if ev.executing {
        statusLine = lipgloss.NewStyle().Foreground(WarningColor).Render("  Executing...")
    } else if ev.flash != "" {
        statusLine = lipgloss.NewStyle().Foreground(MutedColor).Render("  " + ev.flash)
    }

    content := lipgloss.JoinVertical(lipgloss.Left,
        title,
        ev.textarea.View(),
        statusLine,
    )

    return lipgloss.NewStyle().Padding(1, 2).Render(content)
}

func (ev ExecView) update(msg tea.Msg) (view, tea.Cmd) {
    switch msg := msg.(type) {
    case ExecDoneMsg:
        ev.executing = false
        content := msg.Output
        if msg.Err != nil {
            content = "-- Error:\n" + msg.Output
        }
        // Push result overlay and pop exec view
        return ev, func() tea.Msg {
            return execShowResultMsg{content: content, success: msg.Err == nil}
        }

    case tea.KeyMsg:
        if ev.executing {
            return ev, nil // ignore keys during execution
        }

        switch msg.String() {
        case "esc":
            if ev.textarea.Value() == "" {
                return ev, func() tea.Msg { return PopViewMsg{} }
            }
            // if there's content, first Esc clears flash; second Esc closes
            if ev.flash != "" {
                ev.flash = ""
                return ev, nil
            }
            return ev, func() tea.Msg { return PopViewMsg{} }

        case "ctrl+e":
            mdlText := strings.TrimSpace(ev.textarea.Value())
            if mdlText == "" {
                ev.flash = "nothing to execute"
                return ev, nil
            }
            ev.executing = true
            return ev, ev.executeMDL(mdlText)

        case "ctrl+o":
            return ev, ev.openFileDialog()
        }

        // Forward to textarea
        var cmd tea.Cmd
        ev.textarea, cmd = ev.textarea.Update(msg)
        return ev, cmd
    }

    var cmd tea.Cmd
    ev.textarea, cmd = ev.textarea.Update(msg)
    return ev, cmd
}

// executeMDL writes MDL to a temp file and runs `mxcli exec`.
func (ev ExecView) executeMDL(mdlText string) tea.Cmd {
    mxcliPath := ev.mxcliPath
    projectPath := ev.projectPath
    return func() tea.Msg {
        // write to temp file
        tmpFile, err := os.CreateTemp("", "mxcli-exec-*.mdl")
        if err != nil {
            return ExecDoneMsg{Output: fmt.Sprintf("Failed to create temp file: %v", err), Err: err}
        }
        tmpPath := tmpFile.Name()
        defer os.Remove(tmpPath)

        if _, err := tmpFile.WriteString(mdlText); err != nil {
            tmpFile.Close()
            return ExecDoneMsg{Output: fmt.Sprintf("Failed to write temp file: %v", err), Err: err}
        }
        tmpFile.Close()

        args := []string{"exec"}
        if projectPath != "" {
            args = append(args, "-p", projectPath)
        }
        args = append(args, tmpPath)
        out, err := runMxcli(mxcliPath, args...)
        return ExecDoneMsg{Output: out, Err: err}
    }
}

// openFileDialog uses the system file picker or a simple stdin prompt.
// for now, it reads from a well-known env var or prompts via the picker.
func (ev ExecView) openFileDialog() tea.Cmd {
    return func() tea.Msg {
        return execOpenFileMsg{}
    }
}

// execShowResultMsg signals App to show exec result and optionally refresh tree.
type execShowResultMsg struct {
    content string
    success bool
}

// execOpenFileMsg signals App to open the file picker for MDL files.
type execOpenFileMsg struct{}
```

**Step 4: Run test to verify it passes**

Run: `cd /mnt/data_sdd/gh/mxcli-wt-01 && go test ./cmd/mxcli/tui/ -run TestExecView -v`
Expected: PASS

**Step 5: Commit**

```bash
git add cmd/mxcli/tui/execview.go cmd/mxcli/tui/execview_test.go
git commit -m "feat(tui): add ExecView with textarea for MDL input"
```

---

### Task 3: Wire ExecView into App

**Files:**
- Modify: `cmd/mxcli/tui/app.go` (handle key `x` in browser mode, handle ExecDoneMsg, execShowResultMsg, execOpenFileMsg)
- Modify: `cmd/mxcli/tui/help.go` (add exec entry)

**Step 1: Add key `x` handler in `handleBrowserAppKeys`**

In `app.go`, inside `handleBrowserAppKeys`, add before the final `return nil`:

```go
case "x":
    ev := NewExecView(a.mxcliPath, a.activeTabProjectPath(), a.width, a.height)
    a.views.Push(ev)
    return func() tea.Msg { return nil }
```

**Step 2: Handle execShowResultMsg in App.Update**

Add a case in the `update` switch:

```go
case execShowResultMsg:
    // Pop the ExecView
    a.views.Pop()
    // show result in overlay
    content := DetectAndHighlight(msg.Content)
    ov := NewOverlayView("exec Result", content, a.width, a.height, OverlayViewOpts{})
    a.views.Push(ov)
    // if execution succeeded, refresh tree
    if msg.Success {
        return a, a.Init()
    }
    return a, nil
```

**Step 3: Handle execOpenFileMsg in App.Update**

Add a case for file picker:

```go
case execOpenFileMsg:
    // Re-use the existing picker to pick an MDL file, then load its content
    // into the ExecView textarea.
    if execView, ok := a.views.Active().(ExecView); ok {
        return a, execView.pickFile()
    }
    return a, nil
```

Add `pickFile` method to ExecView in `execview.go`:

```go
// pickFile opens a native file dialog or reads path from env.
func (ev ExecView) pickFile() tea.Cmd {
    return func() tea.Msg {
        // Try zenity / kdialog for file selection
        for _, picker := range []struct {
            bin  string
            args []string
        }{
            {"zenity", []string{"--file-selection", "--file-filter=MDL files (*.mdl)|*.mdl"}},
            {"kdialog", []string{"--getopenfilename", ".", "*.mdl"}},
        } {
            if binPath, err := exec.LookPath(picker.bin); err == nil {
                cmd := exec.Command(binPath, picker.args...)
                out, err := cmd.Output()
                if err == nil {
                    path := strings.TrimSpace(string(out))
                    if path != "" {
                        content, err := os.ReadFile(path)
                        if err != nil {
                            return execFileLoadedMsg{Err: err}
                        }
                        return execFileLoadedMsg{path: path, content: string(content)}
                    }
                }
                return execFileLoadedMsg{} // user cancelled
            }
        }
        return execFileLoadedMsg{Err: fmt.Errorf("no file picker available (install zenity or kdialog)")}
    }
}
```

Add message type and handler in `execview.go`:

```go
type execFileLoadedMsg struct {
    path    string
    content string
    Err     error
}
```

And in ExecView.Update, add:

```go
case execFileLoadedMsg:
    if msg.Err != nil {
        ev.flash = fmt.Sprintf("error: %v", msg.Err)
        return ev, nil
    }
    if msg.Content != "" {
        ev.textarea.SetValue(msg.Content)
        ev.flash = fmt.Sprintf("Loaded: %s", msg.Path)
    }
    return ev, nil
```

**Step 4: Update help text**

In `help.go`, add under ACTIONS:

```
    x     execute MDL script
```

**Step 5: Add hint for x in ListBrowsingHints**

In `hintbar.go`, add before `{key: "?", label: "help"}`:

```go
{key: "x", label: "exec"},
```

**Step 6: Build and verify**

Run: `cd /mnt/data_sdd/gh/mxcli-wt-01 && go build ./cmd/mxcli/...`
Expected: PASS

**Step 7: Commit**

```bash
git add cmd/mxcli/tui/app.go cmd/mxcli/tui/execview.go cmd/mxcli/tui/help.go cmd/mxcli/tui/hintbar.go
git commit -m "feat(tui): wire ExecView into App with key 'x', file picker, and tree refresh"
```

---

### Task 4: Handle ExecDoneMsg forwarding in App

**Files:**
- Modify: `cmd/mxcli/tui/app.go`

The `ExecDoneMsg` is dispatched to the active view (ExecView) via the default case. The ExecView then emits `execShowResultMsg` which App handles. This should work with the existing message forwarding pattern.

**Step 1: Verify the message flow**

Verify that `ExecDoneMsg` flows through the default case in `App.Update`:

```go
default:
    updated, cmd := a.views.Active().Update(msg)
    a.views.SetActive(updated)
    return a, cmd
```

This already forwards unknown messages to the active view, so `ExecDoneMsg` will reach `ExecView.Update`.

**Step 2: Add `execShowResultMsg` and `execFileLoadedMsg` to App.Update**

These need explicit cases because they require App-level actions:

```go
case execShowResultMsg:
    // handled in step 3.2 above
case execOpenFileMsg:
    // handled in step 3.3 above
```

`execFileLoadedMsg` can flow through the default case to ExecView since it only modifies ExecView state.

**Step 3: Manual test**

Run: `cd /mnt/data_sdd/gh/mxcli-wt-01 && go run ./cmd/mxcli tui -p /mnt/data_sdd/gh/mxproj-GenAIDemo/App.mpr`

- Press `x` → ExecView should appear with textarea
- Type `show modules;` → Press Ctrl+E → Should execute and show result
- Press `q` to close result → Tree should refresh
- Press `x` again → Press Ctrl+O → File picker should open (if zenity installed)

**Step 4: Commit if any fixes were needed**

```bash
git add -u
git commit -m "fix(tui): fix exec message forwarding"
```

---

### Task 5: Add ExecView window size handling

**Files:**
- Modify: `cmd/mxcli/tui/execview.go`

**Step 1: Handle WindowSizeMsg**

In ExecView.Update, before the default textarea forwarding, add:

```go
case tea.WindowSizeMsg:
    ev.width = msg.Width
    ev.height = msg.Height
    ev.textarea.SetWidth(msg.Width - 4)
    ev.textarea.SetHeight(msg.Height - 6)
    return ev, nil
```

**Step 2: Build and verify**

Run: `cd /mnt/data_sdd/gh/mxcli-wt-01 && go build ./cmd/mxcli/...`

**Step 3: Commit**

```bash
git add cmd/mxcli/tui/execview.go
git commit -m "fix(tui): handle window resize in ExecView"
```

---

### Task 6: Final integration test and cleanup

**Step 1: Run all TUI tests**

Run: `cd /mnt/data_sdd/gh/mxcli-wt-01 && go test ./cmd/mxcli/tui/ -v`
Expected: All pass

**Step 2: Run full build**

Run: `cd /mnt/data_sdd/gh/mxcli-wt-01 && make build`
Expected: Success

**Step 3: Manual E2E test**

1. `./bin/mxcli tui -p /mnt/data_sdd/gh/mxproj-GenAIDemo/App.mpr`
2. Press `x` → verify textarea appears
3. Paste: `show modules;` → Ctrl+E → verify output overlay
4. Press `q` → verify back to browser, tree refreshed
5. Press `x` → Ctrl+O → verify file picker (or graceful error)
6. Press `Esc` → verify returns to browser

**Step 4: Final commit**

```bash
git add -A
git commit -m "feat(tui): MDL execution from TUI (paste or file) - closes #30"
```
