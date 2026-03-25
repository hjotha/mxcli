# Proposal: Unified mxcli & MDL Documentation Site

## Problem

mxcli has extensive documentation — 119 markdown files, a complete language specification, architecture docs, examples, and user guides — but it's scattered across 14 directories with no unified structure. There's no way for a user to go from "what is mxcli?" to "how do I CREATE ENTITY with constraints?" in a logical progression. The docs can't be put online or bundled as a single PDF.

**What we need**: A PostgreSQL-style documentation site that works as:
1. A browsable website (GitHub Pages or custom domain)
2. A single downloadable PDF
3. High-quality LLM training material (plain text, structured, example-rich)

## Design Principles (Learned from PostgreSQL)

PostgreSQL's docs work because they separate concerns:

| Layer | Purpose | PostgreSQL Example | mxcli Equivalent |
|-------|---------|-------------------|-------------------|
| **Tutorial** | Learn by doing | Part I: Tutorial (3 chapters) | Getting started, first project |
| **Conceptual** | Understand features | Part II: SQL Language (11 chapters) | MDL language guide, Mendix concepts |
| **Reference** | Look up syntax | Part VI: SQL Commands (one page per statement) | MDL statement reference |
| **Administration** | Run and configure | Part III: Server Administration | Installation, CLI usage, Docker, LSP |
| **Internals** | Understand implementation | Part VII: Internals | Architecture, BSON format, parser |

Key insight: **never mix tutorial prose with reference syntax**. The conceptual chapters in Part II teach *how* to use CREATE TABLE. The reference page for CREATE TABLE gives the exact syntax, parameters, and edge cases. Both exist, both link to each other.

---

## Proposed Structure

```
mxcli Documentation
│
├── Preface
│   ├── What is mxcli?
│   ├── What is MDL?
│   ├── Mendix Concepts for Newcomers
│   └── Document Conventions
│
├── Part I: Tutorial
│   ├── 1. Setting Up
│   │   ├── Installation (binary, go install, dev container)
│   │   ├── Opening Your First Project
│   │   └── The REPL
│   ├── 2. Exploring a Project
│   │   ├── SHOW MODULES, SHOW ENTITIES
│   │   ├── DESCRIBE, SEARCH
│   │   └── SHOW STRUCTURE
│   ├── 3. Your First Changes
│   │   ├── Creating an Entity
│   │   ├── Creating a Microflow
│   │   ├── Creating a Page
│   │   └── Validating with mxcli check
│   └── 4. Working with AI Assistants
│       ├── Claude Code Integration
│       ├── Cursor / Continue.dev / Windsurf
│       ├── Skills and CLAUDE.md
│       └── The MDL + AI Workflow
│
├── Part II: The MDL Language
│   ├── 5. MDL Basics
│   │   ├── Lexical Structure (keywords, identifiers, literals)
│   │   ├── Qualified Names (Module.Name)
│   │   ├── Comments and Documentation (/** */)
│   │   └── Script Files (.mdl)
│   ├── 6. Data Types
│   │   ├── Primitive Types (String, Integer, Long, Decimal, Boolean, DateTime, ...)
│   │   ├── Constraints (NOT NULL, DEFAULT, UNIQUE)
│   │   ├── Enumerations
│   │   └── Type Mapping (MDL → Mendix → Database)
│   ├── 7. Domain Model
│   │   ├── Entities (persistent, non-persistent, external, view)
│   │   ├── Attributes and Validation Rules
│   │   ├── Associations (Reference, ReferenceSet, ownership, delete behavior)
│   │   ├── Generalization (EXTENDS)
│   │   ├── Indexes
│   │   └── ALTER ENTITY
│   ├── 8. Microflows and Nanoflows
│   │   ├── Structure (parameters, variables, activities, return)
│   │   ├── Activity Types (retrieve, create, change, commit, delete, ...)
│   │   ├── Control Flow (IF/ELSE, LOOP, error handling)
│   │   ├── Expressions
│   │   ├── Nanoflows vs Microflows
│   │   └── Common Patterns (CRUD, validation, batch processing)
│   ├── 9. Pages
│   │   ├── Page Structure (layout, content, data source)
│   │   ├── Widget Types (DataView, DataGrid, Container, TextBox, Button, ...)
│   │   ├── Data Binding (-> operator)
│   │   ├── Snippets
│   │   ├── ALTER PAGE / ALTER SNIPPET
│   │   └── Common Patterns (list page, edit page, master-detail)
│   ├── 10. Security
│   │   ├── Module Roles and User Roles
│   │   ├── Entity Access (CREATE, READ, WRITE, DELETE)
│   │   ├── Microflow, Page, and Nanoflow Access
│   │   ├── GRANT / REVOKE
│   │   └── Demo Users
│   ├── 11. Navigation and Settings
│   │   ├── Navigation Profiles
│   │   ├── Home Pages and Menus
│   │   └── Project Settings
│   ├── 12. Workflows
│   │   ├── Workflow Structure
│   │   ├── Activity Types (user tasks, decisions, parallel splits, ...)
│   │   └── Workflow vs Microflow
│   └── 13. Business Events
│       ├── Event Services
│       └── Publishing and Consuming Events
│
├── Part III: Project Tools
│   ├── 14. Code Navigation
│   │   ├── SHOW CALLERS / CALLEES
│   │   ├── SHOW REFERENCES / IMPACT
│   │   ├── SHOW CONTEXT
│   │   └── Full-Text Search (SEARCH)
│   ├── 15. Catalog Queries
│   │   ├── REFRESH CATALOG
│   │   ├── Available Tables (modules, entities, microflows, pages, ...)
│   │   ├── SQL Queries (SELECT FROM CATALOG.*)
│   │   └── Use Cases (impact analysis, unused elements, complexity metrics)
│   ├── 16. Linting and Reports
│   │   ├── Built-in Rules (14 Go rules)
│   │   ├── Starlark Rules (27 extensible rules)
│   │   ├── Writing Custom Rules
│   │   ├── mxcli lint (JSON, SARIF output)
│   │   └── mxcli report (scored best practices)
│   ├── 17. Testing
│   │   ├── .test.mdl and .test.md Formats
│   │   ├── Test Annotations (@test, @expect)
│   │   ├── Running Tests (mxcli test, Docker requirement)
│   │   └── Diff (mxcli diff, mxcli diff-local)
│   ├── 18. External SQL
│   │   ├── SQL CONNECT (PostgreSQL, Oracle, SQL Server)
│   │   ├── Querying External Databases
│   │   ├── IMPORT FROM ... INTO ... MAP
│   │   ├── Credential Management
│   │   └── Database Connector Generation
│   └── 19. Docker Integration
│       ├── mxcli docker build (PAD)
│       ├── mxcli docker run
│       ├── OQL Queries (mxcli oql)
│       └── Dev Container Setup
│
├── Part IV: IDE Integration
│   ├── 20. VS Code Extension
│   │   ├── Installation
│   │   ├── Syntax Highlighting and Diagnostics
│   │   ├── Completion, Hover, Go-to-Definition
│   │   ├── Project Tree
│   │   └── Context Menu Commands
│   ├── 21. LSP Server
│   │   ├── Protocol (stdio)
│   │   ├── Capabilities
│   │   └── Integration with Other Editors
│   └── 22. mxcli init
│       ├── What Gets Created (.claude/, .devcontainer/, skills)
│       ├── Customizing Skills
│       └── Syncing with Updates
│
├── Part V: Go Library
│   ├── 23. Quick Start
│   │   ├── Installation (go get)
│   │   ├── Reading a Project
│   │   └── Modifying a Project
│   ├── 24. Public API (modelsdk.go)
│   │   ├── Open / OpenForWriting
│   │   ├── Reader Methods
│   │   └── Writer Methods
│   └── 25. Fluent API (api/)
│       ├── ModelAPI Entry Point
│       ├── EntityBuilder, MicroflowBuilder, PageBuilder
│       └── Examples
│
├── Part VI: MDL Statement Reference
│   │
│   │   (One page per statement, PostgreSQL-style:
│   │    Synopsis → Description → Parameters → Notes → Examples → See Also)
│   │
│   ├── Connection Statements
│   │   ├── OPEN PROJECT
│   │   └── CLOSE PROJECT
│   ├── Query Statements
│   │   ├── SHOW MODULES
│   │   ├── SHOW ENTITIES
│   │   ├── SHOW MICROFLOWS / NANOFLOWS
│   │   ├── SHOW PAGES / SNIPPETS
│   │   ├── SHOW ENUMERATIONS
│   │   ├── SHOW ASSOCIATIONS
│   │   ├── SHOW CONSTANTS
│   │   ├── SHOW WORKFLOWS
│   │   ├── SHOW BUSINESS EVENTS
│   │   ├── SHOW STRUCTURE
│   │   ├── SHOW WIDGETS
│   │   ├── DESCRIBE ENTITY
│   │   ├── DESCRIBE MICROFLOW / NANOFLOW
│   │   ├── DESCRIBE PAGE / SNIPPET
│   │   ├── DESCRIBE ENUMERATION
│   │   ├── DESCRIBE ASSOCIATION
│   │   └── SEARCH
│   ├── Domain Model Statements
│   │   ├── CREATE ENTITY
│   │   ├── ALTER ENTITY
│   │   ├── DROP ENTITY
│   │   ├── CREATE ENUMERATION
│   │   ├── DROP ENUMERATION
│   │   ├── CREATE ASSOCIATION
│   │   ├── DROP ASSOCIATION
│   │   └── CREATE CONSTANT
│   ├── Microflow Statements
│   │   ├── CREATE MICROFLOW
│   │   ├── CREATE NANOFLOW
│   │   ├── DROP MICROFLOW / NANOFLOW
│   │   └── CREATE JAVA ACTION
│   ├── Page Statements
│   │   ├── CREATE PAGE
│   │   ├── CREATE SNIPPET
│   │   ├── ALTER PAGE / ALTER SNIPPET
│   │   ├── DROP PAGE / SNIPPET
│   │   └── CREATE LAYOUT
│   ├── Security Statements
│   │   ├── CREATE MODULE ROLE
│   │   ├── CREATE USER ROLE
│   │   ├── GRANT
│   │   ├── REVOKE
│   │   └── CREATE DEMO USER
│   ├── Navigation Statements
│   │   ├── ALTER NAVIGATION
│   │   └── SHOW NAVIGATION
│   ├── Workflow Statements
│   │   ├── CREATE WORKFLOW
│   │   └── DROP WORKFLOW
│   ├── Business Event Statements
│   │   ├── CREATE BUSINESS EVENT SERVICE
│   │   └── DROP BUSINESS EVENT SERVICE
│   ├── Catalog Statements
│   │   ├── REFRESH CATALOG
│   │   ├── SELECT FROM CATALOG
│   │   ├── SHOW CALLERS / CALLEES
│   │   ├── SHOW REFERENCES / IMPACT / CONTEXT
│   │   └── SHOW CATALOG TABLES
│   ├── External SQL Statements
│   │   ├── SQL CONNECT
│   │   ├── SQL DISCONNECT
│   │   ├── SQL (query)
│   │   ├── SQL GENERATE CONNECTOR
│   │   └── IMPORT FROM
│   ├── Settings Statements
│   │   ├── SHOW SETTINGS
│   │   └── ALTER SETTINGS
│   ├── Organization Statements
│   │   ├── CREATE MODULE
│   │   ├── CREATE FOLDER
│   │   └── MOVE
│   └── Session Statements
│       ├── SET
│       └── SHOW STATUS
│
├── Part VII: Architecture & Internals
│   ├── 26. System Architecture
│   │   ├── Layer Diagram (ASCII + Mermaid)
│   │   ├── Package Structure
│   │   └── Design Decisions
│   ├── 27. MPR File Format
│   │   ├── v1 (SQLite) vs v2 (mprcontents/)
│   │   ├── BSON Document Structure
│   │   ├── Storage Names vs Qualified Names
│   │   └── Widget Template System
│   ├── 28. MDL Parser
│   │   ├── ANTLR4 Grammar Design
│   │   ├── Lexer → Parser → AST → Executor Pipeline
│   │   └── Adding New Statements
│   └── 29. Catalog System
│       ├── SQLite Schema
│       ├── FTS5 Full-Text Search
│       └── Reference Tracking
│
├── Part VIII: Appendixes
│   ├── A. MDL Quick Reference (cheat sheet)
│   ├── B. Data Type Mapping Table
│   ├── C. Reserved Words
│   ├── D. Mendix Version Compatibility
│   ├── E. Common Mistakes and Anti-Patterns
│   ├── F. Error Messages Reference
│   ├── G. Glossary (Mendix terms for non-Mendix developers)
│   ├── H. TypeScript SDK Equivalence
│   └── I. Changelog
│
└── Index
```

---

## Per-Statement Reference Format

Every statement in Part VI follows the same template (matching PostgreSQL):

```markdown
# CREATE ENTITY

## Synopsis

    CREATE [OR MODIFY] [PERSISTENT | NON-PERSISTENT] ENTITY module.name
        [EXTENDS parent.entity]
    (
        attr_name: data_type [NOT NULL] [DEFAULT value] [UNIQUE],
        ...
    );

## Description

Creates a new entity in the specified module's domain model. Entities
are the data objects in a Mendix application, similar to database tables.

## Parameters

**PERSISTENT | NON-PERSISTENT**
: Persistent entities are stored in the database. Non-persistent entities
  exist only in memory during a session. Default: PERSISTENT.

**EXTENDS parent.entity**
: Creates a generalization (inheritance) relationship. The new entity
  inherits all attributes from the parent. Must appear before the
  opening parenthesis.

**attr_name: data_type**
: Defines an attribute. See Data Types for available types.

## Notes

- EXTENDS must appear before `(`, not after `)` — this is a common mistake.
- String attributes require an explicit length: `String(200)`, not `String`.
- Use `OR MODIFY` for idempotent scripts that may be re-run.

## Examples

### Basic entity

    CREATE PERSISTENT ENTITY Sales.Customer (
        Name: String(200) NOT NULL,
        Email: String(200),
        IsActive: Boolean DEFAULT true
    );

### Entity with generalization

    CREATE PERSISTENT ENTITY Sales.VIPCustomer EXTENDS Sales.Customer (
        DiscountPercentage: Decimal,
        LoyaltyTier: String(50) DEFAULT 'Silver'
    );

### Idempotent creation

    CREATE OR MODIFY PERSISTENT ENTITY Sales.Customer (
        Name: String(200) NOT NULL,
        Email: String(200),
        Phone: String(50)
    );

## See Also

ALTER ENTITY, DROP ENTITY, CREATE ASSOCIATION, DESCRIBE ENTITY
```

---

## Content Sourcing

Most content already exists. The work is reorganization and gap-filling, not writing from scratch.

| Part | Source | Work Required |
|------|--------|---------------|
| Preface | README.md, mxcli-overview.md | Rewrite as introductory chapters |
| I: Tutorial | New | Write from scratch (~4 chapters). Most important new content. |
| II: MDL Language | 01-language-reference.md, 02-data-types.md, 03-domain-model.md, skill files | Reorganize into chapters, add examples. ~60% exists. |
| III: Project Tools | MDL_QUICK_REFERENCE.md, mxcli-overview.md, CLAUDE.md | Reorganize and expand. ~70% exists. |
| IV: IDE Integration | README.md, vscode-mdl/package.json | Partially exists, needs expansion. ~40% exists. |
| V: Go Library | GO_LIBRARY.md, api/api.go | Exists, needs minor updates. ~80% exists. |
| VI: Statement Reference | 01-language-reference.md, MDL_QUICK_REFERENCE.md | **Major work**: split into per-statement pages, add examples to each. ~30% exists (syntax yes, examples/notes no). |
| VII: Internals | ARCHITECTURE.md, MDL_PARSER_ARCHITECTURE.md, PAGE_BSON_SERIALIZATION.md | Exists, minor reorganization. ~90% exists. |
| VIII: Appendixes | MDL_QUICK_REFERENCE.md, 02-data-types.md, CLAUDE.md, SDK_EQUIVALENCE.md | Mostly exists, needs formatting. ~75% exists. |

**Estimated new writing**: ~40% of total content. Heaviest in Tutorial (Part I) and Statement Reference (Part VI).

---

## Tooling

### Option A: mdBook (Recommended)

[mdBook](https://rust-lang.github.io/mdBook/) is a Rust tool that builds books from Markdown. Used by the Rust Programming Language book.

**Pros**:
- Markdown source (matches existing docs)
- Built-in search (client-side, no server)
- Single-binary, fast builds
- PDF export via `mdbook-pdf` plugin (uses Chrome headless)
- Clean, readable theme
- GitHub Pages deployment via GitHub Actions
- TOC sidebar with collapsible sections
- Previous/Next navigation
- `SUMMARY.md` defines structure (similar to PostgreSQL's hierarchy)

**Cons**:
- Less customizable than full static site generators
- PDF quality depends on Chrome rendering (good enough, not typeset quality)

**Example `SUMMARY.md`**:
```markdown
# Summary

[Preface](preface.md)

# Part I: Tutorial

- [Setting Up](tutorial/setup.md)
- [Exploring a Project](tutorial/exploring.md)
- [Your First Changes](tutorial/first-changes.md)
- [Working with AI Assistants](tutorial/ai-assistants.md)

# Part II: The MDL Language

- [MDL Basics](language/basics.md)
- [Data Types](language/data-types.md)
- [Domain Model](language/domain-model.md)
  ...
```

### Option B: Docusaurus

Facebook's documentation framework (React-based).

**Pros**:
- Versioned docs (useful for Mendix version-specific content)
- Rich plugin ecosystem
- Better PDF via `docusaurus-pdf` or Typst pipeline
- Blog feature (for release announcements)
- Algolia DocSearch integration

**Cons**:
- Node.js dependency (heavier build)
- More configuration overhead
- Overkill for current scope

### Option C: Typst + Custom Pipeline

Already used for `mxcli-overview.typ`. Typst produces high-quality PDFs.

**Pros**:
- Best PDF quality (proper typesetting)
- Already familiar (mxcli-overview.typ exists)
- Programmable (variables, includes, templates)

**Cons**:
- No built-in web output (need separate HTML generation)
- Would need a dual pipeline: Typst for PDF, something else for web
- Smaller ecosystem than mdBook/Docusaurus

### Recommendation: mdBook + Typst Hybrid

- **mdBook** for the website (GitHub Pages) and search
- **Typst** for the PDF (high-quality typeset output)
- **Shared Markdown source** — mdBook reads Markdown natively; a build script converts to Typst
- **GitHub Actions** deploys both on merge to main

```
docs-site/
├── book.toml              # mdBook config
├── SUMMARY.md             # Table of contents
├── src/                   # Markdown source (shared)
│   ├── preface.md
│   ├── tutorial/
│   ├── language/
│   ├── tools/
│   ├── reference/         # Per-statement pages
│   ├── internals/
│   └── appendixes/
├── typst/
│   ├── main.typ           # Typst entry point (includes from src/)
│   └── template.typ       # PDF styling
└── .github/workflows/
    └── docs.yml           # Build + deploy both formats
```

---

## Hosting

### GitHub Pages (Recommended for Start)

- Free, automatic HTTPS
- Deploy via `gh-pages` branch or GitHub Actions
- URL: `mendixlabs.github.io/mxcli/` or custom domain
- No server to maintain

### Custom Domain (Later)

- `docs.mxcli.dev` or `mxcli.mendixlabs.com`
- CNAME record pointing to GitHub Pages
- Configure in repository settings

---

## Implementation Plan

### Phase 1: Skeleton and Tutorial (2-3 days)

1. Set up `docs-site/` with mdBook config
2. Write `SUMMARY.md` with full structure
3. Write Part I: Tutorial (4 chapters) — **this is the critical new content**
4. Stub all other parts with single-line descriptions
5. Deploy to GitHub Pages
6. Verify navigation, search, and linking work

### Phase 2: Reorganize Existing Content (2-3 days)

1. Move language reference content into Part II chapters (split by topic)
2. Move quick reference + CLI features into Part III chapters
3. Move architecture docs into Part VII
4. Move data type tables, reserved words, etc. into Part VIII appendixes
5. Add cross-links between conceptual chapters and statement reference

### Phase 3: Statement Reference (3-5 days)

1. Create per-statement pages for all MDL statements (~50-60 pages)
2. Each page: Synopsis, Description, Parameters, Notes, Examples, See Also
3. Start with the 10 most-used statements, then fill in the rest
4. Extract examples from mdl-examples/ and skill files

### Phase 4: PDF Pipeline (1 day)

1. Set up Typst template matching mdBook structure
2. Build script to generate Typst from Markdown source
3. GitHub Actions workflow to produce PDF on release
4. Add download link to the website

### Phase 5: Polish (1-2 days)

1. Review all cross-links
2. Add glossary (Appendix G) — essential for non-Mendix developers
3. Write Preface (what is mxcli, what is MDL, Mendix concepts)
4. Test PDF output, fix formatting issues
5. Add version selector if multiple Mendix versions need documenting

---

## What Happens to Existing Docs

| Current Location | Fate |
|-----------------|------|
| `docs/05-mdl-specification/` | Content moves to Parts II + VI. Original files become redirects or are removed. |
| `docs/01-project/ARCHITECTURE.md` | Moves to Part VII Chapter 26. |
| `docs/01-project/MDL_QUICK_REFERENCE.md` | Becomes Appendix A. |
| `docs/10-user-docs/mxcli-overview.md` | Content distributed across Parts I, III, IV. Original kept as standalone marketing doc. |
| `docs/GO_LIBRARY.md` | Moves to Part V. |
| `docs/03-development/` | Technical docs move to Part VII. |
| `README.md` | Stays as repo README, links to docs site. |
| `CLAUDE.md` | Stays as AI agent config, unchanged. |
| `docs/11-proposals/` | Stays as-is (internal, not part of user docs). |
| `docs/12-bug-reports/`, `docs/13-vision/`, `docs/14-eval/` | Stay as-is (internal). |
| `mdl-examples/` | Examples are referenced/included from Part VI statement pages. |

---

## Success Criteria

- A new user can go from "what is mxcli?" to running their first MDL command in under 10 minutes using the Tutorial
- Every MDL statement has a dedicated reference page with synopsis, parameters, and at least 2 examples
- The site is searchable (client-side search, no external service required)
- PDF is downloadable and includes all content with proper page numbers and TOC
- Deployed automatically on merge to main via GitHub Actions
- Non-Mendix developers can understand the docs (glossary, Mendix concepts preface)
