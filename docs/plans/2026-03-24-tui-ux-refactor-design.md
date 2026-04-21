# TUI 2.0 UX Refactor Design

**Date**: 2026-03-24
**Status**: Approved
**Branch**: feat/tui-ux-refactor (from feat/tui-diff-view)

## Problem Statement

The current TUI has grown organically with multiple view modes (Miller browser, overlay, compare, diff) stacked on top of each other through implicit boolean/pointer checks. This causes:

1. **Mode confusion** — Users can't tell which view layer they're in; each mode has different keybindings with no visual indicator of the current context
2. **Visual flatness** — No color differentiation between focused/unfocused columns; all areas look the same due to monochrome Bold/Faint/Reverse styling
3. **Navigation inefficiency** — No global jump-to-node; preview debounce missing; `Tab` key overloaded across contexts
4. **God Object** — `app.go` is 760 lines managing all view routing, state sync scattered across 30+ call sites

### Dual Audience

Both humans and LLMs read the TUI output (via tmux capture-pane). The redesign must serve both:
- Humans: visual hierarchy, color, focus indicators
- LLMs: structured text anchors, plain-text yank, parseable mode indicators

## Architecture

### View System

```
┌─────────────────────────────────────────────────┐
│ Chrome (header)                                  │  ← Fixed
│   TabBar | ViewMode Badge | context Summary      │
├─────────────────────────────────────────────────┤
│                                                  │
│                  Active view                     │  ← Replaceable
│  (BrowserView / CompareView / DiffView / etc.)   │
│                                                  │
├─────────────────────────────────────────────────┤
│ Chrome (footer)                                  │  ← Fixed
│   HintBar (context-sensitive from active view)   │
│   StatusBar (breadcrumb + position + mode)       │
└─────────────────────────────────────────────────┘
```

#### Unified View Interface

```go
// view is the interface all TUI views must implement.
type view interface {
    update(tea.Msg) (view, tea.Cmd)
    Render(width, height int) string
    Hints() []Hint
    StatusInfo() StatusInfo
    Mode() ViewMode
}

type ViewMode int
const (
    ModeBrowser ViewMode = iota
    ModeOverlay
    ModeCompare
    ModeDiff
    ModePicker
)

type StatusInfo struct {
    Breadcrumb []string
    position   string  // e.g. "3/47"
    Mode       string  // e.g. "MDL", "NDSL"
    Extra      string  // view-specific info
}
```

#### ViewStack

Replaces the current `a.diffView != nil` / `a.overlay.IsVisible()` / `a.compare.IsVisible()` priority chain:

```go
type ViewStack struct {
    base  view     // always BrowserView
    stack []view   // overlay/compare/diff pushed on top
}

func (vs *ViewStack) Active() view
func (vs *ViewStack) Push(v view)
func (vs *ViewStack) Pop() view
func (vs *ViewStack) depth() int
```

#### Simplified App

`app.go` reduces to ~200 lines:

```go
func (a App) update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // 1. Handle global keys (ctrl+c, tab switching, Space for jump)
    // 2. Delegate to active view
    active := a.views.Active()
    updated, cmd := active.Update(msg)
    a.views.SetActive(updated)
    return a, cmd
}

func (a App) view() string {
    active := a.views.Active()
    header := a.renderHeader(active)
    content := active.Render(a.width, a.contentHeight())
    footer := a.renderFooter(active)
    return header + "\n" + content + "\n" + footer
}
```

Chrome (header/footer) is rendered by App using data from `active.Hints()` and `active.StatusInfo()` — **declarative, no manual sync calls**.

## Visual Design

### Color System (Terminal-Adaptive)

```go
// theme.go — semantic color tokens using AdaptiveColor
var (
    FocusColor   = lipgloss.AdaptiveColor{Light: "62", Dark: "63"}   // blue-purple
    AccentColor  = lipgloss.AdaptiveColor{Light: "214", Dark: "214"} // orange
    MutedColor   = lipgloss.AdaptiveColor{Light: "245", Dark: "243"} // gray
    AddedColor   = lipgloss.AdaptiveColor{Light: "28", Dark: "114"}  // green
    RemovedColor = lipgloss.AdaptiveColor{Light: "124", Dark: "210"} // red
)
```

### Focus Indicators

- **Focused column title**: FocusColor foreground + Bold
- **Focused column left edge**: FocusColor `▎` vertical bar (1 char)
- **Unfocused columns**: entire content rendered with Faint
- **Preview title**: shows MDL/NDSL mode in AccentColor

### Header Enhancement

```
before:  1:App.mpr  2:Other.mpr
after:   ❶ App.mpr  ❷ Other.mpr          Browse │ 3 modules, 47 entities
```

- Left: tabs with FocusColor underline on active tab
- Center: ViewMode badge (Browse / Compare / Diff / Overlay)
- Right: context summary (node counts from project tree)

### Footer Enhancement

```
before:  h:back  l:open  /:filter  Tab:mdl/ndsl  ...
         MyModule > entities > Customer                    3/47  MDL

after:   h back  l open  / filter  ⇥ mdl/ndsl  c compare  ? help
         MyModule › entities › Customer                   3/47 ⎸ MDL
```

- HintBar: remove `:` separator, use space instead (more compact)
- StatusBar: use `›` for breadcrumb, `⎸` for mode separator
- When ViewStack depth > 1, show depth: `[Browse > Compare > Diff]`

## Navigation & Interaction

### Global Fuzzy Jump

Press `Space` or `Ctrl+P` → opens a fuzzy finder overlay:

```
┌──────────────────────────────────────┐
│ > cust_                              │
│   🏢 MyModule.Customer       entity  │
│   📄 MyModule.Customer_Overview page │
│   ⚡ MyModule.ACT_Customer_Create MF │
└──────────────────────────────────────┘
```

- Reuses existing `flattenQualifiedNames()` + fuzzy match logic
- On selection: navigates Miller columns to the node (expands path)
- LLM-friendly: results are plain text, capturable via tmux

### Preview Debounce

Add 150ms debounce to cursor change → preview request:

```go
func (m MillerView) handleCursorChanged(msg CursorChangedMsg) (MillerView, tea.Cmd) {
    m.pendingPreviewNode = msg.Node
    return m, tea.Tick(150*time.Millisecond, func(t time.Time) tea.Msg {
        return previewDebounceMsg{node: msg.Node}
    })
}
```

Prevents flooding mxcli subprocesses during fast j/k scrolling.

### Keybinding Redesign

| Key | Global | Browser | Overlay | Compare | Diff |
|-----|--------|---------|---------|---------|------|
| `q` / `Esc` | Pop ViewStack (quit if empty) | — | — | — | — |
| `Space` | Global fuzzy jump | — | — | — | — |
| `Tab` | — | MDL/NDSL toggle | MDL/NDSL toggle | — | Diff mode cycle |
| `1-9` | Tab switch (Browser) | — | — | Compare mode | — |
| `y` | Copy content | Copy preview | Copy content | Copy diff | Copy diff |

Key changes:
- **`q`/`Esc` unified** as "exit current layer" (ViewStack.Pop)
- **`Space` for global jump** (currently unused)
- **`1-9` context-sensitive**: tab switch in Browser, compare mode in Compare

### Clickable Breadcrumb

Each segment in the status bar breadcrumb registers as a mouse zone. Clicking a segment calls `goBack()` to that navigation level.

## LLM Friendliness

### Structured Anchors

Each view's first rendered line includes a machine-parseable prefix:

```
[mxcli:browse] MyModule > entities > Customer  3/47  MDL
[mxcli:compare] left: Entity.Customer  right: Entity.Order  NDSL|NDSL
[mxcli:diff] unified  +12 -8  3 hunks
```

LLMs can `grep '[mxcli:'` to identify current state.

### Plain Text Yank

Existing `y` → clipboard mechanism preserved. All views implement `PlainText() string` for ANSI-stripped output.

### Future: Machine Mode

Environment variable `MXCLI_TUI_MACHINE=1` strips all ANSI codes. Stretch goal, not in initial scope.

## Implementation Phases

### Phase 1: View System Foundation

**Files**: New `view.go`, `viewstack.go`; rewrite `app.go`

1. Define `view` interface, `ViewMode` enum, `StatusInfo` struct
2. Implement `ViewStack` with Push/Pop/Active
3. Wrap existing `MillerView` as `BrowserView` implementing `view`
4. Wrap existing `Overlay` as `OverlayView` implementing `view`
5. Wrap existing `CompareView` implementing `view`
6. Wrap existing `DiffView` implementing `view`
7. Rewrite `app.go` to use ViewStack — eliminate all `syncXxx()` calls
8. Verify all existing functionality works unchanged

### Phase 2: Visual System

**Files**: Replace `styles.go` → `theme.go`; modify `column.go`, `statusbar.go`, `hintbar.go`, `tabbar.go`

1. Create `theme.go` with AdaptiveColor tokens
2. Add focus color to column titles and left edge indicator
3. Add Faint to unfocused columns
4. Enhance header: ViewMode badge + context summary
5. Enhance footer: compact hint format, breadcrumb with `›`
6. Add ViewStack depth indicator when depth > 1

### Phase 3: Navigation

**Files**: New `jumper.go`; modify `miller.go`, `preview.go`

1. Implement global fuzzy jump (reuse picker logic)
2. Add preview debounce (150ms)
3. Unify `q`/`Esc` as ViewStack.Pop
4. Add `Space` keybinding for global jump
5. Make `1-9` context-sensitive

### Phase 4: Polish

**Files**: Modify `statusbar.go`, `app.go`

1. Add LLM anchor lines to each view's render
2. Implement clickable breadcrumb (mouse zones)
3. Code cleanup: remove dead code, update comments
4. Update help text for new keybindings

## Testing Strategy

- Each phase: `make test` + manual testing with `mxcli tui -p app.mpr`
- Phase 1 is the riskiest (refactoring core routing) — test all view transitions
- Existing TUI tests (if any) must pass at each phase boundary
