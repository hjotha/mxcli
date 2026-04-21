# Bug Report: mxcli fails to create microflow with cross-module entity parameter

## Summary

When creating a microflow with a parameter whose type is an entity from a different module (e.g., `FeedbackModule.Feedback` in a microflow in `MyFirstModule`), mxcli fails with a misleading error message.

## Environment

- **mxcli version**: Current (as of 2026-01-19)
- **Mendix version**: 11.6.0
- **Platform**: macOS (Darwin 25.2.0)

## Steps to Reproduce

1. Connect to a Mendix project:
   ```bash
   ./mxcli -p MesDemoApp.mpr
   ```

2. Try to create a microflow with a cross-module entity parameter:
   ```sql
   create microflow MyFirstModule.TestMicroflow (
     $feedback: FeedbackModule.Feedback
   )
   returns boolean as $IsValid
   begin
     declare $IsValid boolean = true;
     return $IsValid;
   end;
   /
   ```

## Expected Behavior

The microflow should be created with:
- Parameter `$feedback` of type `FeedbackModule.Feedback`
- Return type `boolean`

## Actual Behavior

The command fails with the error:
```
error: entity '.FeedbackModule' not found for parameter 'Feedback'
```

Additionally, parser warnings are shown:
```
line 2:27 mismatched input '.' expecting ')'
```

### Analysis

The error message `entity '.FeedbackModule' not found` suggests the parser is incorrectly splitting the qualified entity name `FeedbackModule.Feedback`:
- It appears to interpret `FeedbackModule` as `.FeedbackModule` (with a leading dot)
- The entity name `feedback` is being parsed separately

The parser seems to expect the parameter type to end at the module name, treating the `.` as an unexpected token.

## Workaround Attempts

### Attempt 1: Unqualified entity name
```sql
$feedback: feedback
```
**Result**: Microflow is created, but parameter type becomes `Void` instead of `FeedbackModule.Feedback`.

### Attempt 2: Single-line command
```bash
./mxcli -p MesDemoApp.mpr -c "create microflow MyFirstModule.TestParam(\$feedback: FeedbackModule.Feedback) returns boolean as \$IsValid begin declare \$IsValid boolean = true; return \$IsValid; end; /"
```
**Result**: Same error - `entity '.FeedbackModule' not found`

### Attempt 3: Script file execution
```bash
./mxcli -p MesDemoApp.mpr -c "execute script '/tmp/test.mdl'"
```
**Result**: Same error

## Evidence

The `mxcli check` command reports syntax as valid:
```
Checking syntax: /tmp/val_feedback_simple.mdl
✓ Syntax OK (1 statements)
```

But execution still fails, indicating a disconnect between the syntax checker and the executor.

## Comparison with Working Cases

Creating a microflow **without** entity parameters works fine:
```sql
create microflow MyFirstModule.TestMicroflow()
returns boolean as $IsValid
begin
  declare $IsValid boolean = true;
  return $IsValid;
end;
/
```
**Result**: `created microflow: MyFirstModule.TestMicroflow`

## Impact

This bug prevents using mxcli to create microflows that:
- Accept entity parameters from other modules
- Are common patterns like validation microflows (VAL_*) in custom modules that validate Marketplace module entities

## Suggested Fix

The parameter type parser should correctly handle fully qualified entity names in the format `Module.Entity`. The `.` should be recognized as a namespace separator, not treated as an unexpected token.

## Related

The skill documentation in `.claude/skills/write-microflows.md` shows the syntax:
```mdl
-- Entity types
$Customer: Module.Entity
```

This confirms the intended syntax supports qualified entity names.
