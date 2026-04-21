# Bug Report: mxcli Issues Discovered During VAT Feature Build

**Date:** 2026-02-10
**Reporter:** User (VAT feature development)

---

## Bugs

### Bug 1: ALTER ENTITY ADD ATTRIBUTE Silently Fails

**Severity:** Critical (silent data loss)
**Symptom:** The command parses, passes reference checks, and reports no error — but the attributes are never written to the `.mpr`. Discovered when VATRate/VATAmount never appeared on Catalogue.Model.

**Root Cause:** The entire ALTER ENTITY pipeline is unimplemented. The ANTLR grammar defines the syntax (so it parses successfully), but three layers are missing:

1. **No AST type** — `AlterEntityStmt` doesn't exist in `mdl/ast/ast_entity.go`
2. **No visitor listener** — `ExitAlterEntityAction()` is never implemented in any visitor file, so no AST node is produced (ANTLR base listener is a no-op)
3. **No executor handler** — `executor.go:executeInner()` has no `case *ast.AlterEntityStmt`

The statement parses, the visitor silently produces nothing, and execution completes with zero errors and zero work done.

**Working reference:** `alter enumeration` is fully implemented across all three layers (`ast_enumeration.go`, `visitor_enumeration.go:37`, `executor.go:118`).

**Fix:** Implement `AlterEntityStmt` AST type, `ExitAlterEntityAction` visitor method, and `execAlterEntity` executor handler following the enumeration pattern.

**Files to change:**
- `mdl/ast/ast_entity.go` — add `AlterEntityStmt` struct
- `mdl/visitor/visitor_entity.go` — add `ExitAlterEntityAction()` method
- `mdl/executor/executor.go` — add `case *ast.AlterEntityStmt` in `executeInner()`
- `mdl/executor/cmd_entities.go` — add `execAlterEntity()` function

---

### Bug 2: DECLARE Inside LOOP Body Causes Go Panic

**Severity:** High (crash)
**Symptom:** `panic: assignment to entry in nil map` triggered in `addCreateVariableAction` when processing DECLARE inside a loop body.

**Root Cause:** In `cmd_microflows_builder.go:821`, a new `flowBuilder` is created for the loop body but the `declaredVars` map is not initialized (defaults to `nil`). When `addCreateVariableAction` at `cmd_microflows_builder_actions.go:18` writes `fb.declaredVars[s.Variable] = typeName`, it panics on the nil map.

The error handler builder at line 66 of the same file correctly copies `declaredVars: fb.declaredVars` — the loop builder omits this.

**Fix (one line):** Add `declaredVars: fb.declaredVars,` to the loop builder struct literal at `cmd_microflows_builder.go:821`.

**Files to change:**
- `mdl/executor/cmd_microflows_builder.go:821` — add missing field initialization

---

### Bug 3: IMAGE Widget Not Supported in Page Builder

**Severity:** Medium
**Symptom:** `create or replace page` fails at execution with `unsupported V3 widget type: image`, even though `--references` passes. Existing pages with IMAGE widgets can't be round-tripped.

**Root Cause:** The grammar defines IMAGE, the lexer has the token (`MDLLexer.g4:301`), and SDK types exist (`dynamicimage`/`staticimage` in `sdk/pages/pages_widgets_display.go`), but `cmd_pages_builder_v3.go:buildWidgetV3()` has no `case "image"` — it falls through to the default error at line 287.

**Fix:** Add a `case "image"` to the switch in `buildWidgetV3()` and implement a `buildImageV3()` function.

**Files to change:**
- `mdl/executor/cmd_pages_builder_v3.go` — add case in `buildWidgetV3()` switch
- `mdl/executor/cmd_pages_builder_v3_widgets.go` — add `buildImageV3()` function

---

### Bug 4: CUSTOMCONTAINER Widget Not Recognized by Parser

**Severity:** Medium
**Symptom:** `mismatched input 'customcontainer'` at syntax-check time. Existing pages using CUSTOMCONTAINER can't be reconstructed.

**Root Cause:** Unlike IMAGE (which has a token but no builder), CUSTOMCONTAINER has no lexer token in `MDLLexer.g4` and no alternative in the `widgetTypeV3` rule in `MDLParser.g4`. The parser treats it as an unknown identifier and fails.

**Fix:**
1. Add `customcontainer` token to `MDLLexer.g4`
2. Add `customcontainer` to `widgetTypeV3` in `MDLParser.g4`
3. Regenerate parser (`make grammar`)
4. Add builder function in the page builder

**Files to change:**
- `mdl/grammar/MDLLexer.g4` — add token definition
- `mdl/grammar/MDLParser.g4` — add to `widgetTypeV3` rule
- `mdl/grammar/parser/` — regenerate
- `mdl/executor/cmd_pages_builder_v3.go` — add case
- `mdl/executor/cmd_pages_builder_v3_widgets.go` or `_layout.go` — add builder function

---

### Bug 5: Reference Check Misses DataSource Microflow Resolution

**Severity:** Medium
**Symptom:** `Main.DSO_GetFilter` passed `--references` but failed at exec time with `microflow not found`. The reference validator doesn't validate DataSource: MICROFLOW references.

**Root Cause:** `validateWithContext` in `executor.go:477` only validates the page's module name for `CreatePageStmtV3`, not any widget properties. The DataSource: MICROFLOW reference is only resolved at execution time in `buildDataSourceV3()` (`cmd_pages_builder_v3.go:399`). The `--references` check never walks widget trees.

**Fix:** Add a recursive widget-property walker in `validateWithContext` for `CreatePageStmtV3` (and `CreateSnippetStmtV3`) that validates:
- DataSource references (microflow, nanoflow, entity)
- Action references (microflow, nanoflow, page, entity)
- Snippet references

**Files to change:**
- `mdl/executor/executor.go` — expand `validateWithContext` for page/snippet statements

---

### Bug 6: CREATE OR REPLACE PERSISTENT ENTITY Fails on Reserved-Word Attribute Names

**Severity:** Medium
**Symptom:** Entity with attribute named `range` (a reserved MDL keyword) can't be reconstructed via MDL. No quoting/escaping mechanism exists for attribute names.

**Root Cause:** The `attributename` rule in `MDLParser.g4:397` only allows `IDENTIFIER` plus a hardcoded whitelist of ~25 keywords. `range` (and many others) aren't in this whitelist. `QUOTED_IDENTIFIER` is not accepted — even though it IS supported for qualified names (`identifierOrKeyword`) and enum value names (`enumValueName`).

**Fix (recommended):** Add `QUOTED_IDENTIFIER` to the `attributename` rule so users can escape any keyword with double-quotes or backticks (`"range"`, `` `range` ``). Apply the same fix to `parameterName`, `indexColumnName`, and `memberAttributeName` for consistency.

**Files to change:**
- `mdl/grammar/MDLParser.g4` — add `QUOTED_IDENTIFIER` to `attributename` and related rules
- `mdl/grammar/parser/` — regenerate
- `mdl/visitor/visitor_helpers.go` — may need to strip quotes from `QUOTED_IDENTIFIER` text

---

## Missing Features

### No IF EXISTS / Conditional DDL

```sql
-- These don't exist:
drop microflow if exists Catalogue.ACT_CalculateVAT;
create entity if not exists Catalogue.VATInfo (...);
```

Without them, scripts aren't safely re-runnable. Any partial execution leaves the project in a state where rerunning fails immediately. Note that `create or replace` provides partial idempotency for creates, but there is no equivalent for `drop`.

**Implementation scope:** Medium. Grammar already has `if` and `exists` tokens. Need to add optional clause to all DROP/CREATE rules, add `IfExists`/`IfNotExists` fields to AST types, and add conditional logic in executors.

### No Transaction / Rollback

Execution is statement-by-statement with immediate SQLite commits. A script that fails at statement 4 of 7 leaves statements 1-3 permanently applied. There's no way to roll back to a clean state or replay atomically.

**Implementation scope:** Hard. The writer layer (`sdk/mpr/writer.go`) would need SQLite transaction support, and the executor would need to wrap script execution in a transaction with rollback-on-error.

### No formatDateTime() in ContentParams

Dynamic text with date formatting requires a static fallback. Affects footer copyright years and any date-displaying widgets.

**Implementation scope:** Small. Add to expression evaluator for ContentParams in dynamic text widgets.

### Limited Widget Surface for Page Creation

CREATE OR REPLACE PAGE supports a subset of Mendix widgets. Widgets common in production apps (IMAGE, CUSTOMCONTAINER, GROUPBOX, and likely others) are not buildable, making full page reconstruction fragile.

**Implementation scope:** Ongoing. Each missing widget requires grammar token + parser rule + builder function. Pattern is consistent but needs widget-by-widget work.

---

## Priority Recommendation

| Priority | Bug | Rationale |
|----------|-----|-----------|
| 1 | Bug 2 (LOOP panic) | One-line fix, highest severity (crash) |
| 2 | Bug 1 (ALTER ENTITY) | Most dangerous to users (silent data loss) |
| 3 | Bug 6 (reserved words) | Grammar-only fix, unblocks entity round-tripping |
| 4 | Bug 5 (reference checker) | Prevents false-positive validation |
| 5 | Bugs 3+4 (widgets) | Incremental, each widget is independent work |
| 6 | IF EXISTS | Most-requested missing feature for script idempotency |
