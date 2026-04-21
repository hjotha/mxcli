# mxcli Parser Bug Report: Validation Microflows

## Summary

When creating validation microflows using MDL, two significant parser issues prevent proper microflow generation:

1. **IF/THEN block inversion** - Code intended for THEN blocks is placed in ELSE blocks
2. **VALIDATION FEEDBACK not recognized** - The statement is not parsed as a valid action inside IF blocks

## Environment

- Project: MesDemoApp.mpr
- Mendix Version: 11.6.0
- mxcli: Local binary (`./mxcli`)

---

## Bug 1: IF/THEN Block Inversion

### Description

When an IF statement contains actions in the THEN block, the parser incorrectly places those actions in an ELSE block instead, leaving the THEN block empty.

### Input MDL

```mdl
create microflow MES.VAL_Product (
  $Product: MES.Product
)
returns boolean as $IsValid
begin
  declare $IsValid boolean = true;

  if trim($Product/Code) = '' then
    log info node 'Validation' 'Code is empty';
    set $IsValid = false;
  end if;

  return $IsValid;
end;
/
```

### Expected Output

The microflow should have:
- THEN branch: Contains LOG and SET actions
- No ELSE branch

### Actual Output

```mdl
create microflow MES.VAL_Product (
  $Product: MES.Product
)
returns boolean as $IsValid
begin
  declare $IsValid boolean = true;
  if trim($Product / Code) = '' then
  else
    log info node 'Validation' 'Code is empty';
    set $IsValid = false;
  end if;
  return $IsValid;
end;
/
```

### Impact

- **Logic is completely inverted**: When Code IS empty, nothing happens. When Code is NOT empty, it's marked invalid.
- Makes validation microflows unusable without manual correction in Studio Pro

### Reproduction Steps

1. Create any MDL file with an IF/THEN block containing 2+ actions
2. Execute with `./mxcli -p MesDemoApp.mpr -c "execute script 'file.mdl'"`
3. Describe the created microflow to see the inverted logic

---

## Bug 2: VALIDATION FEEDBACK Not Recognized

### Description

The `validation feedback` statement is not recognized as a valid action inside IF blocks. The parser expects `else`, `end`, or `elsif` but encounters `validation`.

### Input MDL

```mdl
create microflow MES.VAL_Product (
  $Product: MES.Product
)
returns boolean as $IsValid
begin
  declare $IsValid boolean = true;

  if trim($Product/Code) = '' then
    set $IsValid = false;
    validation feedback $Product/Code message 'Product code cannot be empty';
  end if;

  return $IsValid;
end;
/
```

### Parser Errors

```
line 18:4 mismatched input 'VALIDATION' expecting {else, end, elsif}
line 18:38 extraneous input 'MESSAGE' expecting {<EOF>, DOC_COMMENT, create, ...}
```

### Expected Behavior

`validation feedback` should be recognized as a valid statement that can appear:
- Inside IF/THEN blocks
- Inside IF/ELSE blocks
- At the top level of BEGIN/END

### Syntax Variants Tested

All of these fail with the same error:

```mdl
-- With MESSAGE keyword (as documented in skills)
validation feedback $Product/Code message 'Error message';

-- Without MESSAGE keyword
validation feedback $Product/Code 'Error message';
```

### Impact

- Cannot create validation microflows that show field-level error messages
- Users must manually add ValidationFeedback actions in Studio Pro

---

## Additional Observations

### Operators in Nested Conditions

The `!=` and `<=` operators generate parser warnings but seem to work in simple contexts:

```
line 27:32 extraneous input '!=' expecting {...}
line 28:34 extraneous input '<=' expecting {...}
```

These warnings appear when used in nested IF statements but the operators may still function.

---

## Requested Fixes

### Priority 1: IF/THEN Block Inversion

This is critical - it makes all conditional logic unusable.

**Fix needed**: Ensure actions following `then` are placed in the true branch, not the false branch.

### Priority 2: VALIDATION FEEDBACK Statement

This is essential for validation microflows.

**Fix needed**: Add `validation feedback` to the grammar as a valid action statement that can appear anywhere a regular action (like SET, LOG, etc.) can appear.

**Syntax to support** (as documented in validation-microflows.md skill):

```mdl
-- Basic form
validation feedback $Variable/attribute message 'Error message';

-- With template arguments
validation feedback $Variable/attribute message '{1}' objects [$MessageVar];
```

---

## Workaround (Current)

Until fixed, validation microflows must be:
1. Created with basic structure via mxcli
2. Manually edited in Studio Pro to:
   - Fix inverted IF logic
   - Add ValidationFeedback actions

---

## Test Case

A complete test case is available at: `mdlsource/val-product.mdl`

To reproduce:
```bash
./mxcli -p MesDemoApp.mpr -c "execute script 'mdlsource/val-product.mdl'"
./mxcli -p MesDemoApp.mpr -c "describe microflow MES.VAL_Product"
```
