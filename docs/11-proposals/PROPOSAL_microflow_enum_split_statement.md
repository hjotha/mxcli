# Proposal: Microflow ENUM SPLIT Statement

Status: Draft

## Summary

Add round-trip MDL support for enumeration decisions:

```mdl
split enum $Status
  case Open, Pending
    return true;
  case (empty)
    return false;
  else
    return false;
end split;
```

## Motivation

Studio Pro represents enumeration decisions as exclusive splits whose outgoing sequence flows carry enumeration case values. Without a first-class MDL statement, describe/exec round-trips collapse those structures into boolean-looking decisions or unsupported comments.

## Semantics

`split enum` evaluates an enumeration variable or attribute path. Each `case` lists one or more enumeration values that enter the same branch. `(empty)` represents the Mendix empty enumeration case. `else` is optional and maps to the outgoing flow without an explicit case value.

## Tests And Examples

`mdl-examples/doctype-tests/enum_split_statement.test.mdl` demonstrates parser syntax. Go regression tests cover AST parsing, builder generation of enumeration case flows, and describer output for existing split graphs.

## Open Questions

- Should the builder validate case values against the referenced enumeration when backend metadata is available?
- Should enum value names be emitted fully qualified in ambiguous cross-module cases?
