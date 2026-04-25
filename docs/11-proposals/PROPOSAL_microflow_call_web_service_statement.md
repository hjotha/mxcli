# Microflow Call Web Service Statement

Status: Draft

## Summary

Add MDL support for legacy Mendix SOAP `Microflows$CallWebServiceAction`.

```mdl
$Root = call web service 'sample-service-id'
operation 'FetchSampleItems'
send mapping 'sample-send-mapping-id'
receive mapping 'sample-receive-mapping-id'
timeout 30;

$Root = call web service raw 'AQID';
```

This proposal is primarily about safe round-trip preservation of existing SOAP
actions. New integrations should prefer consumed REST services or inline REST
calls.

## Motivation

Legacy projects can contain SOAP web service calls. Without an MDL
representation, describe output either drops the activity or emits an
unsupported-action comment that cannot be re-executed into the same model.

The immediate goal is therefore fidelity:

- Parse existing `CallWebServiceAction` BSON.
- Emit an MDL statement that can be executed back into the MPR.
- Preserve unsupported or version-specific BSON fields when the structured
  fields are incomplete.

## Syntax

```antlr
callWebServiceStatement
    : (VARIABLE EQUALS)? CALL WEB SERVICE
      (RAW STRING_LITERAL
      | STRING_LITERAL
        (OPERATION STRING_LITERAL)?
        (SEND MAPPING STRING_LITERAL)?
        (RECEIVE MAPPING STRING_LITERAL)?
        (TIMEOUT expression)?)
      onErrorClause?
      SEMICOLON
    ;
```

## Design Notes

The structured form currently uses Mendix `$ID` values for the service and
mapping references. That is a deliberate passthrough limitation, not the final
authoring design. A future iteration should resolve those IDs to stable
qualified names where the MPR contains enough metadata.

The `raw` form is an explicit escape hatch. Its string is base64-encoded BSON
for the complete action payload and is authoritative when re-executed. It exists
so unsupported SOAP fields can be preserved byte-for-byte until the structured
syntax covers them.

## Tests And Examples

- Parser/visitor coverage: `TestCallWebServiceStatement` and
  `TestCallWebServiceRawStatement`.
- Builder/writer coverage: `TestBuildFlowGraph_WebServiceCallCreatesRealAction`,
  `TestBuildFlowGraph_WebServiceCallPreservesRawBSON`, and MPR RawBSON tests.
- Example script: `mdl-examples/doctype-tests/call_web_service.test.mdl`.

## Open Questions

- Should service and mapping references be resolved to `Module.Document` names
  before this becomes a general authoring feature?
- Should the raw payload eventually move to a generic
  `raw microflow action '...'` escape hatch instead of remaining under
  `call web service raw`?

