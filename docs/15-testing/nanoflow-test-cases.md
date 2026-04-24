# Nanoflow Implementation — Test Cases

**Updated:** 2026-04-24
**PR:** [retran/mxcli#10](https://github.com/retran/mxcli/pull/10) (consolidated nanoflow support)
**Test projects:** Demo apps from [Mendix App Gallery](https://appgallery.mendixcloud.com/), developed by Evangelists team:
- **Lato Enquiry Management** (SP 11.4.0, 79 nanoflows) — AI multi-agent workflow, MCP
- **Evora - Factory Management** (SP 10.24.15, 93 nanoflows) — SAP, IoT, AI, Snowflake, Teamcenter
- **Lato Product Inventory** (SP 11.2.0, 51 nanoflows) — dashboards, 3D viewer, GenAI

Total: 223 nanoflows across 3 projects. Download from App Gallery, open in Studio Pro to extract `.mpr`.

## Overview

Test cases for verifying the full mxcli nanoflow implementation. Covers every MDL command, BSON round-trip, validation, catalog, security, formatting, multi-step workflows, failure modes, security cascades, and boundary cases. Derived from code audit of all nanoflow-related source files and brainstorm session (2026-04-24).

## Setup — Configuring Test Projects

### 1. Download test apps

1. Go to [Mendix App Gallery](https://appgallery.mendixcloud.com/)
2. Search for and download each demo app:
   - **Lato Enquiry Management**
   - **Evora - Factory Management**
   - **Lato Product Inventory**
3. Open each downloaded `.mpk` in Studio Pro to extract the project (this creates the `.mpr` and `mprcontents/` directory)

### 2. Build mxcli

```bash
cd ~/workspace/mxcli
git checkout pr4-nanoflows-all   # or branch with nanoflow support merged
make build && make test && make lint-go
```

### 3. Verify connectivity

```bash
# Quick smoke test — should list nanoflows from each project
for mpr in ~/workspace/mendix-apps/*/*.mpr; do
  echo "=== $(basename $(dirname $mpr)) ==="
  echo "show nanoflows;" > /tmp/show-nf.mdl
  mxcli exec /tmp/show-nf.mdl -p "$mpr" 2>&1 | tail -1
done
```

**Expected output:**
```
=== EnquiriesManagement ===
(79 nanoflows)
=== Evora-FactoryManagement-main ===
(93 nanoflows)
=== LatoProductInventory ===
(51 nanoflows)
```

### 4. Interactive testing via REPL

```bash
mxcli repl -p ~/workspace/mendix-apps/EnquiriesManagement/EnquiriesManagement.mpr
```

Then run MDL commands interactively (e.g. `show nanoflows;`, `describe nanoflow Module.Name;`).

### 5. Script-based testing

Create `.mdl` files with test sequences and execute:

```bash
mxcli exec test-sequence.mdl -p ~/workspace/mendix-apps/EnquiriesManagement/EnquiriesManagement.mpr
```

**Note:** Write operations (CREATE, DROP, GRANT/REVOKE) modify the `.mpr` file. Back up or use a copy for destructive tests.

## Prerequisites

- `mxcli` built from branch `pr4-nanoflows-all` (or later with nanoflow support)
- Test `.mpr` files set up per the Setup section above
- `make build && make test && make lint-go` passes before manual testing

---

## 1. SHOW NANOFLOWS

**Source:** `cmd_microflows_show.go:82` — `listNanoflows()`

### 1.1 List all nanoflows
```
show nanoflows;
```
**Expected:** All nanoflows listed across all modules. Verify count matches Studio Pro.

### 1.2 List nanoflows in specific module
```
show nanoflows in MyModule;
```
**Expected:** Only nanoflows from `MyModule`. No microflows mixed in.

### 1.3 Empty module
```
show nanoflows in ModuleWithNoNanoflows;
```
**Expected:** Empty result, no error.

### 1.4 Non-existent module
```
show nanoflows in NonExistentModule;
```
**Expected:** Empty result or clear error.

### 1.5 Activity count accuracy
Pick 5+ nanoflows with known activity counts (verified in Studio Pro). Check that `show nanoflows` column matches.
**Source:** `countNanoflowActivities()` in `catalog/builder_microflows.go:427`

### 1.6 Complexity calculation
Verify complexity values shown for nanoflows with varying numbers of decisions, loops, and nested paths.
**Source:** `calculateNanoflowComplexity()` in `catalog/builder_microflows.go:447`

---

## 2. DESCRIBE NANOFLOW

**Source:** `cmd_microflows_show.go:316` — `describeNanoflow()`

### 2.1 Simple nanoflow (no parameters, no return)
```
describe nanoflow Module.SimpleNanoflow;
```
**Expected:** Valid `create or modify nanoflow` MDL output. Empty body or minimal activities.

### 2.2 Nanoflow with parameters and return type
```
describe nanoflow Module.NanoflowWithParams;
```
**Expected:** Parameters with correct types. Return type shown.

### 2.3 Parameter format variants
The BSON parser handles 3 parameter storage formats: `MicroflowParameterCollection`, `MicroflowParameters`, `Parameters`. Additionally, parameters may be extracted from `ObjectCollection.Objects` as fallback.
**Test:** Find nanoflows in different Studio Pro versions to exercise each path.

### 2.4 Nanoflow with activities — full coverage

Test DESCRIBE on nanoflows containing each of the 25 allowed action types:

| # | Activity | What to verify |
|---|----------|---------------|
| 1 | CreateVariable | Variable name, type, initial value |
| 2 | ChangeVariable | Target variable, new value expression |
| 3 | CreateObject | Entity name, member assignments |
| 4 | ChangeObject | Target object, changed members |
| 5 | CommitObject | With/without events |
| 6 | DeleteObject | Target object |
| 7 | RollbackObject | Target object |
| 8 | Retrieve | Source (association/database), XPath constraint |
| 9 | AggregateList | List, function (sum/avg/count/min/max) |
| 10 | ChangeList | Target list, operation |
| 11 | CreateList | Entity type |
| 12 | ListOperation | Operation type, lists involved |
| 13 | CastObject | Source, target type |
| 14 | ShowPage | Page reference, page parameter object |
| 15 | ClosePage | No arguments |
| 16 | ShowMessage | Message template, blocking/non-blocking |
| 17 | ValidationFeedback | Object, member, message |
| 18 | CallNanoflow | `call nanoflow Module.Name (args)` — NOT `call microflow` |
| 19 | CallMicroflow | `call microflow Module.Name (args)` — server call from nanoflow |
| 20 | CallJavaScriptAction | `call javascript action` syntax, JS action reference |
| 21 | Synchronize | No arguments (nanoflow-only) |
| 22 | LogMessage | Level, message template |
| 23 | ExclusiveSplit | Decision expression, true/false paths |
| 24 | Loop | Iterator variable, list variable |
| 25 | MergeNode | Multiple incoming paths converge |

### 2.5 Nanoflow with error handling
```
describe nanoflow Module.NanoflowWithErrorHandling;
```
**Expected:** Error handler flow shown. `$latestError` predefined variable preserved.
**Source:** `getErrorHandling()` covers 12 statement types — verify IfStmt, LoopStmt, WhileStmt error handlers.

### 2.6 Nanoflow with nested control flow
Test nanoflows with:
- If inside loop
- Loop inside if
- Nested if/else chains
- Error handling inside loop body

### 2.7 Nanoflow not found
```
describe nanoflow Module.DoesNotExist;
```
**Expected:** Clear error message.

### 2.8 Activity count regression
Pick 5+ nanoflows with known activity counts from Studio Pro. Verify DESCRIBE body contains correct number.
**Regression:** Parser previously returned 0 activities due to missing `parseMicroflowObjectCollection()` call.

### 2.9 Documentation and MarkAsUsed properties
Test nanoflow with Documentation string set and MarkAsUsed=true. Verify both appear in DESCRIBE output.

### 2.10 Excluded nanoflow
Test nanoflow with Excluded=true. Verify property appears in output.

---

## 3. CREATE NANOFLOW

**Source:** `cmd_nanoflows_create.go` — `execCreateNanoflow()` (225 lines)

> **Note:** Parentheses are required even for parameterless nanoflows: `create nanoflow M.N () begin end;`. This is grammar-by-design — the parser expects `'(' params? ')'` unconditionally.

### 3.1 Minimal nanoflow
```
create nanoflow MyModule.TestNano ()
begin
end;
```
**Expected:** Created. Listed in `show nanoflows`. Roundtrips via DESCRIBE.

### 3.2 With parameters — primitive types
```
create nanoflow MyModule.TestParams (
  Name : String,
  Count : Integer,
  Active : Boolean,
  Amount : Decimal,
  StartDate : DateTime
) returns String
begin
end;
```
**Expected:** All parameter types preserved.

### 3.3 With entity parameter
```
create nanoflow MyModule.TestEntity (
  Input : MyModule.MyEntity
) returns MyModule.MyEntity
begin
end;
```
**Expected:** Entity reference resolved. Error if entity doesn't exist.

### 3.4 With enumeration parameter
```
create nanoflow MyModule.TestEnum (
  Status : MyModule.StatusEnum
) returns MyModule.StatusEnum
begin
end;
```
**Expected:** Enum reference resolved. Error if enum doesn't exist.

### 3.5 With activities
```
create nanoflow MyModule.TestActivities ()
begin
  log info 'hello';
  log warning 'world';
end;
```
**Expected:** Activities preserved in DESCRIBE. `log info 'text'` renders as `log info node 'Application' 'text'` in output.

### 3.6 With call nanoflow action
```
create nanoflow MyModule.Caller () returns Boolean
begin
  $Result = call nanoflow MyModule.Target ();
end;
```
**Expected:** `NanoflowCallAction` stored in BSON (not `MicroflowCallAction`).

### 3.7 Create or modify — existing nanoflow
```
create or modify nanoflow MyModule.Existing ()
begin
end;
```
**Expected:** Existing nanoflow updated (ID reused). No AlreadyExistsError.

### 3.8 Create duplicate — no CreateOrModify
```
create nanoflow MyModule.TestNano () begin end;
create nanoflow MyModule.TestNano () begin end;
```
**Expected:** Second CREATE fails with AlreadyExistsError.

### 3.9 Module auto-creation
```
create nanoflow NewModule.TestNano () begin end;
```
**Expected:** If `NewModule` doesn't exist, it's created automatically.

### 3.10 Folder resolution
```
create nanoflow MyModule.TestNano () in folder 'SubFolder/Nested' begin end;
```
**Expected:** Nanoflow placed in correct folder. Error if folder path invalid.

### 3.11 consumeDroppedNanoflow — ID reuse
```
create nanoflow MyModule.A () begin end;
drop nanoflow MyModule.A;
create nanoflow MyModule.A () begin end;
```
**Expected:** Second CREATE reuses the ID from the dropped nanoflow (same session).

### 3.12 Default return type
Create nanoflow without explicit return type. DESCRIBE should show VoidType or omit return.

### 3.13 Not-connected-for-write guard
Attempt CREATE without opening a project for writing.
**Expected:** Error about not being connected.

---

## 4. CREATE NANOFLOW — Validation

**Source:** `nanoflow_validation.go` (145 lines)

### 4.1 Disallowed actions — full list
Each must be rejected with clear error when used inside `create nanoflow`:

| # | Disallowed action | BSON/AST type |
|---|-------------------|---------------|
| 1 | ErrorEvent | `ast.ErrorEvent` |
| 2 | Java action call | Java action |
| 3 | Database query | External DB query |
| 4 | REST call | Call REST service |
| 5 | Web service call | Call web service |
| 6 | Import mapping | Import with mapping |
| 7 | Export mapping | Export with mapping |
| 8 | Generate document | Document generation |
| 9 | Show home page | ShowHomePage |
| 10 | Download file | Download file |
| 11 | External action | Call external action |
| 12 | Send external object | Send external object |
| 13 | Delete external object | Delete external object |
| 14 | All workflow actions (9 types) | Create/show/complete/lock/etc |

### 4.2 Binary return type rejected
```
create nanoflow MyModule.Bad () returns Binary begin end;
```
**Expected:** Validation error.

### 4.3 Float return type
`ast.TypeFloat` does not exist in the AST — Float can never be used as a nanoflow return type. No validation needed.

### 4.4 Disallowed actions in nested control flow
```
create nanoflow MyModule.Nested ()
begin
  if (true) then
    call java action SomeModule.JavaAction ();
  end if;
end;
```
**Expected:** Rejected — validation recurses into IfStmt, LoopStmt, WhileStmt bodies.

### 4.5 Disallowed actions in error handling body
**Expected:** Rejected — validation also checks error handling clauses.

### 4.6 Cross-reference validation — nanoflow target
```
create nanoflow MyModule.BadRef () begin
  call nanoflow NonExistent.Flow ();
end;
```
**Expected:** Error — target nanoflow not found.
**Source:** `validate.go:285-293` uses `buildNanoflowQualifiedNames`.

### 4.7 Cross-reference validation — page target
```
create nanoflow MyModule.BadPage () begin
  show page NonExistent.Page ();
end;
```
**Expected:** Error — target page not found.

### 4.8 Cross-reference validation — microflow target from nanoflow
```
create nanoflow MyModule.BadMF () begin
  call microflow NonExistent.Flow ();
end;
```
**Expected:** Error — target microflow not found.

---

## 5. DROP NANOFLOW

**Source:** `cmd_nanoflows_drop.go:14` — `execDropNanoflow()`

### 5.1 Drop existing nanoflow
```
create nanoflow MyModule.ToDrop () begin end;
drop nanoflow MyModule.ToDrop;
```
**Expected:** Removed. Not in `show nanoflows`.

### 5.2 Drop non-existent nanoflow
```
drop nanoflow MyModule.DoesNotExist;
```
**Expected:** Clear error.

### 5.3 Drop referenced nanoflow
Create two nanoflows where one calls the other, drop the callee.
**Expected:** Warning or error about dangling reference.

### 5.4 Drop and recreate (ID reuse)
See 3.11 — verify `consumeDroppedNanoflow` works.

### 5.5 Not-connected-for-write guard
**Expected:** Error if no project open for writing.

---

## 6. CALL NANOFLOW (inside flow body)

**Source:** Flow builder, `validate_microflow.go:357`

Note: `call nanoflow` is an action inside a flow body (`begin`/`end`), not a standalone MDL command.

### 6.1 Call with arguments
```
create nanoflow MyModule.Adder (A : Integer, B : Integer) returns Integer
begin
  $Result = $A + $B;
end;
create microflow MyModule.Caller () returns Integer
begin
  $Result = call nanoflow MyModule.Adder (A = 1, B = 2);
end;
```
**Expected:** Microflow can call nanoflow. Arguments mapped correctly.

### 6.2 Call nanoflow from nanoflow
**Expected:** Uses `NanoflowCallAction` BSON type (not `MicroflowCallAction`).

### 6.3 Call with return value assignment
```
$Result = call nanoflow MyModule.GetValue ();
```
**Expected:** Return value assigned to variable.

### 6.4 Call without return value (void nanoflow)
```
call nanoflow MyModule.DoSomething ();
```
**Expected:** No assignment. No error.

### 6.5 Call with error handling clause
```
$Result = call nanoflow MyModule.Risky () on error continue;
```
**Expected:** `onErrorClause` parsed and preserved.

### 6.6 Recursive call
```
create nanoflow MyModule.Recursive () returns Boolean
begin
  $Result = call nanoflow MyModule.Recursive ();
end;
```
**Expected:** Parses without error (no compile-time recursion check).

---

## 7. GRANT / REVOKE EXECUTE ON NANOFLOW

**Source:** `cmd_security_write.go:674` (grant), `:733` (revoke)

> **Note:** Drop/recreate of the same nanoflow name preserves security roles by design. The executor caches `AllowedModuleRoles` on DROP and restores them on the next CREATE with the same qualified name (`rememberDroppedNanoflow`/`consumeDroppedNanoflow` pattern). Use REVOKE after recreate if roles should change.

### 7.1 Grant to single role
```
grant execute on nanoflow MyModule.TestNano to MyModule.User;
```
**Expected:** `AllowedModuleRoles` updated. Verifiable via `show access on nanoflow`.

### 7.2 Grant to multiple roles
```
grant execute on nanoflow MyModule.TestNano to MyModule.User, MyModule.Admin;
```
**Expected:** Both roles added.

### 7.3 Idempotent grant
Grant same role twice.
**Expected:** Message "already have access" on second grant. No duplicate entries.

### 7.4 Revoke from role
```
revoke execute on nanoflow MyModule.TestNano from MyModule.User;
```
**Expected:** Role removed.

### 7.5 Idempotent revoke
Revoke role that was never granted.
**Expected:** Message "none of the specified roles have access".

### 7.6 Grant on non-existent nanoflow
**Expected:** Clear error — nanoflow not found.

### 7.7 Grant with non-existent role
**Expected:** Error from `validateModuleRole`.

### 7.8 Not-connected-for-write guard
**Expected:** Error if no project open for writing.

---

## 8. SHOW ACCESS ON NANOFLOW

**Source:** `cmd_security.go:405` — `listAccessOnNanoflow()`

### 8.1 Nanoflow with roles
```
show access on nanoflow MyModule.TestNano;
```
**Expected:** Lists all allowed module roles.

### 8.2 Nanoflow without roles
**Expected:** "No access" or empty list.

### 8.3 JSON output format
**Expected:** Valid JSON array of role objects.

### 8.4 Nil name
**Expected:** Validation error (not crash).

### 8.5 Non-existent nanoflow
**Expected:** Clear error.

### 8.6 Role ID format
`AllowedModuleRoles` stored as IDs — `listAccessOnNanoflow` splits on `.` to display `Module.Role`. Verify this works correctly for all role formats.

---

## 9. RENAME NANOFLOW

### 9.1 Simple rename
```
rename nanoflow MyModule.OldName to NewName;
```
**Expected:** Renamed. `show nanoflows` shows new name.

### 9.2 Rename with callers
Rename nanoflow called by another flow. Verify caller's reference updated.

### 9.3 Rename to existing name
**Expected:** Error — name collision.

---

## 10. MOVE NANOFLOW

**Source:** `cmd_move.go:215` — `moveNanoflow()`

### 10.1 Move to another module
```
move nanoflow MyModule.TestNano to TargetModule;
```
**Expected:** Nanoflow moved to `TargetModule`. Qualified name becomes `TargetModule.TestNano`.

### 10.2 Move to non-existent module
**Expected:** Error: `failed to find target module: module not found: X`.

### 20.4 Iterative development (CREATE OR MODIFY loop)
1. `create nanoflow M.Evolving () begin end;`
2. `create or modify nanoflow M.Evolving ($Name : String) begin end;` — add parameter
3. `create or modify nanoflow M.Evolving ($Name : String) returns String begin $Result = $Name; end;` — add body + return
4. `create or modify nanoflow M.Evolving ($Name : String, $Count : Integer) returns String begin $Result = $Name; end;` — add second param
5. After each step: DESCRIBE and verify cumulative changes preserved
6. Final: DESCRIBE → capture MDL → DROP → execute captured MDL → DESCRIBE → compare (full roundtrip)
**Expected:** Each modification preserves prior state except for explicitly changed fields. Roundtrip at end matches last version.

### 20.5 Drop and recreate with different signature
1. CREATE nanoflow with `String` return type and 2 params, GRANT roles
2. DROP
3. CREATE same name with `Integer` return type and 0 params
4. `show access on nanoflow M.Name;` — verify roles NOT carried over (clean slate)
5. DESCRIBE — verify new signature, no remnant of old params/body
**Source:** exercises `consumeDroppedNanoflow` (ID reuse) + security state reset

### 20.6 Cross-module call chain
1. CREATE `ModuleA.Entrypoint` calling `ModuleB.Processor`
2. CREATE `ModuleB.Processor` calling `microflow ModuleC.DataFetcher`
3. DESCRIBE `ModuleA.Entrypoint` — verify cross-module nanoflow call shown
4. DROP `ModuleB.Processor`
5. DESCRIBE `ModuleA.Entrypoint` — verify dangling reference handling
6. Recreate `ModuleB.Processor` — verify caller roundtrips cleanly again
**Source:** tests cross-module references and dangling ref recovery

---

## 21. FAILURE MODES & ERROR RECOVERY

Test error paths, partial state after failures, and silent corruption detection.

### 21.1 Validation failure mid-batch
Execute MDL script with 3 CREATEs where #2 has a disallowed action:
```
create nanoflow M.Good1 () begin end;
create nanoflow M.Bad () begin call java action SomeModule.JavaAction (); end;
create nanoflow M.Good3 () begin end;
```
**Expected:** Good1 created, Bad rejected with clear error, Good3 **not created** — batch aborts on first error.

> **Note:** Batch mode (`mxcli exec`) is fail-fast — the first error aborts all remaining statements. REPL mode (interactive or piped) continues on error per-line. This is consistent across all entity types. `IF EXISTS` / `IF NOT EXISTS` syntax does not exist yet.

### 21.2 CREATE with non-existent entity parameter
```
create nanoflow M.BadParam (Input : NonExistent.Entity) begin end;
```
**Expected:** Clear error. No partial nanoflow left in model. `show nanoflows` does not list `M.BadParam`.

### 21.3 CREATE with non-existent enum return type
```
create nanoflow M.BadReturn () returns NonExistent.MyEnum begin end;
```
**Expected:** Clear error. No partial nanoflow left in model.

### 21.4 DESCRIBE after partial modification failure
1. CREATE nanoflow with valid body
2. Attempt `create or modify` with invalid body (disallowed action)
3. DESCRIBE — verify original version preserved (not corrupted by failed modify)
**Source:** `execCreateNanoflow` should be atomic — either fully applied or fully rejected

### 21.5 BSON roundtrip data integrity check
For 10+ complex nanoflows from test projects (choose ones with error handling, annotations, 10+ activities, multiple parameter types):
1. DESCRIBE → capture MDL
2. DROP
3. Execute captured MDL
4. DESCRIBE → capture again
5. **Diff the two outputs** — any difference is a data loss bug
**Source:** exercises full BSON parse→serialize→parse pipeline against real-world data

### 21.6 Double DROP
```
drop nanoflow M.X;
drop nanoflow M.X;
```
**Expected:** First succeeds, second gives clear "not found" error (not crash or silent failure).

### 21.7 GRANT on just-dropped nanoflow (same session)
1. CREATE nanoflow, then DROP
2. `grant execute on nanoflow M.Dropped to M.User;`
**Expected:** Clear error about nanoflow not found. No crash, no phantom entry created.

### 21.8 CREATE OR MODIFY with completely different body
1. CREATE nanoflow with 5-activity body (variables, if/else, loop)
2. CREATE OR MODIFY same name with completely different 3-activity body
3. DESCRIBE — verify old body fully replaced, no ghost activities from original
**Source:** verifies ObjectCollection is fully replaced, not merged

### 21.9 Corrupt cross-reference after callee drop
1. CREATE nanoflow A calling nanoflow B (both exist)
2. DROP B
3. DESCRIBE A — how is the dangling `call nanoflow M.B` rendered? Verify no crash.
4. `describe nanoflow M.A mermaid;` — verify graceful handling of missing call target
**Expected:** Either error message, placeholder text, or unresolved qualified name — not a panic.

### 21.10 Error message quality audit
For each error scenario, verify the error message includes:
- **What** went wrong (e.g., "nanoflow not found", "disallowed action")
- **Which** nanoflow (qualified name)
- **Actionable guidance** (e.g., "did you mean...", "use `show nanoflows` to list available")

Scenarios: not-found (DESCRIBE, DROP, GRANT, REVOKE, MOVE, SHOW ACCESS), not-connected (CREATE, DROP, GRANT, REVOKE), validation failure (disallowed action, binary return), duplicate (CREATE without OR MODIFY).

### 21.11 Empty string and unicode names
```
create nanoflow MyModule.Nañoflow_テスト () begin end;
```
**Expected:** Parser handles consistently — either accepts and roundtrips correctly, or rejects with clear error. Document actual behavior.

### 21.12 Very long MDL statement
CREATE nanoflow with 100-character name, 10 parameters, 20-line body with nested control flow.
**Expected:** Parses and executes without truncation or buffer issues. DESCRIBE output complete.

---

## 22. SECURITY CASCADES

Extended grant/revoke patterns testing role accumulation, cross-module references, and persistence.

### 22.1 Multi-role accumulation
1. `grant execute on nanoflow M.N to M.RoleA;`
2. `grant execute on nanoflow M.N to M.RoleB;`
3. `grant execute on nanoflow M.N to M.RoleC;`
4. `show access on nanoflow M.N;` — verify all 3 roles present (no overwrite on each grant)
**Source:** `UpdateAllowedRoles` merge logic in `cmd_security_write.go`

### 22.2 Selective revoke
1. GRANT roles A, B, C (per 22.1)
2. `revoke execute on nanoflow M.N from M.RoleB;` — revoke only B
3. `show access on nanoflow M.N;` — verify A and C remain, B removed
**Source:** revoke filters by role, shouldn't affect other roles

### 22.3 Revoke all then re-grant
1. GRANT A, B, C
2. REVOKE A, then REVOKE B, then REVOKE C
3. `show access on nanoflow M.N;` — verify empty
4. GRANT A
5. `show access on nanoflow M.N;` — verify only A
**Source:** tests empty AllowedModuleRoles state + recovery

### 22.4 Cross-module role reference
```
grant execute on nanoflow ModuleA.Nano to ModuleB.UserRole;
```
**Expected:** Cross-module role reference accepted and persisted. SHOW ACCESS displays `ModuleB.UserRole` correctly.
**Source:** `AllowedModuleRoles` stored as IDs — display splits on `.` delimiter

### 22.5 Security persistence through save/reopen
1. CREATE nanoflow, GRANT 2 roles
2. Disconnect from project
3. Reconnect to same project
4. `show access on nanoflow M.N;` — verify roles persisted in BSON
**Source:** verifies `serializeNanoflow` correctly writes AllowedModuleRoles to BSON

### 22.6 Security state after CREATE OR MODIFY
1. CREATE nanoflow, GRANT roles A and B
2. `create or modify nanoflow M.N () begin $x : String = 'changed'; end;` — change body only
3. `show access on nanoflow M.N;` — verify roles A and B preserved (not cleared by modify)
**Source:** CREATE OR MODIFY should reuse existing nanoflow ID and preserve non-body fields

### 22.7 Bulk grant in script
```
grant execute on nanoflow M.N1 to M.User, M.Admin;
grant execute on nanoflow M.N2 to M.User, M.Admin;
grant execute on nanoflow M.N3 to M.User, M.Admin;
show access on nanoflow M.N1;
show access on nanoflow M.N2;
show access on nanoflow M.N3;
```
Execute as batch. Verify all 3 nanoflows have both roles.
**Source:** tests batch execution of security commands

### 22.8 Grant with non-existent module role
```
grant execute on nanoflow M.Nano to M.NonExistentRole;
```
**Expected:** Clear error from `validateModuleRole`. No partial grant applied. SHOW ACCESS unchanged.

---

## 23. BOUNDARY & STRESS CASES

Extreme inputs that push beyond typical usage patterns.

### 23.1 Maximum parameters (20)
CREATE nanoflow with 20 parameters of mixed types:
```
create nanoflow M.ManyParams (
  P1 : String, P2 : Integer, P3 : Boolean, P4 : Decimal, P5 : DateTime,
  P6 : String, P7 : Integer, P8 : Boolean, P9 : Decimal, P10 : DateTime,
  P11 : String, P12 : Integer, P13 : Boolean, P14 : Decimal, P15 : DateTime,
  P16 : String, P17 : Integer, P18 : Boolean, P19 : Decimal, P20 : DateTime
) returns String
begin
end;
```
DESCRIBE — verify all 20 preserved with correct types and names.
Roundtrip — DROP → CREATE from DESCRIBE output → DESCRIBE → compare.

### 23.2 Deeply nested control flow (4+ levels)
```
create nanoflow M.DeepNest ()
begin
  if true then
    if true then
      if true then
        if true then
          log info 'level4';
        end if;
      end if;
    end if;
  end if;
end;
```
DESCRIBE — verify full 4-level nesting preserved (each level with positions, anchors, captions).
**Source:** exercises recursive validation (`validateNanoflowBody`) and ObjectCollection depth

### 23.3 Many activities (30+)
CREATE nanoflow with 30+ sequential log actions:
```
create nanoflow M.ManyActivities ()
begin
  log info 'action_1';
  log info 'action_2';
  ...
  log info 'action_30';
end;
```
DESCRIBE — verify all 30+ activities present (ObjectCollection not truncated).
Check activity count in `show nanoflows;` matches 30+.

### 23.4 Empty body with complex signature
```
create nanoflow M.EmptyComplex (
  A : String, B : Integer, C : Boolean, D : Decimal, E : DateTime
) returns String
begin
end;
```
**Expected:** Empty body accepted. All 5 parameters and return type roundtrip correctly via DESCRIBE.

### 23.5 Nanoflow calling 5+ other nanoflows
CREATE 5 target nanoflows, then one caller that calls all 5 in sequence:
```
create nanoflow M.Caller () returns Boolean
begin
  call nanoflow M.Target1 ();
  call nanoflow M.Target2 ();
  call nanoflow M.Target3 ();
  call nanoflow M.Target4 ();
  call nanoflow M.Target5 ();
end;
```
DESCRIBE — verify all 5 call targets preserved.
MERMAID — verify all 5 call nodes rendered with correct target names.

### 23.6 Multiple error handling clauses
CREATE nanoflow with 3 different `on error continue` clauses:
```
create nanoflow M.MultiError () returns Boolean
begin
  $R1 = call nanoflow M.Risky1 () on error continue;
  $R2 = call nanoflow M.Risky2 () on error continue;
  $R3 = call nanoflow M.Risky3 () on error continue;
end;
```
DESCRIBE — verify all 3 error handling clauses preserved.
**Source:** `getErrorHandling()` in `validate_microflow.go`

### 23.7 Annotations on every statement type
CREATE nanoflow with `@annotation`, `@caption`, `@color`, `@position` on various statement types (declare, set, if, loop, call).
DESCRIBE — verify all annotations roundtrip.
MERMAID — verify annotations rendered where applicable.
**Source:** AnnotationFlows parsing in `parser_nanoflow.go`

### 23.8 Show nanoflows with 100+ results
Run `show nanoflows;` on Evora project (93 nanoflows).
CREATE 10 additional nanoflows to push past 100.
**Expected:** No truncation, formatting stable, all listed.

### 23.9 Rapid CREATE/DROP cycle (10 iterations)
```
-- repeat 10 times:
create nanoflow M.Temp () begin end;
drop nanoflow M.Temp;
```
After all 10 cycles: `show nanoflows;` should show zero temp nanoflows.
**Expected:** No resource leak, no ID collision, no catalog corruption after rapid cycling.
**Source:** exercises `consumeDroppedNanoflow` ID reuse under repeated stress

### 23.10 Nanoflow with all 25 allowed action types
CREATE single nanoflow containing one instance of each allowed action type (where grammar supports it): CreateVariable, ChangeVariable, CreateObject, ChangeObject, CommitObject, DeleteObject, RollbackObject, Retrieve, AggregateList, ChangeList, CreateList, ListOperation, CastObject, ShowPage, ClosePage, ShowMessage, ValidationFeedback, CallNanoflow, CallMicroflow, CallJavaScriptAction, Synchronize, LogMessage, ExclusiveSplit, Loop, MergeNode.
DESCRIBE → compare to original input. This is the ultimate roundtrip stress test.
**Note:** Some action types may not be expressible in current MDL grammar — document which ones work and which don't.

---

## Test Project Coverage Matrix

| Category | Enquiries (79) | Evora Factory (93) | Lato Inventory (51) |
|---|---|---|---|
| SHOW nanoflows count | Verify: 79 | Verify: 93 | Verify: 51 |
| DESCRIBE (sample 10+) | Diverse activities | Diverse activities | Diverse activities |
| DESCRIBE MERMAID (sample 5) | Complex flows | Complex flows | Complex flows |
| SHOW ACCESS (sample 5) | With/without roles | With/without roles | With/without roles |
| Catalog query | Full table | Full table | Full table |
| Roundtrip (sample 10+) | Describe→Drop→Create→Describe | Same | Same |
| Activity coverage | Track 25 allowed types | Same | Same |
| Multi-step workflows (§20) | Use project entities for call chains | Same | Same |
| BSON data integrity (§21.5) | 10+ complex nanoflows | Same | Same |
| Security cascades (§22) | Test with project roles | Same | Same |
| Boundary: 100+ listing (§23.8) | N/A (79) | CREATE extras to reach 100+ | N/A (51) |

---

## Automated Test Coverage Status

| Area | Automated Tests | Status |
|---|---|---|
| Catalog: activity count | `TestCountNanoflowActivities` | Covered |
| Catalog: complexity | `TestCalculateNanoflowComplexity` | Covered |
| Registry: handler registration | `registry_test.go` | Covered |
| CREATE NANOFLOW executor | 13 integration + 4 mock tests | Covered |
| DROP NANOFLOW executor | 2 integration + 1 mock test | Covered |
| GRANT/REVOKE executor | 3 integration + 5 mock tests | Covered |
| SHOW NANOFLOWS executor | 2 integration + 2 mock tests | Covered |
| DESCRIBE NANOFLOW executor | 2 integration + 2 mock tests | Covered |
| SHOW ACCESS executor | 3 mock tests | Covered |
| MOVE NANOFLOW executor | 1 integration + 1 mock test | Covered |
| MERMAID output | 1 integration test | Covered |
| Nanoflow validation | 6 mock tests (disallowed actions, nested, return types) | Covered |
| BSON parser | 5 roundtrip tests (synthetic + with activities) | Covered |
| BSON writer | 5 roundtrip tests (serialize→parse→serialize cycle) | Covered |
| Diff output | None | **Gap** |
| Roundtrip (integration) | 3 integration tests (create→describe→compare) | Covered |
| Multi-step workflows (§20) | None | **Manual only** |
| Failure modes (§21) | Partial: double-drop, not-connected guards in mock tests | **Mostly manual** |
| Security cascades (§22) | Partial: idempotent grant/revoke in mock + integration tests | **Mostly manual** |
| Boundary cases (§23) | None | **Manual only** |

Manual testing priority:
1. Roundtrip all 223 nanoflows across 3 test projects (bulk DESCRIBE→DROP→CREATE→DESCRIBE)
2. Activity type coverage (verify all 25 allowed actions represented)
3. Multi-step workflows (§20) — highest risk for interaction bugs
4. Failure modes (§21) — especially §21.5 BSON data integrity and §21.8 body replacement
5. Diff output with nanoflow changes

---

## Manual Test Report Template

Copy and fill in after running manual tests. Include in PR description under `## Manual Testing`.

```markdown
## Manual Testing

**Date:** YYYY-MM-DD
**Branch:** pr4-nanoflows-all
**Build:** `make build && make test && make lint-go` — PASS

### Test Projects

| App | Studio Pro | Nanoflows | SHOW count | DESCRIBE sample | Mermaid sample | Roundtrip |
|-----|-----------|-----------|------------|-----------------|----------------|-----------|
| Lato Enquiry Management | 11.4.0 | 79 | ✅ 79 | ✅ _n_/79 | ✅ _n_/79 | ✅ _n_/79 |
| Evora Factory Management | 10.24.15 | 93 | ✅ 93 | ✅ _n_/93 | ✅ _n_/93 | ✅ _n_/93 |
| Lato Product Inventory | 11.2.0 | 51 | ✅ 51 | ✅ _n_/51 | ✅ _n_/51 | ✅ _n_/51 |

### Command Coverage

| Command | Tested | Notes |
|---------|--------|-------|
| SHOW NANOFLOWS | ✅/❌ | |
| SHOW NANOFLOWS IN module | ✅/❌ | |
| DESCRIBE NANOFLOW | ✅/❌ | |
| DESCRIBE MERMAID NANOFLOW | ✅/❌ | |
| CREATE NANOFLOW | ✅/❌ | |
| CREATE OR MODIFY NANOFLOW | ✅/❌ | |
| DROP NANOFLOW | ✅/❌ | |
| CALL NANOFLOW (in body) | ✅/❌ | |
| GRANT EXECUTE ON NANOFLOW | ✅/❌ | |
| REVOKE EXECUTE ON NANOFLOW | ✅/❌ | |
| SHOW ACCESS ON NANOFLOW | ✅/❌ | |
| RENAME NANOFLOW | ✅/❌ | |
| MOVE NANOFLOW | ✅/❌ | |

### Bulk Roundtrip Results

> **Note:** Expression whitespace is intentionally normalized during roundtrip. Function arguments get a space after commas: `find($x,'y')` → `find($x, 'y')`. This is by-design normalization for readability, not a fidelity bug.

```
# Command used:
# for each nanoflow: describe → capture → drop → execute captured → describe → diff

Total: _n_ nanoflows tested
Passed: _n_
Failed: _n_ (list failures below)
```

### Activity Type Coverage

_List which of the 25 allowed action types were exercised across test projects._

| # | Activity | Found in test project | Roundtrip OK |
|---|----------|-----------------------|-------------|
| 1 | CreateVariable | | |
| 2 | ChangeVariable | | |
| ... | ... | | |

### Validation Tests

| Scenario | Result | Notes |
|----------|--------|-------|
| Disallowed action rejected | ✅/❌ | |
| Binary return type rejected | ✅/❌ | |
| Nested disallowed action rejected | ✅/❌ | |
| Cross-ref to non-existent target | ✅/❌ | |

### Multi-Step Workflows (§20)

| Scenario | Result | Notes |
|----------|--------|-------|
| 20.1 Scaffold module | ✅/❌ | |
| 20.2 Rename in call chain | ✅/❌ | |
| 20.3 Move and reorganize | ✅/❌ | |
| 20.4 Iterative CREATE OR MODIFY | ✅/❌ | |
| 20.5 Drop/recreate different sig | ✅/❌ | |
| 20.6 Cross-module call chain | ✅/❌ | |

### Failure Modes (§21)

| Scenario | Result | Notes |
|----------|--------|-------|
| 21.1 Validation mid-batch | ✅/❌ | |
| 21.4 DESCRIBE after failed modify | ✅/❌ | |
| 21.5 BSON data integrity (10+) | ✅/❌ | |
| 21.8 Full body replacement | ✅/❌ | |
| 21.9 Dangling cross-reference | ✅/❌ | |

### Security Cascades (§22)

| Scenario | Result | Notes |
|----------|--------|-------|
| 22.1 Multi-role accumulation | ✅/❌ | |
| 22.5 Persistence through save | ✅/❌ | |
| 22.6 Preserved after modify | ✅/❌ | |

### Boundary Cases (§23)

| Scenario | Result | Notes |
|----------|--------|-------|
| 23.1 20 parameters | ✅/❌ | |
| 23.2 5-level nesting | ✅/❌ | |
| 23.9 Rapid CREATE/DROP x10 | ✅/❌ | |
| 23.10 All 25 action types | ✅/❌ | |

### Issues Found

_List any issues discovered during manual testing. For each: command, input, expected vs actual, severity._

1. (none / describe issues here)
```
