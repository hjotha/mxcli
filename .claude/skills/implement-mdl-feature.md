# Implement New MDL Feature Skill

This skill provides a step-by-step workflow for implementing new MDL syntax for Mendix concepts (microflow actions, page widgets, etc.).

## When to Use This Skill

Use this skill when:
- Adding support for a new Mendix action type (e.g., REST call, web service call)
- Adding support for a new widget type
- Extending existing MDL syntax with new clauses or options
- Implementing round-trip support (CREATE + DESCRIBE) for a Mendix concept

## Overview

Implementing a new MDL feature requires changes across multiple layers:

```
┌─────────────────────────────────────────────────────────────────────┐
│                    1. INVESTIGATION                                  │
│    Dump BSON from MPR → Understand structure → Design MDL syntax    │
└─────────────────────────────────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    2. GRAMMAR (MDLLexer.g4 + MDLParser.g4)          │
│    Add tokens → Add parser rules → Regenerate (make grammar)        │
└─────────────────────────────────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    3. AST (mdl/ast/)                                 │
│    Define Go structs representing the parsed syntax                  │
└─────────────────────────────────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    4. VISITOR (mdl/visitor/)                         │
│    Parse ANTLR context → Build AST nodes                            │
└─────────────────────────────────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    5. SDK TYPES (sdk/microflows/ or sdk/pages/)     │
│    Add/update Go structs for the Mendix concept                     │
└─────────────────────────────────────────────────────────────────────┘
                                  │
                  ┌───────────────┴───────────────┐
                  ▼                               ▼
┌─────────────────────────────┐   ┌─────────────────────────────┐
│  6a. PARSER (BSON → Go)     │   │  6b. WRITER (Go → BSON)     │
│  sdk/mpr/parser_*.go        │   │  sdk/mpr/writer_*.go        │
│  For DESCRIBE to work       │   │  For CREATE to work         │
└─────────────────────────────┘   └─────────────────────────────┘
                  │                               │
                  └───────────────┬───────────────┘
                                  ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    7. EXECUTOR (mdl/executor/)                       │
│    cmd_*_builder.go (AST → BSON)  +  cmd_*_show.go (Go → MDL)       │
└─────────────────────────────────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    8. TESTING                                        │
│    Syntax check → Describe roundtrip → Create in project            │
└─────────────────────────────────────────────────────────────────────┘
```

## Part 1: Investigation

### Step 1: Find an Example in a Mendix Project

First, create or find an example of the feature in Mendix Studio Pro:

1. Open a test project in Studio Pro (e.g., `mx-test-projects/test2-go-app/test2-go.mpr`)
2. Create the element you want to support (e.g., a microflow with a REST call action)
3. Save the project
4. Note the qualified name (e.g., `RestTemplate.GetWebpage`)

### Step 2: Find the Document ID

```bash
./bin/mxcli -p mx-test-projects/test2-go-app/test2-go.mpr \
  -c "SELECT Id, QualifiedName FROM CATALOG.MICROFLOWS WHERE QualifiedName = 'RestTemplate.GetWebpage'"
```

### Step 3: Dump Raw BSON

Use `mxcli bson dump` to inspect the raw BSON structure (see [BSON Tooling Guide](../../docs/03-development/BSON_TOOLING_GUIDE.md) for the full tool reference):

```bash
mxcli bson dump -p mx-test-projects/test2-go-app/test2-go.mpr --type microflow --object "RestTemplate.GetWebpage"
```

Example output (abbreviated):
```json
{
  "$Type": "Microflows$Microflow",
  "ObjectCollection": {
    "Objects": [
      {
        "$Type": "Microflows$ActionActivity",
        "Action": {
          "$Type": "Microflows$RestCallAction",
          "HttpConfiguration": {
            "$Type": "Microflows$HttpConfiguration",
            "HttpMethod": "Get",
            "CustomLocationTemplate": {
              "Text": "{1}",
              "Parameters": [
                {"Expression": "'http://example.com'"}
              ]
            },
            "HttpHeaderEntries": [
              {"Key": "Accept", "Value": "'text/html'"}
            ]
          },
          "ResultHandling": {
            "ResultVariableName": "MyPage"
          },
          "ResultHandlingType": "String",
          "TimeOutExpression": "300"
        }
      }
    ]
  }
}
```

### Step 4: Design MDL Syntax

Based on the BSON structure, design MDL syntax that:
- Is readable and intuitive
- Maps clearly to BSON fields
- Follows existing MDL conventions

Example design for REST call:
```sql
$Response = REST CALL GET 'http://api.example.com/data'
  HEADER 'Content-Type' = 'application/json'
  HEADER Accept = 'application/json'
  AUTH BASIC $username PASSWORD $password
  BODY '{"key": "{1}"}' WITH ({1} = $value)
  TIMEOUT 30
  RETURNS String
  ON ERROR ROLLBACK;
```

### Step 5: Create Examples

Add examples to `mdl-examples/doctype-tests/`:

```bash
# Add to 02-microflow-examples.mdl
CREATE MICROFLOW RestExamples.SimpleGet()
RETURNS String AS $Response
BEGIN
  $Response = REST CALL GET 'https://api.example.com/data'
    TIMEOUT 30
    RETURNS String;
  RETURN $Response;
END;
```

## Part 2: Grammar Implementation

### Step 1: Add Tokens (MDLLexer.g4)

Add new keywords if needed:

```antlr
// In MDLLexer.g4, alphabetically ordered
AUTH: A U T H;
BODY: B O D Y;
HEADER: H E A D E R;
MAPPING: M A P P I N G;
PASSWORD: P A S S W O R D;
REST: R E S T;
TIMEOUT: T I M E O U T;
```

### Step 2: Add Parser Rules (MDLParser.g4)

```antlr
// Add to microflowStatement alternatives
microflowStatement
    : // ... existing alternatives
    | restCallStatement
    ;

// Define the new statement
restCallStatement
    : (VARIABLE EQUALS)? REST CALL httpMethod restCallUrl restCallUrlParams?
      restCallHeaderClause*
      restCallAuthClause?
      restCallBodyClause?
      restCallTimeoutClause?
      restCallReturnsClause
      onErrorClause?
    ;

httpMethod
    : GET | POST | PUT | PATCH | DELETE
    ;

restCallUrl
    : STRING_LITERAL | expression
    ;

restCallHeaderClause
    : HEADER (IDENTIFIER | STRING_LITERAL) EQUALS expression
    ;

// ... other clauses
```

### Step 3: Regenerate Parser

```bash
make grammar
```

This runs ANTLR4 and generates the Go parser code in `mdl/grammar/parser/`.

## Part 3: AST Types

Add AST types in `mdl/ast/ast_microflow.go`:

```go
// HttpMethod represents an HTTP method.
type HttpMethod string

const (
    HttpMethodGet    HttpMethod = "Get"
    HttpMethodPost   HttpMethod = "Post"
    HttpMethodPut    HttpMethod = "Put"
    HttpMethodPatch  HttpMethod = "Patch"
    HttpMethodDelete HttpMethod = "Delete"
)

// RestHeader represents a custom HTTP header.
type RestHeader struct {
    Name  string
    Value Expression
}

// RestCallStmt represents a REST call statement.
type RestCallStmt struct {
    OutputVariable string
    Method         HttpMethod
    URL            Expression
    URLParams      []TemplateParam
    Headers        []RestHeader
    Auth           *RestAuth
    Body           *RestBody
    Timeout        Expression
    Result         RestResult
    ErrorHandling  *ErrorHandlingClause
}

func (s *RestCallStmt) isMicroflowStatement() {}
```

## Part 4: Visitor Implementation

Add visitor handler in `mdl/visitor/visitor_microflow_statements.go`:

```go
func buildRestCallStatement(ctx parser.IRestCallStatementContext) *ast.RestCallStmt {
    restCtx := ctx.(*parser.RestCallStatementContext)
    stmt := &ast.RestCallStmt{}

    // Output variable
    if v := restCtx.VARIABLE(); v != nil {
        stmt.OutputVariable = strings.TrimPrefix(v.GetText(), "$")
    }

    // HTTP method
    if m := restCtx.HttpMethod(); m != nil {
        methodCtx := m.(*parser.HttpMethodContext)
        if methodCtx.GET() != nil {
            stmt.Method = ast.HttpMethodGet
        } else if methodCtx.POST() != nil {
            stmt.Method = ast.HttpMethodPost
        }
        // ... other methods
    }

    // URL
    if urlC := restCtx.RestCallUrl(); urlC != nil {
        urlCtx := urlC.(*parser.RestCallUrlContext)
        if strLit := urlCtx.STRING_LITERAL(); strLit != nil {
            stmt.URL = &ast.StringLiteral{Value: unquoteString(strLit.GetText())}
        }
    }

    // ... parse other clauses

    return stmt
}
```

Add to the statement switch in `buildMicroflowStatement`:

```go
case ctx.RestCallStatement() != nil:
    return buildRestCallStatement(ctx.RestCallStatement())
```

## Part 5: SDK Types

Update or add types in `sdk/microflows/microflows_actions.go`:

```go
// RestCallAction represents a REST call action.
type RestCallAction struct {
    model.BaseElement
    HttpConfiguration *HttpConfiguration `json:"httpConfiguration,omitempty"`
    RequestHandling   RequestHandling    `json:"requestHandling,omitempty"`
    ResultHandling    ResultHandling     `json:"resultHandling,omitempty"`
    ErrorHandlingType ErrorHandlingType  `json:"errorHandlingType,omitempty"`
    OutputVariable    string             `json:"outputVariable,omitempty"`
    TimeoutExpression string             `json:"timeoutExpression,omitempty"`
}

func (RestCallAction) isMicroflowAction() {}

// HttpConfiguration represents HTTP configuration for a REST call.
type HttpConfiguration struct {
    model.BaseElement
    HttpMethod        HttpMethod    `json:"httpMethod"`
    LocationTemplate  string        `json:"locationTemplate,omitempty"`
    LocationParams    []string      `json:"locationParams,omitempty"`
    CustomHeaders     []*HttpHeader `json:"customHeaders,omitempty"`
    UseAuthentication bool          `json:"useAuthentication,omitempty"`
    Username          string        `json:"username,omitempty"`
    Password          string        `json:"password,omitempty"`
}
```

## Part 6a: Parser (BSON → Go)

Add parsing logic in `sdk/mpr/parser_microflow.go`:

```go
// Add case in parseActionActivity switch
case "Microflows$RestCallAction":
    return parseRestCallAction(raw)

// Add parser function
func parseRestCallAction(raw map[string]interface{}) *microflows.RestCallAction {
    action := &microflows.RestCallAction{}
    action.ID = model.ID(extractBsonID(raw["$ID"]))
    action.TimeoutExpression = extractString(raw["TimeOutExpression"])
    action.ErrorHandlingType = microflows.ErrorHandlingType(extractString(raw["ErrorHandlingType"]))

    // Parse HttpConfiguration
    if httpConfig, ok := raw["HttpConfiguration"].(map[string]interface{}); ok {
        action.HttpConfiguration = parseHttpConfiguration(httpConfig)
    }

    // Parse ResultHandling
    resultHandlingType := extractString(raw["ResultHandlingType"])
    if resultHandling, ok := raw["ResultHandling"].(map[string]interface{}); ok {
        action.ResultHandling = parseResultHandling(resultHandling, resultHandlingType)
    }

    return action
}
```

## Part 6b: Writer (Go → BSON)

Add serialization logic in `sdk/mpr/writer_microflow.go`:

```go
func serializeRestCallAction(action *microflows.RestCallAction) bson.D {
    doc := bson.D{
        {Key: "$ID", Value: idToBsonBinary(action.ID)},
        {Key: "$Type", Value: "Microflows$RestCallAction"},
        {Key: "ErrorHandlingType", Value: string(action.ErrorHandlingType)},
        {Key: "TimeOutExpression", Value: action.TimeoutExpression},
        {Key: "UseRequestTimeOut", Value: action.TimeoutExpression != ""},
    }

    // Serialize HttpConfiguration
    if action.HttpConfiguration != nil {
        doc = append(doc, bson.E{Key: "HttpConfiguration",
            Value: serializeHttpConfiguration(action.HttpConfiguration)})
    }

    // ... serialize other fields

    return doc
}
```

## Part 7: Executor

### Builder (AST → BSON)

In `mdl/executor/cmd_microflows_builder.go`:

```go
// Add case in addStatement switch
case *ast.RestCallStmt:
    return mb.addRestCallAction(stmt)

// Add builder method
func (mb *MicroflowBuilder) addRestCallAction(stmt *ast.RestCallStmt) error {
    action := &microflows.RestCallAction{
        ErrorHandlingType: mb.mapErrorHandling(stmt.ErrorHandling),
    }

    // Build HttpConfiguration
    action.HttpConfiguration = &microflows.HttpConfiguration{
        HttpMethod: microflows.HttpMethod(stmt.Method),
    }

    // Evaluate URL expression
    if stmt.URL != nil {
        urlVal, err := mb.evaluateExpression(stmt.URL)
        if err != nil {
            return err
        }
        action.HttpConfiguration.LocationTemplate = "{1}"
        action.HttpConfiguration.LocationParams = []string{urlVal}
    }

    // ... build other fields

    return mb.addActivity(action, stmt.OutputVariable, stmt.ErrorHandling)
}
```

### Show (Go → MDL)

In `mdl/executor/cmd_microflows_show.go`:

```go
// Add case in formatActionStatement switch
case *microflows.RestCallAction:
    return e.formatRestCallAction(a)

// Add formatter
func (e *Executor) formatRestCallAction(a *microflows.RestCallAction) string {
    var sb strings.Builder

    // Output variable
    if outputVar := getOutputVariable(a); outputVar != "" {
        sb.WriteString("$")
        sb.WriteString(outputVar)
        sb.WriteString(" = ")
    }

    sb.WriteString("REST CALL ")

    // HTTP method
    if a.HttpConfiguration != nil {
        sb.WriteString(strings.ToUpper(string(a.HttpConfiguration.HttpMethod)))
    }
    sb.WriteString(" ")

    // URL with parameters
    if a.HttpConfiguration != nil {
        sb.WriteString("'")
        sb.WriteString(a.HttpConfiguration.LocationTemplate)
        sb.WriteString("'")
        if len(a.HttpConfiguration.LocationParams) > 0 {
            sb.WriteString(" WITH (")
            for i, param := range a.HttpConfiguration.LocationParams {
                if i > 0 {
                    sb.WriteString(", ")
                }
                sb.WriteString(fmt.Sprintf("{%d} = %s", i+1, param))
            }
            sb.WriteString(")")
        }
    }

    // ... format other clauses

    sb.WriteString(";")
    return sb.String()
}
```

## Part 8: Testing

### Step 1: Syntax Check

```bash
./bin/mxcli check mdl-examples/doctype-tests/02-microflow-examples.mdl
```

Expected: `Syntax OK (N statements)`

### Step 2: Describe Roundtrip

```bash
./bin/mxcli -p mx-test-projects/test2-go-app/test2-go.mpr \
  -c "DESCRIBE MICROFLOW RestTemplate.GetWebpage"
```

Expected: Valid MDL output that could be re-parsed.

### Step 3: Create and Verify

```bash
# Create via MDL
./bin/mxcli -p test-project.mpr -c "
CREATE MICROFLOW Test.RestExample()
RETURNS String AS \$R
BEGIN
  \$R = REST CALL GET 'http://example.com' TIMEOUT 30 RETURNS String;
  RETURN \$R;
END;"

# Describe to verify
./bin/mxcli -p test-project.mpr -c "DESCRIBE MICROFLOW Test.RestExample"

# Open in Studio Pro to verify it works
```

### Step 4: Run Existing Tests

```bash
make test
```

## Checklist

Before considering the implementation complete:

- [ ] Investigation complete - BSON structure understood
- [ ] MDL syntax designed and documented in examples
- [ ] Grammar tokens added (`MDLLexer.g4`)
- [ ] Grammar rules added (`MDLParser.g4`)
- [ ] Parser regenerated (`make grammar`)
- [ ] AST types added (`mdl/ast/`)
- [ ] Visitor implemented (`mdl/visitor/`)
- [ ] SDK types added/updated (`sdk/microflows/` or `sdk/pages/`)
- [ ] BSON parser added (`sdk/mpr/parser_*.go`)
- [ ] BSON writer added (`sdk/mpr/writer_*.go`)
- [ ] Executor builder added (`mdl/executor/cmd_*_builder.go`)
- [ ] Executor show/describe added (`mdl/executor/cmd_*_show.go`)
- [ ] Syntax check passes
- [ ] Describe roundtrip works
- [ ] CREATE works and opens in Studio Pro
- [ ] Build succeeds (`make build`)
- [ ] Tests pass (`make test`)

## File Reference

| Layer | Files | Purpose |
|-------|-------|---------|
| Grammar | `mdl/grammar/MDLLexer.g4` | Token definitions |
| Grammar | `mdl/grammar/MDLParser.g4` | Parser rules |
| AST | `mdl/ast/ast_microflow.go` | Microflow statement types |
| AST | `mdl/ast/ast_page.go` | Page/widget statement types |
| Visitor | `mdl/visitor/visitor_microflow_statements.go` | Microflow parsing |
| Visitor | `mdl/visitor/visitor_page.go` | Page parsing |
| SDK | `sdk/microflows/microflows_actions.go` | Action Go types |
| SDK | `sdk/pages/pages.go` | Widget Go types |
| Parser | `sdk/mpr/parser_microflow.go` | BSON → Go (microflows) |
| Parser | `sdk/mpr/parser_page.go` | BSON → Go (pages) |
| Writer | `sdk/mpr/writer_microflow.go` | Go → BSON (microflows) |
| Writer | `sdk/mpr/writer_widgets.go` | Go → BSON (widgets) |
| Executor | `mdl/executor/cmd_microflows_builder.go` | AST → microflow BSON |
| Executor | `mdl/executor/cmd_microflows_show.go` | Microflow → MDL |
| Executor | `mdl/executor/cmd_pages_builder.go` | AST → page BSON |
| Executor | `mdl/executor/cmd_pages_show.go` | Page → MDL |
| Debug | `cmd/debug/main.go` | Raw BSON dump tool |
| Examples | `mdl-examples/doctype-tests/` | MDL examples |

## Related Documentation

- [BSON Tooling Guide](../../docs/03-development/BSON_TOOLING_GUIDE.md) - Which BSON tool to use when (dump, compare, discover, TUI, Python)
- [Debug BSON](./debug-bson.md) - Fixing BSON serialization issues (CE errors, widget templates)
- [Write Microflows](./write-microflows.md) - MDL microflow syntax reference
