# Microflow Call Web Service Statement

Status: Draft

## Summary

Add MDL support for legacy Mendix SOAP `Microflows$CallWebServiceAction`.

```mdl
$Root = call web service 'SampleSOAP.OrderService'
operation 'FetchSampleItems'
send mapping 'SampleSOAP.OrderRequest'
receive mapping 'SampleSOAP.OrderResponse'
timeout 30;

$Root = call web service 'dangling-service-id'
operation 'FetchSampleItems'
send mapping 'dangling-send-mapping-id'
receive mapping 'dangling-receive-mapping-id';

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
    ;
```

## Design Notes

The structured form prefers stable qualified names for the imported web service
and mapping references. During `describe`, mxcli resolves known
`WebServices$ImportedWebService`, `ExportMappings$ExportMapping`, and
`ImportMappings$ImportMapping` IDs through the backend and emits
`Module.DocumentName`.

If a reference is dangling or the backend cannot resolve it, mxcli deliberately
falls back to the raw ID string so unsupported legacy projects still round-trip.

The `raw` form is an explicit escape hatch. Its string is base64-encoded BSON
for the complete action payload and is authoritative when re-executed. It exists
so unsupported SOAP fields can be preserved byte-for-byte until the structured
syntax covers them.

## Tests And Examples

- Parser/visitor coverage for structured and raw forms.
- Builder/writer coverage for real `WebServiceCallAction` construction and raw
  BSON preservation.
- Formatter coverage for qualified-name resolution and raw-ID fallback.
- Example script: `mdl-examples/doctype-tests/call_web_service.test.mdl`.

## Resolved Questions

- Service and mapping references are emitted as `Module.Document` names when
  the backend can resolve them. Raw IDs remain the fallback for dangling
  references and incomplete project metadata.

## Open Questions

- Should the raw payload eventually move to a generic
  `raw microflow action '...'` escape hatch instead of remaining under
  `call web service raw`?
