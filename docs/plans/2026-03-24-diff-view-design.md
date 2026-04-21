# Diff View TUI Component Design

**Issue**: [engalar/mxcli#17](https://github.com/engalar/mxcli/issues/17)
**Date**: 2026-03-24

## Overview

A reusable, generic TUI diff component for mxcli's Bubble Tea interface. Accepts any two text inputs, computes line-level and word-level diffs, and renders them with syntax highlighting and interactive navigation. Supports both Unified and Side-by-Side view modes with keyboard toggle.

## Requirements

- **Generic**: Decoupled from specific data sources (MDL, git, BSON) — accepts `(oldText, newText, language)`.
- **Two view modes**: Unified (single column, +/- markers) and Side-by-Side (left/right panes), switchable via `Tab`.
- **Word-level inline diff**: Within changed lines, highlight the specific words/characters that changed (deep background color), not just the entire line.
- **Syntax highlighting**: Equal lines use Chroma; changed lines use Lipgloss segment rendering (avoids ANSI Reset conflict).
- **Interactive**: Vim-style scrolling, search, hunk navigation (`]c`/`[c`).
- **Integration**: Opens as an overlay in `app.go` via `DiffOpenMsg`, closes with `q`/`Esc`.

## Architecture

### Data Flow

```
(oldText, newText, lang) → diffengine.ComputeDiff() → []DiffLine → diffrender.Render*() → string
```

### Data Model (`diffengine.go`)

```go
type DiffLineType int

const (
    DiffEqual  DiffLineType = iota
    DiffInsert
    DiffDelete
)

type DiffSegment struct {
    text    string
    changed bool // true = this specific segment was modified
}

type DiffLine struct {
    type      DiffLineType
    OldLineNo int          // 0 for insert lines
    NewLineNo int          // 0 for delete lines
    content   string       // raw text
    Segments  []DiffSegment // word-level breakdown (insert/delete only)
}

type DiffResult struct {
    Lines []DiffLine
    Stats DiffStats // additions, deletions, changes counts
}
```

### Diff Engine (`diffengine.go`, ~150 lines)

Uses `github.com/sergi/go-diff/diffmatchpatch`.

**Two-pass strategy:**

1. **Line-level diff**: Use `DiffLinesToChars()` + `DiffMain()` + `DiffCharsToLines()` to get line-level differences. This avoids character-level fragmentation that breaks TUI layout.

2. **Word-level diff** (second pass): For adjacent Delete/Insert line pairs, run `DiffMain()` on the raw text to identify which words/characters changed. Map results to `[]DiffSegment` with `changed` flags.

```go
func ComputeDiff(oldText, newText string) *DiffResult
func computeWordSegments(oldLine, newLine string) (oldSegs, newSegs []DiffSegment)
```

### Rendering (`diffrender.go`, ~250 lines)

#### Chroma ANSI Reset Conflict Resolution

**Problem**: Chroma inserts `\x1b[0m` after each token, which erases any outer background color set by Lipgloss.

**Solution**: Don't use Chroma on changed lines.

| Line Type | Rendering Strategy |
|-----------|--------------------|
| Equal | Chroma syntax highlighting (via existing `DetectAndHighlight`) |
| Insert | Lipgloss per-segment: unchanged segments = light green foreground, changed segments = white on dark green background |
| Delete | Lipgloss per-segment: unchanged segments = light red foreground, changed segments = white on dark red background |

#### Color Palette

```
// insert (added)
AddedFg         = "#00D787"  // light green — unchanged segments
AddedChangedFg  = "#FFFFFF"  // white — changed word text
AddedChangedBg  = "#005F00"  // dark green — changed word background
AddedGutter     = "#00D787"  // gutter "+"

// delete (removed)
RemovedFg       = "#FF5F87"  // light red — unchanged segments
RemovedChangedFg = "#FFFFFF" // white — changed word text
RemovedChangedBg = "#5F0000" // dark red — changed word background
RemovedGutter   = "#FF5F87"  // gutter "-"

// Equal
EqualGutter     = "#626262"  // dim gray — gutter "│"
```

#### Unified View Renderer

```
func RenderUnified(result *DiffResult, width int, lang string) []string
```

Output format per line:
```
[gutter] [old_lineno] [new_lineno] [content]
```

- Gutter: `+` (green), `-` (red), `│` (gray)
- Line numbers: dual column (old/new), dim when not applicable

#### Side-by-Side Renderer

```
func RenderSideBySide(result *DiffResult, width int, lang string) (left, right []string)
```

- Each pane = `(width - 3) / 2` characters (3 = divider + padding)
- Delete lines on left, Insert lines on right, Equal on both
- Blank lines pad the shorter side to maintain alignment

### DiffView Component (`diffview.go`, ~400 lines)

```go
type DiffViewMode int
const (
    DiffViewUnified DiffViewMode = iota
    DiffViewSideBySide
)

type DiffView struct {
    // Input
    oldText, newText string
    language         string
    title            string

    // Computed
    result    *DiffResult
    rendered  []string   // unified: rendered lines; side-by-side: not used
    leftLines []string   // side-by-side only
    rightLines []string  // side-by-side only
    hunkStarts []int     // line indices where hunks begin (for ]c/[c)

    // view state
    viewMode    DiffViewMode
    yOffset     int
    width       int
    height      int

    // Side-by-side state
    syncScroll  bool  // default true
    focus       int   // 0=left, 1=right
    leftOffset  int
    rightOffset int

    // search
    searching   bool
    searchInput textinput.Model
    searchQuery string
    matchLines  []int
    matchIdx    int
}
```

#### Key Bindings

| Key | Action |
|-----|--------|
| `j` / `↓` | Scroll down |
| `k` / `↑` | Scroll up |
| `d` / `PgDn` | Page down |
| `u` / `PgUp` | Page up |
| `g` | Jump to top |
| `G` | Jump to bottom |
| `Tab` | Toggle Unified ↔ Side-by-Side |
| `h` / `l` | Switch focus left/right (side-by-side) |
| `/` | Start search |
| `n` / `N` | Next / previous search match |
| `]c` | Jump to next hunk |
| `[c` | Jump to previous hunk |
| `q` / `Esc` | Close diff view |
| Mouse wheel | Scroll |

#### Messages

```go
type DiffOpenMsg struct {
    OldText  string
    NewText  string
    Language string // "sql", "go", "ndsl", "" (auto-detect)
    title    string
}

type DiffCloseMsg struct{}
```

### Integration (`app.go` changes)

Add `diffView *DiffView` field to the root `App` model.

When `diffView != nil`:
- `update()` delegates all input to `diffView.Update()`
- `view()` renders `diffView.View()` as full-screen overlay
- On `DiffCloseMsg`, set `diffView = nil`

## New Dependencies

```
github.com/sergi/go-diff v1.3.0
```

## File Plan

| File | Type | Lines | Purpose |
|------|------|-------|---------|
| `cmd/mxcli/tui/diffengine.go` | New | ~150 | Diff computation (line + word level) |
| `cmd/mxcli/tui/diffrender.go` | New | ~250 | Unified + Side-by-Side rendering |
| `cmd/mxcli/tui/diffview.go` | New | ~400 | Bubble Tea component |
| `cmd/mxcli/tui/app.go` | Modify | +30 | DiffView overlay integration |
| `go.mod` / `go.sum` | Modify | +2 | Add sergi/go-diff dependency |

## Not in MVP

- Integration with `mxcli diff` / `mxcli diff-local` CLI commands (separate issue)
- Context folding (collapsing unchanged regions)
- Hunk accept/reject (edit mode)
- Side-by-side synchronized line alignment for multi-line insertions
