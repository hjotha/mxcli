# Microflow Variable Alias Collision

Status: Draft

## Summary

Document the round-trip-only aliasing rule used when `describe` encounters
multiple implicit output variables with the same name at the same microflow
position.

When the builder detects a duplicate implicit output at the same canvas point,
the later output is renamed with a numeric suffix:

```mdl
$Item = create SampleModule.Item ();
$Item = create SampleModule.Item (); -- becomes Item_2 internally
```

References emitted after the aliased activity are rewritten to the generated
name (`$Item_2`, `$Item_3`, and so on) so the generated Mendix model remains
valid.

## Motivation

Some legacy projects contain duplicated or ambiguous implicit output variables
that Studio Pro can keep in the model but MDL cannot represent as repeated
variables in the same scope without ambiguity. Failing the round-trip would
block describe/exec use on those projects. Silently reusing the first variable
would also be wrong because later changes, returns, or association paths would
target the wrong object.

The aliasing rule preserves a valid model while making the generated MDL
deterministic and reviewable.

## Semantics

- Aliasing is position-scoped. A duplicate implicit output only aliases when
  the same variable name is produced at the same `@position(x, y)`.
- The first output keeps its original name.
- Later outputs are renamed to the first available suffix:
  `Foo_2`, `Foo_3`, and so on.
- Subsequent references in variables, paths, and preserved source expressions
  are rewritten to the active alias.
- Moving to a different position resets the alias for that variable name.

This is primarily a describe/round-trip preservation rule. Authored MDL should
prefer explicit unique variable names.

## Tests And Examples

- Builder coverage verifies duplicate implicit outputs are emitted as
  `SelectedItem` and `SelectedItem_2`, and that downstream references follow
  the alias.
- Example script:
  `mdl-examples/doctype-tests/variable_alias_collision.test.mdl`.

## Open Questions

- Should the builder fail with an explicit disambiguation error instead of
  aliasing silently when authored MDL contains this pattern?
- Should `describe` emit a comment near generated aliases so users can
  distinguish model-preservation aliases from names authored manually?
