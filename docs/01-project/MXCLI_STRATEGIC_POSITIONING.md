# mxcli / MDL — Strategic Positioning vs PED

## Context

Two parallel approaches have emerged for AI-agent-driven editing of Mendix applications:

- **PED** (Progressive Element Disclosure) — a Studio Pro MCP server exposing JSON-based tools (`ped_read_document`, `ped_get_schema`, `ped_update_document`, `ped_check_errors`). Live connection to an open project. Agent-turn-based, schema-on-demand, one-shot error-fix protocol. Optimized for interactive, single-project pair-modelling.
- **mxcli + MDL** — an offline CLI binary operating on `.mpr` files, driven by a SQL-like text DSL. Script artifacts, per-statement atomicity, pre-execution validation layers, Starlark-based custom lint rules, local `catalog.db` (SQLite) of project metadata. Ships with a VS Code extension providing both visual and textual editing, with multi-root workspaces enabling multi-app/solution development.

This memo captures a strategic conversation about where mxcli/MDL should invest to avoid duplicating PED and instead own the territory where it has structural advantages.

---

## Where PED wins — concede this ground

- Single-project interactive pair-modelling inside Studio Pro.
- Beginner and intermediate Mendix developers who benefit from immediate visual feedback.
- Ad-hoc "help me add this widget / fix this microflow" edits.
- Onboarding and teaching scenarios where the live editor loop is the best teacher.

Any feature whose value only materialises inside one open project session is likely a place where PED has the ergonomic edge and chasing parity is low-ROI.

---

## Where mxcli / MDL has structural advantages

### 1. Portfolio-scale customer use cases

The VS Code + textual MDL + multi-root workspace + (proposed) multi-document semantic search/replace stack addresses problems that are structurally impossible for a session-bound, single-project agent protocol. Highest-value segments:

- **Cross-project refactoring** — rename or restructure an entity used across 30 pages, 50 microflows, 4 modules, 3 apps. Visual editing is unusable; PED runs out of context. MDL + semantic search/replace turns hours into minutes.
- **Module promotion and monolith splitting** — move entities/microflows/pages between modules and apps, bump shared module versions across consumers.
- **Migrations and modernisation** — K2, Nintex, OutSystems, Oracle Forms, legacy Mendix. One-time, high-value, high-willingness-to-pay work with consulting-partner economics. Skills already exist in the demo1 project.
- **Mass model generation** — bootstrap apps from OpenAPI/ERD/database-schema inputs. PED's turn-by-turn cost makes 40-entity bootstraps impractical; MDL expresses the same in one script.
- **Governance and audit** — lint packs as PR/CI gates, `mxcli report` as fleet-wide health dashboard, MDL diffs as auditable change records. PED has no equivalent — it is by nature interactive.
- **Power-user textual editing** — dense microflows, bulk page tweaks, refactoring-heavy sprints. Retains senior developers who feel boxed in by visual-only tooling.

The buyer pattern matters: these are **platform / CoE / governance teams** at enterprises with 5+ Mendix apps, delivery partners running migrations, and ISVs shipping to many customer instances — not individual developers asking an AI to help with a widget.

### 2. Token and context efficiency — compounding at scale

mxcli has a structural token advantage that grows with project size, edit breadth, and session length.

**Four mechanisms:**

1. **Document reads scale with project size; MDL edits don't.** PED reads whole documents into context to edit them; a mature microflow/page is 5–20k tokens of JSON. An `ALTER PAGE … { INSERT AFTER txtEmail … }` is ~30 tokens and never needs the whole page.
2. **Catalog queries replace document walks.** Answers over `catalog.db` are 50-token SQL queries returning 200-token tables. The PED equivalent requires loading every candidate document. The gap widens linearly with project size.
3. **Schemas amortise differently.** PED fetches schemas per session, per type. MDL's grammar lives in the parser; the agent pays once for a bounded skill-file set (~2–10k tokens) that amortises across edits.
4. **Error loops are asymmetric.** `mxcli check --references` validates before apply; failed attempts cost only script size (~100–500 tokens). PED failures force document re-reads for diagnosis with a one-shot fix budget.

**Back-of-envelope for a realistic refactor** (rename `Customer.Email` → `Customer.EmailAddress` across 1 entity + 3 microflows + 2 pages + 4 filters):

- PED: ~50–80k tokens of tool I/O (9 document reads + 9 updates + check + potential re-reads).
- MDL: ~1–2k tokens (catalog query + one ~400-token script + check + exec results).

**Long-term compounding via LLM pretraining.** As MDL is published and used publicly, frontier models will learn the grammar natively — analogous to SQL. Implications:

- Skill files shift from teaching syntax to encoding higher-level patterns (architectural invariants, CRUD conventions, migration recipes, governance rules).
- Same skill-file budget holds 5–10x more patterns; density per token rises.
- Effective attention at any context size improves because less is spent on framework overhead.

**Strategic implication:** the gap is not just "nicer" — at enterprise scale it becomes a **capability gap**. There will be problems an MCP-bounded agent runs out of context trying to solve, that a CLI-equipped agent completes cleanly.

**Caveat:** realising the long-term pretraining advantage requires treating LLM-corpus seeding as an explicit product investment (public repos, community Q&A, reference datasets, open patterns), not assuming it happens automatically. Training lag to frontier models is ~12–24 months.

### 3. CLI composition — the biggest structural moat

This is qualitatively different from the token-efficiency point. It reframes the agent architecture.

- **MCP model:** the LLM is the compute node. Tool results flow *into* its context. It reasons over the data, emits edits.
- **CLI model:** the LLM is an **orchestrator**. Pipes, filters, and subprocesses do the compute. Only judgment-requiring results enter context.

Most questions about a Mendix project don't need judgment — they need filtering, counting, joining, grep. Pushing that to the shell is essentially free.

**Concrete capabilities this unlocks:**

- **Data that never enters context.** `mxcli catalog query "… WHERE type='RestCall' AND url LIKE '%legacy%'" | wc -l` — the agent reads "47." Via MCP the equivalent scans every microflow.
- **Intermediate files as durable memory.** `mxcli report … --format json > report.json; jq '…' report.json` — the agent can re-query later without re-materialising. MCP results are ephemeral.
- **Composition with the entire Unix ecosystem.** `jq`, `rg`, `fd`, `awk`, `sqlite3`, `git`, `gh`, `curl`, `xargs`, `parallel`, Python one-liners. Fifty years of network effects for free. MCP servers only compose with other MCP servers a client has been explicitly configured for.
- **Parallelism for one token of invocation.** `parallel mxcli validate ::: apps/*.mpr | grep ERROR`. MCP is turn-based.
- **SQLite as universal interop surface.** Any tool, language, or BI system can hit `catalog.db` — no custom protocol needed. Quiet moat.
- **The LLM doesn't need to be the smartest component.** A pipeline plus a deterministic lint pass does 80% of the work. Agent decides intent; shell does arithmetic. Cheaper and more reliable.

**Honest counterpoints:**

- MCP has cleaner safety boundaries (typed args, no bash accidents) — matters for novice users, not for platform teams in their own shells.
- Shell fluency requires capable agents — aligned with the target segment, which already uses frontier-class agents.
- Discoverability is weaker; invest in `mxcli --help` that reads well to an LLM and `man` pages an agent can grep.

### 4. Test-informed safety and verification — the force multiplier

Automated testing is the keystone that upgrades the pitch from "mxcli is the right tool for portfolio-scale Mendix engineering" to "mxcli is the only tool that makes agentic Mendix engineering *safe* at portfolio scale." Three mechanisms carry this.

**Testing reshapes the edit loop.** PED's one-shot fix protocol exists because there is no way to verify an edit beyond structural `ped_check_errors`. With real tests the loop becomes:

```
agent writes MDL  →  check (syntax)  →  exec (apply)  →  test (behaviour)  →  iterate
```

Every stage except "agent writes MDL" runs outside the agent's context. The agent consumes pass/fail + focused error excerpts — not logs, not document re-reads, not re-fetched schemas. This is the TDD-style agentic loop Claude Code and Cursor are built around. mxcli fits it naturally; PED cannot participate because Mendix testing requires CI or a build, neither of which is inside Studio Pro.

**Tests unlock operations PED is structurally forbidden from doing.** The Maia safety rules prohibit radical operations (rename, delete multiple elements, restructure) without explicit approval — correct, because without a correctness check the blast radius is unbounded. With a deterministic test suite the equation inverts: an agent can attempt a radical refactor and the test suite decides whether it was correct. This means the single biggest class of high-value tasks PED cannot safely attempt — large-scale refactoring, migrations, shared-module evolution — becomes safe in the mxcli workflow because the correctness check is mechanical, not judgment-based.

**The cost equation sharpens further.** Most mistakes are caught by deterministic test runs (~100 tokens of output) rather than LLM diagnosis (thousands of tokens of re-reading). Failed attempts cost runtime + a summary, not context bloat. For complex refactors, this is often the difference between "completes in three iterations" and "runs out of context."

**Portfolio-scale regression is a testing story.** `parallel mxcli test ::: apps/*.mpr | grep FAIL` makes "did the shared-module bump break any of 20 consumer apps?" trivial. This puts mxcli into the critical path of enterprise release processes — a sticky position.

**Skill-file implications.** As LLMs learn MDL syntax natively, skills shift toward "how to test and what to assert" — where real domain expertise lives (validation patterns, boundary conditions, invariants, regression traps). Syntax content decays; testing content compounds.

**Honest counterpoints:**

- Mendix's testing ecosystem is historically weak. The UnitTesting module works, but a "my Mendix app has meaningful tests" culture is not widespread. This is a product opportunity but also a gating variable — expect to invest in evangelism and scaffolding, not just tooling.
- Generated tests have the "tests the mock" risk. Coverage metrics will be flattering; semantic value can be low. Invest in property-based and behavioural testing patterns explicitly, not just "a test per microflow."
- PED could in principle drive tests too — but only from inside Studio Pro, which isn't where CI lives. The structural gap persists.

---

## Honest caveats on the whole thesis

- **PED will not stand still.** Models will get better at JSON/schema traversal; per-op costs narrow. The *absolute* token gap will shrink. What doesn't narrow are the structural edges (catalog, text artifacts, composition, offline execution, CI gating) — anchor the story there.
- **Training lag is real.** Publishing docs doesn't automatically seed pretraining. Requires volume: public repos, Stack-Overflow-style Q&A, blog content, open reference corpora. Treat as product investment.
- **MDL v1 becomes a permanent commitment** once it lands in model weights. Backward-compat discipline is now strategic, not just developer-courtesy. SQL's conservatism is the right model; early-JavaScript churn is the antipattern.
- **Name collision risk.** "MDL" has other meanings. Make sure the distinctive training signal is the grammar (`CREATE PERSISTENT ENTITY …`, `.mdl` extension, statement shape), not the acronym.
- **These arguments assume capable agents.** Not a broad-consumer story; this is platform-team and power-user territory.
- **catalog.db coverage must be invested in** for the query-as-composition argument to hold. Shallow metadata limits the reach.
- **Mendix testing maturity is a gating variable.** The testing advantage only lands for customers who invest in test suites. Meeting customers where they are — scaffolding, generators, patterns — is part of the product, not just a documentation problem.

---

## Recommended investments, roughly prioritised

1. **Bidirectional `.mpr` ↔ MDL** — `mxcli diff`, `mxcli merge`, `mxcli apply`. Every downstream capability (reviewable history, portable migrations, semantic merge, replay across fleet) compounds off this.
2. **Semantic multi-document search/replace** — reference-aware and type-aware, not text-grep. The wedge for cross-project refactoring.
3. **`mxcli query` as first-class SQL surface over `catalog.db`** — machine-queryable, composable with `jq`/`awk`/Pandas/BI. Extend catalog coverage aggressively.
4. **Stable pipe-friendly I/O** — `--format json|csv|tsv` everywhere, no decoration on stdout, structured stderr, meaningful exit codes.
5. **Machine-readable `--help`** — agent-planning-friendly JSON description of commands and flags.
6. **LLM-corpus seeding programme** — public GitHub corpus, partner blog content, open reference migrations, explicit "Rosetta" datasets (intent → MDL) suitable for pretraining ingestion.
7. **MDL backward-compat governance** — keyword-adding discipline, deprecation policy, versioning story, published conformance tests.
8. **Starlark lint ecosystem** — shareable rule packs (`@security/owasp`, `@governance/gdpr`, per-customer packs), registry, CI integrations. Network-effect moat.
9. **Fleet operations as a first-class product** — `mxcli fleet apply`, cross-project impact analysis, portfolio-level reports.
10. **Pipeline recipes published as skills** — reframe agent behaviour from "load and reason" to "compose and decide."
11. **`mxcli test` as a first-class command** with structured output modes (JUnit XML, TAP, JSON). Wrap the existing Mendix UnitTesting module + UI-test runners; don't reinvent. This is what turns the CI loop from "check and apply" into "check, apply, verify" and unlocks safe radical operations.
12. **Test generation paired with model generation.** When mxcli generates 40 CRUD microflows from a schema, generate their tests in the same pass. Closes the trust gap on mass generation.
13. **Test-informed safety policy.** Codify that radical operations are permitted when pre- and post-tests both pass. Make it a customer-facing promise, not an internal agent convention.
14. **Fleet test runner.** `mxcli fleet test --projects apps/*.mpr --since <sha>`. Cross-project regression becomes a product, not a feature.
15. **CI templates** (GitHub Actions / GitLab / Azure DevOps) wiring check → exec → test → lint → report. Lowers adoption cost to near zero for platform teams.

---

## Where NOT to spend effort

- Duplicating PED's single-project interactive lane.
- Beginner/onboarding guidance.
- Rebuilding general Mendix knowledge the agent can get from any source.
- Feature parity for its own sake — test every feature proposal against *"Does this get better as the customer's app count goes from 1 to 20?"* If no, it's in PED's lane.

---

# Summary for slides

A focused four-slide deck. Slides 1–3 set up the territory, the efficiency case, and the architectural moat. Slide 4 is the capstone that upgrades the pitch from "right tool" to "only safe tool."

## Slide 1 — Focus on the Portfolio, Not the Project

**Thesis:** PED is for a developer with one open project. mxcli/MDL is for a team with a portfolio. Don't duplicate; diverge.

**PED's lane (concede):**
- Single-project, interactive pair-modelling.
- Beginner onboarding and immediate visual feedback.
- Ad-hoc widget or microflow edits in the open editor.

**mxcli/MDL's territory (own it):**
- Cross-project refactoring — rename/restructure across 30+ documents in multiple apps.
- Migrations — legacy modernisation, version bumps, mass model generation.
- Governance — lint-as-CI, MDL diffs as audit artifacts, fleet-wide reports.
- Power-user textual editing in VS Code, including multi-root multi-app workspaces.

**Buyer profile:** platform / CoE teams, delivery partners, ISVs — not individual developers.

**Decision test for every feature:** *Does it get better as the customer's app count goes from 1 to 20?*

---

## Slide 2 — Context Efficiency Is a Capability Gap at Scale

**Thesis:** mxcli's token advantage isn't a convenience — at enterprise scale it's the difference between problems an agent can solve and problems it can't.

**Four mechanisms compound:**
- Document reads scale with project size; MDL `ALTER` statements don't.
- Catalog (SQLite) queries replace document walks.
- MDL grammar amortises into LLM weights as the corpus grows publicly.
- `check --references` fails fast, before apply — failures cost script size, not re-reads.

**Back-of-envelope for a cross-doc rename (9 affected files):**
- PED: ~50–80k tokens of tool I/O.
- MDL: ~1–2k tokens.

**Long-term:** as MDL lands in LLM pretraining, skill files shift from teaching syntax to encoding patterns, lifting density per token 5–10x and freeing reasoning headroom for hard problems.

**Caveat:** realising the pretraining advantage requires active corpus seeding (public repos, blogs, Q&A, reference datasets) and MDL backward-compat discipline — both are product investments.

---

## Slide 3 — CLI Composition Is the Real Moat

**Thesis:** MCP forces every byte through the LLM's context. CLI pipes move work out of the context entirely. This is a structural architecture difference, not a feature difference.

**The reframe:**
- MCP: the LLM is the compute node.
- CLI: the LLM is an orchestrator; pipes, filters, and subprocesses do the compute.

**What this unlocks:**
- Data that never enters context (`… | wc -l`, `… | jq`, `… | grep`).
- Intermediate files as durable session memory.
- Composition with the entire Unix ecosystem — `jq`, `rg`, `sqlite3`, `git`, `parallel`, Python. Fifty years of network effects, free.
- Real parallelism (`parallel mxcli validate ::: apps/*.mpr`).
- `catalog.db` as SQLite = universal interop surface for future tools, dashboards, BI.

**Strategic positioning:** this makes mxcli not "a CLI for Mendix" but **Mendix's place in the Unix ecosystem** — a much bigger and more defensible position than anything MCP-bounded can become.

**Required investments to realise it:** stable pipe-friendly output modes, `mxcli query` as first-class SQL surface, machine-readable `--help`, aggressive catalog coverage.

---

## Slide 4 — Testing Is What Makes Agentic Mendix Engineering Safe at Scale

**Thesis:** without tests, an agent editing a portfolio is a liability. With tests, the same agent is a force multiplier. mxcli's CI-native loop is the only place this can land — PED is session-bound and can't participate in CI.

**The edit loop changes shape:**

```
agent writes MDL  →  check  →  exec  →  test  →  iterate
```

Every stage except "agent writes MDL" runs outside the agent's context. Pass/fail + focused errors are what enter context — not logs, not re-reads.

**What tests unlock that PED structurally cannot:**
- Radical operations become safe — rename, restructure, monolith-splitting, shared-module evolution. The correctness check is mechanical, not judgment-based.
- Portfolio regression becomes trivial — `parallel mxcli test ::: apps/*.mpr | grep FAIL`.
- Mass-generation trust — generated tests paired with generated microflows close the "did the agent get this right?" gap.

**Cost implications reinforce Slide 2:** most mistakes caught by a ~100-token test result rather than thousands of tokens of LLM diagnosis. Difference between "completes in three iterations" and "runs out of context" on complex refactors.

**Positioning upgrade:** from "mxcli is the right tool for portfolio-scale Mendix engineering" to **"mxcli is the only tool that makes agentic Mendix engineering safe at portfolio scale."** That is categorically stronger and architecturally out of PED's reach.

**Caveat + opportunity:** Mendix testing culture is historically thin. The scaffolding, generators, and patterns to change that are a product investment — and also a differentiated one, because nobody else is solving it.
