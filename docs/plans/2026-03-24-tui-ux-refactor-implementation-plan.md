# TUI 2.0 UX Refactor — Implementation Plan

**Date**: 2026-03-24
**Design Doc**: `docs/plans/2026-03-24-tui-ux-refactor-design.md`
**Branch**: `feat/tui-ux-refactor` (from `feat/tui-diff-view`)

## Current Architecture Summary

| File | Lines | Role |
|------|-------|------|
| `app.go` | 760 | Root Bubble Tea model; routes all messages; manages overlay/compare/diff via booleans & pointers |
| `miller.go` | 862 | Three-column Miller view; navigation stack; preview rendering |
| `column.go` | 511 | Scrollable list column with filter; emits `CursorChangedMsg` |
| `compare.go` | 634 | Side-by-side comparison with built-in fuzzy picker |
| `diffview.go` | 666 | Interactive diff viewer (unified/side-by-side/plain modes) |
| `overlay.go` | 135 | Fullscreen modal wrapping `ContentView` |
| `contentview.go` | 409 | Scrollable content viewer with line numbers + search |
| `styles.go` | 45 | Monochrome style tokens |
| `hintbar.go` | 128 | Context-sensitive key hints |
| `statusbar.go` | 66 | Bottom status line (breadcrumb + position + mode) |
| `tabbar.go` | 104 | Horizontal tab bar with click zones |
| `tab.go` | 105 | Tab struct (Miller + nodes + project path) |
| `picker.go` | 443 | Project path picker (standalone + embedded) |
| `preview.go` | 242 | Async preview engine with cache + cancellation |
| `help.go` | 50 | Help overlay text |

### Current Routing in `app.go`

The `update()` method uses a priority chain of `if` checks:
1. `a.picker != nil` → delegate to picker
2. `a.diffView != nil` → delegate to DiffView
3. `a.compare.IsVisible()` → delegate to CompareView
4. `a.overlay.IsVisible()` → delegate to Overlay
5. `a.showHelp` → dismiss help
6. default → `updateNormalMode()` → Miller view

State sync is manual: every branch calls `a.syncTabBar()`, `a.syncStatusBar()`, `a.syncHintBar()` — 30+ call sites total.

### Key Coupling Points

- `App` holds `overlay Overlay`, `compare CompareView`, `diffView *DiffView`, `picker *PickerModel`, `showHelp bool` — all direct struct fields
- `App.View()` has the same priority chain as `update()`
- `syncHintBar()` maps the priority chain to hint sets
- `syncStatusBar()` reads directly into `tab.Miller.preview.mode`, `tab.Miller.current.cursor`
- Compare/Overlay/Diff each render their own chrome (title bar, status bar, hint bar) — not shared
- `CompareView` has a built-in fuzzy picker (`picker bool`, `pickerInput`, etc.)
- `MillerView` directly renders preview content inside its `view()` method

---

## Phase 1: View System Foundation

**Goal**: Introduce `view` interface + `ViewStack`; rewrite `app.go` to use them. All existing functionality preserved.

### Step 1.1: Define core types — `view.go` (NEW, ~80 lines)

```go
// view.go
package tui

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
    position   string  // "3/47"
    Mode       string  // "MDL", "NDSL"
    Extra      string  // view-specific
}

type Hint struct { key string; label string }  // already exists in hintbar.go — REUSE it

type view interface {
    update(tea.Msg) (view, tea.Cmd)
    Render(width, height int) string
    Hints() []Hint
    StatusInfo() StatusInfo
    Mode() ViewMode
}
```

**Notes**:
- `Hint` struct already exists in `hintbar.go`. Do NOT duplicate — import from same package.
- `View.Update()` returns `(view, tea.Cmd)` instead of `(tea.Model, tea.Cmd)` — this is intentional to stay within the `tui` package type system.
- `View.Render(width, height int) string` replaces `view()` — explicit dimensions for layout composability.

### Step 1.2: Implement ViewStack — `viewstack.go` (NEW, ~60 lines)

```go
// viewstack.go
package tui

type ViewStack struct {
    base  view     // always BrowserView
    stack []view   // pushed views (overlay/compare/diff)
}

func NewViewStack(base view) ViewStack
func (vs *ViewStack) Active() view        // top of stack, or base
func (vs *ViewStack) Push(v view)
func (vs *ViewStack) Pop() (view, bool)   // returns popped view + ok; no-op if stack empty
func (vs *ViewStack) depth() int          // len(stack) + 1 (for base)
func (vs *ViewStack) SetActive(v view)    // replace top of stack (or base if empty)
```

**Notes**:
- `SetActive()` is needed because `update()` returns a new `view` value (Go value semantics).
- `Pop()` returns the popped view so the caller can inspect it if needed.

### Step 1.3: Wrap MillerView as BrowserView — `browserview.go` (NEW, ~160 lines)

This is a **wrapper**, not a rewrite. `BrowserView` embeds the existing `MillerView` and adds the `view` interface.

```go
type BrowserView struct {
    miller      MillerView
    tab         *Tab
    allNodes    []*TreeNode
    mxcliPath   string
    projectPath string
}
```

**Interface methods**:
- `update(msg)` — delegates to `MillerView.Update()` for key/mouse/cursor/preview messages. Handles node-action keys (`b`, `m`, `c`, `d`, `y`) that currently live in `app.go:updateNormalMode()`. Returns `tea.Cmd` that may emit `PushViewMsg` (new message type) for overlay/compare/diff.
- `Render(w, h)` — calls `miller.SetSize(w, h)` + `miller.View()`.
- `Hints()` — returns `ListBrowsingHints` or `FilterActiveHints` based on `miller.focusedColumn().IsFilterActive()`.
- `StatusInfo()` — reads breadcrumb, position, mode from `miller`.
- `Mode()` — returns `ModeBrowser`.

**New message types** (in `view.go`):
```go
type PushViewMsg struct { view view }
type PopViewMsg  struct{}
```

These replace the implicit `a.overlay.Show()` / `a.compare.Show()` / `a.diffView = &dv` patterns. `App` handles them by calling `ViewStack.Push()`/`Pop()`.

**What moves out of `app.go` into `BrowserView`**:
- `updateNormalMode()` keys: `b`, `m`, `c`, `d`, `y`, `r`, `z`, `/` (filter delegate)
- Helper methods: `runBsonOverlay()`, `runMDLOverlay()`, `loadBsonNDSL()`, `loadMDL()`, `loadForCompare()`, `openDiagram()`
- Overlay state: `overlayQName`, `overlayNodeType`, `overlayIsNDSL`

**What stays in `app.go`**:
- Tab management keys: `t`, `T`, `W`, `1-9`, `[`, `]`
- Global keys: `q`, `?`, `ctrl+c`
- `LoadTreeMsg` handling (needs access to tabs array)
- Tab bar / chrome rendering

### Step 1.4: Wrap Overlay as OverlayView — modify `overlay.go` (~40 lines added)

Add `view` interface methods to existing `Overlay` struct:

```go
func (o Overlay) update(msg tea.Msg) (view, tea.Cmd)  // wraps existing o.Update()
func (o Overlay) Render(w, h int) string               // wraps existing o.View()
func (o Overlay) Hints() []Hint                        // returns OverlayHints
func (o Overlay) StatusInfo() StatusInfo               // title + scroll position
func (o Overlay) Mode() ViewMode                       // ModeOverlay
```

**Key change**: When overlay closes (`esc`/`q` pressed), instead of setting `o.visible = false`, it returns a `tea.Cmd` that emits `PopViewMsg{}`.

**Tab switching** (NDSL/MDL toggle): Currently handled by `app.go` reading `a.overlayQName`. This state needs to live in the `Overlay` (or a wrapping `OverlayView`). Two options:

- **Option A**: Add `qname`, `nodeType`, `isNDSL`, `switchable` fields to `Overlay` and a callback `func(nodeType, qname string) tea.Cmd` for reloading.
- **Option B (recommended)**: Create a thin `OverlayView` wrapper struct that embeds `Overlay` and holds the reload context.

Going with **Option B**: Create `OverlayView` struct in `overlay.go`:

```go
type OverlayView struct {
    overlay     Overlay
    qname       string
    nodeType    string
    isNDSL      bool
    mxcliPath   string
    projectPath string
}
```

The `OverlayView.Update()` handles `tab` for NDSL/MDL switching internally (moves logic from `app.go:260-285`).

### Step 1.5: Wrap CompareView — modify `compare.go` (~30 lines added)

Add `view` interface methods to existing `CompareView`:

```go
func (c CompareView) update(msg tea.Msg) (view, tea.Cmd)  // wraps existing
func (c CompareView) Render(w, h int) string               // sets size, calls view()
func (c CompareView) Hints() []Hint                        // CompareHints
func (c CompareView) StatusInfo() StatusInfo
func (c CompareView) Mode() ViewMode                       // ModeCompare
```

**Key change**: `esc`/`q` emits `PopViewMsg` instead of `c.visible = false`.

**Issue**: `CompareView` currently stores `visible bool` and the `view()` method checks it. With ViewStack, visibility is implicit (it's on the stack or not). The `visible` field becomes redundant but can stay for backward compat during transition.

**Diff launch**: When `D` is pressed in compare, it currently emits `DiffOpenMsg` which `app.go` handles by creating a `DiffView` and setting `a.diffView = &dv`. After refactor, it should emit `PushViewMsg{view: NewDiffViewWrapped(...)}`.

### Step 1.6: Wrap DiffView — modify `diffview.go` (~30 lines added)

Add `view` interface methods:

```go
func (dv DiffView) UpdateView(msg tea.Msg) (view, tea.Cmd)  // wraps existing, named differently to avoid conflict
func (dv DiffView) Render(w, h int) string
func (dv DiffView) Hints() []Hint                           // DiffViewHints
func (dv DiffView) StatusInfo() StatusInfo
func (dv DiffView) Mode() ViewMode                          // ModeDiff
```

**Naming conflict**: `DiffView` already has `func (dv DiffView) update(msg tea.Msg) (DiffView, tea.Cmd)`. The `view` interface needs `update(tea.Msg) (view, tea.Cmd)`. Solution: rename existing to `updateInternal` and add the interface method.

**Key change**: `q`/`esc` emits `PopViewMsg` instead of `DiffCloseMsg`.

### Step 1.7: Wrap PickerModel — `pickerview.go` (NEW, ~50 lines)

```go
type PickerView struct {
    picker PickerModel
}
```

Implements `view` interface. When picker completes, emits `PopViewMsg` + action message.

### Step 1.8: Rewrite `app.go` (~250 lines, down from 760)

The new `App` struct:

```go
type App struct {
    tabs      []Tab
    activeTab int
    nextTabID int

    width     int
    height    int
    mxcliPath string

    views     ViewStack  // replaces overlay, compare, diffView, picker, showHelp
    showHelp  bool       // help is special (rendered as overlay on top of chrome)

    tabBar        TabBar
    statusBar     StatusBar
    hintBar       HintBar
    previewEngine *PreviewEngine
}
```

**New `update()` flow**:

```go
func (a App) update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // 1. Handle PushViewMsg / PopViewMsg
    // 2. Handle global keys (ctrl+c, tab management, help toggle)
    // 3. Delegate to active view
    active := a.views.Active()
    updated, cmd := active.Update(msg)
    a.views.SetActive(updated)
    return a, cmd
}
```

**New `view()` flow**:

```go
func (a App) view() string {
    active := a.views.Active()

    // Chrome
    header := a.renderHeader(active)   // tab bar + mode badge + context
    footer := a.renderFooter(active)   // hint bar + status bar

    contentH := a.height - chromeHeight
    content := active.Render(a.width, contentH)

    return header + "\n" + content + "\n" + footer
}
```

**Chrome rendering** is now declarative:
- `a.hintBar.SetHints(active.Hints())` — no more `syncHintBar()`
- Status bar reads from `active.StatusInfo()` — no more `syncStatusBar()`
- Tab bar sync only needed on tab changes — no more scattered `syncTabBar()` calls

**Message routing changes**:

| Old | New |
|-----|-----|
| `DiffOpenMsg` → `a.diffView = &dv` | `DiffOpenMsg` → `PushViewMsg{NewDiffView(...)}` |
| `DiffCloseMsg` → `a.diffView = nil` | `PopViewMsg` |
| `OpenOverlayMsg` → `a.overlay.Show(...)` | `PushViewMsg{NewOverlayView(...)}` |
| `ComparePickMsg` → `a.compare.SetLoading()` | Stays within `CompareView.Update()` |
| `CompareLoadMsg` → `a.compare.SetContent()` | Stays within `CompareView.Update()` |
| `CompareReloadMsg` → reload both panes | Stays within `CompareView.Update()` |

**What `App` still handles directly**:
- `tea.WindowSizeMsg` — resize + propagate to active view
- `LoadTreeMsg` — update tab nodes, set items on BrowserView
- `PickerDoneMsg` — create new tab
- `CmdResultMsg` — push overlay with result
- Tab management keys (`t`, `T`, `W`, `1-9`, `[`, `]`)

### Step 1.9: Handle message forwarding for async results

**Problem**: `CompareLoadMsg`, `CompareReloadMsg`, `ComparePickMsg`, `PreviewReadyMsg`, `PreviewLoadingMsg` are emitted by async commands and need to reach the correct view. Currently `app.go` routes them explicitly.

**Solution**: `App.Update()` forwards all non-global messages to `active.Update()`. The active view (CompareView, BrowserView, etc.) handles its own async messages. If a message doesn't match, it's a no-op.

**Edge case**: `CompareLoadMsg` arrives after user has popped CompareView. This is harmless — the message hits BrowserView which ignores it.

### Step 1.10: Verification

1. `make build` — compile succeeds
2. `make test` — all existing tests pass
3. Manual test matrix:
   - [ ] Launch TUI, navigate Miller columns (h/l/j/k)
   - [ ] Press `b` → BSON overlay opens, `Tab` switches NDSL/MDL
   - [ ] Press `m` → MDL overlay opens
   - [ ] Press `Esc` → overlay closes
   - [ ] Press `c` → compare view opens
   - [ ] In compare: `/` picker, `1/2/3` mode switch, `D` diff
   - [ ] In compare: `Esc` → closes
   - [ ] In diff: `Tab` mode cycle, `q` → closes back to compare
   - [ ] Tab management: `t` clone, `T` new project, `W` close, `1-9` switch
   - [ ] Filter: `/` in column, type, `Enter`/`Esc`
   - [ ] Mouse: click columns, scroll wheel, tab bar click
   - [ ] Resize terminal window
   - [ ] `y` yank in all views

### Risk Areas — Phase 1

1. **Value semantics**: Go's Bubble Tea pattern uses value receivers. `view` interface methods must handle value/pointer semantics correctly. `MillerView` uses value receivers (`func (m MillerView) update()`), so `BrowserView` wrapping it must copy correctly. `CompareView` already uses value receivers too.

2. **Overlay state**: Moving `overlayQName/nodeType/isNDSL` from `App` to `OverlayView` means the overlay reload command needs `mxcliPath` and `projectPath`. These are passed at construction time.

3. **Compare → Diff flow**: Currently `CompareView.Update()` emits `DiffOpenMsg` which `App` handles. After refactor, `CompareView` should emit `PushViewMsg` with a constructed `DiffView`. But `CompareView` doesn't have access to the `view` constructor. **Solution**: `CompareView` still emits `DiffOpenMsg`; `App.Update()` intercepts it before delegating to active view and does `ViewStack.Push(NewDiffView(...))`.

4. **Compare needs `mxcliPath`/`projectPath`**: For `ComparePickMsg` handling, `App` currently calls `a.loadForCompare()` which uses `a.mxcliPath` and `tab.ProjectPath`. After refactor, `CompareView` needs these at construction time, or `BrowserView` handles `ComparePickMsg` and forwards the loaded content.

   **Solution**: `CompareView` stores `mxcliPath` and `projectPath` (passed at construction). Its `update()` handles `ComparePickMsg` internally using these fields. `CompareLoadMsg` is already handled by `CompareView`.

---

## Phase 2: Visual System

**Goal**: Replace monochrome styles with semantic color tokens. Add focus indicators. Enhance chrome.

**Prerequisite**: Phase 1 complete.

### Step 2.1: Create `theme.go` (REPLACE `styles.go`, ~100 lines)

```go
// theme.go
package tui

var (
    // Semantic color tokens — AdaptiveColor for light/dark terminal support
    FocusColor   = lipgloss.AdaptiveColor{Light: "62", Dark: "63"}
    AccentColor  = lipgloss.AdaptiveColor{Light: "214", Dark: "214"}
    MutedColor   = lipgloss.AdaptiveColor{Light: "245", Dark: "243"}
    AddedColor   = lipgloss.AdaptiveColor{Light: "28", Dark: "114"}
    RemovedColor = lipgloss.AdaptiveColor{Light: "124", Dark: "210"}

    // Derived styles (same names as styles.go for minimal diff)
    SeparatorChar  = "│"
    SeparatorStyle = lipgloss.NewStyle().Foreground(MutedColor)

    ActiveTabStyle   = lipgloss.NewStyle().Bold(true).Foreground(FocusColor).Underline(true)
    InactiveTabStyle = lipgloss.NewStyle().Foreground(MutedColor)

    ColumnTitleStyle     = lipgloss.NewStyle().Bold(true)
    FocusedTitleStyle    = lipgloss.NewStyle().Bold(true).Foreground(FocusColor)
    FocusedEdgeChar      = "▎"
    FocusedEdgeStyle     = lipgloss.NewStyle().Foreground(FocusColor)

    SelectedItemStyle = lipgloss.NewStyle().Reverse(true)
    DirectoryStyle    = lipgloss.NewStyle().Bold(true)
    LeafStyle         = lipgloss.NewStyle()

    // ... rest of existing styles migrated with color tokens
)
```

**Migration**: Delete `styles.go`, create `theme.go`. All style variable names stay the same — zero downstream impact.

### Step 2.2: Add focus indicators to columns — modify `column.go` (~20 lines changed)

In `Column.View()`:
- When `c.focused`: render title with `FocusedTitleStyle`, prepend each line with `FocusedEdgeStyle.Render(FocusedEdgeChar)` (1 char, deduct from content width)
- When `!c.focused`: wrap entire output with `lipgloss.NewStyle().Faint(true)` (gray out)

Changes to `column.go`:
1. `view()` method: check `c.focused` flag
2. Title rendering: `FocusedTitleStyle` vs `ColumnTitleStyle`
3. Line rendering: add `FocusedEdgeChar` prefix when focused
4. Width calculation: subtract 1 for edge char when focused

### Step 2.3: Add Faint to unfocused columns — modify `miller.go` (~10 lines)

In `MillerView.View()`:
- Parent column already has `SetFocused(false)` when current is focused
- Preview child column: add `SetFocused(false)`
- The Faint styling is handled by `Column.View()` from Step 2.2

### Step 2.4: Enhance preview mode label — modify `miller.go` (~5 lines)

In `renderPreview()`: use `AccentColor` for the MDL/NDSL mode label instead of plain bold.

### Step 2.5: Enhance header — modify `app.go` or new `chrome.go` (~60 lines)

Create `renderHeader()` in `app.go` (or separate `chrome.go`):

```
❶ App.mpr  ❷ Other.mpr          Browse │ 3 modules, 47 entities
```

- Left: tabs with numbered circles (❶❷❸...) instead of `[1]`
- Center: ViewMode badge from `active.Mode()` → string
- Right: context summary — count modules/entities from `tab.AllNodes`

`renderContextSummary(nodes []*TreeNode) string` — walks tree, counts by type.

### Step 2.6: Enhance footer — modify `hintbar.go` + `statusbar.go` (~30 lines total)

**HintBar** changes:
- Remove `:` separator between key and label (currently `key:label`, change to `key label`)
- This is a 1-line change in `view()`: `HintKeyStyle.Render(hint.Key) + " " + HintLabelStyle.Render(hint.Label)`

**StatusBar** changes:
- Use `›` instead of ` > ` for breadcrumb separator
- Use `⎸` before mode indicator
- Add ViewStack depth indicator: `[Browse > Compare > Diff]` when depth > 1
- `StatusBar` needs to accept `ViewStack.Depth()` and mode names — add `SetViewStackInfo(depth int, modes []string)`

### Step 2.7: Verification

1. `make build` + `make test`
2. Visual inspection:
   - [ ] Focused column has blue title + blue left edge
   - [ ] Unfocused columns are faint
   - [ ] Tab bar shows colored underline on active tab
   - [ ] Header shows mode badge + context summary
   - [ ] Footer uses `›` breadcrumb, compact hints
   - [ ] Dark terminal: colors readable
   - [ ] Light terminal: colors readable (if testable)

---

## Phase 3: Navigation

**Goal**: Add global fuzzy jump, preview debounce, unified keybindings.

**Prerequisite**: Phase 1 + Phase 2 complete.

### Step 3.1: Implement global fuzzy jump — `jumper.go` (NEW, ~180 lines)

```go
type JumperView struct {
    input     textinput.Model
    items     []PickerItem          // all qualified names from tree
    matches   []pickerMatch         // filtered + scored
    cursor    int
    offset    int
    width     int
    height    int
    maxShow   int                   // 12
}
```

Implements `view` interface.

**Reuse**: Copy fuzzy scoring from `compare.go:fuzzyScore()`. The `PickerItem` type is already defined in `compare.go`. Consider extracting `fuzzyScore` and `PickerItem` to a shared file (`fuzzy.go`) to avoid duplication.

**On selection**: `JumperView.Update()` on `enter` emits a new `JumpToNodeMsg{QName string, NodeType string}`. `BrowserView.Update()` handles this by navigating the Miller columns:
1. Walk `allNodes` to find the path to the node
2. Reset nav stack
3. Drill in step by step to reach the target

**Rendering**: Centered modal box (same pattern as compare picker and overlay):
```
┌──────────────────────────────────────┐
│ > cust_                              │
│   🏢 MyModule.Customer       entity  │
│   📄 MyModule.Customer_Overview page │
│   ⚡ MyModule.ACT_Customer_Create MF │
└──────────────────────────────────────┘
```

### Step 3.2: Extract fuzzy logic — `fuzzy.go` (NEW, ~50 lines)

Move from `compare.go`:
- `type PickerItem struct` — stays in `compare.go` (or move here, update imports)
- `func fuzzyScore(target, query string) (bool, int)` → move to `fuzzy.go`
- `type pickerMatch struct` → move to `fuzzy.go`

This avoids duplication between `JumperView` and `CompareView`'s picker.

### Step 3.3: Add `navigateToNode()` to BrowserView — modify `browserview.go` (~60 lines)

```go
func (bv *BrowserView) navigateToNode(qname string) tea.Cmd
```

Algorithm:
1. Walk `allNodes` recursively to find path: `[root, module, category, node]`
2. Reset Miller to root (`SetRootNodes`)
3. For each step in path: call `drillIn()` programmatically (set cursor to matching child, then drill)
4. Return final preview request command

### Step 3.4: Add preview debounce — modify `miller.go` (~20 lines)

In `handleCursorChanged()`:

```go
type previewDebounceMsg struct {
    node    *TreeNode
    counter int
}

func (m MillerView) handleCursorChanged(msg CursorChangedMsg) (MillerView, tea.Cmd) {
    m.pendingPreviewNode = msg.Node
    m.debounceCounter++
    counter := m.debounceCounter
    return m, tea.Tick(150*time.Millisecond, func(t time.Time) tea.Msg {
        return previewDebounceMsg{node: msg.Node, counter: counter}
    })
}
```

Add `pendingPreviewNode *TreeNode` and `debounceCounter int` fields to `MillerView`.

Handle `previewDebounceMsg` in `update()`:
- If `msg.counter != m.debounceCounter`, ignore (superseded by newer cursor move)
- Otherwise, proceed with preview request

**Child column preview** (non-leaf nodes): show immediately without debounce (purely local, no subprocess).

### Step 3.5: Unify `q`/`Esc` as ViewStack.Pop — already done in Phase 1

Each wrapped view already emits `PopViewMsg` on `q`/`Esc`. `App.Update()` handles `PopViewMsg` by calling `ViewStack.Pop()`.

For base view (BrowserView): `q` still quits the application (current behavior). `Esc` does nothing special in Browser mode (or could deactivate filter — already handled by Column).

### Step 3.6: Add `Space` keybinding for global jump — modify `app.go` (~10 lines)

In `App.Update()` global key handler:

```go
case " ", "ctrl+p":
    // build item list from active tab's nodes
    items := flattenQualifiedNames(tab.AllNodes)
    jumper := NewJumperView(items, a.width, a.height)
    a.views.Push(jumper)
    return a, nil
```

### Step 3.7: Make `1-9` context-sensitive — modify `app.go` (~15 lines)

In `App.Update()`:

```go
case "1", "2", ..., "9":
    if a.views.Active().Mode() == ModeBrowser {
        // Tab switch (existing behavior)
        a.switchToTab(idx)
    }
    // in other modes: let active view handle (compare uses 1/2/3 for mode)
    // Already handled by delegation to active.Update()
```

Actually, current flow already works: if active view is CompareView, `App.Update()` delegates to it, and CompareView handles `1/2/3`. If active is BrowserView, `App.Update()` checks for tab management keys first. The key is to ensure `App` only intercepts `1-9` when active mode is Browser.

### Step 3.8: Verification

1. `make build` + `make test`
2. Manual testing:
   - [ ] Press `Space` → fuzzy jump opens
   - [ ] Type partial name → results filter
   - [ ] Press `Enter` → navigates to node in Miller
   - [ ] Press `Esc` → jump closes
   - [ ] Fast j/k scrolling → preview updates with 150ms debounce (no subprocess flooding)
   - [ ] Slow j/k → preview still updates
   - [ ] `q` in overlay/compare/diff → pops back to previous view
   - [ ] `q` in browser → quits app

---

## Phase 4: Polish

**Goal**: LLM anchors, clickable breadcrumb, help update, code cleanup.

**Prerequisite**: Phase 1-3 complete.

### Step 4.1: Add LLM anchor lines — modify each view's `Render()` (~5 lines each)

Each view prepends a machine-parseable line to its rendered output:

| View | Anchor |
|------|--------|
| BrowserView | `[mxcli:browse] MyModule > entities > Customer  3/47  MDL` |
| OverlayView | `[mxcli:overlay] BSON: MyModule.Customer  NDSL` |
| CompareView | `[mxcli:compare] left: Entity.Customer  right: Entity.Order  NDSL\|NDSL` |
| DiffView    | `[mxcli:diff] unified  +12 -8  3 hunks` |
| JumperView  | `[mxcli:jump] > query_text  12 matches` |

The anchor is the **first line** of `Render()` output, replacing one line from `chromeHeight`.

### Step 4.2: Implement clickable breadcrumb — modify `statusbar.go` + `app.go` (~40 lines)

Use Bubble Tea's zone manager or manual mouse zone tracking:

```go
type BreadcrumbClickMsg struct {
    depth int // which breadcrumb segment was clicked (0 = root)
}
```

In `StatusBar.View()`:
- Track column ranges for each breadcrumb segment
- Store zones in `StatusBar.zones []breadcrumbZone`

In `App.Update()` for `tea.MouseMsg`:
- Check if Y == last line (status bar)
- Call `statusBar.HitTest(x)` → `BreadcrumbClickMsg`
- If hit: call `miller.goBack()` N times to reach the clicked depth

**Implementation detail**: `goBack()` N times means popping N entries from `MillerView.navStack`. Add `goBackToDepth(depth int)` method to `MillerView` that pops multiple levels.

### Step 4.3: Update help text — modify `help.go` (~20 lines)

Update `helpText` constant to reflect new keybindings:
- Add `Space` for fuzzy jump
- Remove `Tab` from browser mode (if changed)
- Update overlay/compare/diff sections

### Step 4.4: Code cleanup — across multiple files

1. **Remove dead code**:
   - `a.syncHintBar()` calls — replaced by declarative `active.Hints()`
   - `a.syncStatusBar()` calls — replaced by `active.StatusInfo()`
   - `a.syncTabBar()` — only called on tab changes now
   - `overlay.visible` field — visibility managed by ViewStack
   - `compare.visible` field — visibility managed by ViewStack
   - `showHelp` bool if help becomes a View

2. **Remove `app.go` helper methods** that moved to views:
   - `runBsonOverlay()`, `runMDLOverlay()` → `OverlayView`
   - `loadBsonNDSL()`, `loadMDL()`, `loadForCompare()` → `CompareView` or `BrowserView`
   - `openDiagram()` → `BrowserView`

3. **Delete `styles.go`** (replaced by `theme.go` in Phase 2)

4. **Update comments** in files that reference old routing logic

### Step 4.5: Verification

1. `make build` + `make test`
2. LLM anchor testing:
   - [ ] `tmux capture-pane -p | grep '\[mxcli:'` shows current state
   - [ ] Each view mode produces correct anchor
3. Breadcrumb clicking:
   - [ ] Click middle segment → navigates back to that level
   - [ ] Click root → goes to root
4. Help text accurate

---

## Implementation Order Summary

```
Phase 1 (Foundation) — ~600 lines new/changed
  1.1  view.go (NEW)         — types
  1.2  viewstack.go (NEW)    — stack logic
  1.3  browserview.go (NEW)  — Miller wrapper
  1.4  overlay.go (modify)   — OverlayView wrapper
  1.5  compare.go (modify)   — view interface
  1.6  diffview.go (modify)  — view interface
  1.7  pickerview.go (NEW)   — Picker wrapper
  1.8  app.go (REWRITE)      — ViewStack-based routing
  1.9  message routing       — async result forwarding
  1.10 verification          — full manual test

Phase 2 (Visual) — ~200 lines new/changed
  2.1  theme.go (NEW, replaces styles.go)
  2.2  column.go (modify)    — focus indicators
  2.3  miller.go (modify)    — faint unfocused
  2.4  miller.go (modify)    — accent preview label
  2.5  app.go (modify)       — enhanced header
  2.6  hintbar.go + statusbar.go (modify)
  2.7  verification

Phase 3 (navigation) — ~300 lines new/changed
  3.1  jumper.go (NEW)       — fuzzy jump view
  3.2  fuzzy.go (NEW)        — extracted fuzzy logic
  3.3  browserview.go (modify) — navigateToNode
  3.4  miller.go (modify)    — preview debounce
  3.5  (already done in P1)
  3.6  app.go (modify)       — Space keybinding
  3.7  app.go (modify)       — context-sensitive 1-9
  3.8  verification

Phase 4 (Polish) — ~150 lines new/changed
  4.1  all views (modify)    — LLM anchors
  4.2  statusbar.go + app.go — clickable breadcrumb
  4.3  help.go (modify)      — updated help text
  4.4  cleanup               — dead code removal
  4.5  verification
```

## File Impact Matrix

| File | Phase 1 | Phase 2 | Phase 3 | Phase 4 |
|------|---------|---------|---------|---------|
| `view.go` | **NEW** | — | — | — |
| `viewstack.go` | **NEW** | — | — | — |
| `browserview.go` | **NEW** | — | MODIFY | MODIFY |
| `pickerview.go` | **NEW** | — | — | — |
| `jumper.go` | — | — | **NEW** | MODIFY |
| `fuzzy.go` | — | — | **NEW** | — |
| `theme.go` | — | **NEW** | — | — |
| `app.go` | **REWRITE** | MODIFY | MODIFY | MODIFY |
| `miller.go` | — | MODIFY | MODIFY | — |
| `column.go` | — | MODIFY | — | — |
| `compare.go` | MODIFY | — | — | MODIFY |
| `diffview.go` | MODIFY | — | — | MODIFY |
| `overlay.go` | MODIFY | — | — | MODIFY |
| `styles.go` | — | **DELETE** | — | — |
| `hintbar.go` | — | MODIFY | — | — |
| `statusbar.go` | — | MODIFY | — | MODIFY |
| `tabbar.go` | — | MODIFY | — | — |
| `help.go` | — | — | — | MODIFY |

## Test Strategy

- **Unit tests**: Add `viewstack_test.go` for Push/Pop/Active/Depth logic
- **Existing tests**: `preview_test.go` must continue to pass at every phase boundary
- **Manual tests**: Full test matrix at each phase end (see verification sections)
- **Build gate**: `make build && make test` after every step, not just phase boundaries
- **Regression signal**: If any existing keybinding stops working, it's a regression — fix before proceeding
