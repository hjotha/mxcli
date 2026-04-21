# Bug Report: `mxcli test` crashes when testing microflows with VALIDATION FEEDBACK

## Summary

The `mxcli test` runner crashes the Mendix runtime when a test calls a microflow that uses `validation feedback`. After the first test that triggers validation feedback passes its assertion, the runtime crashes before the next test can execute. This makes it impossible to test validation microflows that use `validation feedback` beyond a single negative test case.

## Environment

- **mxcli version**: Latest (Feb 2026)
- **Mendix version**: 11.6.3
- **Platform**: Linux (aarch64, Docker)

## Steps to Reproduce

### 1. Create a validation microflow that uses VALIDATION FEEDBACK

```mdl
create microflow Formula1.VAL_Driver_NewEdit (
  $Driver: Formula1.Driver
)
returns boolean as $IsValid
folder 'Validation'
begin
  declare $IsValid boolean = true;

  if trim($Driver/Forename) = '' then
    set $IsValid = false;
    validation feedback $Driver/Forename message 'Forename is required';
  end if;

  if trim($Driver/Surname) = '' then
    set $IsValid = false;
    validation feedback $Driver/Surname message 'Surname is required';
  end if;

  return $IsValid;
end;
/
```

### 2. Create a test file with multiple negative test cases

```mdl
/**
 * @test Empty forename fails validation
 * @expect $result1 = false
 * @cleanup none
 */
$driver1 = create Formula1.Driver (Forename = '', Surname = 'Hamilton');
$result1 = call microflow Formula1.VAL_Driver_NewEdit($Driver = $driver1);
delete $driver1;
/

/**
 * @test Empty surname fails validation
 * @expect $result2 = false
 * @cleanup none
 */
$driver2 = create Formula1.Driver (Forename = 'Lewis', Surname = '');
$result2 = call microflow Formula1.VAL_Driver_NewEdit($Driver = $driver2);
delete $driver2;
/
```

### 3. Run the tests

```bash
./mxcli test tests/driver-validation.test.mdl -p Formula1Demo.mpr --color
```

## Expected Behavior

Both tests should pass: the validation microflow correctly returns `false` for invalid input, and the `@expect` assertions match.

## Actual Behavior

- **Test 1** (Empty forename): **PASS** - The assertion is captured correctly.
- **Test 2** (Empty surname): **ERROR** - "Test was not executed (runtime may have crashed before reaching it)"

The Mendix runtime crashes after the first test that triggers `validation feedback`. The crash occurs because `validation feedback` is a UI-oriented action that does not work correctly in the headless after-startup context used by the test runner.

The pending validation feedback on the entity object causes the runtime to throw an exception (e.g., `object id: X, validation errors: (member: Forename, message: )`) before subsequent test activities can execute.

## Workaround

Reorder tests so that:
1. All "positive" tests (valid input, no VALIDATION FEEDBACK triggered) run first
2. Place only ONE "negative" test (triggers VALIDATION FEEDBACK) at the very end

This allows the final negative test's assertion to be captured before the crash, since there are no subsequent tests.

**Result with workaround**: 4/8 tests pass instead of 2/8.

## Attempted Mitigations (None Worked)

| Approach | Result |
|----------|--------|
| `delete $entity` after validation call | Still crashes |
| `rollback $entity` after validation call | ROLLBACK itself triggers the crash |
| `@cleanup none` annotation | No effect |
| Unique variable names per test | Fixed a separate issue, but didn't fix this crash |

## Suggested Fix

The test runner should isolate VALIDATION FEEDBACK side effects between tests. Possible approaches:

1. **Wrap each test block in error handling** so that validation feedback exceptions don't propagate to subsequent tests
2. **Clear validation feedback state** on entity objects between test executions
3. **Run each test in a separate transaction/context** to prevent state leakage
4. **Suppress VALIDATION FEEDBACK actions** in test mode (capture the feedback intent but don't execute the UI-side action)

## Secondary Bug: Segfault during cleanup

After tests complete, `mxcli test` attempts to restore the after-startup setting. This consistently crashes with a nil pointer dereference:

```
warning: could not restore after-startup: exit status 2: panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x58 pc=0x83a134]

goroutine 1 [running]:
github.com/mendixlabs/mxcli/mdl/visitor.(*Builder).ExitAlterSettingsClause(...)
  /users/andrej.koelewijn/GitHub/ModelSDKGo/mdl/visitor/visitor_settings.go:48 +0x9f4
```

This appears to be a nil pointer in `visitor_settings.go:48` when there was no previous after-startup microflow to restore.

## Test Results Output

```
Test Results: driver-validation
============================================================
  PASS  Valid driver passes validation (0s)
  PASS  Valid 3-char code passes validation (0s)
  PASS  empty code is acceptable (optional field) (0s)
  PASS  empty forename fails validation
  error  empty surname fails validation
         Test was not executed (runtime may have crashed before reaching it)
  error  Whitespace-only forename fails validation
         Test was not executed (runtime may have crashed before reaching it)
  error  Invalid code length (2 chars) fails validation
         Test was not executed (runtime may have crashed before reaching it)
  error  both names empty fails validation
         Test was not executed (runtime may have crashed before reaching it)
------------------------------------------------------------
Total: 8  Passed: 4  Failed: 4  Skipped: 0
```
