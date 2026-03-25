# Proposal: SHOW/DESCRIBE Rules

## Overview

**Document type:** `Microflows$Rule`
**Prevalence:** 49 across test projects (9 Enquiries, 28 Evora, 12 Lato)
**Priority:** Medium — decision logic used in microflow split conditions

Rules are structurally identical to Microflows but must return Boolean. They are used in split conditions (exclusive splits) as an alternative to expressions. Rules have parameters, activities, and flows — the same internal structure as microflows.

## What Already Exists

| Layer | Status | Location |
|-------|--------|----------|
| **Go type** | No | — |
| **Parser** | No | Rules share BSON structure with Microflows but no parser exists |
| **Reader** | No | — |
| **Generated metamodel** | Yes | `generated/metamodel/types.go` line 4791 |

## BSON Structure (from test projects)

```
Microflows$Rule:
  Name: string
  Documentation: string
  Excluded: bool
  ExportLevel: string
  ApplyEntityAccess: bool
  MarkAsUsed: bool
  MicroflowReturnType: DataTypes$BooleanType (always Boolean)
  ReturnVariableName: string
  ObjectCollection: Microflows$MicroflowObjectCollection
    Objects: []*MicroflowObject (activities, parameters, etc.)
  Flows: []*Microflows$SequenceFlow
```

The internal structure (ObjectCollection, Flows) is identical to `Microflows$Microflow`.

## Proposed MDL Syntax

### SHOW RULES

```
SHOW RULES [IN Module]
```

| Qualified Name | Module | Name | Parameters | Activities |
|----------------|--------|------|------------|------------|

### DESCRIBE RULE

```
DESCRIBE RULE Module.Name
```

Output format (mirrors DESCRIBE MICROFLOW but with RULE keyword):

```
/**
 * Checks if the customer is eligible for a discount
 */
RULE MyModule.IsEligibleForDiscount
  PARAMETER $Customer: MyModule.Customer
  RETURNS Boolean
BEGIN
  RETRIEVE $orders FROM DATABASE
    WHERE MyModule.Customer_Order/MyModule.Order/Status = 'Completed';
  IF $orders/length > 5 THEN
    RETURN true;
  ELSE
    RETURN false;
  END IF;
END;
/
```

## Implementation Steps

### 1. Add Model Type and Parser

Since Rules share the same BSON structure as Microflows, the existing microflow parser (`parser_microflow.go`) can be reused with minimal changes:

```go
type Rule struct {
    // Same fields as Microflow
    microflows.Microflow // embed or duplicate
}
```

Alternatively, parse Rules as `Microflow` instances with a `Kind: "Rule"` marker.

### 2. Add Reader

```go
func (r *Reader) ListRules() ([]*microflows.Microflow, error) {
    return r.listUnitsByType("Microflows$Rule", parseMicroflow)
}
```

### 3. Add AST, Grammar, Visitor, Executor

Grammar tokens: `RULE` (may already exist), `RULES`.

The DESCRIBE handler can delegate to `describeMicroflow()` internally, with the output keyword changed from `MICROFLOW` to `RULE`.

### 4. Add Autocomplete

```go
func (e *Executor) GetRuleNames(moduleFilter string) []string
```

## Design Decision

**Option A: Reuse Microflow infrastructure** — Parse rules as microflows with a `Kind` field. DESCRIBE outputs `RULE` keyword but reuses the microflow formatter. This is simpler but less explicit.

**Option B: Separate type** — Dedicated `Rule` type and handlers. More code but cleaner separation.

**Recommendation:** Option A — rules ARE microflows with a Boolean return constraint. Reusing the infrastructure minimizes code.

## Testing

- Verify against Evora project (28 rules — most comprehensive)
