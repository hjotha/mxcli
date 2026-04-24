# mxcli / MDL — Strategic Positioning vs PED

## Context

Two parallel approaches have emerged for AI-agent-driven editing of Mendix applications:

- **PED** (Progressive Element Disclosure) — a Studio Pro MCP server exposing JSON-based tools (`ped_read_document`, `ped_get_schema`, `ped_update_document`, `ped_check_errors`). Live connection to an open Studio Pro project. Agent-turn-based, schema-on-demand, one-shot error-fix protocol. Optimised for interactive, single-project pair-modelling.
- **mxcli + MDL** — a CLI binary driven by a SQL-like text DSL, recently refactored to support **multiple backends**: one that manipulates `.mpr` files directly (offline, CI-native), and one that uses the ModelAPI inside Studio Pro (live, interactive). Same DSL, same skills, same composition model across both. Ships with a VS Code extension providing visual and textual editing, multi-root workspaces for multi-app solutions, `catalog.db` (SQLite) for project metadata, and Starlark-based custom lint rules.

This memo captures a strategic conversation about where mxcli/MDL should invest to avoid duplicating PED and instead own the territory where it has structural advantages.

---

## The real axis of comparison — interaction protocol, not connection target

The dual-backend refactor **collapses the "live vs offline" distinction** that used to separate PED from mxcli. Both can now edit a running Studio Pro project interactively. What remains — and what is durable — is a different axis:

**How does the agent interact with the model?**

- **PED:** MCP tool calls. Each operation is a JSON request/response inside the agent's context window. The LLM is the compute node; tools deliver data into context and emit edits.
- **mxcli:** CLI invocations over a text DSL. The agent composes scripts, pipes, and shell commands. The LLM is an orchestrator; compute happens in subprocesses outside context.

Every durable strategic difference — text artifacts vs transcripts, CLI composition vs in-context reasoning, CI-native vs session-bound, catalog-query vs document-walk, training-corpus compounding vs runtime-only — flows from this axis. It is not affected by the backend target.

---

## Where PED wins — the narrowed concession

With the dual-backend in place, the set of PED-preferred scenarios shrinks but doesn't disappear. Honest list:

**Durable PED advantages:**
- **MCP-native distribution.** Drops into Claude Desktop, Cursor, and similar clients with zero shim. Capability is easy to match; distribution is not.
- **Typed, self-describing tool surface.** An agent that has never seen PED can still use it. mxcli requires the agent to know MDL (closing as LLMs learn it, but not yet closed).
- **Opinionated safety rails as protocol.** One-shot fix protocol, radical-operation prohibitions, abstract-type navigation. Safety is steered by the protocol itself, not by convention.
- **First-party Studio Pro integration.** PED ships inside Studio Pro and gets internal model changes first. mxcli's live backend depends on ModelAPI stability and release cadence.
- **Smaller envelope for trivial single-element edits** when already inside an MCP client.

**Contingent (no longer PED-unique):**
- Live immediacy — mxcli-live matches it.
- Visual pair-modelling in the open editor — both can drive it.

**Rule of thumb:** in an MCP client, for a trivial ad-hoc edit, PED is still the lower-overhead choice until an `mxcli-mcp` wrapper exists. Those edits are low-value and fine to concede — they are the on-ramp, not the destination.

---

## PED's structural weaknesses (unchanged by the dual-backend refactor)

- **No text artifact.** No diff, no review, no commit, no replay. The change exists only as a transcript.
- **No CLI composition.** Every byte of work passes through the agent's context window. No `jq`, no `wc -l`, no `parallel`, no SQL over a catalog.
- **Session-bounded.** No fleet operations, no CI, no multi-project reasoning.
- **Scales poorly with project size.** At 100+ modules the context-window ceiling is hit hard regardless of backend.
- **One-shot fix protocol is fragile.** No pre-apply validation; post-apply repairs have a tight budget.

These are properties of the **protocol**, not the connection. They persist whether PED is talking to a live project or (hypothetically) an offline one.

---

## New mxcli weaknesses introduced by the live backend

The refactor adds strategic reach but also adds real engineering concerns that didn't exist with `.mpr`-only operation. Surface these honestly — customers at scale will ask.

- **Two backends, two failure modes.** Live backend has to cope with Studio Pro caches, parallel user edits, ModelAPI lock contention, process restarts. Offline `.mpr` editing sidesteps these.
- **Catalog freshness against live edits.** If `catalog.db` was refreshed from the `.mpr` but the ModelAPI is making changes, the agent's queries can go stale mid-session. Cache invalidation against live state is a new correctness problem.
- **Semantic consistency across backends.** Per-statement atomicity on `.mpr` may not match the live-backend's atomicity guarantees. Customers who rely on offline semantics need to not be surprised when running live.
- **ModelAPI surface coverage.** Some operations may work against `.mpr` but not ModelAPI, or vice-versa. The "same script, any backend" promise depends on maintaining parity.

---

## Net-new capabilities the dual-backend unlocks

Worth naming these explicitly — they are strategic differentiators that neither PED nor previous-mxcli could claim.

- **Recordable live editing.** Developer uses Studio Pro naturally; mxcli records the MDL equivalent of each change. Session ends with a reviewable, replayable script. Permanently solves "how did this project get into this state?" — and it is impossible in PED (no artifact) and in Model SDK (no live hook).
- **Same-script, multi-target execution.** Author MDL once; run against `.mpr` in CI, against live ModelAPI during pair-modelling, against a cached copy for analysis. PED has one target by construction.
- **Bi-modal sessions.** Start live for exploration, switch to offline scripted mode for the bulk change, commit the script. Same tool, same DSL, no context switch.
- **Live dry-run.** `mxcli check` against the live backend previews consistency errors before the user sees them in Studio Pro.
- **Live lint and test inside Studio Pro.** Real-time CI-quality feedback during editing, not batch-after-save.

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

## At enterprise scale: a capability ceiling, not just a cost ceiling

Mendix customers exist with projects comprising **100+ modules, 1,000+ entities, 1,000+ microflows, and 500+ pages**. At this scale the thesis stops being about efficiency and becomes about what is mathematically possible inside a single agent session.

**Rough token sizes of a project at that scale:**

| Element | JSON tokens (est.) | MDL tokens (est.) | Compression |
|---|---|---|---|
| Entity, 10 attributes | 1–2k | 50–200 | 10–20× |
| Simple microflow (5 activities) | 3–5k | 300–1,000 | ~10× |
| Complex microflow (30 activities) | 15–30k | 1,500–5,000 | ~8× |
| Form page | 5–15k | 500–3,000 | ~5–10× |
| Grid page with complex layout | 20–50k | 2,000–8,000 | ~5–8× |
| Domain model, 50 entities | 50–150k | 10–30k | ~5× |

**Full-project corpus at this scale:**

- Full JSON: ~5–15M tokens.
- Full MDL: ~500k–1.5M tokens.
- Frontier context windows today: 200k–2M tokens.

Neither representation fits fully in context. **The project is physically unobservable as a whole from inside an agent session.** The question becomes: how does the agent know what to load without loading everything?

**PED has no external index.** The only way to learn about a microflow is to read it. At 1,000 microflows, questions like "what writes to `Customer.Email`?", "are naming conventions consistent across modules?", or "is this refactor safe?" cannot be answered because the answer requires information the session can't hold. This is a **capability ceiling** — not a cost problem an agent can work through with patience.

**mxcli has `catalog.db`.** The project's structure lives on disk. The agent issues SQL queries:

```sql
SELECT microflow FROM activities WHERE type='CommitObject' AND entity='Customer';
SELECT caller FROM microflow_calls WHERE callee='Sales.VAL_Customer';
```

Query cost is ~500 tokens in, ~200–2k out, regardless of project size. The agent reasons over answers, not over documents.

**Three concrete scenarios at enterprise scale:**

1. **Bootstrapping a 50-entity app from a spec.** MDL: one script, ~10–30k output + ~5k skill load, single shell call. PED: 50 × (schema + payload + response) ≈ 35–80k tokens I/O plus domain-model re-reads. Realistic: **50–120k per session**, and hits boundaries well before 200 entities.
2. **One edit in a 1,000-microflow project.** MDL: catalog query for conventions (~500 tokens) + MDL (~500) + check/exec (~500) = **~2k total**. PED: needs exemplars — 3 reference microflows × ~10k = 30k baseline, plus schemas and the edit = **40–60k** for a single microflow addition, because there is no cheap way to learn the project's conventions.
3. **"Add audit-log fields to all entities in a 1,000-entity portfolio."** MDL: agent writes a template + catalog query, produces a ~100k-token script **on disk** (zero agent context), sees pass/fail summary. **Agent footprint ~5–10k total.** PED: 1,000 update calls minimum = **500k–1M tokens**; impossible in one session; multi-session handoffs compound error rates. Essentially not doable as a single task.

**What an enterprise actually feels:** a platform team running daily governance sweeps across a 100-module portfolio spends real money on agent tokens — potentially millions per day. The delta is:

- **10–100× on API cost** (bounded vs. linear in project size).
- **Unbounded on latency** — many PED tasks simply do not complete at this scale.
- **Quality degradation** — attention thins at high context fill even when it technically fits. An mxcli agent has reasoning headroom an equivalent PED agent does not.

**The enterprise pitch is not "mxcli is cheaper." It is "mxcli can do things your current setup cannot do at all."**

**Caveats that need real investment to hold at this scale:**

- **Catalog coverage must be deep.** If `catalog.db` doesn't index microflow internals (activity types, attribute reads/writes, call edges, XPath constraints), the agent falls back to reading documents and the advantage collapses. Cataloguing is the highest-leverage backend investment.
- **MDL expressiveness for complex microflows.** If MDL for a 30-activity microflow is 20k tokens (close to JSON), the compression argument weakens. Dense, pattern-aware MDL (`FOREACH`, sub-flow references, reusable fragments) preserves the ratio.
- **Template-first generation.** At 1,000-entity scale, the agent must *generate a generator* (template + catalog query), not enumerate statements. This is a skill-file investment — teach the pattern explicitly.
- **Validation at volume.** 1,000 ALTER statements = 1,000 possible partial-failure points. `CREATE OR REPLACE` idempotence, per-statement safety, and rollback semantics matter more as N grows.

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
16. **`mxcli-mcp` wrapper.** Expose `mxcli check`, `mxcli exec`, `mxcli query`, `mxcli test` as MCP tools so mxcli drops into Claude Desktop / Cursor / Claude Code without shim. Neutralises PED's strongest remaining claim (MCP-native distribution) and is technically trivial.
17. **Recordable live editing as a headline feature.** The dual-backend enables live edits that *also* produce reviewable MDL scripts. Name it, brand it, demo it — nobody else has this. Turns every Studio Pro session into a git-reviewable change set.
18. **Live-backend catalog freshness.** Event-driven refresh (or a ModelAPI-backed live catalog) so queries don't go stale mid-session. This is the hardest technical investment but it's load-bearing for the composition story when the backend is live.
19. **Backend-semantics documentation.** Explicitly document atomicity, error handling, lock behaviour, rollback across `.mpr` and live backends. Avoid customers discovering semantic differences by incident.

---

## Where NOT to spend effort

- Duplicating PED's single-project interactive lane.
- Beginner/onboarding guidance.
- Rebuilding general Mendix knowledge the agent can get from any source.
- Feature parity for its own sake — test every feature proposal against *"Does this get better as the customer's app count goes from 1 to 20?"* If no, it's in PED's lane.

---

# Summary for slides

A focused six-slide deck. Slide 1 sets the axis — the dual-backend refactor means "live vs offline" is no longer what separates the tools. Slide 2 frames the territory. Slide 3 makes the efficiency case. Slide 4 is the architectural moat. Slide 5 is the safety upgrade. Slide 6 is the enterprise closer.

## Slide 1 — Interaction Protocol Is the Real Axis

**Thesis:** the dual-backend refactor collapses "live vs offline." Both PED and mxcli can now edit a running Studio Pro project interactively. What remains — and what is durable — is **how the agent interacts with the model**.

**The two protocols:**
- **PED:** MCP tool calls. Each op is a JSON request/response *inside* the agent's context. The LLM is the compute node.
- **mxcli:** CLI over a text DSL. The agent composes scripts and pipes. Compute happens in subprocesses *outside* context. The LLM is an orchestrator.

**Everything durable follows from this axis:**
- Text artifacts vs. transcripts.
- CLI composition vs. in-context reasoning.
- CI-native vs. session-bound.
- Catalog queries vs. document walks.
- Training-corpus compounding vs. runtime-only.

**Consequence:** mxcli now matches PED on every connection target *and* retains every composition, artifact, and scale advantage. The strategic position is strictly stronger after the refactor, not merely equivalent.

**New capability neither tool had before the refactor:** **recordable live editing** — Studio Pro edits that also produce a reviewable MDL script. Unique to mxcli; impossible in PED.

**Honest concession:** in an MCP client, for trivial ad-hoc single-element edits, PED is still lower-overhead until an `mxcli-mcp` wrapper exists. Concede the on-ramp; own everything beyond it.

---

## Slide 2 — Focus on the Portfolio, Not the Project

**Thesis:** PED is for a developer with one open project. mxcli/MDL is for a team with a portfolio. Don't duplicate; diverge.

**PED's lane (concede):**
- Single-project, interactive pair-modelling in an MCP client.
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

## Slide 3 — Context Efficiency Is a Capability Gap at Scale

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

## Slide 4 — CLI Composition Is the Real Moat

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

## Slide 5 — Testing Is What Makes Agentic Mendix Engineering Safe at Scale

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

---

## Slide 6 — The Capability Ceiling at Enterprise Scale

**Thesis:** at 100+ modules / 1,000+ entities / 1,000+ microflows, the conversation stops being about efficiency. PED hits a **mathematical wall** — some questions cannot be answered from inside an agent session at all. mxcli's catalog + CLI composition are the only way through.

**The scale reality:**
- Full project in JSON: ~5–15M tokens. Full project in MDL: ~500k–1.5M. Frontier context windows: 200k–2M. **Neither fits.** The project is unobservable as a whole.
- PED has no external index — the only way to learn about a microflow is to read it. At 1,000 microflows, this is structural blindness to the whole.
- mxcli has `catalog.db`. Agents issue SQL queries; cost is constant regardless of project size. The project lives on disk; the agent reasons over answers.

**Three scenarios that clarify the gap:**
- **Bootstrap 50 entities:** MDL ~15–35k tokens. PED: 50–120k per session, hits boundaries before ~200 entities.
- **One edit in a 1,000-microflow app:** MDL ~2k tokens (catalog query locates conventions). PED: 40–60k, because exemplars must be read to learn what "idiomatic" looks like.
- **Audit-log pattern across 1,000 entities:** MDL ~5–10k agent tokens (template + catalog → 100k script on disk → summary). PED: 500k–1M tokens minimum; not doable in one session.

**What the enterprise feels:**
- **10–100× on API cost** — bounded vs. linear in project size.
- **Unbounded latency** — many PED tasks simply don't complete at this scale.
- **Quality degradation** even when context technically fits — attention thins at high fill.

**The pitch that closes the deal:** *"mxcli can do things your current setup cannot do at all."* Not cheaper — **possible.**

**What must hold for this to land:**
- Deep `catalog.db` coverage (activity types, attribute reads/writes, call edges, XPath).
- MDL expressiveness that preserves compression for complex microflows.
- Template-first generation patterns taught as first-class skills.
- Idempotent, partial-failure-safe execution for volume operations.

**Target buyer for this slide:** platform engineering lead at a bank, insurer, telco, or large government — the person whose nightmare is "how do I maintain consistency across 40 apps built by 80 developers over 6 years." This is the slide they remember.
