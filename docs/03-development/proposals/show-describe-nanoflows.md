# Proposal: DESCRIBE Nanoflow (Enhancement)

## Overview

**Document type:** `Microflows$Nanoflow`
**Prevalence:** 227 across test projects (79 Enquiries, 97 Evora, 51 Lato)
**Priority:** High — nanoflows are heavily used, SHOW works but DESCRIBE is missing

Nanoflows execute client-side in the browser or native app. They share the same BSON structure as microflows but run on the client. Currently `SHOW NANOFLOWS` works but `DESCRIBE NANOFLOW` does not.

## What Already Exists

| Layer | Status | Location |
|-------|--------|----------|
| **Go type** | Yes | `sdk/microflows/microflows.go` — `Nanoflow` struct |
| **Parser** | Yes | `sdk/mpr/parser_nanoflow.go` — full parsing |
| **Reader** | Yes | `ListNanoflows()`, `GetNanoflow()` |
| **SHOW** | Yes | `showNanoflows()` in executor |
| **DESCRIBE** | **No** | No `DescribeNanoflow` AST type or handler |
| **DROP** | **No** | No `DropNanoflowStmt` |

## What's Missing

The infrastructure is all there — nanoflows are fully parsed. The gap is purely in the AST/Grammar/Executor wiring for DESCRIBE (and DROP).

## Proposed MDL Syntax

### DESCRIBE NANOFLOW

```
DESCRIBE NANOFLOW Module.Name
```

Output format (same as DESCRIBE MICROFLOW but with NANOFLOW keyword):

```
/**
 * Validates the customer form before saving
 */
NANOFLOW MyModule.ValidateCustomerForm
  PARAMETER $Customer: MyModule.Customer
  RETURNS Boolean
BEGIN
  IF $Customer/Name = '' THEN
    VALIDATION FEEDBACK $Customer ATTRIBUTE Name MESSAGE 'Name is required';
    RETURN false;
  END IF;
  RETURN true;
END;
/
```

### DROP NANOFLOW (optional, lower priority)

```
DROP NANOFLOW Module.Name;
```

## Implementation Steps

### 1. Add AST Types (mdl/ast/ast_query.go)

```go
// In DescribeObjectType enum:
DescribeNanoflow

// Add String() case:
case DescribeNanoflow:
    return "NANOFLOW"
```

For DROP (optional):
```go
// In ast_microflow.go or similar:
type DropNanoflowStmt struct {
    Name QualifiedName
}
```

### 2. Add Grammar Rules (MDLParser.g4)

The grammar likely already has `DESCRIBE NANOFLOW` syntax — the visitor just needs to wire it to a new AST type instead of silently ignoring it.

### 3. Add Visitor Mapping

Map `DESCRIBE NANOFLOW qualifiedName` to `DescribeStmt{ObjectType: DescribeNanoflow}`.

### 4. Add Executor Handler (mdl/executor/cmd_microflows_show.go or similar)

```go
func (e *Executor) describeNanoflow(name ast.QualifiedName) error {
    // Reuse describeMicroflow logic but look up nanoflows instead
    nanoflows, err := e.reader.ListNanoflows()
    // ... find by qualified name ...
    // Output using same formatter as microflows, with "NANOFLOW" keyword
}
```

The key insight is that the DESCRIBE formatter for microflows can be reused — just change the header keyword from `MICROFLOW` to `NANOFLOW`.

### 5. Wire into Executor Dispatcher

```go
case ast.DescribeNanoflow:
    return e.describeNanoflow(s.Name)
```

## Complexity

**Very low** — all infrastructure exists. This is purely a wiring task:
- 1 new AST constant
- 1 new grammar rule mapping (or fix existing)
- 1 new executor handler that delegates to microflow formatter
- ~30 lines of code total

## Testing

- Add `DESCRIBE NANOFLOW` examples to existing test files
- Verify against all 3 test projects
