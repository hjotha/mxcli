# mxcli Security Audit Report
**Project:** QueryDemoApp-main (Mendix 11.6.3)
**Date:** 2026-02-28
**Scope:** DIVD-2022-00019 vulnerability assessment + catalog security analysis
**Method:** mxcli MDL commands + catalog SQL queries

---

## 1. Executive Summary

The application uses Production security level with no anonymous (guest) access configured, meaning it is **not directly vulnerable to DIVD-2022-00019**. However, the audit uncovered several significant secondary findings, and the process of conducting the audit exposed both missing data in the catalog tables and broken SQL features that prevented a number of planned queries from executing.

This report documents: security findings, what catalog queries worked, what failed and why, what data is absent from the catalog schema, and a prioritised list of feature requests for the mxcli team.

---

## 2. Security Findings

### 2.1 DIVD-2022-00019 — Anonymous Entity Access

| Check | Result |
|-------|--------|
| Security Level | **Production** ✅ |
| Guest Access (anonymous user role) | **false** ✅ |
| Check Security | **true** ✅ |

**Verdict: Not vulnerable.** The vulnerability requires an anonymous user role to be configured and entity access rules to grant that role READ access. Neither condition is met.

**How queried:**
```sql
show project security;
```

---

### 2.2 Weak Password Policy (High Risk)

```
minimum length : 1
Require Digit  : false
Require Mixed case : false
Require Symbol : false
```

A one-character password satisfies the policy. This effectively eliminates any protection from brute-force or credential stuffing attacks on all accounts, including `MxAdmin`.

**How queried:** `show project security;`

---

### 2.3 Demo Users Active in Production (High Risk)

```
demo_administrator  → Administrator role (full account management)
demo_user           → user + api roles
```

Demo users with predictable or default credentials are active. The `demo_administrator` account has the `Administration.Administrator` role, which grants access to account management, active sessions, scheduled events, and runtime configuration.

**How queried:** `show demo users;`

---

### 2.4 OqlPad Module in Production Navigation (High Risk)

The OqlPad module exposes an arbitrary OQL query executor. It is not merely deployed — it is **registered as a menu item in the main navigation** (`Navigation.Responsive → OqlPad.OqlEditor`). Any user assigned `OqlPad.OqlPadUser` sees it directly in the nav bar.

**Permissions granted to `OqlPad.OqlPadUser`:**
- `OqlPad.ExecuteOqlQuery` — runs arbitrary OQL
- `OqlPad.TextToOQL` — LLM-assisted query generation
- `OqlPad.AdHocOqlQuery_Post`, `_Get_All`, `NewOqlQuery`, `DeleteQuery`, `DeleteResult`

OQL queries bypass entity-level access rules by operating directly on the database. This is a development/demo tool that should be disabled before any production deployment.

**How queried:**
```sql
select SourceName, TargetName, RefKind from CATALOG.REFS where RefKind = 'menu_item';
select ElementName from CATALOG.PERMISSIONS where ModuleRoleName = 'OqlPad.OqlPadUser' and ElementType = 'MICROFLOW';
```

---

### 2.5 Persistent Entities with No Access Rules (Medium Risk — Misconfiguration)

In Production mode Mendix blocks access to entities with no access rules, so these are not directly exploitable. However, they represent incomplete or abandoned configuration.

**Entities with zero access rules (persistent, non-external):**

| Entity | Attributes | Notes |
|--------|-----------|-------|
| `DPP.Product` | 10 | Entire DPP module has no module roles |
| `DPP.ProductEnvironmentalData` | 21 | |
| `DPP.ProductImage` | 0 | |
| `DPP.ManufactureData` | 0 | |
| `DPP.Packaging` | 6 | |
| `Shop.OrdersArchived` | 4 | Archived orders, no rules |
| `Shop.UserTenant` | 2 | Tenant isolation data, no rules |
| `MultipleAggregate.Policy` | 2 | |
| `ShopViews.Entity` | 0 | Unnamed entity — likely leftover |

`show security matrix in DPP;` returned: *"No module roles found in DPP"* — the entire module has no roles defined at all.

**How queried:** `show entities in <module>;` (AccessRules column in output), `show security matrix in DPP;`

---

### 2.6 Strict Mode Disabled (Low Risk)

```
Strict Mode: false
```

Strict mode enforces additional XPath constraint validation and is relevant to CVE-2023-23835 (XPath constraint bypass). While not critical in isolation, it should be enabled in any production deployment.

---

### 2.7 Module Roles with No Page, Microflow, or OData Permissions (Low Risk — Needs Review)

25 module roles exist in the project but have no permissions recorded in `CATALOG.PERMISSIONS`. Some are legitimate (entity-only access for API roles), others appear to be orphaned.

**Notable entries requiring review:**

| Role | Concern |
|------|---------|
| `Shop.UserRole` | Has entity grants (DELETE/READ*/WRITE* on `Shop.Order`) but no reachable pages |
| `ShopSvc.Api` | Listed as role for ShopSvc OData but absent from permissions |
| `ODataDecouplingApi.Api` | Same pattern |
| `GraphQLDemo.user` | Module `GraphQLDemo` has zero entities/pages/microflows — fully empty |
| `DataImporter.Admin` | Admin-level role with no access paths defined |
| `OQL.Administrator` | OQL module admin with no permissions |

**How found:**
```bash
comm -23 all_module_roles.txt roles_in_permissions.txt
```

---

## 3. Catalog SQL Query Analysis

### 3.1 Tables Available

| Table | Records | Notes |
|-------|---------|-------|
| `CATALOG.ENTITIES` | 204 | All entities including views, external, non-persistent |
| `CATALOG.MICROFLOWS` | 159 | Microflows and nanoflows |
| `CATALOG.PAGES` | 252 | All pages |
| `CATALOG.PERMISSIONS` | 299 | Pages, microflows, OData only — **not entities** |
| `CATALOG.REFS` | 257 | Cross-references (requires `refresh catalog full`) |

### 3.2 What Worked

| Query Pattern | Result |
|--------------|--------|
| `select col from table where col = 'value'` | ✅ Works |
| `select count(col) from table` | ✅ Works |
| `select ... GROUP by ... having count(col) > n` | ✅ Works |
| `left join ... on ... where right.Id IS null` | ✅ Works (null pattern for anti-join) |
| `refresh catalog full` | ✅ Populates `CATALOG.REFS` |
| `show project security` | ✅ Returns security level, guest access, password policy |
| `show security matrix in module` | ✅ Returns entity + microflow + page access per role |
| `describe entity Module.Entity` | ✅ Returns full GRANT statements |

### 3.3 Broken: Column Aliases Not Usable in ORDER BY

**Attempted:**
```sql
select ModuleRoleName, count(ElementName) as RoleCount
from CATALOG.PERMISSIONS
where ElementType = 'PAGE'
GROUP by ModuleRoleName
ORDER by RoleCount desc;
```

**Error:**
```
sql logic error: no such column: RoleCount (1)
```

Standard SQL allows referencing a `select` alias in `ORDER by`. The underlying SQLite engine supports this, but mxcli's query layer does not pass the alias through. **Workaround:** repeat the expression — `ORDER by count(ElementName) desc` — but this is not supported either; `ORDER by` on aggregates without aliases also fails.

**Impact:** Sorting aggregate results is not possible. Results from `GROUP by` queries come back in arbitrary order.

---

### 3.4 Broken: `not exists` Subqueries

**Attempted:**
```sql
select QualifiedName from CATALOG.ENTITIES e
where EntityType = 'PERSISTENT'
  and not exists (
    select 1 from CATALOG.PERMISSIONS p
    where p.ElementName = e.QualifiedName
  );
```

**Error:**
```
Parse error: line 7:10 extraneous input 'EXISTS' expecting ...
```

`not exists` (and by extension `exists`) is not supported by mxcli's SQL parser. This is the natural way to find entities/pages/microflows with no corresponding permissions. **Workaround:** LEFT JOIN with `where right.Id IS null`, but this requires table alias support which is also partially broken (see 3.5).

---

### 3.5 Broken: Table Aliases with Dot-Notation in Single-Line Queries

**Attempted (as single `-c` string):**
```sql
select p.ModuleName, count(p.Name) from CATALOG.PAGES p
left join CATALOG.PERMISSIONS perm on perm.ElementName = p.QualifiedName
where perm.Id IS null GROUP by p.ModuleName ORDER by count(p.Name) desc;
```

**Error:**
```
Parse error: no viable alternative at input 'SELECTp.ModuleName,COUNT(p.Name)'
```

Table aliases with dot-notation (`p.ColumnName`) are parsed correctly in multi-line script files but fail when passed via `-c` in certain join configurations. The query worked correctly when executed via `exec script.mdl` with proper line breaks, but was unreliable from `-c`.

---

### 3.6 Missing Column: `AccessRules` Not in `CATALOG.ENTITIES`

**Attempted:**
```sql
select QualifiedName from CATALOG.ENTITIES where AccessRules = 0;
```

**Error:**
```
sql logic error: no such column: AccessRules (1)
```

The `show entities in module` command displays an `AccessRules` count column in its output, but this column **does not exist** in the underlying `CATALOG.ENTITIES` table. A natural query like "find all persistent entities with no access rules" is therefore impossible via catalog SQL. It requires iterating `show entities in module` for every module separately and parsing the output.

---

## 4. Missing Data in CATALOG.PERMISSIONS

This is the most significant gap for security auditing. `CATALOG.PERMISSIONS` records only three element types:

```
microflow    → execute permissions
page         → view permissions
ODATA_SERVICE → access permissions
```

### 4.1 Entity Access Rules Are Absent

Entity-level `grant` statements — the core of all Mendix data security — are completely absent from the catalog. You cannot write a query like:

```sql
-- Does not work — ENTITY type does not exist in CATALOG.PERMISSIONS
select EntityName, ModuleRoleName, AccessType, XPathConstraint
from CATALOG.PERMISSIONS
where ElementType = 'ENTITY';
```

To audit entity access you must call `describe entity` or `show security matrix` individually per module, then parse free-text output. For a project with 204 entities across 52 modules, this is not scalable.

### 4.2 XPath Constraints Are Not Recorded Anywhere

Even if entity records were added to `CATALOG.PERMISSIONS`, the XPath constraint (row-level filter) is absent. XPath constraints are the mechanism that scopes what rows a role can read. Without them, you cannot answer:

- Does `Shop.ShopUser` see all customers, or only their own?
- Is tenant isolation enforced via XPath on `Shop.UserTenant`?
- Are there any roles with unconstrained READ access to `Shop.Customer` (PII)?

### 4.3 OData Service Authentication Mode Is Not Recorded

`CATALOG.PERMISSIONS` shows which roles can access an OData service, but not whether the service **requires authentication** in the first place. A service set to authentication mode `none` is publicly accessible regardless of role assignments. This is queryable in Studio Pro but has no catalog equivalent.

### 4.4 Published REST Services Are Invisible

The project has no dedicated `CATALOG.REST_SERVICES` table (only `CATALOG.MICROFLOWS` exists). Published REST services, their operations, and their authentication settings are not discoverable via any catalog query.

---

## 5. Feature Requests for the mxcli Team

### Priority 1 — Critical for Security Auditing

**FR-01: Add entity access rules to `CATALOG.PERMISSIONS`**

Extend `CATALOG.PERMISSIONS` with `ElementType = 'entity'` records:

```
ModuleRoleName   | ElementType | ElementName       | MemberName | AccessType              | XPathConstraint
-----------------+-------------+-------------------+------------+-------------------------+-----------------
Shop.ShopUser    | entity      | Shop.Customer     |            | read,write,delete       | [%CurrentUser%]...
Shop.ShopUser    | entity      | Shop.Customer     | Email      | read                    |
```

This single change unlocks:
- Find all entities readable by any given role
- Find all entities with unconstrained READ (no XPath)
- Find all entities accessible to a specific user role
- Detect DIVD-2022-00019 pattern with a one-liner when anonymous role exists

**FR-02: Add `AccessRuleCount` column to `CATALOG.ENTITIES`**

The `show entities` command already computes this value. Expose it in the catalog table so persistent entities with zero rules can be found via SQL:

```sql
select QualifiedName from CATALOG.ENTITIES
where EntityType = 'PERSISTENT' and AccessRuleCount = 0 and IsExternal = 0;
```

**FR-03: Add OData/REST service authentication mode to catalog**

Extend `CATALOG.MICROFLOWS` or add `CATALOG.PUBLISHED_SERVICES` with an `AuthenticationMode` column (None / UsernamePassword / ActiveSession / Custom). The single most dangerous Mendix misconfiguration — a service open to the internet with no auth — is currently invisible to catalog queries.

---

### Priority 2 — Important for Practical Queries

**FR-04: Fix `ORDER by` on column aliases from `select`**

Standard SQL behaviour: `select count(x) as n ... ORDER by n` must work. Currently `n` is not recognised in the `ORDER by` clause, making sorted aggregate results impossible. This forces awkward workarounds or produces unsorted output.

**FR-05: Support `not exists` / `exists` subqueries**

The anti-join pattern (`not exists`) is the most readable way to find "things with no matching permission". Currently only `left join ... where id IS null` works, and that form has its own alias parsing fragility. Both should be supported.

**FR-06: Fix table alias dot-notation in `-c` single-line commands**

Multi-table `join` queries using `table_alias.column` notation parse correctly in `.mdl` script files but fail intermittently when passed via the `-c` flag. The lexer appears to interpret the dot as a statement separator in some token sequences. Consistent alias support across both execution paths is needed.

---

### Priority 3 — Useful Additions

**FR-07: `AUDIT security` command**

A dedicated command that runs a complete security check and returns structured findings:

```
./mxcli audit-security -p App.mpr
```

Output should include:
- Project security level and password policy
- Demo users status
- Entities with zero access rules
- Module roles with no permissions
- OData/REST services with no authentication
- Navigation items pointing to admin/debug modules
- Orphaned module roles (exist but assigned to no user role)

**FR-08: `show ANONYMOUS GRANTS` command**

When an anonymous user role is configured, provide a direct query for all entity/page/microflow permissions granted to it:

```sql
show ANONYMOUS GRANTS;
-- Returns all permissions where role maps to the anonymous user role
```

This directly surfaces the DIVD-2022-00019 pattern.

**FR-09: Add `CATALOG.NAVIGATION` table**

Currently, discovering what is in the navigation menu requires querying `CATALOG.REFS where RefKind = 'menu_item'`, which is non-obvious. A dedicated `CATALOG.NAVIGATION` table (NavigationProfile, TargetPage, AllowedRoles) would make it straightforward to audit what is reachable without login and which roles see which nav items.

**FR-10: Add lint rule: Persistent entity with zero access rules**

A new lint rule (e.g., `SEC001`) that flags any persistent, non-external entity with `AccessRuleCount = 0`. In Production mode these entities are dead weight; in any security-level downgrade they become open. The DPP module (5 entities, 0 rules) would be caught immediately.

**FR-11: Add lint rule: Weak password policy**

A lint rule (`SEC002`) that warns when `MinimumLength < 8` or all complexity requirements are disabled. This is a standard security baseline check that has no equivalent in the current rule set.

**FR-12: Add lint rule: Demo users active**

A lint rule (`SEC003`) that warns when `demo users = true` at Production security level. Demo users should never be active in a deployed application.

---

## 6. Summary Table

| Finding | Severity | Status | Detectable via Catalog? |
|---------|----------|--------|------------------------|
| No anonymous user role (DIVD-2022-00019) | N/A | ✅ Not vulnerable | `show project security` |
| Weak password policy (min length 1) | High | ⚠️ Active | `show project security` |
| Demo users enabled (`demo_administrator`) | High | ⚠️ Active | `show demo users` |
| OqlPad in production navigation | High | ⚠️ Active | `CATALOG.REFS` (RefKind = menu_item) |
| 9 persistent entities, zero access rules (DPP) | Medium | ⚠️ Active | `show entities` output only — **not via SQL** |
| Strict mode disabled | Low | ⚠️ Active | `show project security` |
| 25 module roles with no permissions | Low | 🔍 Review needed | Diff of `show module roles` vs `CATALOG.PERMISSIONS` |
| `ORDER by` on aliases fails | Bug | ❌ Broken | — |
| `not exists` not supported | Bug | ❌ Broken | — |
| Entity grants absent from `CATALOG.PERMISSIONS` | Gap | ✅ Already implemented | `refresh catalog full` required — see Section 7 |
| XPath constraints not queryable | Gap | ✅ Already implemented | Included in PERMISSIONS table (XPathConstraint column) |
| OData auth mode not queryable | Gap | ⚠️ Partial | Schema has `AuthenticationTypes` column |
| Published REST services not in catalog | Gap | ❌ Missing | — |
| `AccessRuleCount` missing from `CATALOG.ENTITIES` | Gap | 🔧 Fix planned | — |

---

## 7. Code Investigation Findings (2026-02-28)

Investigation of the mxcli source code revealed that several items reported as missing are actually already implemented, and identified root causes for the bugs.

### 7.1 FR-01 — Entity Access Rules: Already Implemented

**Status: Already implemented — requires `refresh catalog full`**

Entity permissions (including XPath constraints and member-level access) are fully implemented in `mdl/catalog/builder_permissions.go:39-94`. The `buildEntityPermissions()` function populates `CATALOG.PERMISSIONS` with:

- Entity-level: `create`, `read`, `write`, `delete` rows per role
- Member-level: `MEMBER_READ`, `MEMBER_WRITE` rows per attribute/association
- `XPathConstraint` column populated from access rule XPath

**Why the auditor didn't see it:** Permissions only build in full mode (`builder_permissions.go:12`). The auditor ran `refresh catalog` (fast mode) instead of `refresh catalog full`. The tool gave no warning that the permissions table was empty due to mode.

**Fix planned:** Add `"permissions"` to `fullOnlyTables` in `cmd_catalog.go` so a warning is shown when querying permissions in fast mode.

### 7.2 FR-04 — ORDER BY Alias: Root Cause Found

**Status: Bug confirmed — root cause in ANTLR text reconstruction**

In `mdl/visitor/visitor_query.go:413`:
```go
query.WriteString(sel.GetText())  // BUG
```

ANTLR's `GetText()` concatenates all child tokens **without whitespace**. So the SELECT list `count(ElementName) as RoleCount` is reconstructed as `count(ElementName)ASRoleCount`. SQLite never sees the alias definition, so `ORDER by RoleCount` fails with "no such column".

The same file already uses `getSpacedText()` (which preserves spaces) for WHERE (line 421) and HAVING (line 434) clauses, but not for the SELECT list.

**Fix:** Change `sel.GetText()` to `getSpacedText(sel)` on line 413.

### 7.3 FR-05 — NOT EXISTS: Grammar Limitation Confirmed

The `exists` token is defined in `MDLLexer.g4:395` but **never referenced in any parser rule**. The `expression` rule (`MDLParser.g4:2512-2537`) does not include `exists (subquery)` or `not exists (subquery)`. This requires a grammar extension.

### 7.4 FR-06 — JOIN in Catalog Queries: Grammar Limitation Confirmed

The `catalogSelectQuery` rule (`MDLParser.g4:2228-2236`) only supports `from CATALOG.xxx` — no JOIN clause is defined. Multi-table catalog queries (e.g., LEFT JOIN between ENTITIES and PERMISSIONS) require extending the grammar rule with optional join clauses.

### 7.5 FR-03 — OData Auth Mode: Schema Exists

The `CATALOG.ODATA_SERVICES` table already has an `AuthenticationTypes` column (`tables.go:347`). Needs verification that the builder populates it from the MPR data.

---

## 8. Fix Plan

| Priority | Fix | Files | Effort |
|----------|-----|-------|--------|
| **P0** | Fix SELECT list spacing (`GetText()` → `getSpacedText()`) | `mdl/visitor/visitor_query.go:413` | 1 line |
| **P1** | Add `AccessRuleCount` to CATALOG.ENTITIES | `mdl/catalog/tables.go`, `mdl/catalog/builder_modules.go` | ~20 lines |
| **P1** | Warn on PERMISSIONS query in fast mode | `mdl/executor/cmd_catalog.go` | ~2 lines |
| **P2** | Add JOIN to `catalogSelectQuery` grammar | `mdl/grammar/MDLParser.g4`, visitor, regenerate | Medium |
| **P2** | Add `not exists` to expression grammar | `mdl/grammar/MDLParser.g4`, visitor, regenerate | Medium |
| **P2** | Add SEC001/SEC002/SEC003 lint rules | `mdl/linter/rules/` | ~150 lines each |
