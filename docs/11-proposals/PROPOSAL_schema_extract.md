# Proposal: `mxcli schema extract` — Empirical Metamodel Schema Extraction

## Problem

The BSON Schema Registry proposal (see `BSON_SCHEMA_REGISTRY_PROPOSAL.md`) requires accurate
per-version type metadata to drive serialization, validation, and default-value filling. That
metadata currently comes from two sources, both with reliability problems:

1. **Reflection data** (`reference/mendixmodellib/reflection-data/`) — extracted from the
   TypeScript `mendixmodelsdk` npm package. Stale at Mendix 11.6. Requires manual update each
   release. Uses TypeScript SDK names that differ from BSON storage names (e.g. SDK says
   `CreateObjectAction`, BSON stores `CreateChangeAction`).

2. **`supplements.json`** (PR #335) — hand-maintained bridge between SDK type names and BSON
   reality. Grows with every release. Gaps are discovered at runtime when Studio Pro rejects
   written output.

Neither source is self-updating, and neither is verifiable against actual Studio Pro behavior
without opening a project and checking for errors.

## Insight

When Studio Pro opens a Mendix project, its MCP server exposes the full metamodel through the
PED (Progressive Element Discovery) API. By combining three observations:

1. `ped_get_schema` returns the complete property structure and valid enum values for any
   element type — including the exhaustive `allowedTypes` list for every abstract extension
   point (e.g. all 40+ microflow action types).

2. Documents created via PED are serialised to `.mxunit` BSON files on disk immediately.
   Decoding those files reveals the exact `$Type` storage name and BSON field encoding that
   Studio Pro actually writes — no inference needed.

3. The BSON value type is deterministic: a `{"Subtype": N, "Data": "..."}` binary value is a
   BY_ID reference; a plain string is BY_NAME; a leading integer in an array is the list
   encoding type (1 = compact, 2 = key-value, 3 = object array).

These three observations together produce a complete, verifiable schema for every element type
reachable from the Mendix model — sourced directly from Studio Pro rather than from a generated
TypeScript artifact.

## Proposed Command

```
mxcli schema extract [--output dir] [--version label] [--domains domain1,domain2,...]
```

Connects to a running Studio Pro instance via MCP, creates minimal example documents for every
reachable element type, decodes the resulting `.mxunit` files, and writes a structured schema
JSON file.

```bash
# Extract schema for whatever Studio Pro version is currently open
mxcli schema extract --output reference/mendixmodellib/reflection-data/

# Extract only microflow and domain model domains
mxcli schema extract --domains microflows,domainmodels --output ./schemas/

# Dry run: print coverage report without writing files
mxcli schema extract --dry-run
```

## Extraction Paths

The command uses four distinct extraction paths depending on the element category.

### Path 1: Microflow and nanoflow activities

**Enumeration**: `ped_get_schema(["Microflows$ActionActivity"])` returns the full
`allowedTypes` map for the `action` property — the authoritative list of all action types
(currently 40+).

**Schema extraction**: For each action type:
1. Create a minimal `Microflows$Microflow` in a scratch module with one `ActionActivity`
   containing that action, plus required Start/End events.
2. Decode the resulting `.mxunit` file.
3. Walk every field in the decoded BSON and classify:
   - `{"Subtype": N, "Data": "..."}` → `referenceKind: BY_ID`
   - plain string → `referenceKind: BY_NAME`
   - dict with `$Type` → `referenceKind: PART`
   - `[1, ...]` → `listEncoding: compact`
   - `[2, ...]` → `listEncoding: keyValue`
   - `[3, ...]` → `listEncoding: objectArray`
4. Record the `$Type` storage name (which may differ from the PED type name, e.g.
   `Microflows$CreateObjectAction` → `Microflows$CreateChangeAction`).

**Nanoflow subset**: Repeat with `Microflows$Nanoflow`. Actions that are disallowed in
nanoflows fail `ped_check_errors` with a recognisable error; all others produce a valid schema.
The disallowed set is recorded as metadata on each action type.

**Preconditions**: Some action types require an existing entity, microflow, or page. The
extractor creates stubs automatically (a `_SchemaExtract` scratch module with a bare entity, a
trivial nanoflow, etc.) and tears them down after extraction.

### Path 2: Built-in widgets (`Forms$` types)

**Enumeration**: The 78 `Forms$` widget types are known from the existing codebase. PED's page
schema does not enumerate them; instead the extractor uses the known list as its input.

**Schema extraction**: For each widget type:
1. Use the pagegen tools to create a minimal page containing one instance of the widget.
2. Decode the resulting `.mxunit` file.
3. Apply the same field classification as Path 1.

Because the pagegen API accepts widget types by name, this is fully automatable without manual
page construction.

### Path 3: Pluggable widgets (`.mpk` packages)

**Enumeration**: Scan the project's `widgets/` directory for `.mpk` files.

**Schema extraction**: Each `.mpk` is a ZIP archive. Extract `{WidgetName}.xml` — a structured
property definition that is the *source* from which Studio Pro generates the BSON
`CustomWidgetType` PropertyTypes blob. Parse it directly:

```xml
<property key="source" type="enumeration" defaultValue="context" required="true">
<property key="attributeEnumeration" type="attribute" required="true">
<property key="optionsSourceDatabaseDataSource" type="datasource" isList="true">
```

This is preferable to the mxunit roundtrip for widgets because the XML is the canonical
definition — decoding BSON would only recover a derivative representation of it.

The extractor records the widget ID, version, and the full property tree. This replaces the
current `sdk/widgets/templates/` embedded templates with a live, project-accurate schema.

### Path 4: Domain model and other document types

**Enumeration**: `ped_get_schema` for all top-level document types and their nested element
types, traversing the schema graph from each document root.

**Schema extraction**: Create a minimal instance of each type (entity, enumeration,
association, etc.), decode the mxunit, and classify fields using the same algorithm.

## Output Format

The extractor writes one JSON file per Mendix version:

```
reference/mendixmodellib/reflection-data/
  11.9.0-extracted.json    ← new format, one file replaces -structures + -storageNames pair
  11.8.0-extracted.json
  ...
```

Schema JSON structure (one entry per element type):

```json
{
  "version": "11.9.0",
  "extractedAt": "2026-05-02T14:00:00Z",
  "types": {
    "Microflows$CreateChangeAction": {
      "pedName": "Microflows$CreateObjectAction",
      "storageName": "Microflows$CreateChangeAction",
      "properties": {
        "Entity": {
          "referenceKind": "BY_NAME",
          "referredType": "DomainModels$Entity",
          "listEncoding": null,
          "default": ""
        },
        "Items": {
          "referenceKind": "PART",
          "referredType": "Microflows$ChangeActionItem",
          "listEncoding": "keyValue",
          "default": []
        },
        "Commit": {
          "referenceKind": null,
          "referredType": null,
          "listEncoding": null,
          "enumValues": ["Yes", "YesWithoutEvents", "No"],
          "default": "No"
        }
      },
      "allowedInNanoflow": false,
      "introducedIn": null,
      "removedIn": null
    }
  },
  "widgets": {
    "com.mendix.widget.web.combobox.Combobox": {
      "version": "2.5.0",
      "source": "mpk",
      "properties": { ... }
    }
  }
}
```

## Version Lifecycle

A single extraction run produces the schema for the version currently open in Studio Pro. To
populate version lifecycle fields (`introducedIn`, `removedIn`), run extraction against
multiple Studio Pro versions and diff the outputs:

```bash
# Run against Studio Pro 10.24, then 11.9
mxcli schema extract --version 10.24.0 --output schemas/
mxcli schema extract --version 11.9.0  --output schemas/

# Diff to compute property lifecycle
mxcli schema diff schemas/10.24.0-extracted.json schemas/11.9.0-extracted.json
```

This is the same information the `.version.mxcore` files encode in the internal appdev
monorepo — but derived empirically rather than from a canonical source. For most practical
purposes (knowing whether a property exists for a given project version) this is sufficient.

## Relationship to Existing Proposals

### BSON Schema Registry (PR #335)

The schema registry proposal requires accurate per-version type metadata but doesn't address
how to keep that metadata current. `mxcli schema extract` is the data pipeline that feeds the
registry. The registry's `LoadRegistry(version)` function would load `{version}-extracted.json`
instead of the current `-structures.json` + `-storageNames.json` pair.

### PR #335 `supplements.json`

Most entries in `supplements.json` exist because the JS-parsed SDK doesn't capture BSON storage
names, list encodings, or reference kinds. The extracted schema contains all of these directly.
Once extraction covers the relevant domains, `supplements.json` shrinks to only the genuinely
exotic cases.

### retran's `.mxcore` suggestion (PR #335 review)

The `.mxcore` DSL in the internal appdev monorepo is the canonical source for the same
metadata. `mxcli schema extract` is the external-contributor-accessible equivalent: it derives
the same structural facts empirically. If Mendix publishes a minimal derived artifact from
`.mxcore`, the extraction pipeline could be simplified or replaced; the output format is
designed to be compatible with either source.

## Coverage

| Domain | Types | Extraction path | Completeness |
|---|---|---|---|
| Microflow actions | 40+ | PED allowedTypes + mxunit | Full |
| Nanoflow actions | subset of above | Try + ped_check_errors | Full |
| Domain model | ~15 | PED schema traversal + mxunit | Full |
| Built-in widgets | 78 | Known list + pagegen + mxunit | Full |
| Pluggable widgets | project-dependent | `.mpk` XML parse | Full (per project) |
| Other document types | ~40 domains | PED schema traversal + mxunit | Partial (reachable types only) |

## Implementation Plan

### Phase 1: Core extractor (microflow actions)
- Connect to Studio Pro MCP, read PED type catalog for `ActionActivity`
- Create scratch module, create one microflow per action type
- Decode mxunit, classify fields, write `{version}-extracted.json`
- Validate: compare extracted storage names against current CLAUDE.md table

### Phase 2: Domain model and nanoflow subset
- Extend to domain model types (entity, attribute, association, enumeration)
- Extend to nanoflow: run same list, record disallowed set
- Add precondition scaffold (stub entity, stub nanoflow, etc.)

### Phase 3: Built-in widgets
- Implement pagegen → mxunit decode path for `Forms$` widget types
- Add widget fields to extracted schema

### Phase 4: Pluggable widget `.mpk` parsing
- Implement ZIP + XML parser for widget property definitions
- Record widget ID, version, property tree
- Emit as `widgets` section of extracted JSON

### Phase 5: Integration with schema registry
- Update `LoadRegistry` to accept the new extracted format
- Remove or shrink `supplements.json` based on extraction coverage
- Add `mxcli schema diff` for version lifecycle computation

## Open Questions

1. **MCP availability**: The extractor requires a running Studio Pro instance. Should there be
   a fallback that uses the existing reflection data when no MCP connection is available, or
   should extraction always be an explicit developer operation?

2. **Scratch module cleanup**: The extractor creates a `_SchemaExtract` module and deletes it
   after. What happens if extraction is interrupted mid-run? Should it be idempotent (detect
   and reuse an existing scratch module)?

3. **Frequency**: Schema extraction only needs to run when a new Studio Pro version is
   targeted. Should this be a manual developer command, a CI step, or triggered automatically
   on `mxcli setup mxbuild`?

4. **Non-extractable types**: Some types can't be instantiated without complex preconditions
   (published REST services, business events, OData contracts). How should the extractor handle
   types it cannot reach? Flag as `"source": "manual"` and fall back to existing reflection
   data for those domains.
