# Ellipsis Placeholder Expression

Status: Draft

## Summary

Add a single-token expression `...` that represents an unbound /
intentionally-empty argument value in microflow call statements.

```mdl
$Total = call java action SampleModule.Recalculate(
  CompanyId       = ...,
  RecalculateAll  = true,
  ItemList        = ...
);
```

`...` produces a parameter binding with an empty `Argument` string in the
serialized BSON (`Microflows$BasicCodeActionParameterValue.Argument = ""`).
Re-executing a script that contains `...` reproduces the same empty
binding byte-for-byte, so describe → exec → describe stays symmetric for
existing Studio Pro projects that have unbound code-action parameters.

## Motivation

Studio Pro's Java-action call dialog allows a developer to leave individual
parameters empty — for example, when a Java action declares a parameter that
the calling microflow does not yet have a meaningful value for, or when an
external mapping is expected to fill the slot at runtime. The on-disk
representation is a `Microflows$JavaActionParameterMapping` whose `Value` is a
`BasicCodeActionParameterValue` with `Argument: ""`.

Before `...` existed, the describer had two options for these empty bindings:

1. Emit `''` (empty string literal). On re-exec, the visitor would round-trip
   to a non-empty single-quote literal whose `Argument` was `''`, not `""`,
   and Studio Pro would render the parameter as the literal string `''`.
2. Drop the parameter entirely. Studio Pro would then add a back a
   placeholder mapping with a generated value, breaking the round-trip.

Both lose information. `...` lets the describer round-trip the empty binding
without inventing a fake value.

## Syntax

```antlr
atomicExpression
    : literal
    | ELLIPSIS
    | ...
    ;
```

Where `ELLIPSIS` is the lexer token `'...'`. The token is reserved for this
single use; it is not valid in arithmetic / boolean / comparison
expressions.

## Semantics

- `...` is recognised by the builder via `isPlaceholderExpression` in
  `mdl/executor/cmd_microflows_builder_calls.go`.
- Inside a Java-action `callArgument`, `...` produces a
  `BasicCodeActionParameterValue` with `Argument: ""`.
- Outside of `callArgument` lists, `...` parses but the builder rejects it
  (it never resolves to a runtime value). Future statements may extend the
  set of contexts that accept `...` — see Open Questions.

## Examples

```mdl
-- Java action call with two unbound and one bound argument
$Total = call java action SampleModule.Recalculate(
  CompanyId       = ...,
  RecalculateAll  = true,
  ItemList        = ...
);
```

The Mendix BSON for the unbound arguments is:

```
JavaActionParameterMapping {
  Parameter: 'SampleModule.Recalculate.CompanyId',
  Value: BasicCodeActionParameterValue { Argument: '' }
}
```

## Tests And Examples

- Builder coverage: `TestBuildJavaAction_PlaceholderArgumentPreservesEmptyBasicValue`
  in `mdl/executor/cmd_microflows_builder_java_action_test.go`.
- Visitor coverage: `atomicExpression`'s `ELLIPSIS` arm produces
  `ast.SourceExpr{Source: "..."}` (see
  `mdl/visitor/visitor_microflow_expression.go`).
- Example script: `mdl-examples/doctype-tests/ellipsis_placeholder.test.mdl`.

## Open Questions

- Should `...` be allowed as an argument to `call microflow` and
  `call nanoflow` calls as well? Today only Java actions consume the
  `BasicCodeActionParameterValue` form, so there is no symmetric BSON
  representation, but a future proposal could extend this.
- Should we explicitly document `...` as round-trip-only and warn the linter
  when an authored microflow uses `...` outside of a known describe-emitted
  context? This would prevent users from authoring scripts that produce
  Studio Pro warnings ("unbound parameter") on import.
- Should the surface syntax be `...` or a clearer keyword like
  `unspecified` / `default` / `unbound`? `...` was chosen because it is
  visually distinct, short, and matches existing "this is intentionally
  blank" conventions in other tools (Python's `Ellipsis`, TypeScript's
  `never` placeholder, etc.).
