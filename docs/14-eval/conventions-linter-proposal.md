# Proposal: Mendix Best Practices Linter Rules

**Date:** 2026-02-26
**Source:** Conventions.pdf (Squad Apps best practices document)
**Goal:** Automatically assess how well a Mendix project follows established best practices, using `mxcli lint` rules and a new `mxcli report` command.

---

## 1. Executive Summary

The Conventions.pdf document defines 14 categories of Mendix project best practices covering naming, folder structure, security, maintainability, performance, error handling, and more. Today, `mxcli lint` covers only a fraction of these rules through 10 built-in Go rules (MDL001–MDL007, SEC001–SEC003) and 16 Starlark rules (ARCH, DESIGN, QUAL, SEC series).

This proposal identifies **17 new convention rules** that can be implemented primarily as Starlark lint rules, plus enhancements to existing rules and a new `mxcli report` command for holistic project health reporting.

---

## 2. Current Coverage

### 2.1 Built-in Go Rules

| Rule | What It Checks |
|------|---------------|
| MDL001 | Naming conventions (PascalCase; 6 microflow prefixes: ACT\_, SUB\_, DS\_, VAL\_, SCH\_, IVK\_) |
| MDL002 | Empty microflows (zero activities) |
| MDL003 | Domain model size (>15 persistent entities per module) |
| MDL004 | Empty validation feedback messages |
| MDL005 | Unconfigured image widgets |
| MDL006 | Empty containers (runtime crash risk) |
| MDL007 | Page navigation security (pages in nav without allowed roles) |
| SEC001 | Persistent entities with no access rules |
| SEC002 | Weak password policy (<8 characters) |
| SEC003 | Demo users active in production security level |

### 2.2 Starlark Rules

| Rule | What It Checks |
|------|---------------|
| ARCH001–003 | Cross-module data access, data changes through microflows, business keys |
| DESIGN001 | Entity attribute count (>10) |
| QUAL001–004 | McCabe complexity, documentation, long microflows (>25), orphaned elements |
| SEC004–009 | Guest access, strict mode, PII exposure, unconstrained access |

---

## 3. Gap Analysis

The table below maps each convention from the PDF to existing coverage and identifies gaps.

### 3.1 Naming Conventions (PDF §1)

| Convention | Status | Details |
|-----------|--------|---------|
| Microflow prefix naming | **Partial** | MDL001 recognizes 6 prefixes. Missing 11: BCO\_, ACO\_, BCR\_, ACR\_, BDE\_, ADE\_, BRO\_, ARO\_, OCH\_, SE\_, DL\_, PWS\_, ASU\_, NAV\_, LOGIN\_ |
| Entity PascalCase, no underscores | Covered | MDL001 |
| Attribute PascalCase | Covered | MDL001 (implicit) |
| Boolean naming (IsX, HasX) | **Gap** | No rule checks boolean attribute naming patterns |
| No default values on string/numeric attributes | **Gap** | No rule inspects attribute default values |
| Entity names should be singular | **Gap** | Would require NLP / heuristic plural detection |
| Association rename = suffix only | **Gap** | Hard to enforce automatically |
| Page naming suffixes (\_NewEdit, \_View, \_Overview) | **Gap** | MDL001 only checks PascalCase, not semantic suffixes |
| Enumeration prefix (ENUM\_) | **Gap** | MDL001 checks PascalCase only |
| Snippet prefix (SNIPPET\_) | **Gap** | No snippet naming rule exists |
| Java action prefix (JA\_) | **Gap** | No Java action naming rule exists |

### 3.2 Folder Structure (PDF §2)

| Convention | Status | Details |
|-----------|--------|---------|
| Each module has [Objects] and [UI] folders | **Gap** | Catalog does not currently track folder hierarchy |
| ACT\_ and DS\_ microflows live in [UI] folder | **Gap** | Requires folder-aware catalog |
| Entity subfolders under [Objects] | **Gap** | Requires folder-aware catalog |

### 3.3 Security (PDF §3)

| Convention | Status | Details |
|-----------|--------|---------|
| 1:1 module-role-to-user-role mapping | **Gap** | Role mapping data is available via reader |
| Entity access rules on persistent entities | Covered | SEC001 |
| No create/delete entity access rights | **Gap** | Permissions data exists but rule not written |
| XPath constraints on all entity access | **Partial** | SEC007 covers anonymous only; needs extension to all roles |
| Set default rights to None | **Gap** | Requires BSON inspection of entity access defaults |
| No entity validation rules | **Gap** | Validation rule count not currently in catalog |
| No entity event handlers | **Gap** | Event handler data not currently in catalog |

### 3.4 Maintainability (PDF §4)

| Convention | Status | Details |
|-----------|--------|---------|
| Caption on exclusive splits | **Gap** | Requires activity-level BSON inspection |
| Never change microflow activity captions | N/A | Not automatically detectable |
| Remove all warnings | N/A | Studio Pro concern, not model-level |
| Don't change imported modules | **Gap** | Could detect modifications to marketplace modules |
| ACT\_ microflows: only page activities + submicroflow calls | **Gap** | Requires activity-level data (FULL catalog) |
| Max 15 objects per microflow | **Partial** | QUAL003 uses 25 limit; needs configurable threshold |
| Delete unused/excluded items | Covered | QUAL004 (orphaned elements) |
| Do not duplicate code | N/A | Requires clone detection (complex) |

### 3.5 Performance (PDF §5)

| Convention | Status | Details |
|-----------|--------|---------|
| No calculated attributes | **Gap** | Attribute metadata may include IsCalculated flag |
| No event handlers | **Gap** | Same as §3 gap above |
| No commit in loops | **Gap** | Requires activity + loop structure analysis |
| Use batching for >100 objects | N/A | Intent-level; hard to enforce automatically |
| No Export to Excel (use CSV) | **Gap** | Could detect ExportToExcel widget usage |

### 3.6 Error Handling (PDF §9)

| Convention | Status | Details |
|-----------|--------|---------|
| Error handling on service calls / Java actions | **Gap** | Requires BSON inspection of call activities |
| Never use error handling "Continue" | **Gap** | Requires BSON inspection of error handlers |

---

## 4. Proposed New Rules

### 4.1 Tier 1 — Works with Current Catalog Data

These rules can be implemented immediately as Starlark rules using existing catalog APIs.

| Rule ID | Name | Category | Severity | What It Checks |
|---------|------|----------|----------|----------------|
| **CONV001** | BooleanNaming | naming | warning | Boolean attributes must start with `Is`, `Has`, `Can`, `Should`, `Was`, or `Will` |
| **CONV002** | NoEntityDefaultValues | quality | warning | String and numeric attributes should not have non-empty default values |
| **CONV003** | PageNamingSuffix | naming | info | Pages should follow `Entity_NewEdit`, `Entity_View`, `Entity_Overview` pattern |
| **CONV004** | EnumerationPrefix | naming | info | Enumerations should be prefixed with `ENUM_` |
| **CONV005** | SnippetPrefix | naming | info | Snippets should be prefixed with `SNIPPET_` |
| **CONV006** | NoCreateDeleteRights | security | warning | Entity access rules should not grant Create or Delete rights (except non-persistent) |
| **CONV007** | XPathOnAllAccess | security | warning | All persistent entity access rules for non-reference-data should have XPath constraints |
| **CONV008** | ModuleRoleMapping | security | warning | Each module role should map to exactly one user role |
| **CONV009** | MaxMicroflowObjects | quality | warning | Microflows should have at most 15 objects (configurable, convention says 15) |

### 4.2 Tier 2 — Requires FULL Catalog or BSON Inspection

These rules need activity-level data from `refresh catalog full` or raw BSON reader access.

| Rule ID | Name | Category | Severity | What It Checks |
|---------|------|----------|----------|----------------|
| **CONV010** | ACTMicroflowContent | architecture | warning | ACT\_ microflows should only contain page activities (ShowPage, ClosePage, ShowHomePage, ShowMessage, DownloadFile) and SubMicroflow calls |
| **CONV011** | NoCommitInLoop | performance | warning | Commit actions should not appear inside loop containers |
| **CONV012** | ExclusiveSplitCaption | quality | warning | Exclusive split activities must have a non-empty caption |
| **CONV013** | ErrorHandlingOnCalls | quality | warning | REST call, web service call, and Java action activities should have custom error handling |
| **CONV014** | NoContinueErrorHandling | quality | error | Error handling should never use "Continue" mode |

### 4.3 Tier 3 — Requires Catalog Enhancement

These rules need new data points to be added to the catalog builder.

| Rule ID | Name | Category | Severity | What It Checks |
|---------|------|----------|----------|----------------|
| **CONV015** | NoEntityValidationRules | quality | warning | Entities should not have validation rules defined |
| **CONV016** | NoEventHandlers | performance | warning | Entities should not have before/after commit/create/delete/rollback event handlers |
| **CONV017** | NoCalculatedAttributes | performance | warning | Attributes should not be calculated (virtual) |

### 4.4 Enhancement to MDL001

Extend the existing Go naming rule to recognize all 17 microflow prefixes from the conventions:

```
Current:  ACT_, SUB_, DS_, VAL_, SCH_, IVK_
add:      BCO_, ACO_, BCR_, ACR_, BDE_, ADE_, BRO_, ARO_, OCH_, SE_, DL_, PWS_, ASU_, NAV_, LOGIN_
```

---

## 5. Report Command

### 5.1 Concept

A new `mxcli report` command that runs all convention rules and produces a structured **Best Practices Report** with per-category scoring.

```bash
# generate report
mxcli report -p app.mpr --format markdown
mxcli report -p app.mpr --format json
mxcli report -p app.mpr --format html
```

### 5.2 Report Structure

```
╔══════════════════════════════════════════════════╗
║          Mendix Best Practices Report            ║
║          MyProject — 2026-02-26                  ║
╠══════════════════════════════════════════════════╣
║                                                  ║
║  Overall Score:  72/100                          ║
║                                                  ║
║  Category Scores:                                ║
║  ├── Naming Conventions     ████████░░  82%      ║
║  ├── security               ██████░░░░  60%      ║
║  ├── Maintainability        ████████░░  78%      ║
║  ├── Performance            █████████░  90%      ║
║  ├── Architecture           ███████░░░  70%      ║
║  └── Quality                ████████░░  75%      ║
║                                                  ║
╚══════════════════════════════════════════════════╝
```

Each category section would list:
- Total elements checked
- Violations found (grouped by severity)
- Top 5 actionable recommendations
- Trend comparison (if previous report exists)

### 5.3 Scoring Method

- Each rule contributes to its category score
- Violations reduce the score proportionally:
  - Error: −10 points per violation
  - Warning: −3 points per violation
  - Info: −1 point per violation
  - Hint: −0.5 points per violation
- Score is capped at 0–100, normalized by the number of elements checked
- Overall score is a weighted average of category scores

---

## 6. Implementation Plan

### Phase 1: Quick Wins (Starlark rules, no code changes)

1. Write CONV001–CONV009 as Starlark rules in `.claude/lint-rules/`
2. Extend MDL001 prefix list (small Go change)
3. Ship via `mxcli init` into projects

**Effort:** ~2 days
**Impact:** Covers naming, basic security, and quality conventions

### Phase 2: Deep Analysis Rules (BSON inspection)

1. Implement CONV010–CONV014 as Go rules requiring reader access
2. Add activity-type filtering to catalog queries

**Effort:** ~3 days
**Impact:** Covers ACT\_ architecture, loop commits, error handling, split captions

### Phase 3: Catalog Enhancement

1. Add validation rule count, event handler flags, calculated attribute flag to entity catalog
2. Implement CONV015–CONV017

**Effort:** ~2 days
**Impact:** Covers entity-level performance/quality conventions

### Phase 4: Report Command

1. Implement `mxcli report` with category grouping and scoring
2. Support markdown, JSON, and HTML output formats
3. Add trend tracking (compare against previous report)

**Effort:** ~3 days
**Impact:** Holistic project health dashboard

---

## 7. Conventions Not Automatically Enforceable

Some conventions from the PDF are guidance-oriented and cannot be reliably detected by static analysis:

| Convention | Reason |
|-----------|--------|
| "Don't optimize prematurely" | Intent-based guidance |
| "Use understandable variable names" | Subjective / requires NLP |
| "Do not duplicate code" | Requires clone detection |
| "Never change activity captions" | Requires change tracking / history |
| "Use SASS, not CSS" | File-system concern outside MPR |
| Branching strategy | Git/SVN operational concern |
| Migration template pattern | Intent-based; many valid implementations |
| "Use error handling for the right reasons" | Requires understanding intent |
| Association rename = suffix only | Requires knowing original auto-generated name |

These are best addressed through **developer training, code review checklists, or Claude skills** that guide developers during development (e.g., the existing `write-microflows.md` skill).

---

## 8. Delivery Mechanism

| Mechanism | Purpose |
|-----------|---------|
| **Starlark rules** in `.claude/lint-rules/` | Convention-specific checks, shipped via `mxcli init` |
| **Go rules** in `mdl/linter/rules/` | Rules requiring BSON inspection or high performance |
| **lint-config.yaml** | Per-project configuration (enable/disable, severity overrides, thresholds) |
| **`mxcli report`** | Aggregated best-practices report with scoring |
| **Claude skills** | Developer guidance during authoring (existing mechanism) |

All Starlark rules would be **enabled by default** but configurable. Projects can disable rules that don't match their specific conventions via `lint-config.yaml`.
