# Enhanced Theme System — Extensions Beyond Current Styling Support

**Date:** 2026-03-27
**Status:** Draft
**Extends:** `docs/11-proposals/page-styling-support.md` (Phases 1-3 shipped)

## Context

The existing styling proposal (`page-styling-support.md`) defined three phases. Phases 1 and 3 are fully implemented:
- Phase 1: `class` and `style` inline properties
- Phase 3: `show design properties`, `describe styling`, `alter styling`, inline `designproperties:`, `update widgets` bulk styling

Phase 2 (`DynamicClasses`) remains open.

This document proposes **new features beyond the original proposal scope**: SCSS variable management and theme presets.

---

## Feature 1: Structured SCSS Variable Management

### Motivation

Currently there is no MDL interface for modifying `theme/web/custom-variables.scss`. Users must manually edit the SCSS file. An MDL interface would let AI assistants adjust brand colors, font sizes, and spacing without file-level editing.

### Syntax

```sql
-- List all custom variables (from theme/web/custom-variables.scss)
show THEME variables;

-- List atlas_core default variables (read-only reference)
show THEME variables default;

-- Search variables by pattern
show THEME variables like '%brand%';

-- Set a single variable
alter THEME VARIABLE '$brand-primary' = '#FF6B35';

-- Batch set
alter THEME variables
  '$brand-primary' = '#FF6B35',
  '$brand-secondary' = '#2D3748',
  '$font-size-default' = '15px';

-- Reset to atlas_core default (remove from custom-variables.scss)
alter THEME VARIABLE '$brand-primary' RESET;
```

### Implementation Strategy

**SCSS Variable Parser** — Simple line-based parser (no full AST needed):
- Regex: `^\s*(\$[\w-]+)\s*:\s*(.+?)\s*(!default)?\s*;\s*(//.*)?$`
- Preserves non-variable lines (comments, blank lines, imports) for write-back
- Surgical line edits only — never rewrites the entire file

**Validation** — Only accepts variable names that exist in atlas_core defaults. Override with `force` keyword for truly custom variables.

### Implementation Files

| File | Change |
|------|--------|
| `mdl/grammar/MDLParser.g4` | Add SHOW/ALTER THEME VARIABLE(S) rules |
| `mdl/ast/ast_styling.go` | Add `ShowThemeVariablesStmt`, `AlterThemeVariableStmt` |
| `mdl/visitor/visitor_styling.go` | Add visitor methods |
| `mdl/executor/theme_variables.go` | **New**: SCSS parser, variable read/write |
| `mdl/executor/cmd_styling.go` | Add executor dispatch |

---

## Feature 2: Theme Presets via MDL Script Templates

### Motivation

Theme presets package common styling recipes (dark mode, high-contrast, brand color schemes) as reusable MDL scripts. Since `alter THEME variables` (Feature 1) provides the primitive, presets are a thin wrapper around script execution.

### Syntax

```sql
-- List available presets (built-in + project-local)
show THEME PRESETS;

-- Preview what a preset will do (dry run)
describe THEME PRESET 'dark';

-- Apply a preset (executes the MDL script)
apply THEME PRESET 'dark';
```

### Preset File Format

Standard `.mdl` files with comment metadata headers:

```sql
-- preset: dark
-- description: Dark color scheme with light text on dark backgrounds
-- author: mxcli
-- version: 1.0

alter THEME variables
  '$brand-default' = '#1a1a2e',
  '$brand-primary' = '#6c63ff',
  '$bg-color' = '#16213e',
  '$font-color-default' = '#edf2f7';
```

### Storage and Discovery

| Location | Purpose | Priority |
|----------|---------|----------|
| `cmd/mxcli/presets/*.mdl` | Built-in presets (via `go:embed`) | Lowest |
| `theme/presets/*.mdl` | Project-specific presets | Highest |

### Implementation Files

| File | Change |
|------|--------|
| `cmd/mxcli/presets/*.mdl` | **New**: Built-in preset MDL scripts |
| `mdl/grammar/MDLParser.g4` | Add SHOW/DESCRIBE/APPLY THEME PRESET rules |
| `mdl/ast/ast_styling.go` | Add AST nodes |
| `mdl/executor/theme_presets.go` | **New**: Preset discovery, metadata parsing |
| `mdl/executor/cmd_styling.go` | Add `execApplyThemePreset()` (loads + executes script) |

---

## Implementation Priority

| Phase | Feature | Dependency |
|-------|---------|------------|
| 1 | SCSS Variable Management | Independent |
| 2 | Theme Presets | Depends on Phase 1 (presets use `alter THEME variables`) |

Phase 2 (`DynamicClasses` from the original proposal) can be implemented independently at any time.
