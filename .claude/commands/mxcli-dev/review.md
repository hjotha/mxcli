# /mxcli-dev:review — PR Review

Run a structured review of the current branch's changes against the CLAUDE.md
checklist, then check the recurring findings table below for patterns that have
burned us before.

## Steps

1. Run `gh pr view` and `gh pr diff` (or `git diff main...HEAD`) to read the change.
2. Work through the CLAUDE.md "PR / Commit Review Checklist" in full.
3. Then check every row in the Recurring Findings table below — flag any match.
4. Report: blockers first, then moderate issues, then minor. Include a concrete fix
   option for every blocker (not just "this is wrong").
5. After the review: **add a row** to the Recurring Findings table for any new
   pattern not already covered.

---

## Recurring Findings

Patterns caught in real reviews. Each row is a class of mistake worth checking
proactively. Add a row after every review that surfaces something new.

| # | Finding | Category | Canonical fix |
|---|---------|----------|---------------|
| 1 | Formatter emits a keyword not present in `MDLParser.g4` → DESCRIBE output won't re-parse (e.g. `RANGE(...)`) | DESCRIBE roundtrip | Grep grammar before assuming a keyword is valid; if construct can't be expressed yet, emit `-- TypeName(field=value) — not yet expressible in MDL` |
| 2 | Output uses `$currentObject/Attr` prefix — non-idiomatic; Studio Pro uses bare attribute names | Idiomatic output | Verify against a real Studio Pro BSON sample before choosing a prefix convention |
| 3 | Malformed BSON field (missing key, wrong type) produces silent garbage output (e.g. `RANGE($x, , )`) | Error handling | Default missing numeric fields to `"0"`; or emit `-- malformed <TypeName>` rather than broken MDL |
| 4 | No DESCRIBE roundtrip test — grammar gap went undetected until human review | Test coverage | Add roundtrip test: format struct → MDL string → parse → confirm no error |
| 5 | Hardcoded personal path in committed file (e.g. `/c/Users/Ylber.Sadiku/...`) | Docs quality | Use bare commands (`go test ./...`) without absolute paths in any committed doc or skill |
| 6 | Docs-only PR cites an unmerged PR as a "model example" — cited PR had blockers | Docs quality | Only cite merged, verified PRs; or annotate with known gaps if citing in-flight work |
| 7 | Skill/doc table references a function that doesn't exist (e.g. `formatActionStatement()` vs `formatAction()`) | Docs quality | Grep function names before writing: `grep -r "func formatA" mdl/executor/` |
| 8 | "Always X" rule is too absolute for trivial edge cases (e.g. "always write failing test first" for one-char typos) | Docs quality | Soften to "prefer X" or add an exception clause; include the reasoning so readers can judge edge cases |
| 9 | Doc comment promises a fallback/feature that doesn't exist in the code (e.g., "raw-map fallback in the client" when no such fallback was implemented) | Docs quality | Grep for function/type names referenced in doc comments to confirm they exist before committing |
| 10 | BSON array items decoded by mongo driver are `primitive.D`, not `map[string]any` — bare type assertion `item.(map[string]any)` always fails silently, causing silent data loss (e.g. Languages not parsed, issue #480) | BSON parsing | Always use `extractBsonMap(item)` instead of `item.(map[string]any)`; write a parser unit test with `primitive.D` items to catch this class of bug |
| 11 | `execShow` switch missing a case for a new `ShowXxx` constant — executor handler is wired but never dispatched, command silently does nothing | Dispatch gap | After adding a new `Show*` constant and handler, grep `executor_query.go` to confirm the case is present; add a mock test that calls the handler directly |

---

## After Every Review

- [ ] All blockers have a concrete fix option stated.
- [ ] Recurring Findings table updated with any new pattern.
- [ ] If docs-only PR: every function name, path, and PR reference verified against
      live code before approving.
