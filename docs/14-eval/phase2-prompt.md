# Eval Framework Phase 2: Automated Claude Invocation + Docker Checks

## Goal

Extend the eval framework (Phase 1 already implemented) to automate the full loop: copy a template project, invoke Claude Code CLI with the eval prompt, run validation checks, handle iteration scenarios, and produce scored reports. No manual steps required.

## What Already Exists (Phase 1)

The `cmd/mxcli/evalrunner/` package is already implemented with:

- **`parser.go`** — Parses eval test definitions from Markdown+YAML frontmatter files (like `docs/14-eval/eval-1.md`). Key types: `EvalTest`, `EvalIteration`, `check`.
- **`checks.go`** — Runs 8 check types against an MPR file by shelling out to `mxcli`: `entity_exists`, `entity_has_attribute`, `page_exists`, `page_has_widget`, `microflow_exists`, `navigation_has_item`, `mx_check_passes`, `lint_passes`. Uses wildcard pattern matching (`*.Book` matches `MyModule.Book`).
- **`results.go`** — Result types (`CheckResult`, `PhaseResult`, `EvalResult`, `RunSummary`) and scoring logic.
- **`report.go`** — Console, JSON, and Markdown report generation.
- **`cmd_eval.go`** — Cobra commands: `eval check` (validate project against criteria) and `eval list` (list tests).

The existing `eval check` command validates an already-built project:
```bash
mxcli eval check docs/14-eval/eval-1.md -p app.mpr
```

Phase 2 adds `eval run` which automates the entire pipeline.

## What Needs to Be Built

### 1. `eval run` command (`cmd/mxcli/cmd_eval.go`)

Add a new `evalRunCmd` subcommand:

```bash
# run all eval tests with automated Claude invocation
mxcli eval run docs/14-eval/ --template mx-test-projects/template-app-116/TemplateApp116.mpr

# run specific test
mxcli eval run docs/14-eval/eval-1.md --template template.mpr --test APP-001

# use specific model
mxcli eval run docs/14-eval/ --template template.mpr --model opus

# Skip Docker checks (L0-L2 only)
mxcli eval run docs/14-eval/ --template template.mpr --skip-docker

# Output reports
mxcli eval run docs/14-eval/ --template template.mpr --output eval-results/
```

Key flags:
- `--template` (required) — Path to a clean Mendix project to copy for each test
- `--model` — Claude model to use (default: `sonnet`)
- `--test` — Run only a specific test ID
- `--skip-docker` — Skip Docker build/run (L3+)
- `--output` — Report output directory
- `--max-turns` — Max Claude turns (default: 50)
- `--timeout` — Per-test timeout (default: from eval YAML or 10m)
- `--color` — Colored console output

### 2. Runner orchestrator (`cmd/mxcli/evalrunner/runner.go`)

Main orchestration function:

```go
type RunOptions struct {
    EvalFiles    []string      // Eval test file(s) or directory
    TemplatePath string        // path to template .mpr to copy
    TestID       string        // Optional: run only this test
    model        string        // Claude model (default: "sonnet")
    MaxTurns     int           // max Claude turns (default: 50)
    SkipDocker   bool          // Skip L3+ checks
    SkipMxCheck  bool          // Skip mx check
    OutputDir    string        // Report output directory
    MxCliPath    string        // path to mxcli binary
    Color        bool
    timeout      time.Duration
    Stdout       io.Writer
    Stderr       io.Writer
}

func run(opts RunOptions) (*RunSummary, error)
```

Pipeline per test:
1. **Copy template project** to a temp directory (entire project dir, not just the .mpr)
2. **Initialize for Claude** — Run `mxcli init` on the copied project
3. **Invoke Claude Code** with the prompt from the eval test
4. **Run L0-L2 checks** (reuse existing `RunChecks()`)
5. **Optionally run L3** — `mxcli docker run -p app.mpr --wait`
6. **Score and report**
7. **Iteration** — If the eval test has an `## Iteration` section:
   - Invoke Claude again with the iteration prompt (same session via `--continue`)
   - Re-run checks (iteration checks + re-check initial checks)
   - Score iteration phase separately
8. **Cleanup** — Stop Docker if started, optionally keep or remove temp project

### 3. Claude invocation wrapper (`cmd/mxcli/evalrunner/claude.go`)

```go
type ClaudeOptions struct {
    WorkDir    string   // project directory
    Prompt     string   // user prompt
    model      string   // "sonnet", "opus", etc.
    MaxTurns   int
    SessionID  string   // for --continue on iteration
    continue   bool     // use --continue flag
    timeout    time.Duration
}

type ClaudeResult struct {
    ExitCode   int
    Stdout     string
    Stderr     string
    Duration   time.Duration
}

func InvokeClaude(opts ClaudeOptions) (*ClaudeResult, error)
```

The invocation should use:
```bash
claude -p "$PROMPT" \
  --print \
  --model $MODEL \
  --max-turns $MAX_TURNS \
  --session-id $SESSION_ID \
  --dangerously-skip-permissions \
  --output-format json \
  2>&1
```

Key considerations:
- **`--print`** — Non-interactive mode, prints response and exits
- **`--dangerously-skip-permissions`** — Eval runs in controlled environments, skip permission prompts
- **`--session-id`** — Use a deterministic session ID (e.g., `eval-APP-001`) so iteration can `--continue` the same session
- **`--output-format json`** — Capture structured output for the transcript
- **Working directory** — `cmd.Dir` should be set to the copied project directory so Claude finds the `.mpr` and `.claude/` skills
- **Timeout** — Use `context.WithTimeout` to kill Claude if it exceeds the limit
- **Transcript capture** — Save full stdout/stderr to `eval-results/APP-001/claude-transcript.json`

For iteration:
```bash
claude -p "$ITERATION_PROMPT" \
  --print \
  --continue \
  --session-id eval-APP-001 \
  --dangerously-skip-permissions \
  ...
```

### 4. Project copying

Copy the entire template project directory (not just the .mpr) to a temp location. For MPR v2 projects (Mendix >= 10.18), the project includes `mprcontents/`, `widgets/`, `themesource/`, etc. alongside the `.mpr` file.

```go
func CopyProject(templateMprPath string, destDir string) (string, error)
```

- Copy the entire parent directory of the template .mpr
- Return the path to the new .mpr file
- Use `os.CopyFS` or walk + copy

### 5. Docker check types

Add two new check types to `checks.go`:

- `docker_starts: true` — Run `mxcli docker run -p app.mpr --wait`, pass if exit code 0
- `docker_no_startup_errors: true` — Parse Docker logs for error markers

These are expensive (~1-3 minutes), so they're gated behind `--skip-docker`.

## Existing Infrastructure to Reuse

| Need | Existing | Location |
|------|----------|----------|
| Eval parsing | `ParseEvalFile()`, `ParseEvalDir()` | `evalrunner/parser.go` |
| Check execution | `RunChecks()` | `evalrunner/checks.go` |
| Scoring | `PhaseResult.ComputeScore()`, `EvalResult.ComputeOverallScore()` | `evalrunner/results.go` |
| Reports | `PrintResult()`, `WriteJSONReport()`, `WriteMarkdownReport()` | `evalrunner/report.go` |
| Docker build/run | `docker.Run()` | `cmd/mxcli/docker/runtime.go` |
| Project init | `mxcli init` command | `cmd/mxcli/init.go` |
| Template project | `mx-test-projects/template-app-116/TemplateApp116.mpr` | Already exists |

## Example End-to-End Flow

```
$ mxcli eval run docs/14-eval/eval-1.md \
    --template mx-test-projects/template-app-116/TemplateApp116.mpr \
    --model sonnet --output eval-results/

Eval run: 1 test(s), model: sonnet
============================================================

[APP-001] Copying template project...
[APP-001] Initializing project for Claude Code...
[APP-001] Invoking Claude with initial prompt...
  Prompt: "create an app to manage my bookstore inventory..."
  Claude completed in 45s (32 turns)
[APP-001] Running initial checks...
  [PASS] entity_exists *.Book — found: Bookstore.Book
  [PASS] entity_has_attribute *.Book.Title string
  [PASS] entity_has_attribute *.Book.Author string
  [PASS] page_exists *overview*
  [PASS] page_exists *Edit*
  [PASS] navigation_has_item true
  [PASS] mx_check_passes true
  Score: 10/10 (100%)

[APP-001] Invoking Claude with iteration prompt...
  Prompt: "add a category field to the books..."
  Claude completed in 20s (12 turns)
[APP-001] Running iteration checks...
  [PASS] entity_has_attribute *.Book.Category
  Score: 1/1 (100%)

============================================================
Overall: APP-001 = 100% (11/11)
Reports written to eval-results/run-2026-02-25T16-00/
```

## Files to Create/Modify

| File | Action | Description |
|------|--------|-------------|
| `cmd/mxcli/evalrunner/runner.go` | Create | Main orchestrator |
| `cmd/mxcli/evalrunner/claude.go` | Create | Claude CLI invocation wrapper |
| `cmd/mxcli/cmd_eval.go` | Modify | Add `evalRunCmd` subcommand |
| `cmd/mxcli/main.go` | Modify | Register run command flags |

## Testing

After implementation, test with:

```bash
# build
make build

# run eval against the template project
./bin/mxcli eval run docs/14-eval/eval-1.md \
  --template mx-test-projects/template-app-116/TemplateApp116.mpr \
  --skip-docker --output /tmp/eval-results/

# Verify: check that Claude was invoked, checks ran, reports generated
cat /tmp/eval-results/run-*/APP-001/score.json
cat /tmp/eval-results/run-*/summary.md
```

## Important Notes

- The `claude` CLI must be available in PATH
- The template project should be a clean Mendix 11.6+ project (the template-app-116 in mx-test-projects works)
- `--dangerously-skip-permissions` is needed because eval runs non-interactively
- Session IDs should be deterministic per test so iteration `--continue` works
- Capture the Claude transcript (stdout) for debugging and analysis
- Set `MXCLI_QUIET=1` when invoking mxcli from within the runner to suppress banners
- The runner should work with the current mxcli binary (use `os.Executable()` to find self)
