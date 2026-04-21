# Bug Report: REVOKE and DROP MODULE ROLE fail to remove multi-role access rules

## Summary

Two related bugs in mxcli security management:

1. **`revoke` silently fails** on entity access rules that have multiple module roles assigned to a single rule (the pattern Studio Pro uses). It reports success ("Revoked access on ...") but the rule is unchanged.
2. **`drop module role` does not cascade-delete** entity access rules that reference the dropped role, leaving orphaned rules that reference non-existent roles.

## Impact

Orphaned access rules referencing non-existent module roles remain in the domain model. These cannot be removed via mxcli, requiring manual cleanup in Studio Pro. The orphaned rules may also interfere with runtime behavior — in our case, datagrids in a Mendix 11.6.3 app showed "no access to attribute" errors that could not be resolved via mxcli.

## Environment

- mxcli version: local binary (latest as of 2026-02-22)
- Mendix version: 11.6.3
- OS: Linux (devcontainer)

## Root Cause Analysis

Studio Pro creates access rules with **multiple roles on a single rule**:
```
rule 1: F1.User, F1.UserRole
  Rights: create, delete
  default member access: ReadWrite
```

mxcli `grant` creates **one rule per role**:
```
rule 2: F1.RoleA
  Rights: read
rule 3: F1.RoleB
  Rights: read
```

`revoke` only matches and removes **single-role rules** (the mxcli pattern). It does not match rules with multiple roles assigned.

## Reproduction Steps

### Bug 1: REVOKE silently fails on multi-role rules

**Prerequisite:** An entity with a multi-role access rule (created via Studio Pro or equivalent). In our case, `F1.Season` had:
```
rule 1: F1.User, F1.UserRole
  Rights: create, delete
  default member access: ReadWrite
```

```sql
-- Both roles exist as module roles
show module roles in F1;
-- Output: F1.User, F1.UserRole

-- Attempt to revoke
revoke F1.User on F1.Season;
-- Output: "Revoked access on F1.Season from F1.User"

-- Check: rule is UNCHANGED
show access on F1.Season;
-- Output: Rule 1: F1.User, F1.UserRole (still present, no change)
```

**Expected:** Either remove F1.User from the rule (leaving just F1.UserRole), or remove the entire rule.
**Actual:** Reports success but rule is unchanged.

### Bug 2: DROP MODULE ROLE does not cascade-delete access rules

```sql
-- Setup: create role and grant access
create module role F1.TestRole;
grant F1.TestRole on F1.Season (read *);

-- Verify rule exists
show access on F1.Season;
-- Output includes: Rule 2: F1.TestRole, Rights: READ

-- Drop role WITHOUT revoking first
drop module role F1.TestRole;
-- Output: "Dropped module role: F1.TestRole"

-- Check: access rule is STILL PRESENT (orphaned)
show access on F1.Season;
-- Output includes: Rule 2: F1.TestRole, Rights: READ
-- F1.TestRole no longer exists as a module role!
```

**Expected:** `drop module role` should cascade-delete all entity access rules, page access, and microflow access that reference the dropped role.
**Actual:** Role is dropped but access rules referencing it remain as orphans.

**Note:** For single-role orphans, a subsequent `revoke F1.TestRole on F1.Season` DOES work (even though the role doesn't exist). However, for multi-role orphans (Bug 1), there is no way to remove them via mxcli.

### Comparison: mxcli-created rules work correctly

```sql
-- GRANT creates single-role rules
create module role F1.RoleA;
create module role F1.RoleB;
grant F1.RoleA on F1.Season (read *);
grant F1.RoleB on F1.Season (read *);

-- SHOW ACCESS shows two separate rules:
-- Rule 2: F1.RoleA, Rights: READ
-- Rule 3: F1.RoleB, Rights: READ

-- REVOKE works correctly on these:
revoke F1.RoleA on F1.Season;
-- Rule 2 is properly removed

-- REVOKE also works on orphaned single-role rules:
drop module role F1.RoleB;  -- leaves orphan
revoke F1.RoleB on F1.Season;
-- Rule 3 is properly removed
```

## Suggested Fixes

### Fix 1: REVOKE should handle multi-role access rules

When `revoke F1.User on F1.Entity` encounters a rule like `[F1.User, F1.UserRole]`:
- Remove `F1.User` from the rule's role list
- If the rule has no remaining roles, delete the entire rule
- If the rule still has other roles, keep it with the remaining roles

### Fix 2: DROP MODULE ROLE should cascade-delete

`drop module role F1.RoleName` should:
1. Remove the role from all entity access rules (same as REVOKE for each entity)
2. Remove the role from all page access grants
3. Remove the role from all microflow access grants
4. Then delete the role itself

Alternatively, add a `cascade` option: `drop module role F1.RoleName cascade`

### Fix 3: REVOKE should not report success when nothing changed

If REVOKE cannot find or modify a matching rule, it should report a warning or error rather than "Revoked access on ... from ..."
