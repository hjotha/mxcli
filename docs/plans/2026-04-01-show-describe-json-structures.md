# Plan: SHOW/DESCRIBE/CREATE/DROP JSON Structures

## Context

JSON Structures are a Mendix document type (`JsonStructures$JsonStructure`) used for REST integrations and import/export mappings. They define JSON schemas with a snippet and a parsed element tree that maps JSON fields to types. The proposal (`docs/11-proposals/show-describe-json-structures.md`) requests SHOW/DESCRIBE support; we extend this to also include CREATE/DROP for full read-write support.

The generated metamodel already defines Go types (`JsonStructuresJsonStructure`, `JsonStructuresJsonElement` in `generated/metamodel/types.go`). The lexer already has `json` and `structure` tokens. No parser, reader, executor, or grammar rules exist yet.

## Real-World BSON Reference

Extracted from a real Mendix 11.6.3 MPR v2 project — a paginated API search response JSON structure.

### Top-Level Document Fields

```
$ID: <binary 16 bytes UUID>
$type: "JsonStructures$JsonStructure"
Name: "SearchResponse"
documentation: ""
Excluded: false
ExportLevel: "Hidden"
JsonSnippet: <raw json string>
Elements: [2, ...elements]  // bson.A starts with int version marker (2)
```

### Element Fields (each `JsonStructures$JsonElement`)

```
$ID: <binary 16 bytes UUID>
$type: "JsonStructures$JsonElement"
ExposedName: "Root"            // display name
ExposedItemName: ""            // for arrays: name of individual items
path: "(object)"              // json path using (object)|(Array) notation
ElementType: "object"         // "object", "Array", "value"
PrimitiveType: "Unknown"      // "string", "integer", "boolean", "decimal", "Unknown"
MinOccurs: 1
MaxOccurs: 1                  // -1 = unbounded (arrays)
Nillable: true
IsDefaultType: false
MaxLength: -1                 // -1 = unset
FractionDigits: -1            // -1 = unset
TotalDigits: -1               // -1 = unset
OriginalValue: ""             // Original json sample value
ErrorMessage: ""
WarningMessage: ""
Children: [2, ...children]    // Recursive, same bson.A format with version prefix
```

### Example Element Tree (Paginated API Response)

```
Root: object
  page: integer
  ResultsPerPage: integer
  Total: integer
  Results: Array
    ResultsItem: object[0..*]
      Id: string
      Name: string
      Category: string
      Address: object
        Street: string
        City: string
        ZipCode: string
      _type: string
      Links: Array
        link: object[0..*]
          Rel: string
          Href: string
  Links_2: Array
    Links_2Item: object[0..*]
      Rel: string
      Href: string
```

### Key Observations from Real Data

1. **Version prefix**: `Elements` and `Children` arrays start with an integer `2` before the actual elements — must be skipped during parsing (same pattern as other BSON arrays in this codebase)
2. **Root element**: Always present as the first element, with `path: "(object)"` and `ExposedName: "Root"`
3. **Array items**: An Array element (e.g., `Resultaten`) contains a child Object element with `MaxOccurs: -1` representing the array item
4. **Negative sentinels**: `-1` means "unset" for `MaxLength`, `FractionDigits`, `TotalDigits`, and "unbounded" for `MaxOccurs`
5. **PrimitiveType for containers**: Object and Array elements have `PrimitiveType: "Unknown"`
6. **ExposedName generation**: Mendix derives ExposedName by capitalizing the JSON key (e.g., `name` → `Name`). For arrays, items get suffix `item` (e.g., `results` → `ResultsItem`)
7. **Duplicate key handling**: When JSON has duplicate keys at different levels (e.g., `links`), Mendix appends `_2`, `_3` etc. (e.g., `Links_2`)
8. **Path format**: `(object)|fieldName|(Array)|(object)|nestedField` — path segments use `(object)` and `(Array)` type markers

## Scope

**In scope:** SHOW, DESCRIBE, CREATE, CREATE OR REPLACE, DROP JSON STRUCTURE, autocomplete, catalog table
**Edit flow:** `create or replace json structure` handles updates — drops existing structure and recreates from new snippet. This is idempotent, atomic, and ideal for AI agents that generate the full JSON snippet. No ALTER needed since the element tree is always derived from the snippet.

## Implementation Steps

Following the established pattern from Image Collections (`cmd_imagecollections.go`, `writer_imagecollection.go`) and Business Events (`cmd_businessevents.go`, `writer_businessevents.go`).

### Step 1: Model Type — `sdk/mpr/reader_types.go`

Add after `ImageCollection` (line 283):

```go
type JsonStructure struct {
    model.BaseElement
    ContainerID   model.ID       `json:"containerId"`
    Name          string         `json:"name"`
    documentation string         `json:"documentation,omitempty"`
    JsonSnippet   string         `json:"jsonSnippet,omitempty"`
    Elements      []*JsonElement `json:"elements,omitempty"`
    Excluded      bool           `json:"excluded,omitempty"`
    ExportLevel   string         `json:"exportLevel,omitempty"`
}

type JsonElement struct {
    ExposedName     string         `json:"exposedName"`
    ExposedItemName string         `json:"exposedItemName,omitempty"`
    path            string         `json:"path"`
    ElementType     string         `json:"elementType"`     // "object", "Array", "value"
    PrimitiveType   string         `json:"primitiveType"`   // "string", "integer", "boolean", "decimal", "Unknown"
    MinOccurs       int            `json:"minOccurs"`
    MaxOccurs       int            `json:"maxOccurs"`       // -1 = unbounded
    Nillable        bool           `json:"nillable,omitempty"`
    IsDefaultType   bool           `json:"isDefaultType,omitempty"`
    MaxLength       int            `json:"maxLength"`       // -1 = unset
    FractionDigits  int            `json:"fractionDigits"`  // -1 = unset
    TotalDigits     int            `json:"totalDigits"`     // -1 = unset
    OriginalValue   string         `json:"originalValue,omitempty"`
    Children        []*JsonElement `json:"children,omitempty"`
}
```

Add `GetName()`, `GetContainerID()` methods and `ListJsonStructures()` reader method.

### Step 2: Parser (BSON → Go) — `sdk/mpr/parser_misc.go`

Add after `parseImageCollection` (line 643):

```go
func (r *Reader) parseJsonStructure(unitID, containerID string, contents []byte) (*JsonStructure, error)
func parseJsonElement(raw map[string]any) *JsonElement  // recursive
```

Key details:
- Use `r.resolveContents(unitID, contents)` for v2 format
- Extract `Elements` from `bson.A`, skip leading version int (`2`)
- Recursively parse `Children` with same version-int skip
- Use `extractString`, `extractBool`, `extractInt32` helpers

### Step 3: Writer (Go → BSON) — new file `sdk/mpr/writer_jsonstructure.go`

Following `writer_imagecollection.go` pattern:

```go
func (w *Writer) CreateJsonStructure(js *JsonStructure) error
func (w *Writer) DeleteJsonStructure(id string) error
func serializeJsonStructure(js *JsonStructure) ([]byte, error)
func serializeJsonElement(elem *JsonElement) bson.D  // recursive
```

Key serialization details:
- Unit type: `"JsonStructures$JsonStructure"`
- Containment name: `"Documents"`
- Elements array: `bson.A{int32(2), ...serialized elements}` (version prefix 2)
- Children array: same `bson.A{int32(2), ...}` recursive format
- Generate UUIDs for each element `$ID`
- Set all fields including defaults (`ErrorMessage: ""`, `WarningMessage: ""`, etc.)

### Step 4: JSON Snippet → Element Tree Builder — `sdk/mpr/writer_jsonstructure.go`

The CREATE command accepts a JSON snippet and must auto-build the element tree (same as Studio Pro):

```go
func buildJsonElementsFromSnippet(snippet string) ([]*JsonElement, error)
```

Algorithm:
1. Parse JSON snippet with `encoding/json`
2. Walk the JSON tree recursively
3. For each key-value pair, create a `JsonElement` with:
   - `ExposedName`: capitalize first letter of JSON key
   - `path`: build from parent path + `|keyName`
   - `ElementType`: "Object" for `{}`, "Array" for `[]`, "Value" for primitives
   - `PrimitiveType`: infer from JSON value type (string→"String", number with `.`→"Decimal", integer→"Integer", bool→"Boolean")
   - `OriginalValue`: JSON-encoded sample value
   - `MaxOccurs`: 1 for objects/values, -1 for array items
4. Handle duplicate keys by appending `_2`, `_3` suffixes
5. For arrays, create intermediate Array element + Object child with `MaxOccurs: -1`
6. Always wrap in a Root element with `path: "(object)"`

### Step 5: AST — `mdl/ast/ast_query.go` + new `mdl/ast/ast_jsonstructure.go`

**In `ast_query.go`:**
- Add `ShowJsonStructures` to `ShowObjectType` iota (after `ShowConstantValues`, line 82)
- Add `"json structures"` to `ShowObjectType.String()`
- Add `DescribeJsonStructure` to `DescribeObjectType` iota (after `DescribeContractMessage`, line 256)
- Add `"json structure"` to `DescribeObjectType.String()`

**New `ast_jsonstructure.go`** (following `ast_imagecollection.go` pattern):
```go
type CreateJsonStructureStmt struct {
    Name          QualifiedName
    JsonSnippet   string   // Raw json snippet
    ExportLevel   string   // "Hidden" (default) or "Public"
    documentation string
    CreateOrReplace bool
}
func (s *CreateJsonStructureStmt) isStatement() {}

type DropJsonStructureStmt struct {
    Name QualifiedName
}
func (s *DropJsonStructureStmt) isStatement() {}
```

### Step 6: Grammar — `mdl/grammar/MDLLexer.g4` + `MDLParser.g4`

**Lexer** — add `structures` token (alphabetically, ~line 545):
```antlr
structures: S T R U C T U R E S;
```

**Parser** — add to `showStatement` alternatives (after line 2501):
```antlr
| show json structures (in (qualifiedName | IDENTIFIER))?
```

Add to `describeStatement` alternatives (after line 2627):
```antlr
| describe json structure qualifiedName
```

Add to `dropStatement` alternatives (after line 258):
```antlr
| drop json structure qualifiedName
```

Add `createJsonStructureStatement` to `createStatement` alternatives (after line 100):
```antlr
| createJsonStructureStatement
```

Add new production:
```antlr
createJsonStructureStatement
    : json structure qualifiedName jsonStructureOptions? snippet STRING_LITERAL
    ;

jsonStructureOptions
    : jsonStructureOption+
    ;

jsonStructureOption
    : export level STRING_LITERAL
    | comment STRING_LITERAL
    ;
```

MDL syntax:
```sql
-- Create new (fails if exists)
create json structure Module.Name snippet '{...json...}';

-- Create with metadata
create json structure Module.Name export level 'Public' comment 'API response schema' snippet '{...}';

-- Create or update (idempotent — preferred for AI agents)
create or replace json structure Module.Name snippet '{...}';
```

Then run `make grammar` to regenerate.

### Step 7: Visitor — `mdl/visitor/visitor_query.go` + new `mdl/visitor/visitor_jsonstructure.go`

**In `visitor_query.go`:**

Add to `ExitShowStatement` (after IMAGE COLLECTION block, ~line 468):
```go
} else if ctx.JSON() != nil && ctx.STRUCTURES() != nil {
    stmt := &ast.ShowStmt{ObjectType: ast.ShowJsonStructures}
    // parse in module filter
    b.statements = append(b.statements, stmt)
}
```

Add to `ExitDescribeStatement` (after IMAGE COLLECTION block, ~line 809):
```go
} else if ctx.JSON() != nil && ctx.STRUCTURE() != nil {
    b.statements = append(b.statements, &ast.DescribeStmt{
        ObjectType: ast.DescribeJsonStructure,
        Name:       buildQualifiedName(qn),
    })
}
```

**New `visitor_jsonstructure.go`** (following `visitor_imagecollection.go` pattern):
```go
func (b *Builder) ExitCreateJsonStructureStatement(ctx *parser.CreateJsonStructureStatementContext)
```

### Step 8: Executor — new file `mdl/executor/cmd_jsonstructures.go`

**`showJsonStructures(moduleName string) error`**
Table: `| Qualified Name | Elements | source |`

**`describeJsonStructure(name ast.QualifiedName) error`**
Output: Element tree with indentation + JSON Snippet as comment block.

Rendering rules:
- Indent children by 2 spaces per nesting level
- Type = PrimitiveType for Value elements, ElementType for Object/Array
- Show `[min..max]` only when MaxOccurs != 1 (use `*` for -1)
- JSON Snippet rendered as `--`-prefixed comment lines
- Terminated with `/` (MDL statement separator)

**`execCreateJsonStructure(s *ast.CreateJsonStructureStmt) error`**
1. Find or auto-create module
2. Check if already exists:
   - If exists and `CreateOrReplace` is true → delete existing first, then create new
   - If exists and `CreateOrReplace` is false → return error "already exists"
3. Call `buildJsonElementsFromSnippet(s.JsonSnippet)` to generate element tree
4. Call `writer.CreateJsonStructure(js)`
5. Invalidate hierarchy cache

This makes `create or replace` the idempotent edit operation — AI agents can always use it without checking existence first.

**`execDropJsonStructure(s *ast.DropJsonStructureStmt) error`**
Same pattern as `execDropImageCollection`.

**`findJsonStructure(moduleName, structName string) *mpr.JsonStructure`**
Helper, same pattern as `findImageCollection`.

### Step 9: Executor Dispatch — `mdl/executor/executor.go`

In `execShow()` (after line 784):
```go
case ast.ShowJsonStructures:
    return e.showJsonStructures(s.InModule)
```

In `execDescribe()` (after line 855):
```go
case ast.DescribeJsonStructure:
    return e.describeJsonStructure(s.Name)
```

In the main statement switch (near line 278):
```go
case *ast.CreateJsonStructureStmt:
    return e.execCreateJsonStructure(s)
case *ast.DropJsonStructureStmt:
    return e.execDropJsonStructure(s)
```

### Step 10: Autocomplete — `mdl/executor/autocomplete.go` + `mdl/repl/repl.go`

Add `GetJsonStructureNames(moduleFilter string) []string` following `GetBusinessEventServiceNames` (line 315).

Register in REPL completions for:
- `"describe json structure "` → `GetJsonStructureNames`
- `"drop json structure "` → `GetJsonStructureNames`

### Step 11: Catalog Table — `mdl/catalog/tables.go`

Add `json_structures` catalog table:
- Columns: QualifiedName, Module, Name, ElementCount, HasSnippet, Documentation, ExportLevel
- Populate in catalog refresh function

### Step 12: Test Examples — `mdl-examples/doctype-tests/18-json-structure-examples.mdl`

```sql
-- JSON Structure examples (read)
show json structures;
show json structures in MyModule;
describe json structure MyModule.CustomerResponse;

-- JSON Structure examples (create)
create json structure MyModule.CustomerResponse snippet '{
  "id": 1,
  "name": "John",
  "email": "john@example.com",
  "addresses": [{"street": "Main St", "city": "Springfield", "zipCode": "62701"}],
  "active": true
}';

-- JSON Structure examples (edit — idempotent, preferred for AI agents)
create or replace json structure MyModule.CustomerResponse snippet '{
  "id": 1,
  "name": "John",
  "email": "john@example.com",
  "phone": "+1-555-0123",
  "addresses": [{"street": "Main St", "city": "Springfield", "zipCode": "62701"}],
  "active": true
}';

-- JSON Structure examples (delete)
drop json structure MyModule.CustomerResponse;
```

## Files to Modify

| File | Change |
|------|--------|
| `sdk/mpr/reader_types.go` | Add JsonStructure/JsonElement types + ListJsonStructures() |
| `sdk/mpr/parser_misc.go` | Add parseJsonStructure() + parseJsonElement() |
| **`sdk/mpr/writer_jsonstructure.go`** | **New file** — CreateJsonStructure, DeleteJsonStructure, serialization, snippet→element builder |
| `mdl/ast/ast_query.go` | Add ShowJsonStructures + DescribeJsonStructure enum values + String() cases |
| **`mdl/ast/ast_jsonstructure.go`** | **New file** — CreateJsonStructureStmt, DropJsonStructureStmt |
| `mdl/grammar/MDLLexer.g4` | Add STRUCTURES token |
| `mdl/grammar/MDLParser.g4` | Add SHOW/DESCRIBE/CREATE/DROP JSON STRUCTURE rules |
| `mdl/grammar/parser/` | Regenerated (make grammar) |
| `mdl/visitor/visitor_query.go` | Add JSON STRUCTURES/STRUCTURE handlers |
| **`mdl/visitor/visitor_jsonstructure.go`** | **New file** — ExitCreateJsonStructureStatement |
| **`mdl/executor/cmd_jsonstructures.go`** | **New file** — show, describe, create, drop, find handlers |
| `mdl/executor/executor.go` | Add dispatch cases for SHOW/DESCRIBE/CREATE/DROP |
| `mdl/executor/autocomplete.go` | Add GetJsonStructureNames() |
| `mdl/repl/repl.go` | Register completions |
| `mdl/catalog/tables.go` | Add json_structures catalog table |
| **`mdl-examples/doctype-tests/18-json-structure-examples.mdl`** | **New file** — test examples |

## Verification

1. `make grammar` — regenerate parser from modified grammar
2. `make build` — verify compilation
3. `make test` — existing tests pass
4. `./bin/mxcli check mdl-examples/doctype-tests/18-json-structure-examples.mdl` — syntax OK
5. Test against any Mendix project with JSON structures:
   ```bash
   ./bin/mxcli -p app.mpr -c "show json structures"
   ./bin/mxcli -p app.mpr -c "describe json structure Module.StructureName"
   ```
6. Test CREATE roundtrip:
   ```bash
   ./bin/mxcli -p app.mpr -c "create json structure MyModule.Test snippet '{\"name\": \"John\", \"age\": 30}'"
   ./bin/mxcli -p app.mpr -c "describe json structure MyModule.Test"
   ```
7. Test CREATE OR REPLACE (edit flow):
   ```bash
   ./bin/mxcli -p app.mpr -c "create or replace json structure MyModule.Test snippet '{\"name\": \"John\", \"age\": 30, \"email\": \"john@test.com\"}'"
   ./bin/mxcli -p app.mpr -c "describe json structure MyModule.Test"
   # Verify element tree now includes Email field
   ```
8. Test DROP:
   ```bash
   ./bin/mxcli -p app.mpr -c "drop json structure MyModule.Test"
   ```
9. Validate with `mx check` that created JSON structures are valid Mendix format
