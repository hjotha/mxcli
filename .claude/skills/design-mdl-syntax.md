# Design New MDL Syntax

This skill provides guardrails for designing new MDL statements. Read this **before** writing grammar rules, AST types, or executor code for any new MDL feature.

## When to Use This Skill

- Adding a new document type to MDL (e.g., scheduled events, message definitions, REST services)
- Adding a new action type to microflows (e.g., new activity, new operation)
- Extending existing syntax with new clauses or keywords
- Reviewing a PR that adds or modifies MDL syntax
- Resolving a syntax design disagreement

## Core Principles (Priority Order)

When principles conflict, higher-priority ones win.

### 1. Read Like English

MDL targets citizen developers and business analysts, not software engineers. Statements should read as natural English sentences.

- Use keyword words (`FROM`, `WHERE`, `IN`), not symbols (`->`, `|>`, `=>`)
- Spell out full words (`MICROFLOW`, `ASSOCIATION`), not abbreviations (`MF`, `ASSOC`)
- Use prepositions to clarify relationships: `GRANT READ ON Entity TO Role`

**Test**: Read it aloud. A business analyst should understand on first hearing.

### 2. One Way to Do Each Thing

Reuse existing patterns. Never create a second syntax for the same concept.

| Operation | Pattern | Example |
|-----------|---------|---------|
| Create | `CREATE [MODIFIERS] <TYPE> Module.Name (...)` | `CREATE PERSISTENT ENTITY Shop.Product (...)` |
| Modify | `ALTER <TYPE> Module.Name <OPERATION>` | `ALTER ENTITY Shop.Product ADD (...)` |
| Remove | `DROP <TYPE> Module.Name` | `DROP ENTITY Shop.Product` |
| List | `SHOW <TYPE>S [IN Module]` | `SHOW ENTITIES IN Shop` |
| Inspect | `DESCRIBE <TYPE> Module.Name` | `DESCRIBE ENTITY Shop.Product` |
| Security | `GRANT/REVOKE <perm> ON <target> TO/FROM <role>` | `GRANT READ ON Shop.Product TO Shop.User` |

Do NOT use alternative verbs: `ADD` instead of `CREATE`, `REMOVE` instead of `DROP`, `LIST` instead of `SHOW`, `VIEW` instead of `DESCRIBE`.

### 3. Optimize for LLMs

- Keep patterns regular so one example is sufficient for generation
- Statements must be self-contained (no implicit state from prior statements)
- Use consistent keyword order: `<VERB> [MODIFIERS] <TYPE> <NAME> [CLAUSES] [BODY]`
- Prefer flat statement sequences over deeply nested structures

### 4. Make Diffs Reviewable

- One property per line in multi-property constructs
- Allow trailing commas
- `DESCRIBE` output uses deterministic property order
- Default values omitted unless non-obvious

### 5. Token Efficiency (Without Sacrificing Clarity)

- Omit noise words: `CREATE ENTITY` not `CREATE A NEW ENTITY`
- Support `OR MODIFY` to avoid check-then-create
- Allow type inference for obvious cases: `DECLARE $Count = 0`
- Do NOT use symbols to save tokens at the cost of readability

## Design Workflow

Follow these steps when designing syntax for a new MDL feature.

### Step 1: Check Existing Patterns

Read the MDL Quick Reference: `docs/01-project/MDL_QUICK_REFERENCE.md`

Does an existing pattern cover this? If yes, extend it. Don't invent new syntax.

```
New feature: "image collections"
Existing pattern: CREATE/ALTER/DROP/SHOW/DESCRIBE
Design: CREATE IMAGE COLLECTION Module.Name (...)
        DESCRIBE IMAGE COLLECTION Module.Name
        SHOW IMAGE COLLECTIONS [IN Module]
```

### Step 2: Pick the Statement Shape

Every MDL statement fits one of these shapes:

```
DDL:   <VERB> [MODIFIERS] <TYPE> <QualifiedName> [CLAUSES] [BODY];
DML:   <ACTION> <TARGET> [CLAUSES];
DQL:   <QUERY-VERB> <TYPE>S [FILTERS];
```

If your feature doesn't fit any shape, it may belong as a CLI command (`mxcli <subcommand>`) rather than MDL syntax.

### Step 3: Choose Keywords

1. Reuse existing keywords first (check reserved words in grammar)
2. Use SQL/DDL verbs: `CREATE`, `ALTER`, `DROP`, `SHOW`, `DESCRIBE`, `GRANT`, `REVOKE`, `SET`
3. Use Mendix terminology: `ENTITY` not `TABLE`, `MICROFLOW` not `FUNCTION`, `PAGE` not `VIEW`
4. Prepositions clarify structure: `FROM`, `TO`, `IN`, `ON`, `BY`, `WITH`, `AS`, `WHERE`, `INTO`

### Step 4: Write the Property List

All property-bearing constructs use this format:

```mdl
CREATE <TYPE> Module.Name (
    Property1: value,
    Property2: value,
);
```

Rules:
- Parentheses `()` delimit property lists
- Colon `:` separates key from value
- Comma `,` separates properties
- Trailing comma allowed
- One property per line (single line acceptable for 1-2 properties)

#### Colon `:` vs `AS` — When to Use Each

Use **colon** for property definitions (assigning a value to a named property):

```mdl
CREATE ENTITY Shop.Product (
    Name: String(200),          -- property: type/value
    Price: Decimal,
);
TEXTBOX txtName (Label: 'Name', Attribute: Title)
```

Use **`AS`** for name-to-name mappings (renaming, aliasing, mapping one name to another):

```mdl
CUSTOM NAME MAP (
    'kvkNummer' AS 'ChamberOfCommerceNumber',   -- old name AS new name
    'naam' AS 'CompanyName',
)
ALTER ENTITY Shop.Product RENAME Code AS ProductCode   -- old attr AS new attr
```

**Rule of thumb**: if the left side is a *fixed property key* defined by the syntax, use `:`. If the left side is a *user-provided name* being mapped to another name, use `AS`.

### Step 5: Validate

Run these checks before finalizing syntax design:

1. **Read aloud test** — Does it read as English? Can a business analyst understand it?
2. **LLM generation test** — Give one example to an LLM, ask for a variant. Does it get it right?
3. **Diff test** — Change one property. Is the diff exactly one line?
4. **Pattern test** — Does it follow CREATE/ALTER/DROP/SHOW/DESCRIBE? If not, why?
5. **Roundtrip test** — Can `DESCRIBE` output be fed back as input?

## Anti-Patterns (DO NOT)

### Custom Verbs for Standard Operations

```mdl
-- WRONG: custom verb
SCHEDULE EVENT Shop.Cleanup ...
REGISTER WEBHOOK Shop.OnOrder ...

-- RIGHT: standard CREATE
CREATE SCHEDULED EVENT Shop.Cleanup (...)
CREATE WEBHOOK Shop.OnOrder (...)
```

### Implicit Module Context

```mdl
-- WRONG: implicit state
USE MODULE Shop;
CREATE ENTITY Customer (...);

-- RIGHT: explicit qualified name
CREATE ENTITY Shop.Customer (...);
```

### Symbolic Syntax

```mdl
-- WRONG: requires learning symbol meanings
$items |> filter($.active) |> map($.name)

-- RIGHT: keyword-based
FILTER $Items WHERE Active = true
```

### Positional Arguments

```mdl
-- WRONG: meaning unclear without docs
CREATE RULE Shop Process Order ACT_ProcessOrder

-- RIGHT: labeled properties
CREATE RULE Shop.ProcessOrder (
    Type: Validation,
    Microflow: Shop.ACT_ProcessOrder,
);
```

### Keyword Overloading

```mdl
-- CAUTION: SET already means variable assignment in microflows
-- Don't reuse it to mean property modification elsewhere unless established
```

## Checklist

Before merging any PR that adds new MDL syntax, verify:

- [ ] Follows `CREATE`/`ALTER`/`DROP`/`SHOW`/`DESCRIBE` pattern
- [ ] Uses `Module.Element` qualified names (no bare names)
- [ ] Property lists use `( Key: value, ... )` format
- [ ] Keywords are full English words (no abbreviations)
- [ ] Statement reads as English (aloud test passed)
- [ ] One example sufficient for LLM generation
- [ ] Small change = one-line diff
- [ ] No new keyword overloading
- [ ] No implicit context dependency
- [ ] `DESCRIBE` roundtrips to valid MDL
- [ ] Grammar regenerated (`make grammar`)
- [ ] Quick reference updated (`docs/01-project/MDL_QUICK_REFERENCE.md`)
- [ ] Full-stack wired: grammar, AST, visitor, executor, DESCRIBE

## Related Resources

- Full design rationale: `docs/11-proposals/PROPOSAL_mdl_syntax_design_guidelines.md`
- MDL Quick Reference: `docs/01-project/MDL_QUICK_REFERENCE.md`
- Implementation workflow: `.claude/skills/implement-mdl-feature.md`
- Existing syntax proposals: `docs/11-proposals/PROPOSAL_mdl_syntax_improvements.md`
- Grammar file: `mdl/grammar/MDLParser.g4`
