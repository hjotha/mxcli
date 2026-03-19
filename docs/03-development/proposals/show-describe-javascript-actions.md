# Proposal: SHOW/DESCRIBE JavaScript Actions

## Overview

**Document type:** `JavaScriptActions$JavaScriptAction`
**Prevalence:** 283 across test projects (96 Enquiries, 81 Evora, 106 Lato)
**Priority:** High — present in every project, used heavily for nanoflow logic

JavaScript Actions are the nanoflow equivalent of Java Actions. They define callable functions with typed parameters and return types, targeting Web, Native, or All platforms.

## What Already Exists

| Layer | Status | Location |
|-------|--------|----------|
| **Go type** | Minimal | `sdk/mpr/reader_types.go` — `JavaScriptAction{Name, Documentation, ReturnType}` |
| **Parser** | Minimal | `sdk/mpr/reader_types.go` — parses Name, Documentation, ReturnType only |
| **Reader** | Yes | `ListJavaScriptActions()` in `sdk/mpr/reader_types.go` |
| **Generated metamodel** | Full | `generated/metamodel/types.go` line 3103 |
| **AST** | No | No `ShowJavaScriptActions` / `DescribeJavaScriptAction` |
| **Executor** | No | No show/describe handlers |
| **Grammar** | No | No `JAVASCRIPT ACTION` keywords |

## BSON Structure (from test projects)

```
JavaScriptActions$JavaScriptAction:
  Name: string
  Documentation: string
  Excluded: bool
  ExportLevel: string ("API" | "Hidden")
  Platform: string ("All" | "Web" | "Native")
  JavaReturnType: polymorphic CodeActions type
  Parameters: []*JavaScriptActionParameter
    - Name: string
    - Description: string
    - Category: string
    - IsRequired: bool
    - ParameterType: polymorphic CodeActions type
  TypeParameters: []*CodeActionsTypeParameter
    - Name: string
  MicroflowActionInfo: nullable
    - Caption: string
    - Category: string
```

## Proposed MDL Syntax

### SHOW JAVASCRIPT ACTIONS

```
SHOW JAVASCRIPT ACTIONS [IN Module]
```

Output table columns:

| Qualified Name | Module | Name | Platform | Return Type | Parameters | Exposed |
|----------------|--------|------|----------|-------------|------------|---------|

### DESCRIBE JAVASCRIPT ACTION

```
DESCRIBE JAVASCRIPT ACTION Module.Name
```

Output format (modeled after `DESCRIBE JAVA ACTION`):

```
/**
 * Documentation text here
 */
CREATE JAVASCRIPT ACTION MyModule.MyAction
  PLATFORM Web
  PARAMETER $input: String
  PARAMETER $count: Integer (REQUIRED)
  RETURNS Boolean
  EXPOSED AS 'My Custom Action' IN 'Data';
/
```

## Implementation Steps

### 1. Enhance Parser (sdk/mpr/reader_types.go)

Extend `JavaScriptAction` struct and parser to capture:
- `Platform` (string)
- `Excluded` (bool)
- `ExportLevel` (string)
- `Parameters` (slice of parameter structs with Name, Type, IsRequired, Description)
- `MicroflowActionInfo` (Caption, Category — for exposed actions)

### 2. Add AST Types (mdl/ast/ast_query.go)

```go
// In ShowObjectType enum:
ShowJavaScriptActions

// In DescribeObjectType enum:
DescribeJavaScriptAction

// Add String() cases
```

### 3. Add Grammar Rules (MDLLexer.g4 / MDLParser.g4)

Add `JAVASCRIPT` token and rules:
```antlr
JAVASCRIPT: 'JAVASCRIPT';

// In showStmt:
| SHOW JAVASCRIPT ACTIONS (IN qualifiedName)?
// In describeStmt:
| DESCRIBE JAVASCRIPT ACTION qualifiedName
```

### 4. Add Visitor (mdl/visitor/)

Map parse tree to `ShowStmt{ObjectType: ShowJavaScriptActions}` and `DescribeStmt{ObjectType: DescribeJavaScriptAction}`.

### 5. Add Executor (mdl/executor/cmd_javascript_actions.go)

- `showJavaScriptActions(moduleName string)` — list in markdown table
- `describeJavaScriptAction(name QualifiedName)` — output MDL format

Follow patterns from `cmd_javaactions.go`.

### 6. Add Autocomplete (mdl/executor/autocomplete.go)

```go
func (e *Executor) GetJavaScriptActionNames(moduleFilter string) []string
```

### 7. Wire into Executor Dispatcher (mdl/executor/executor.go)

Add cases in `execShow` and `execDescribe` switch statements.

## Testing

- Add entries to `mdl-examples/doctype-tests/07-java-action-examples.mdl` (or new file)
- Verify against all 3 test projects
