# Proposal: Agent Document Type Support in MDL

**Status:** Draft
**Date:** 2026-04-13

## Problem Statement

Mendix 11.9 introduces **Agents** as a first-class concept for building agentic AI applications. Agents are defined in Studio Pro through the Agent Editor extension and stored as `CustomBlobDocuments$CustomBlobDocument` with `CustomDocumentType = "agenteditor.agent"`. The ecosystem includes five marketplace modules:

| Module | Role |
|--------|------|
| **GenAICommons** | Core AI types: Request/Response, DeployedModel, Tool, ToolCall, Trace, KnowledgeBase |
| **MxGenAIConnector** | Mendix Cloud AI backend (Bedrock), model config, embeddings, knowledge base collections |
| **AgentCommons** | Agent management: versioned agents, tools, knowledge bases, MCP, test cases |
| **AgentEditorCommons** | Bridge between Studio Pro Agent Editor extension and AgentCommons runtime entities |
| **MCPClient** | Model Context Protocol client: server connections, tool/prompt discovery, execution |
| **ConversationalUI** | Chat widgets, message rendering, tool approval UI, trace monitoring, token dashboards |

Currently, `mxcli` has no visibility into agent documents. `show structure` reports the Agents module as empty because it only contains `CustomBlobDocument` units, which are not parsed. An AI coding agent cannot discover, inspect, or create agents via MDL.

## BSON Structure

All four agent-editor document types (Agent, Model, Knowledge Base, Consumed MCP Service) share the same outer wrapper — a generic `CustomBlobDocument` with a JSON payload in `Contents`. They're distinguished by the `CustomDocumentType` field.

### Outer Wrapper (common to all four types)

```
CustomBlobDocuments$CustomBlobDocument:
  $ID: bytes
  $type: "CustomBlobDocuments$CustomBlobDocument"
  Name: string
  Contents: string (json payload — schema depends on CustomDocumentType)
  CustomDocumentType: "agenteditor.agent"
                    | "agenteditor.model"
                    | "agenteditor.knowledgebase"
                    | "agenteditor.consumedMCPService"
  documentation: string
  Excluded: bool
  ExportLevel: "Hidden"
  Metadata:
    $ID: bytes
    $type: "CustomBlobDocuments$CustomBlobDocumentMetadata"
    CreatedByExtension: "extension/agent-editor"
    ReadableTypeName: "agent" | "model" | "knowledge base" | "consumed mcp service"
```

Key observations about the wrapper:
- `CustomDocumentType` is the discriminator for the inner JSON schema
- `Contents` is a JSON string (not nested BSON) — the agent editor extension owns the inner schema
- `Metadata.ReadableTypeName` is a human-friendly label (also used as the UI badge in Studio Pro)
- `Excluded` is `false` for documents created in the new Agent Editor extension; observed as `true` on the 4 older agents in the project (likely pre-release format)

### MODEL — `Contents` JSON schema

Observed in `Agents.MyFirstModel`:

```json
{
  "type": "",
  "name": "",
  "displayName": "",
  "provider": "MxCloudGenAI",
  "providerFields": {
    "environment": "",
    "deepLinkURL": "",
    "keyId": "",
    "keyName": "",
    "resourceName": "",
    "key": {
      "documentId": "51b85be5-f040-4562-bf4c-086347d387a9",
      "qualifiedName": "Agents.LLMKey"
    }
  }
}
```

- `provider` is a top-level discriminator — `"MxCloudGenAI"` is the only observed value. `providerFields` shape depends on `provider`.
- `providerFields.key` references a **String constant** (holding the Mendix Cloud GenAI Portal key) by `{documentId, qualifiedName}`.
- `type`, `name`, `displayName`, `environment`, `deepLinkURL`, `keyId`, `keyName`, `resourceName` are all empty in the sample — they're **Portal-populated** when the user clicks **Test Key** in Studio Pro, not user-set fields.

### KNOWLEDGE BASE — `Contents` JSON schema

Observed in `Agents.Knowledge_base`:

```json
{
  "name": "",
  "provider": "MxCloudGenAI",
  "providerFields": {
    "environment": "",
    "deepLinkURL": "",
    "keyId": "",
    "keyName": "",
    "modelDisplayName": "",
    "modelName": "",
    "key": {
      "documentId": "51b85be5-f040-4562-bf4c-086347d387a9",
      "qualifiedName": "Agents.LLMKey"
    }
  }
}
```

Same shape as Model, but `providerFields` includes embedding-model info (`modelDisplayName`, `modelName`) instead of `resourceName`. The `key` reference points to the same String constant.

### CONSUMED MCP SERVICE — `Contents` JSON schema

Observed in `Agents.Consumed_MCP_service`:

```json
{
  "protocolVersion": "v2025_03_26",
  "documentation": " fqwef qwec qwefc",
  "version": "0.0.1",
  "connectionTimeoutSeconds": 30
}
```

Notably **absent**: no endpoint URL and no credentials microflow reference. Those are presumably configured at runtime via the `MCPClient.ConsumedMCPService` entity (see the `Agents.MCP_Server_Endpoint` String constant in the same project — used by a runtime microflow, not embedded in the document). This matches real-world deployment: endpoints typically differ across dev/staging/prod.

Enum values for `protocolVersion`: `"v2024_11_05"` or `"v2025_03_26"`.

### AGENT — `Contents` JSON schema

Simple agent (observed in `AgentEditorCommons.TranslationAgent`):

```json
{
  "description": "",
  "systemPrompt": "Translate the given text into {{description}}.",
  "userPrompt": "...",
  "usageType": "task",
  "variables": [
    { "key": "description", "isAttributeInEntity": true }
  ],
  "tools": [],
  "knowledgebaseTools": [],
  "entity": {
    "documentId": "83d81a7b-4a84-416e-a64f-0ffa981c8408",
    "qualifiedName": "System.Language"
  }
}
```

Fully-populated agent (observed in `Agents.Agent007`):

```json
{
  "description": "doing your stuff for you",
  "systemPrompt": "Do you intereesting and useful stuff that makes me money",
  "userPrompt": "Just do it",
  "usageType": "task",
  "variables": [],
  "tools": [
    {
      "id": "044bc8c2-8ca6-4166-b8f0-9d2245aba8c7",
      "name": "",
      "description": "",
      "enabled": true,
      "toolType": "mcp",
      "document": {
        "qualifiedName": "Agents.Consumed_MCP_service",
        "documentId": "47c9987a-e922-44eb-a389-e641f325ce15"
      }
    }
  ],
  "knowledgebaseTools": [
    {
      "id": "20980c3d-399b-409e-8292-49df7d0ab533",
      "name": "My_mem",
      "description": "My memory of useful stuff",
      "enabled": true,
      "toolType": "",
      "document": {
        "qualifiedName": "Agents.Knowledge_base",
        "documentId": "cccc0b5b-7600-47a9-8f6e-5761ce2fc620"
      },
      "collectionIdentifier": "agent1-collection",
      "maxResults": 3
    }
  ],
  "model": {
    "documentId": "3addaaa1-8bd3-4654-8cc9-2c886d0a01e9",
    "qualifiedName": "Agents.MyFirstModel"
  },
  "maxTokens": 16384,
  "toolChoice": "Auto"
}
```

Agent schema observations:
- **`model`**: Reference to a Model document by `{documentId, qualifiedName}`. Optional in the older samples (model set at runtime); present on new Agent Editor agents.
- **`tools[]`**: Array of tool references. Each entry has a UUID `id`, name, description, `enabled` boolean, and a `toolType` discriminator. Observed `toolType` values: `"mcp"` (a whole MCP service attached as tools). A microflow-tool sample is still missing — likely `"microflow"` with a `microflow` reference instead of `document`.
- **`knowledgebaseTools[]`**: Array of KB references. Same base fields plus `collectionIdentifier` and `maxResults`. No `minSimilarity` observed in current schema.
- **`variables[]`**: Empty in `Agent007`; populated in older samples with `{key, isAttributeInEntity}`.
- **`entity`**: Optional. Present on older agents with `isAttributeInEntity: true` variables; absent on `Agent007`.
- **`maxTokens`**, **`toolChoice`**: Agent-level inference parameters. Enum values for `toolChoice` observed: `"Auto"` (capitalized, not the lowercase `auto` used by `GenAICommons.ENUM_ToolChoice` at runtime). Other likely values: `"none"`, `"Any"`, `"tool"`.
- **`temperature`**, **`topP`**: Not observed in any sample — omitted when not set.
- **No `UserAccessApproval`/`access` field** on tools. That's a runtime-only concern (set on `AgentCommons.Tool` entity, not the document). **This is a correction to earlier versions of this proposal.**

### Observed documents in test3 project

| Document | Type | Notes |
|---|---|---|
| `AgentEditorCommons.InformationExtractorAgent` | Agent | Older format, `Excluded: true`, no model reference |
| `AgentEditorCommons.SummarizationAgent` | Agent | Older format |
| `AgentEditorCommons.TranslationAgent` | Agent | Older format, bound to `System.Language` |
| `AgentEditorCommons.ProductDescription` | Agent | Older format, bound to `AgentCommons.ProductDescriptionGenerator_EXAMPLE` |
| `Agents.Agent007` | Agent | New format with model reference, MCP tool, KB tool |
| `Agents.MyFirstModel` | Model | Provider `MxCloudGenAI`, key → `Agents.LLMKey` |
| `Agents.Knowledge_base` | Knowledge Base | Provider `MxCloudGenAI`, key → `Agents.LLMKey` |
| `Agents.Consumed_MCP_service` | Consumed MCP Service | Protocol v2025_03_26, timeout 30s |

## Proposed MDL Syntax

### SHOW AGENTS

```sql
show agents [in module]
```

Output:

| Qualified Name | Module | Name | UsageType | Entity | Variables |
|----------------|--------|------|-----------|--------|-----------|
| AgentEditorCommons.InformationExtractorAgent | AgentEditorCommons | InformationExtractorAgent | Task | AgentCommons.InformationExtractor_EXAMPLE | Information |
| AgentEditorCommons.SummarizationAgent | AgentEditorCommons | SummarizationAgent | Task | | |
| AgentEditorCommons.TranslationAgent | AgentEditorCommons | TranslationAgent | Task | System.Language | Description |
| AgentEditorCommons.ProductDescription | AgentEditorCommons | ProductDescription | Task | AgentCommons.ProductDescriptionGenerator_EXAMPLE | ProductName, Keywords |

### DESCRIBE AGENT

```sql
describe agent AgentEditorCommons.TranslationAgent
```

Output (round-trippable MDL):

```sql
create agent AgentEditorCommons."TranslationAgent" (
  UsageType: task,
  entity: System.Language,
  variables: ("description": EntityAttribute),
  SystemPrompt: 'Translate the given text into {{Description}}.',
  UserPrompt: 'What is a multi-agent AI system?...'
);
/
```

### CREATE AGENT

The syntax follows the same shape as `create rest client`: top-level configuration in `(...)` followed by a `{...}` body containing one block per attached resource (`tool`, `knowledge base`, `mcp service`). Simple agents with no resources omit the body entirely.

**Simple task agent (no body needed):**

```sql
create agent MyModule."SentimentAnalyzer" (
  UsageType: task,
  entity: MyModule.FeedbackItem,
  variables: ("FeedbackText": EntityAttribute),
  model: MyModule.GPT4Model,
  SystemPrompt: 'Analyze the sentiment of {{FeedbackText}}. Classify as positive, negative, or neutral.',
  UserPrompt: '{{FeedbackText}}'
);
```

**Agent with tools, knowledge bases, and MCP services (matches `Agents.Agent007`):**

```sql
create agent Agents."Agent007" (
  UsageType: task,
  model: Agents.MyFirstModel,
  MaxTokens: 16384,
  ToolChoice: Auto,
  description: 'doing your stuff for you',
  SystemPrompt: 'Do you intereesting and useful stuff that makes me money',
  UserPrompt: 'Just do it'
)
{
  mcp service Agents.Consumed_MCP_service {
    Enabled: true
  }

  knowledge base My_mem {
    source: Agents.Knowledge_base,
    collection: 'agent1-collection',
    MaxResults: 3,
    description: 'My memory of useful stuff',
    Enabled: true
  }
};
```

**Block-level property reference:**

Each block maps to one entry in the agent's `Contents` JSON (`tools[]` for TOOL/MCP SERVICE, `knowledgebaseTools[]` for KNOWLEDGE BASE). Block IDs (the `id` UUID field in JSON) are auto-generated by the writer.

| Block | Referenced by | Properties | Maps to JSON field |
|---|---|---|---|
| `mcp service <QualifiedName> { ... }` | ConsumedMCPService document | `Enabled`, `description` | `tools[]` entry with `toolType: "mcp"`, `document: {...}` |
| `tool <Name> { microflow: ... }` | microflow name | `microflow`, `Enabled`, `description` | `tools[]` entry with `toolType: "microflow"` *(shape speculative — see Open Questions)* |
| `knowledge base <Name> { source: ... }` | KB document via `source:` | `source` (required), `collection`, `MaxResults`, `description`, `Enabled` | `knowledgebaseTools[]` entry |

### DROP AGENT

```sql
drop agent MyModule."SentimentAnalyzer"
```

### What Goes Where: Design-Time vs. Call-Time

Everything inside `create agent` — including the properties in the top-level `(...)` — is **design-time configuration that is stored in the agent document**. None of it is an invocation parameter. Runtime inputs are supplied at the `call agent` site.

The layering follows the same pattern as `create rest client`:

| Layer | REST CLIENT | AGENT |
|---|---|---|
| **Document-level static config** (stored in document) | `BaseUrl`, `authentication` | `UsageType`, `description`, `entity`, `model`, `MaxTokens`, `ToolChoice`, `SystemPrompt`, `UserPrompt` |
| **Input contract** (what the caller must bring at call time) | `parameters: ($id: string)` on each operation | `variables: ("Topic": string, ...)` on the agent |
| **Attached resources** (body blocks) | `operation` blocks | `tool` / `knowledge base` / `mcp service` blocks |
| **Runtime invocation** (values supplied at call site) | `send rest request Mod.Api.GetItems (id = $x)` | `call agent with HISTORY $agent request $req context $obj` |

In other words:

- `UsageType`, `entity`, `SystemPrompt`, `UserPrompt` are the same kind of property as `BaseUrl` on a REST client — baked into the document, changed by editing the document.
- `variables: (...)` is the same kind of property as `parameters: (...)` on a REST operation — it declares the **schema** of what the caller must supply, not the values. Actual values arrive at runtime: for `EntityAttribute` variables, they're read from matching attributes on the `context` object; for free-form variables (future extension), they'd be passed directly.
- `tool`, `knowledge base`, `mcp service` blocks describe **capabilities the agent carries with it** — the LLM can invoke them autonomously at runtime, but they aren't something the caller passes in.

Example — all of this is stored in the agent document:

```sql
create agent Reviews."SentimentAnalyzer" (
  UsageType: task,                                             -- design-time mode
  entity: Reviews.ProductReview,                               -- context entity contract
  variables: ("ProductName": EntityAttribute,                  -- input contract
              "ReviewText": EntityAttribute),
  SystemPrompt: 'Analyze the review for {{ProductName}}.',     -- prompt template
  UserPrompt: '{{ReviewText}}'                                 -- prompt template
);
```

And this is the runtime call — the only place values flow in:

```sql
call agent without HISTORY $agent context $Review into $response;
-- $Review is a Reviews.ProductReview instance;
-- its ProductName and ReviewText attributes satisfy the Variables contract.
```

### Syntax Design Rationale

| Decision | Rationale |
|----------|-----------|
| `agent` as document type keyword | Matches `Metadata.ReadableTypeName = "agent"` and Mendix UI terminology |
| Top-level `(key: value)` config + `{...}` body with singular blocks | Mirrors `create rest client ... (...) { operation Name {...} }` exactly — same shape, same mental model |
| `model: <QualifiedName>` in top-level config | Agent documents can reference a Model document directly via the `model` JSON field (confirmed in `Agent007`). Making it a peer of `UsageType` mirrors how the Agent Editor UI presents it |
| `ToolChoice: Auto` PascalCase enum literal | Matches the real JSON value (`"Auto"`), which differs from the lowercase `auto` used by `GenAICommons.ENUM_ToolChoice` at runtime. Values: `Auto`, `none`, `Any`, `tool` |
| `MaxTokens: <int>` on the agent | Matches the JSON `maxTokens` field; agent-level inference parameter |
| `tool`, `knowledge base`, `mcp service` as singular block types | Matches the `operation` singular used in REST CLIENT; each block defines one resource |
| `mcp service <QualifiedName> { Enabled, description }` | The name is the qualified name of a ConsumedMCPService document (the whole service is attached as a bundle of tools) |
| `knowledge base <Name> { source: <doc>, collection, MaxResults, ... }` | `<Name>` is the per-agent identifier stored in JSON `name`; `source:` references the KB document. Matches `Agent007`'s `My_mem` KB entry |
| `tool <Name> { microflow: ..., description, Enabled }` | Speculative: microflow-tool JSON shape not yet observed (test3 only has MCP tools). Final shape TBD when we capture a sample |
| `variables: (...)` is the input-schema analog of REST CLIENT's `parameters: (...)` | Declares what the caller must supply; values flow in via the `context` object at the `call agent` site. Inline form matches REST CLIENT's `parameters: ($id: string)` |
| No `access:` on tool blocks | `UserAccessApproval` is NOT stored in the agent document JSON — it's a runtime-only concern on the `AgentCommons.Tool` entity. (Earlier drafts of this proposal incorrectly placed it on the block.) |
| Body omitted when there are no tools/KB/MCP | Same concession REST CLIENT makes implicitly — empty bodies are awkward; drop them |
| Prompts as string literals | Consistent with other MDL string properties; `{{var}}` placeholders are just text |
| Auto-generated `id` UUIDs on block entries | Each tool/KB entry has a UUID `id` in the JSON (Studio Pro-generated). The MDL writer will generate these; they round-trip stably through `describe` |

## What Already Exists

| Layer | Status | Location |
|-------|--------|----------|
| **Go type** | No | — |
| **BSON parser** | No (CustomBlobDocuments not parsed) | — |
| **Reader/Catalog** | No | — |
| **Grammar** | No | — |
| **AST** | No | — |
| **Visitor** | No | — |
| **Executor** | No | — |
| **Generated metamodel** | Partial | `generated/metamodel/types.go` has `CustomBlobDocuments$CustomBlobDocument` |

## Implementation Plan

### Phase 1: Read Support (SHOW/DESCRIBE)

#### 1.1 Add Model Types

In a new file `sdk/agents/types.go` (or extend an existing domain):

```go
package agents

import "github.com/nicholasgasior/modelsdk-go/model"

type agent struct {
    ContainerID   model.ID
    ID            model.ID
    Name          string
    documentation string

    // Parsed from Contents json
    description  string
    SystemPrompt string
    UserPrompt   string
    UsageType    string        // "task", "Conversational"
    variables    []Variable
    Tools        []ToolRef     // tools[] array
    KBTools      []KBToolRef   // knowledgebaseTools[] array
    model        *DocRef       // optional, points to a model document
    entity       *EntityRef    // optional, points to a domain entity
    MaxTokens    *int          // optional
    ToolChoice   string        // optional: "Auto", "none", "Any", "tool"
    Temperature  *float64      // optional, not yet observed
    TopP         *float64      // optional, not yet observed
}

type Variable struct {
    key                 string
    IsAttributeInEntity bool
}

type EntityRef struct {
    DocumentID    string // UUID of the entity's domain model
    QualifiedName string // Module.EntityName
}

type DocRef struct {
    DocumentID    string // UUID of the referenced CustomBlobDocument
    QualifiedName string // Module.DocumentName
}

// Entry in the agent's tools[] array
type ToolRef struct {
    ID          string  // per-tool UUID (generated by writer)
    Name        string
    description string
    Enabled     bool
    ToolType    string  // "mcp" | "microflow" (microflow shape TBD)
    Document    *DocRef // set when ToolType=="mcp", references ConsumedMCPService
    microflow   string  // set when ToolType=="microflow" (speculative)
}

// Entry in the agent's knowledgebaseTools[] array
type KBToolRef struct {
    ID                   string
    Name                 string
    description          string
    Enabled              bool
    ToolType             string  // empty string in observed sample
    Document             *DocRef // references KnowledgeBase document
    CollectionIdentifier string
    MaxResults           int
}

// Peer document types (same wrapper, different Contents json)

type model struct {
    ContainerID   model.ID
    ID            model.ID
    Name          string
    documentation string

    type        string                 // Portal-populated, usually empty
    DisplayName string                 // Portal-populated
    Provider    string                 // "MxCloudGenAI"
    Fields      map[string]interface{} // providerFields — shape depends on provider
    KeyConstant *ConstantRef           // providerFields.key → string constant
}

type KnowledgeBase struct {
    ContainerID   model.ID
    ID            model.ID
    Name          string
    documentation string

    Provider    string                 // "MxCloudGenAI"
    Fields      map[string]interface{} // providerFields (includes modelDisplayName, modelName)
    KeyConstant *ConstantRef
}

type ConsumedMCPService struct {
    ContainerID              model.ID
    ID                       model.ID
    Name                     string
    documentation            string

    ProtocolVersion          string // "v2024_11_05" | "v2025_03_26"
    version                  string // app-specified version
    InnerDocumentation       string // Contents.documentation (free text)
    ConnectionTimeoutSeconds int
}

type ConstantRef struct {
    DocumentID    string
    QualifiedName string // e.g. "Agents.LLMKey"
}
```

#### 1.2 Add BSON Parser

In `sdk/mpr/parser_customblob.go` (generic — handles all four document types):

- Parse `CustomBlobDocuments$CustomBlobDocument` documents
- Dispatch by `CustomDocumentType`:
  - `"agenteditor.agent"` → decode Contents JSON as `agent`
  - `"agenteditor.model"` → decode as `model`
  - `"agenteditor.knowledgebase"` → decode as `KnowledgeBase`
  - `"agenteditor.consumedMCPService"` → decode as `ConsumedMCPService`
  - unknown → store raw Contents, warn
- Store in per-type maps on the reader

The parser should be tolerant: unknown JSON fields in `Contents` are ignored (the agent editor extension may add fields in future versions).

#### 1.3 Add Reader Methods

```go
func (r *Reader) agents() []*agenteditor.Agent
func (r *Reader) AgentByQualifiedName(name string) *agenteditor.Agent

func (r *Reader) models() []*agenteditor.Model
func (r *Reader) ModelByQualifiedName(name string) *agenteditor.Model

func (r *Reader) KnowledgeBases() []*agenteditor.KnowledgeBase
func (r *Reader) KnowledgeBaseByQualifiedName(name string) *agenteditor.KnowledgeBase

func (r *Reader) ConsumedMCPServices() []*agenteditor.ConsumedMCPService
func (r *Reader) ConsumedMCPServiceByQualifiedName(name string) *agenteditor.ConsumedMCPService
```

#### 1.4 Add Catalog Tables

- `CATALOG.AGENTS` (module, name, qualified_name, usage_type, entity, model, variables, tool_count, kb_count)
- `CATALOG.MODELS` (module, name, qualified_name, provider, key_constant)
- `CATALOG.KNOWLEDGE_BASES` (module, name, qualified_name, provider, key_constant)
- `CATALOG.CONSUMED_MCP_SERVICES` (module, name, qualified_name, protocol_version, timeout_seconds)

#### 1.5 Add Grammar/AST/Visitor/Executor

- Grammar: `show {agents | models | knowledge bases | consumed mcp services} [in module]`
- Grammar: `describe {agent | model | knowledge base | consumed mcp service} qualifiedName`
- AST: `ShowCustomBlobStmt`, `DescribeCustomBlobStmt` (discriminated by type enum)
- Executor: format output using standard table/MDL patterns

**Recommended implementation order** (matches user preference to start with MODEL):
1. Generic wrapper parser + `model` type + `show models` + `describe model` (smallest Contents JSON)
2. `ConsumedMCPService` (also small)
3. `KnowledgeBase` (similar shape to Model)
4. `agent` (largest, depends on the other three for resolving `model`/`document` references in its body)

### Phase 2: Write Support (CREATE/DROP)

#### 2.1 Add BSON Writer

In `sdk/mpr/writer_customblob.go` (generic wrapper for all four types):

- Serialize any of `agent` / `model` / `KnowledgeBase` / `ConsumedMCPService` structs to a `CustomBlobDocuments$CustomBlobDocument` BSON
- Set `CustomDocumentType` per type (`agenteditor.agent`, `agenteditor.model`, etc.)
- Set `Metadata.CreatedByExtension = "extension/agent-editor"`
- Set `Metadata.ReadableTypeName` per type (`"agent"`, `"model"`, `"knowledge base"`, `"consumed mcp service"`)
- Serialize `Contents` as a JSON string (per-type encoder)
- Set `Excluded = false`, `ExportLevel = "Hidden"` (matches the new Agent Editor defaults)
- Generate stable UUIDs for `$ID` and `Metadata.$ID`
- For `agent`: generate UUIDs for `id` field on each `tools[]` and `knowledgebaseTools[]` entry
- For `model` / `KnowledgeBase`: resolve the `key` constant reference to `{documentId, qualifiedName}` by looking up the String constant in the reader

#### 2.2 Add Grammar/AST/Visitor/Executor for CREATE/DROP

- Grammar: `create agent qualifiedName properties variablesClause?`
- AST: `CreateAgentStmt`, `DropAgentStmt`
- Executor: validate, write BSON, register in module

#### 2.3 Validation

- Entity reference must exist (if specified)
- Variables marked `EntityAttribute` must correspond to attributes on the referenced entity
- `UsageType` must be a known value (`task` or `Conversational`)
- Variable names used in `{{...}}` in prompts should match declared variables (warning, not error)

### Phase 3: Integration & Catalog

#### 3.1 Catalog Integration

- Include agents in `show structure` output
- Add `CATALOG.AGENTS` table for SQL queries
- Include agent references in `show references` / `show impact`
- Wire into `refresh catalog` (both fast and full modes)

#### 3.2 Version Gating

- Agent documents (`CustomBlobDocuments$CustomBlobDocument`) require Mendix 11.x
- Add to `sdk/versions/mendix-11.yaml`:
  ```yaml
  agents:
    agent_document:
      min_version: "11.9.0"
      mdl: "create agent Module.Name (...) { tool ... { ... } ... }"
      notes: "Requires AgentEditorCommons marketplace module"
  ```
- Executor pre-check: `checkFeature("agent_document")` before CREATE

#### 3.3 LSP Support

- Hover on agent names shows system prompt summary
- Go-to-definition navigates to agent document
- Completion for `describe agent` with agent names

### Phase 4: Supporting Document Types and Microflow Activities

Full agent support requires MDL coverage of related `CustomBlobDocument` types and the new "Call Agent" microflow activity. These are split into sub-phases but all are needed for the examples in this proposal to work end-to-end.

#### 4.1 `create model` Document

Models are peer `CustomBlobDocument`s that reference a Mendix Cloud GenAI Portal key stored in a **String constant**. The minimum input from the user is the provider and the constant reference — Portal metadata (`displayName`, `keyId`, `keyName`, `environment`, `resourceName`, etc.) is filled by Studio Pro when the user clicks **Test Key** and shouldn't be user-set in MDL.

Matches the observed BSON for `Agents.MyFirstModel`:

```sql
create model Agents."MyFirstModel" (
  Provider: MxCloudGenAI,
  key: Agents.LLMKey
);
```

`describe model` may show Portal-populated fields when present (round-trip preserves them, but they're not user-editable in MDL):

```sql
-- What DESCRIBE produces for a model that has been activated against the Portal
create model Agents."MyFirstModel" (
  Provider: MxCloudGenAI,
  key: Agents.LLMKey,
  DisplayName: 'GPT-4 Turbo',          -- Portal-populated, read-only in MDL
  KeyName: 'prod-gpt4',                -- Portal-populated, read-only in MDL
  Environment: 'production'            -- Portal-populated, read-only in MDL
);
```

At runtime, `AgentEditorCommons.ASU_AgentEditor` reads the constant and creates the corresponding `GenAICommons.DeployedModel`.

**JSON output shape:**
```json
{
  "type": "",
  "name": "",
  "displayName": "<portal-populated or empty>",
  "provider": "MxCloudGenAI",
  "providerFields": {
    "environment": "", "deepLinkURL": "", "keyId": "", "keyName": "", "resourceName": "",
    "key": { "documentId": "<uuid>", "qualifiedName": "Agents.LLMKey" }
  }
}
```

#### 4.2 `create knowledge base` Document

Same shape as Model, but `providerFields` carries embedding-model info instead of model-resource info. User-settable fields are just `Provider` and `key`:

```sql
create knowledge base Agents."Knowledge_base" (
  Provider: MxCloudGenAI,
  key: Agents.LLMKey
);
```

`describe` can round-trip Portal-populated fields:

```sql
create knowledge base Agents."Knowledge_base" (
  Provider: MxCloudGenAI,
  key: Agents.LLMKey,
  ModelDisplayName: 'text-embedding-3-large',   -- Portal-populated
  ModelName: 'text-embedding-3-large'           -- Portal-populated
);
```

Referenced from agents via `knowledge base <Name> { source: <QualifiedName>, ... }` blocks inside the agent body.

**JSON output shape:**
```json
{
  "name": "",
  "provider": "MxCloudGenAI",
  "providerFields": {
    "environment": "", "deepLinkURL": "", "keyId": "", "keyName": "",
    "modelDisplayName": "", "modelName": "",
    "key": { "documentId": "<uuid>", "qualifiedName": "Agents.LLMKey" }
  }
}
```

#### 4.3 `create consumed mcp service` Document

Matches the observed BSON for `Agents.Consumed_MCP_service`. Endpoint and credentials are **not** part of the document — they're runtime configuration on the `MCPClient.ConsumedMCPService` entity. The document only carries protocol version, app-level version, timeout, and documentation.

```sql
create consumed mcp service Agents."Consumed_MCP_service" (
  ProtocolVersion: v2025_03_26,
  version: '0.0.1',
  ConnectionTimeoutSeconds: 30,
  documentation: 'Description of what this MCP service provides'
);
```

Referenced from agents via `mcp service <QualifiedName> { ... }` blocks inside the agent body.

**JSON output shape:**
```json
{
  "protocolVersion": "v2025_03_26",
  "documentation": "description of what this mcp service provides",
  "version": "0.0.1",
  "connectionTimeoutSeconds": 30
}
```

> **Note:** `documentation` as a top-level property maps to the JSON `documentation` field (inside `Contents`), not to the outer BSON `documentation` field on the CustomBlobDocument wrapper. Two different fields with the same name — the MDL writer sets both consistently.

#### 4.4 `call agent` / `NEW CHAT for agent` Microflow Activities

New MDL microflow statements mapping to the Agents Kit toolbox actions (see the "New MDL Statement" section under "Building Smart Apps" for syntax):

| MDL Statement | Java Action | Purpose |
|---------------|-------------|---------|
| `call agent with HISTORY $agent request $req [context $obj] into $response` | `AgentCommons.Agent_Call_WithHistory` | Call a conversational agent with chat history |
| `call agent without HISTORY $agent [context $obj] [request $req] [FILES $fc] into $response` | `AgentCommons.Agent_Call_WithoutHistory` | Call a single-call (Task) agent |
| `NEW CHAT for agent $agent action microflow <mf> [context $obj] [model $dm] into $ChatContext` | `AgentCommons.ChatContext_Create_ForAgent` | Create a ChatContext wired to an agent |

These need a new BSON activity type (or mapping to the generic Java action call — see Open Question 4).

#### 4.5 ALTER AGENT

Follows the same shape as `alter page` — in-place modifications to top-level properties and body blocks:

```sql
alter agent MyModule."SentimentAnalyzer" {
  set SystemPrompt = 'New prompt with {{Variable}}.';
  set variables = ("FeedbackText": EntityAttribute, "NewVar": string);
  set model = MyModule.OtherModel;
  set ToolChoice = none;

  insert mcp service MyModule.NewMCPService {
    Enabled: true
  };

  insert knowledge base NewKB {
    source: MyModule.OtherKB,
    collection: 'other-collection',
    MaxResults: 5
  };

  drop mcp service MyModule.OldMCPService;
};
```

## Building Smart Apps with MDL: End-to-End Examples

This section demonstrates how MDL agent support, combined with the existing agentic marketplace modules (GenAICommons, AgentCommons, MCPClient, MCPServer, ConversationalUI, and one of the connectors such as MxGenAIConnector), enables building complete AI-powered applications entirely from MDL scripts.

> **Sources:** This section follows the official Mendix documentation for the [Agent Editor](https://docs.mendix.com/appstore/modules/genai/genai-for-mx/agent-editor/), [Agent Commons](https://docs.mendix.com/appstore/modules/genai/genai-for-mx/agent-commons/), [Conversational UI](https://docs.mendix.com/appstore/modules/genai/genai-for-mx/conversational-ui/), [Mendix Cloud GenAI Connector](https://docs.mendix.com/appstore/modules/genai/mx-cloud-genai/MxGenAI-connector/), [MCP Client](https://docs.mendix.com/appstore/modules/genai/mcp-modules/mcp-client/), and [MCP Server](https://docs.mendix.com/appstore/modules/genai/mcp-modules/mcp-server/).

### Architecture Overview

A "smart app" in Mendix typically has these layers, all expressible in MDL:

```
┌─────────────────────────────────────────────────────────────────┐
│                     Conversational UI                           │
│         Chat widget, tool approval, trace monitoring            │
├─────────────────────────────────────────────────────────────────┤
│                      agent Layer                                │
│  agent documents (create agent) — prompts, variables,           │
│           tools, knowledge bases, mcp servers                   │
├──────────────┬────────────────────┬─────────────────────────────┤
│    Tools     │   knowledge bases  │      mcp services           │
│  microflows  │   RAG retrieval    │  external tool servers      │
├──────────────┴────────────────────┴─────────────────────────────┤
│              model Documents + MxGenAIConnector                 │
│      model key constant → DeployedModel at runtime              │
├─────────────────────────────────────────────────────────────────┤
│                    Domain model                                 │
│           entities, associations, enumerations                  │
└─────────────────────────────────────────────────────────────────┘
```

### Key Correction: How Agents Work in Mendix

Unlike the initial draft of this proposal, agents in Mendix are **not** wired up by building the request manually in an action microflow. The correct flow is:

1. **Studio Pro design time** — the developer creates agent documents (and model documents) in Studio Pro. Tools, knowledge bases, and MCP servers are **attached to the agent in the agent document itself** (not added at runtime).
2. **Model key** — a Mendix Cloud GenAI Portal key is stored in a String constant on the model document. At runtime, `ASU_AgentEditor` (registered as after-startup microflow) reads the key and auto-creates the corresponding `GenAICommons.DeployedModel`.
3. **Call Agent activity** — in a microflow, a single **"Call Agent With History"** or **"Call Agent Without History"** toolbox action does everything: resolve the agent's in-use version, select its deployed model, replace variable placeholders from the context object, wire in tools/knowledge bases/MCP servers declared on the agent, and call the LLM.
4. **Conversational UI** — to use the agent in a chat, call **"New Chat for Agent"** which creates a `ChatContext` pre-configured with the agent's deployed model, system prompt, and action microflow. The action microflow for chat just calls **"Call Agent With History"** with the request built by `default Preprocessing`.

### Prerequisites

Before any agent can be used, the following one-time setup is required (see the [Agent Editor prerequisites](https://docs.mendix.com/appstore/modules/genai/genai-for-mx/agent-editor/)):

```sql
-- 1. Encryption key must be 32 characters (App > Settings > Configuration).
--    The Encryption module is a prerequisite; its setup is outside MDL.

-- 2. Register ASU_AgentEditor as an after-startup microflow
alter settings model AfterStartupMicroflow = AgentEditorCommons.ASU_AgentEditor;

-- 3. A model key constant must hold a Mendix Cloud GenAI Portal key.
--    Use one constant per model. The agent editor references this constant
--    in the model document.
create constant "MyApp"."DefaultModelKey" (
  type: string,
  DefaultValue: ''   -- set via environment config or configuration UI
);

-- 4. Ensure the required module roles are assigned.
--    MxGenAIConnector.Administrator is needed to configure the connector.
--    AgentCommons.AgentAdmin is needed to manage agents in the runtime UI.

-- 5. Exclude the auto-created /agenteditor folder from version control.
--    (This is handled in .gitignore, outside MDL.)
```

### New MDL Statement: `call agent` Microflow Activity

Because "Call Agent" is a first-class Mendix toolbox activity (distinct from a generic Java action call), this proposal also introduces a corresponding MDL microflow statement:

```
call agent with HISTORY <agent> request <request> [context <obj>] into $response
call agent without HISTORY <agent> [context <obj>] [request <req>] [FILES <fc>] into $response
NEW CHAT for agent <agent> action microflow <microflow> [context <obj>] [model <dm>] into $ChatContext
```

These map directly to the `AgentCommons.Agent_Call_WithHistory`, `AgentCommons.Agent_Call_WithoutHistory`, and `AgentCommons.ChatContext_Create_ForAgent` Java actions (all exposed in the **"Agents Kit"** toolbox category). The MDL form exists so these show up as the actual "Call Agent" activity in Studio Pro rather than as opaque Java action calls.

---

### Example 1: Customer Support Agent with Tools

A conversational agent that helps customer support reps by looking up orders, checking shipment status, and drafting responses. Tools (microflows) are attached to the agent in the agent document — at runtime the "Call Agent" activity handles tool invocation automatically.

#### Step 1: Domain Model

```sql
-- Domain model for the support system

@position(100, 100)
create persistent entity Support."Customer" (
  "Name": string(200) not null error 'Name is required',
  "Email": string(200),
  "Phone": string(50),
  "AccountTier": enumeration(Support.AccountTier)
);

@position(350, 100)
create persistent entity Support."Order" (
  "OrderNumber": string(50) not null error 'Order number is required',
  "OrderDate": datetime,
  "TotalAmount": decimal,
  "status": enumeration(Support.OrderStatus)
);

@position(600, 100)
create persistent entity Support."SupportTicket" (
  "Subject": string(200),
  "description": string(unlimited),
  "Priority": enumeration(Support.TicketPriority),
  "Resolution": string(unlimited),
  "IsResolved": boolean default false
);

create association Support."Order_Customer"
  from Support."Order" to Support."Customer";

create association Support."SupportTicket_Customer"
  from Support."SupportTicket" to Support."Customer";

create association Support."SupportTicket_Order"
  from Support."SupportTicket" to Support."Order";

create enumeration Support."OrderStatus" (
  Pending = 'Pending',
  Shipped = 'Shipped',
  Delivered = 'Delivered',
  Returned = 'Returned'
);

create enumeration Support."TicketPriority" (
  Low = 'Low',
  Medium = 'Medium',
  High = 'High',
  Urgent = 'Urgent'
);

create enumeration Support."AccountTier" (
  Standard = 'Standard',
  Premium = 'Premium',
  Enterprise = 'Enterprise'
);
```

#### Step 2: Tool Microflows

Tool microflows take the **Mendix data types that the LLM should fill in** as input parameters and must return a `string` (which becomes the tool result shown to the model). The Agent Editor infers the tool's JSON schema from the microflow signature — so parameter names and types are what the LLM sees.

```sql
/**
 * Tool microflow: Look up a customer by email address.
 * The Agent Editor will expose this as a tool with input parameter "Email".
 */
create microflow Support."Tool_LookupCustomer" (
  $Email: string
)
returns string
begin
  retrieve $Customer from database Support.Customer
    where Email = $Email limit 1;

  if $Customer != empty then
    return 'Customer: ' + $Customer/Name
      + ', Tier: ' + getKey($Customer/AccountTier)
      + ', Phone: ' + $Customer/Phone;
  else
    return 'No customer found with email: ' + $Email;
  end if;
end;
/

/**
 * Tool microflow: Look up recent orders for a customer by name.
 */
create microflow Support."Tool_GetOrders" (
  $CustomerName: string
)
returns string
begin
  retrieve $Customer from database Support.Customer
    where Name = $CustomerName limit 1;

  if $Customer = empty then
    return 'Customer not found: ' + $CustomerName;
  end if;

  retrieve $OrderList from database Support.Order
    where Support.Order_Customer = $Customer;

  declare $Result string = '';
  loop $Order in $OrderList
  begin
    set $Result = $Result + 'Order ' + $Order/OrderNumber
      + ' (' + formatDateTime($Order/OrderDate, 'yyyy-MM-dd') + ')'
      + ' - ' + formatDecimal($Order/TotalAmount, 2) + ' EUR'
      + ' - Status: ' + getKey($Order/status) + '\n';
  end loop;

  return if $Result = '' then 'No orders found' else $Result;
end;
/

/**
 * Tool microflow: Create a support ticket.
 * Multiple primitive parameters become structured input for the tool.
 */
create microflow Support."Tool_CreateTicket" (
  $Subject: string,
  $description: string,
  $Priority: enum Support.TicketPriority
)
returns string
begin
  $Ticket = create Support.SupportTicket (
    Subject = $Subject,
    description = $description,
    Priority = $Priority,
    IsResolved = false
  );
  commit $Ticket;

  return 'Ticket created (ID: ' + toString($Ticket/System.id) + ').';
end;
/
```

#### Step 3: Agent Document

Tools, knowledge bases, and MCP servers are declared in the **agent document itself** — not attached at runtime in the action microflow. This is the key fix vs. the earlier draft.

```sql
-- The agent definition — stored as a CustomBlobDocument in the project
create agent Support."CustomerSupportAgent" (
  UsageType: Conversational,
  description: 'Customer support agent with lookup and ticketing tools',
  SystemPrompt: 'You are a helpful customer support agent for an e-commerce company.

Your capabilities:
- Look up customer information by email
- check order history and shipment status
- create support tickets for unresolved issues

Guidelines:
- Always verify the customer identity before sharing order details
- for Premium and Enterprise customers, prioritize their requests
- if you cannot resolve an issue, create a support ticket
- Be empathetic and professional in your responses'
)
{
  tool LookupCustomer {
    microflow: Support.Tool_LookupCustomer,
    description: 'Look up a customer by their email address',
    access: VisibleForUser
  }

  tool GetOrders {
    microflow: Support.Tool_GetOrders,
    description: 'Get recent orders for a customer by name',
    access: VisibleForUser
  }

  tool CreateTicket {
    microflow: Support.Tool_CreateTicket,
    description: 'Create a new support ticket with the given subject, description, and priority',
    access: UserConfirmationRequired
  }
};
```

The `access` property maps to `GenAICommons.ENUM_UserAccessApproval`:
- `HiddenForUser` — tool executes silently
- `VisibleForUser` — tool call is shown in the chat UI but executes automatically
- `UserConfirmationRequired` — tool call is shown and user must approve before execution

#### Step 4: Runtime Wiring — Action Microflow for Chat

With tools declared on the agent, the action microflow becomes a simple two-step: preprocess, then call the agent. The "Call Agent With History" activity handles tool invocation, knowledge base retrieval, and the LLM round-trips internally.

```sql
/**
 * Action microflow for the Customer Support chat.
 * Wired up by "New Chat for Agent" when the chat is created.
 *
 * @param $ChatContext The conversation context from the chat widget
 * @returns Boolean indicating success
 */
create microflow Support."Chat_CustomerSupport" (
  $ChatContext: ConversationalUI.ChatContext
)
returns boolean
begin
  -- 1. Default Preprocessing: extract user message, build Request with history
  $request = call microflow ConversationalUI.ChatContext_Preprocessing(
    ChatContext = $ChatContext
  ) on error rollback;

  if $request = empty then
    return false;
  end if;

  -- 2. Retrieve the agent (created automatically from the agent document
  --    when ASU_AgentEditor runs at startup)
  retrieve $agent from database AgentCommons.Agent
    where _QualifiedName = 'Support.CustomerSupportAgent' limit 1;

  -- 3. Call Agent With History — single activity that:
  --    - Selects the in-use version + its deployed model
  --    - Wires in the agent's tools, knowledge bases, MCP servers
  --    - Calls Chat Completions
  --    - Handles tool-call round-trips
  call agent with HISTORY $agent request $request into $response
    on error rollback;

  -- 4. Update the chat UI with the response (same as any ConversationalUI flow)
  if $response != empty and $response/GenAICommons.Response_Message != empty then
    $message = call microflow ConversationalUI.ChatContext_UpdateAssistantResponse(
      ChatContext = $ChatContext,
      MessageStatus = ConversationalUI.ENUM_MessageStatus.Success,
      response = $response
    ) on error rollback;
    return true;
  else
    return false;
  end if;
end;
/
```

#### Step 5: Page with "Start Chat" Button

The chat page uses `New Chat for agent` to create a `ChatContext` pre-configured with the agent's model, system prompt, and action microflow. This replaces the manual ProviderConfig wiring from the earlier draft.

```sql
/**
 * Microflow that opens a support chat. Called from a "Start Chat" button.
 * Uses "New Chat for Agent" to create a ChatContext configured with this agent.
 */
create microflow Support."ACT_StartSupportChat" ()
begin
  retrieve $agent from database AgentCommons.Agent
    where _QualifiedName = 'Support.CustomerSupportAgent' limit 1;

  NEW CHAT for agent $agent
    action microflow Support.Chat_CustomerSupport
    into $ChatContext
    on error rollback;

  show page Support.SupportChat($ChatContext = $ChatContext);
end;
/

/**
 * Customer support chat page.
 * Data source is the ChatContext passed from ACT_StartSupportChat.
 */
create page Support."SupportChat" (
  title: 'Customer Support',
  layout: Atlas_Core.Atlas_Default
) {
  header h1 {
    dynamictext title (caption: 'AI Customer Support')
  }
  dataview chatView (datasource: context ConversationalUI.ChatContext) {
    -- The ConversationalUI chat snippet renders the conversation,
    -- send box, tool call approvals, and message history
    snippetcall chatWidget (snippet: ConversationalUI.Snippet_Output_WithHistory)
  }
};
/
```

#### Step 6: Security

```sql
-- Module roles
create module role Support."user";
create module role Support."Admin";

-- Entity access
grant Support.User on Support.Customer (read *);
grant Support.User on Support.Order (read *);
grant Support.User on Support.SupportTicket (create, read *, write *);
grant Support.Admin on Support.Customer (create, delete, read *, write *);
grant Support.Admin on Support.Order (create, delete, read *, write *);
grant Support.Admin on Support.SupportTicket (create, delete, read *, write *);

-- Microflow access
grant execute on microflow Support.ACT_StartSupportChat to Support.User;
grant execute on microflow Support.Chat_CustomerSupport to Support.User;
-- Tool microflows must be callable because the agent invokes them
grant execute on microflow Support.Tool_LookupCustomer to Support.User;
grant execute on microflow Support.Tool_GetOrders to Support.User;
grant execute on microflow Support.Tool_CreateTicket to Support.User;

-- Page access
grant view on page Support.SupportChat to Support.User;
```

---

### Example 2: MCP-Powered Research Agent

An agent that connects to external MCP servers to access tools like web search, file reading, and database queries. The key insight: the consumed MCP service is a **document** (created in Studio Pro alongside the agent), and it's attached to the agent document directly — no runtime wiring needed.

#### Step 1: Domain Model

```sql
create persistent entity Research."ResearchProject" (
  "title": string(200),
  "Objective": string(unlimited),
  "status": enumeration(Research.ProjectStatus),
  "Summary": string(unlimited)
);

create enumeration Research."ProjectStatus" (
  InProgress = 'In Progress',
  Completed = 'Completed',
  OnHold = 'On Hold'
);
```

#### Step 2: Credentials Microflow

Before defining the consumed MCP service, create a microflow that returns the HTTP headers needed to authenticate to the server. The microflow must take no parameters and return `list<System.HttpHeader>`.

```sql
/**
 * Returns HTTP headers used to authenticate to the research MCP server.
 * Referenced from the ConsumedMCPService document.
 */
create microflow Research."MCP_GetCredentials" ()
returns list of System.HttpHeader
begin
  declare $headers list of System.HttpHeader = empty;
  $AuthHeader = create System.HttpHeader (
    key = 'Authorization',
    value = 'Bearer ' + @Research.ResearchMCPToken
  );
  set $headers = $headers + $AuthHeader;
  return $headers;
end;
/
```

#### Step 3: Consumed MCP Service Document

The ConsumedMCPService is a model document (like the agent itself), not a runtime-created entity. In this proposal's Phase 4 future extensions, we would also add MDL for it:

```sql
-- Proposed (future extension): declare a consumed MCP service as a document
create consumed mcp service Research."ResearchTools" (
  Endpoint: 'https://mcp.example.com/research',
  ProtocolVersion: v2025_03_26,
  GetCredentialsMicroflow: Research.MCP_GetCredentials,
  ConnectionTimeOutInSeconds: 30
);
```

At runtime, `ASU_AgentEditor` syncs this document into a `MCPClient.ConsumedMCPService` entity and discovers its tools.

#### Step 4: Agent with MCP Service Attached

The agent references the consumed MCP service by qualified name. At runtime, "Call Agent" automatically adds all (enabled) tools from the attached MCP server to the request.

```sql
create agent Research."ResearchAssistant" (
  UsageType: Conversational,
  description: 'Research assistant with web search and document analysis via MCP',
  entity: Research.ResearchProject,
  variables: ("title": EntityAttribute, "Objective": EntityAttribute),
  SystemPrompt: 'You are a research assistant helping with project: {{Title}}.

Objective: {{Objective}}

use the available tools to:
1. search the web for relevant information
2. read and analyze documents
3. Summarize findings

Always cite your sources. Present findings in a structured format.'
)
{
  mcp service Research.ResearchTools {
    access: VisibleForUser
  }
};
```

#### Step 5: Action Microflow — Same Simple Pattern

Because tools and MCP services are declared on the agent document, the action microflow stays minimal. "Call Agent With History" takes care of wiring MCP tools into the request automatically.

```sql
/**
 * Action microflow for the research chat.
 * Because MCP services are attached to the agent document, no MCP-specific
 * code is needed here — "Call Agent" handles it.
 */
create microflow Research."Chat_Research" (
  $ChatContext: ConversationalUI.ChatContext
)
returns boolean
begin
  $request = call microflow ConversationalUI.ChatContext_Preprocessing(
    ChatContext = $ChatContext
  ) on error rollback;

  if $request = empty then
    return false;
  end if;

  retrieve $agent from database AgentCommons.Agent
    where _QualifiedName = 'Research.ResearchAssistant' limit 1;

  -- Retrieve the context object passed from the page (ResearchProject).
  -- Variables "Title" and "Objective" are replaced from this object's
  -- attributes by Call Agent automatically.
  retrieve $ProjectList from $ChatContext/ConversationalUI.ChatContext_Owner
    /System.User; -- simplified; real apps pass project via extension entity
  declare $project Research.ResearchProject;

  call agent with HISTORY $agent request $request context $project into $response
    on error rollback;

  if $response != empty and $response/GenAICommons.Response_Message != empty then
    call microflow ConversationalUI.ChatContext_UpdateAssistantResponse(
      ChatContext = $ChatContext,
      MessageStatus = ConversationalUI.ENUM_MessageStatus.Success,
      response = $response
    ) on error rollback;
    return true;
  else
    return false;
  end if;
end;
/
```

---

### Example 3: Single-Call Agent for Data Processing

Not all agents need a chat interface. A `task` (single-call) agent processes one request and returns a result — useful for batch operations, background processing, and microflow-embedded AI. The "Call Agent Without History" activity handles everything in one step.

#### Step 1: Domain Model

```sql
@position(100, 100)
create persistent entity Reviews."ProductReview" (
  "ProductName": string(200),
  "ReviewText": string(unlimited),
  "Sentiment": string(50),
  "KeyThemes": string(unlimited),
  "IsProcessed": boolean default false
);
```

#### Step 2: Agent Document

```sql
create agent Reviews."SentimentAnalyzer" (
  UsageType: task,
  description: 'Single-call agent that extracts sentiment and themes from a product review',
  entity: Reviews.ProductReview,
  variables: ("ProductName": EntityAttribute, "ReviewText": EntityAttribute),
  SystemPrompt: 'Analyze the following product review for {{ProductName}}.

Extract:
1. Overall sentiment (Positive, Negative, Neutral, Mixed)
2. key themes mentioned (comma-separated)

Respond in this exact format:
Sentiment: <sentiment>
Themes: <theme1>, <theme2>, <theme3>',
  UserPrompt: '{{ReviewText}}'
);
```

#### Step 3: Processing Microflow — One Activity

The context object (`$Review`) carries the attribute values that replace `{{ProductName}}` and `{{ReviewText}}` in the prompts. "Call Agent Without History" resolves everything and returns the `response` in a single activity.

```sql
/**
 * Process a single product review using the SentimentAnalyzer agent.
 * Called from a batch microflow, a button action, or a scheduled event.
 *
 * @param $Review The review to analyze — its attributes replace prompt variables
 */
create microflow Reviews."ProcessReview" (
  $Review: Reviews.ProductReview
)
begin
  retrieve $agent from database AgentCommons.Agent
    where _QualifiedName = 'Reviews.SentimentAnalyzer' limit 1;

  -- One activity: resolve version, deployed model, prompts, and call the LLM
  call agent without HISTORY $agent context $Review into $response
    on error rollback;

  if $response = empty or $response/GenAICommons.Response_Message = empty then
    log warning node 'Reviews' 'Sentiment analysis failed for review: '
      + toString($Review/System.id);
    return;
  end if;

  declare $ResponseText string = call microflow
    GenAICommons.Response_GetModelResponseString(response = $response);

  change $Review (
    Sentiment = $ResponseText,
    IsProcessed = true
  );
  commit $Review;
end;
/

/**
 * Batch process all unprocessed reviews.
 */
create microflow Reviews."ProcessAllReviews" (
  $ReviewList: list of Reviews.ProductReview
)
begin
  loop $Review in $ReviewList
  begin
    if $Review/IsProcessed = false then
      call microflow Reviews.ProcessReview(Review = $Review) on error continue;
    end if;
  end loop;
end;
/
```

---

### Example 4: Agent with User-Approved Tools

A finance agent where sensitive tool calls (approving expenses) require user confirmation. The user-confirmation flow is **configured declaratively on the agent's tools**, not built in the action microflow — ConversationalUI renders the approval dialog automatically.

#### Step 1: Tool Microflows

```sql
/**
 * Read-only tool: Look up pending expense reports.
 * Safe for auto-execution (VisibleForUser).
 */
create microflow Finance."Tool_LookupExpenses" (
  $Department: string
)
returns string
begin
  retrieve $ExpenseList from database Finance.ExpenseReport
    where Department = $Department and status = Finance.ExpenseStatus.Pending;

  declare $Result string = '';
  loop $Expense in $ExpenseList
  begin
    set $Result = $Result + 'Expense #' + $Expense/ReportNumber
      + ' by ' + $Expense/SubmittedBy
      + ' - ' + formatDecimal($Expense/Amount, 2) + ' EUR: '
      + $Expense/description + '\n';
  end loop;

  return if $Result = '' then 'No pending expenses' else $Result;
end;
/

/**
 * Write tool: Approve an expense report.
 * Requires user confirmation before execution.
 */
create microflow Finance."Tool_ApproveExpense" (
  $ReportNumber: string,
  $ApprovalNote: string
)
returns string
begin
  retrieve $Expense from database Finance.ExpenseReport
    where ReportNumber = $ReportNumber limit 1;

  if $Expense = empty then
    return 'Expense report not found: ' + $ReportNumber;
  end if;

  change $Expense (
    status = Finance.ExpenseStatus.Approved,
    ApprovalNote = $ApprovalNote,
    ApprovedDate = [%CurrentDateTime%]
  );
  commit $Expense;

  return 'Expense ' + $ReportNumber + ' approved.';
end;
/
```

#### Step 2: Agent with Mixed Access Levels

The `access` modifier per tool controls what ConversationalUI does at runtime:
- `VisibleForUser` — tool call shown in chat, executes automatically
- `UserConfirmationRequired` — chat shows an approval dialog, user clicks Approve/Decline
- `HiddenForUser` — executes silently (use for internal/lookup tools)

```sql
create agent Finance."ExpenseApprovalAgent" (
  UsageType: Conversational,
  description: 'Review and approve expense reports with user confirmation for writes',
  SystemPrompt: 'You are a financial assistant that helps managers review and approve expense reports.

You have access to tools that can:
- Look up expense report details (auto-executes)
- Approve expense reports (requires user confirmation)

IMPORTANT: Always show the expense details before recommending approval.
Never approve expenses that exceed typical department limits without explicit user instruction.'
)
{
  tool LookupExpenses {
    microflow: Finance.Tool_LookupExpenses,
    description: 'List pending expense reports for a department',
    access: VisibleForUser
  }

  tool ApproveExpense {
    microflow: Finance.Tool_ApproveExpense,
    description: 'Approve a specific expense report by report number',
    access: UserConfirmationRequired
  }
};
```

#### Step 3: Action Microflow — Unchanged

Because tool approval is declared on the agent, the action microflow is identical to Example 1 — `call agent with History` handles tool-call round-trips and cooperates with ConversationalUI's approval widget automatically.

```sql
create microflow Finance."Chat_ExpenseApproval" (
  $ChatContext: ConversationalUI.ChatContext
)
returns boolean
begin
  $request = call microflow ConversationalUI.ChatContext_Preprocessing(
    ChatContext = $ChatContext
  ) on error rollback;

  if $request = empty then return false; end if;

  retrieve $agent from database AgentCommons.Agent
    where _QualifiedName = 'Finance.ExpenseApprovalAgent' limit 1;

  call agent with HISTORY $agent request $request into $response
    on error rollback;

  if $response != empty and $response/GenAICommons.Response_Message != empty then
    call microflow ConversationalUI.ChatContext_UpdateAssistantResponse(
      ChatContext = $ChatContext,
      MessageStatus = ConversationalUI.ENUM_MessageStatus.Success,
      response = $response
    ) on error rollback;
    return true;
  else
    return false;
  end if;
end;
/
```

---

### Example 5: Knowledge Base RAG Agent

An agent that uses a knowledge base (vector store) for Retrieval-Augmented Generation. As with tools and MCP services, the knowledge base is attached to the agent **document** — "Call Agent" performs retrieval automatically before invoking the LLM, and source references flow through to the chat UI.

#### Step 1: Knowledge Base Document

A knowledge base is a separate model document that references a Mendix Cloud GenAI Knowledge Base resource via its key. The underlying `GenAICommons.ConsumedKnowledgeBase` is auto-created by `ASU_AgentEditor` at startup from the document.

```sql
-- Proposed (future extension): declare a knowledge base as a document
create knowledge base HelpDesk."ProductDocsKB" (
  DisplayName: 'Product Documentation',
  Architecture: 'MxCloud',
  KeyConstant: HelpDesk.ProductDocsKBKey     -- String constant with the KB resource key
);
```

#### Step 2: Agent with Knowledge Base Attached

```sql
create agent HelpDesk."ProductExpert" (
  UsageType: Conversational,
  description: 'Answers product questions from the documentation knowledge base',
  SystemPrompt: 'You are a product expert for our software platform.

Answer questions using ONLY the information from the provided knowledge base context.
if the knowledge base does not contain relevant information, say so clearly.
Always include the source document reference in your answer.

Do not make up information that is not in the context.'
)
{
  knowledge base HelpDesk.ProductDocsKB {
    collection: 'product-documentation',
    MaxResults: 5,
    MinSimilarity: 0.7
  }
};
```

#### Step 3: Action Microflow — Identical to the Simple Pattern

Because the knowledge base is attached to the agent, RAG retrieval happens inside "Call Agent With History". Source references are automatically added to the `response/message`, and `ChatContext_UpdateAssistantResponse` already handles rendering them (it calls `Source_Create` internally).

```sql
create microflow HelpDesk."Chat_ProductExpert" (
  $ChatContext: ConversationalUI.ChatContext
)
returns boolean
begin
  $request = call microflow ConversationalUI.ChatContext_Preprocessing(
    ChatContext = $ChatContext
  ) on error rollback;

  if $request = empty then return false; end if;

  retrieve $agent from database AgentCommons.Agent
    where _QualifiedName = 'HelpDesk.ProductExpert' limit 1;

  -- RAG retrieval happens inside "Call Agent With History" automatically
  call agent with HISTORY $agent request $request into $response
    on error rollback;

  if $response != empty and $response/GenAICommons.Response_Message != empty then
    call microflow ConversationalUI.ChatContext_UpdateAssistantResponse(
      ChatContext = $ChatContext,
      MessageStatus = ConversationalUI.ENUM_MessageStatus.Success,
      response = $response
    ) on error rollback;
    return true;
  else
    return false;
  end if;
end;
/
```

Notice that Examples 1, 2, 4, and 5 all have **the same shape** for the action microflow — only the agent reference changes. This is the power of declarative agent documents: the capabilities (tools, KB, MCP) are metadata the "Call Agent" activity consumes, not code the developer writes.

---

### Example 6: Building an MCP Server in Mendix

Mendix apps can also act as MCP servers, exposing their microflows as tools that external AI systems (Claude, ChatGPT, or another Mendix app) can call. This is done via the **MCPServer** marketplace module — not a custom Published REST Service. The module provides `create mcp Server` and `add tool` toolbox actions that handle the MCP protocol.

Each tool microflow must accept **primitives or an `MCPServer.Tool` object** as input and return either `string` or `TextContent`. The Agent Editor / MCP Server infers the JSON schema from the signature.

```sql
-- Tool microflow: The MCP server will expose this as a tool named "lookup_customer"
create microflow Support."MCP_LookupCustomer" (
  $Email: string
)
returns string
begin
  retrieve $Customer from database Support.Customer
    where Email = $Email limit 1;
  if $Customer = empty then
    return 'No customer found';
  end if;
  return 'Customer: ' + $Customer/Name + ', Tier: ' + getKey($Customer/AccountTier);
end;
/

create microflow Support."MCP_GetOrderStatus" (
  $OrderNumber: string
)
returns string
begin
  retrieve $Order from database Support.Order
    where OrderNumber = $OrderNumber limit 1;
  if $Order = empty then
    return 'Order not found: ' + $OrderNumber;
  end if;
  return 'Order ' + $Order/OrderNumber + ' status: ' + getKey($Order/status);
end;
/

/**
 * Set up the MCP server at startup and register the tools.
 * Register this as (part of) the after-startup microflow.
 */
create microflow Support."ASU_SetupMCPServer" ()
begin
  -- 1. Create the MCP server instance (Mendix runtime listens for MCP requests)
  $Server = call java action MCPServer.CreateMCPServer(
    Name = 'CustomerSupportMCP',
    version = '1.0',
    ProtocolVersion = MCPServer.ENUM_ProtocolVersion.v2025_03_26
  ) on error rollback;

  -- 2. Expose each tool microflow. The MCP Server module builds the JSON
  --    schema from each microflow's signature.
  call java action MCPServer.AddTool(
    Server = $Server,
    Name = 'lookup_customer',
    description = 'Look up customer information by email address',
    microflow = 'Support.MCP_LookupCustomer'
  ) on error rollback;

  call java action MCPServer.AddTool(
    Server = $Server,
    Name = 'get_order_status',
    description = 'Get the status of an order by order number',
    microflow = 'Support.MCP_GetOrderStatus'
  ) on error rollback;

  log info node 'MCP' 'MCP server started with 2 tools';
end;
/
```

External agents (a Claude Desktop client, another Mendix app using MCPClient, or any MCP-compliant system) can now connect to this server, discover the tools, and call them. The MCPServer module handles protocol framing, authentication, and dispatch to the appropriate microflow.

---

### Example 7: Full Smart App Script — IT Help Desk

This script creates a complete AI-powered IT help desk application in a single MDL file, showing all the correct pieces end-to-end.

```sql
-- =============================================================
-- IT Help Desk Smart App
-- A complete AI-powered help desk built with MDL
-- Prerequisites: Encryption module configured, model key in constant
-- =============================================================

-- 1. Module
create module ITHelp;

-- 2. Domain Model
@position(100, 100)
create persistent entity ITHelp."Ticket" (
  "Subject": string(200) not null error 'Subject is required',
  "description": string(unlimited),
  "Category": enumeration(ITHelp.Category),
  "status": enumeration(ITHelp.TicketStatus),
  "AssignedTo": string(200),
  "Resolution": string(unlimited)
);

@position(100, 300)
create persistent entity ITHelp."KBArticle" (
  "title": string(200),
  "content": string(unlimited),
  "Category": enumeration(ITHelp.Category),
  "ViewCount": integer default 0
);

create enumeration ITHelp."Category" (
  Network = 'Network',
  Hardware = 'Hardware',
  Software = 'Software',
  access = 'Access & Permissions',
  Other = 'Other'
);

create enumeration ITHelp."TicketStatus" (
  New = 'New',
  InProgress = 'In Progress',
  WaitingOnUser = 'Waiting on User',
  Resolved = 'Resolved',
  Closed = 'Closed'
);

-- 3. Model Key Constant (set via environment or Configuration_Overview page)
create constant ITHelp."ModelKey" (
  type: string,
  DefaultValue: ''
);

-- 4. Tool microflows — signatures become the tool JSON schemas
create microflow ITHelp."Tool_SearchKB" ($query: string)
returns string
begin
  retrieve $Articles from database ITHelp.KBArticle
    where contains(title, $query) or contains(content, $query);

  declare $Result string = '';
  loop $Article in $Articles
  begin
    set $Result = $Result + '## ' + $Article/title + '\n'
      + $Article/content + '\n\n';
  end loop;

  return if $Result = '' then 'No articles found for: ' + $query else $Result;
end;
/

create microflow ITHelp."Tool_CreateTicket" (
  $Subject: string,
  $description: string,
  $Category: enum ITHelp.Category
)
returns string
begin
  $Ticket = create ITHelp.Ticket (
    Subject = $Subject,
    description = $description,
    Category = $Category,
    status = ITHelp.TicketStatus.New
  );
  commit $Ticket;
  return 'Ticket ' + toString($Ticket/System.id) + ' created.';
end;
/

create microflow ITHelp."Tool_GetTicketStatus" ($TicketId: string)
returns string
begin
  retrieve $Ticket from database ITHelp.Ticket where System.id = $TicketId limit 1;
  if $Ticket = empty then return 'Ticket not found'; end if;
  return 'Status: ' + getKey($Ticket/status)
    + ', Assigned to: ' + $Ticket/AssignedTo;
end;
/

-- 5. Agent document — tools declared here, not at runtime
create agent ITHelp."ITSupportAgent" (
  UsageType: Conversational,
  description: 'AI-powered first-line IT support',
  SystemPrompt: 'You are an IT support agent for a corporate help desk.

Capabilities (use these tools):
1. search the knowledge base for solutions to common problems
2. create support tickets when issues need escalation
3. check the status of existing tickets

Always try the knowledge base first before creating a ticket.
Be patient and ask clarifying questions when the issue is unclear.
for password resets and access requests, always create a ticket.'
)
{
  tool SearchKB {
    microflow: ITHelp.Tool_SearchKB,
    description: 'Search the knowledge base for articles matching a query',
    access: VisibleForUser
  }

  tool CreateTicket {
    microflow: ITHelp.Tool_CreateTicket,
    description: 'Create a new support ticket with subject, description, and category',
    access: UserConfirmationRequired
  }

  tool GetTicketStatus {
    microflow: ITHelp.Tool_GetTicketStatus,
    description: 'Get current status of an existing support ticket by ID',
    access: VisibleForUser
  }
};

-- 6. Chat action microflow — uniform pattern with Call Agent
create microflow ITHelp."Chat_ITSupport" (
  $ChatContext: ConversationalUI.ChatContext
)
returns boolean
begin
  $request = call microflow ConversationalUI.ChatContext_Preprocessing(
    ChatContext = $ChatContext
  ) on error rollback;

  if $request = empty then return false; end if;

  retrieve $agent from database AgentCommons.Agent
    where _QualifiedName = 'ITHelp.ITSupportAgent' limit 1;

  call agent with HISTORY $agent request $request into $response
    on error rollback;

  if $response != empty and $response/GenAICommons.Response_Message != empty then
    call microflow ConversationalUI.ChatContext_UpdateAssistantResponse(
      ChatContext = $ChatContext,
      MessageStatus = ConversationalUI.ENUM_MessageStatus.Success,
      response = $response
    ) on error rollback;
    return true;
  else
    return false;
  end if;
end;
/

-- 7. Entry-point microflow uses "New Chat for Agent"
create microflow ITHelp."ACT_StartHelpChat" ()
begin
  retrieve $agent from database AgentCommons.Agent
    where _QualifiedName = 'ITHelp.ITSupportAgent' limit 1;

  NEW CHAT for agent $agent
    action microflow ITHelp.Chat_ITSupport
    into $ChatContext
    on error rollback;

  show page ITHelp.HelpDesk($ChatContext = $ChatContext);
end;
/

-- 8. Pages
create page ITHelp."home" (
  title: 'IT Help Desk',
  layout: Atlas_Core.Atlas_Default
) {
  header h1 { dynamictext t (caption: 'IT Help Desk') }
  container c {
    actionbutton startChat (
      caption: 'Start Chat with IT Support',
      action: microflow ITHelp.ACT_StartHelpChat()
    )
  }
};
/

create page ITHelp."HelpDesk" (
  title: 'IT Support Chat',
  layout: Atlas_Core.Atlas_Default
) {
  header h1 { dynamictext t (caption: 'IT Support') }
  dataview chatDv (datasource: context ConversationalUI.ChatContext) {
    snippetcall chat (snippet: ConversationalUI.Snippet_Output_WithHistory)
  }
};
/

create page ITHelp."TicketOverview" (
  title: 'Support Tickets',
  layout: Atlas_Core.Atlas_Default
) {
  header h1 { dynamictext t (caption: 'Support Tickets') }
  datagrid ticketGrid (datasource: database ITHelp.Ticket) {
    column col1 (attribute: Subject, caption: 'Subject')
    column col2 (attribute: Category, caption: 'Category')
    column col3 (attribute: status, caption: 'Status')
    column col4 (attribute: AssignedTo, caption: 'Assigned To')
  }
};
/

-- 9. Security
create module role ITHelp."user";
create module role ITHelp."Admin";

grant ITHelp.User on ITHelp.Ticket (create, read *, write (ITHelp.Ticket.Description));
grant ITHelp.User on ITHelp.KBArticle (read *);
grant ITHelp.Admin on ITHelp.Ticket (create, delete, read *, write *);
grant ITHelp.Admin on ITHelp.KBArticle (create, delete, read *, write *);

grant execute on microflow ITHelp.ACT_StartHelpChat to ITHelp.User;
grant execute on microflow ITHelp.Chat_ITSupport to ITHelp.User;
grant execute on microflow ITHelp.Tool_SearchKB to ITHelp.User;
grant execute on microflow ITHelp.Tool_CreateTicket to ITHelp.User;
grant execute on microflow ITHelp.Tool_GetTicketStatus to ITHelp.User;
grant view on page ITHelp.Home to ITHelp.User;
grant view on page ITHelp.HelpDesk to ITHelp.User;
grant view on page ITHelp.TicketOverview to ITHelp.User, ITHelp.Admin;

-- 10. After-startup microflow registration
alter settings model AfterStartupMicroflow = AgentEditorCommons.ASU_AgentEditor;
-- (Add custom setup to a composite microflow if needed.)

-- 11. Navigation
create or replace navigation Responsive_web
  home page ITHelp.Home for ITHelp.User
  menu (
    item 'Help Desk' page ITHelp.Home,
    item 'Tickets' page ITHelp.TicketOverview,
    item 'Agent Admin' page AgentCommons.Agent_Overview
  );
```

---

### Summary: What MDL Agent Support Enables

| Capability | Without MDL Agent Support | With MDL Agent Support |
|------------|---------------------------|------------------------|
| **Discover agents** | Open Studio Pro, navigate to Agent Editor | `show agents` in CLI or script |
| **Inspect agent prompts** | Click through Agent Editor UI | `describe agent Module.Name` |
| **Create agents** | Only via Studio Pro Agent Editor | `create agent` in MDL scripts |
| **Version control** | Binary CustomBlobDocument diffs | Human-readable MDL diffs |
| **AI-assisted development** | AI cannot see or create agents | AI generates complete smart apps |
| **Batch operations** | Manual, one agent at a time | Script creates multiple agents |
| **Code review** | Cannot review agent changes in PR | MDL changes are reviewable text |
| **Migration** | Manual recreation in new project | Copy/paste MDL scripts |
| **Documentation** | Screenshots of Agent Editor | `describe agent` produces docs |
| **Testing** | Manual testing in Studio Pro | Scriptable test cases with mxcli |

The combination of `create agent` (document definition), tool microflows (business logic), MCP connections (external tools), knowledge bases (RAG), and ConversationalUI (chat interface) means an AI coding agent can scaffold an entire smart app from a natural-language description — creating all layers from domain model to navigation in a single MDL session.

## Open Questions

1. **CustomBlobDocument extensibility** *(answered)*: Mendix uses `CustomBlobDocument` as a general extension pattern. Four `CustomDocumentType` values observed so far: `agenteditor.agent`, `agenteditor.model`, `agenteditor.knowledgebase`, `agenteditor.consumedMCPService`. The parser dispatches by `CustomDocumentType` rather than hardcoding agent-specific logic. Future types (other extensions, other agent-editor documents) plug in naturally.

2. **Contents JSON schema for tools/KB/MCP** *(answered for MCP tools and KB tools, still open for microflow tools)*: The `Agents.Agent007` document in the test3 project gave us the schema for MCP-type tools and knowledge base tools (see BSON Structure section). **Still open**: the JSON shape for a microflow-type tool — none observed yet. Expected: `toolType: "microflow"` plus a `microflow: { qualifiedName, microflowId }` reference, but the exact key names and nesting need a real sample. Implementation should capture one before finalizing the microflow-tool writer.

3. **Separate document types for Model, Knowledge Base, and MCP Service** *(answered)*: Confirmed. Phase 4 of the implementation plan covers `create model`, `create knowledge base`, `create consumed mcp service` with schemas matching the observed BSON.

4. **`call agent` activity BSON format**: The proposed `call agent with HISTORY` / `call agent without HISTORY` / `NEW CHAT for agent` MDL statements need to map to a Studio Pro microflow activity. Is this a dedicated activity type in BSON, or does Studio Pro render a generic `JavaActionCallAction` (pointing at `AgentCommons.Agent_Call_WithHistory`) as "Call Agent"? If it's the latter, MDL can emit a standard Java action call; if the former, we need to identify the new activity BSON `$type`. **Action:** inspect a Studio Pro microflow that uses "Call Agent" to resolve.

5. **ASU_AgentEditor behavior**: `AgentEditorCommons.ASU_AgentEditor` is the after-startup microflow that syncs agent documents to runtime `AgentCommons.Agent` entities. Does `create agent` via MDL need to trigger this sync, or does it happen automatically on next app startup? What happens if MDL creates an agent and the app is already running?

6. **Module placement** *(partially answered)*: The new documents in test3 live in a user-created `agents` module (not in `AgentEditorCommons`), confirming that users **can** place agent-editor documents in their own modules. The older 4 agents in `AgentEditorCommons` appear to be samples shipped with the marketplace module.

7. **Cross-document UUID stability**: Agent documents reference model/KB/MCP documents by both `qualifiedName` AND `documentId` (UUID). When MDL creates a document, the generated UUID must be stable so that subsequent `create agent` statements can correctly fill the `documentId` field. If an `alter` or re-create changes the UUID, all referring agents break. The writer must either (a) preserve existing UUIDs on update or (b) allow the agent's `documentId` field to be left empty and resolved at app-startup time by qualified name.

8. **Portal-populated fields on Model/KB**: Fields like `displayName`, `keyId`, `keyName`, `environment`, `resourceName`, `modelName`, `modelDisplayName` are populated by Studio Pro after the user clicks "Test Key". Should MDL `create model` write them as empty strings (letting Studio Pro fill them on next open), preserve them if provided by `describe` round-trip, or outright reject user-supplied values? Current proposal: accept-and-round-trip but document them as read-only.

7. **Non-Mendix-Cloud model providers (e.g., OpenRouter)**: The [Agent Editor docs](https://docs.mendix.com/appstore/modules/genai/genai-for-mx/agent-editor/) state that model documents require "a String constant that contains the key for a **Text Generation resource**... obtained in the **Mendix Cloud GenAI Portal**" — so the model document format is currently locked to Mendix Cloud GenAI. Meanwhile, `GenAICommons.DeployedModel` is provider-agnostic (it's just `DisplayName` + `Architecture` + a `microflow` pointer), and marketplace connectors exist for OpenAI, Amazon Bedrock, Google Gemini, and Mistral. This creates a split:
   - Users who want OpenAI-compatible endpoints like **OpenRouter** (including its free models: `google/gemini-flash-1.5-8b:free`, `mistralai/mistral-7b-instruct:free`, etc.) cannot use the Agent Editor's model documents today. Workarounds: (a) reconfigure the OpenAI Connector's base URL to OpenRouter; (b) build a custom microflow-based `DeployedModel`; (c) skip the Agent Editor and create `AgentCommons.Agent` / `version` entities at runtime instead.
   - Option (c) means losing the design-time benefits of agent documents (MDL support, version control in the project, LLM-friendly static configuration). `create agent` in MDL therefore won't help these users until Mendix opens the model document format to other providers.
   - **Implications for this proposal**: The proposed `create model` document (Phase 4) should not hard-code `Architecture: 'MxCloud'`. If/when Mendix supports third-party architectures in model documents, the `create model` syntax must accept `Architecture: 'OpenAI' | 'OpenRouter' | 'Bedrock' | ...` and a connector-specific configuration block. The `create agent` body is already model-provider-agnostic (it references a model document by name, not by architecture), so no changes needed there.
   - **Track this externally**: Monitor Mendix release notes for the Agent Editor opening to additional providers. If that happens, the MDL grammar already has room for it — we'd just add more valid `Architecture` values to `create model`.

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Microflow-tool JSON shape in `Contents.tools[]` still unobserved | Medium | Medium | Observed MCP-tool and KB-tool shapes; capture a microflow-tool sample before finalizing that writer path |
| `call agent` activity is a new BSON `$type` | Medium | Medium | Inspect a Studio Pro microflow that uses "Call Agent" before implementing |
| Cross-document UUIDs become stale when documents are re-created | High | High | Preserve UUIDs on update; validate referring agents on `create`/`drop` of a referenced document |
| Contents JSON schema changes in future Mendix versions | Medium | Medium | Parse tolerantly (ignore unknown fields), version-gate new fields |
| CustomBlobDocument format changes | Low | High | Monitor Mendix release notes, BSON schema comparison |
| Studio Pro fails to open MDL-created documents | Medium | High | Test with `mx check` and Studio Pro after creation; compare BSON byte-for-byte with editor-created documents (`Agents.MyFirstModel` etc.) |
| Portal-populated fields overwritten by MDL round-trip | Medium | Medium | On `create model` / `create knowledge base`, preserve any existing Portal fields if the document already exists; write empty strings only on fresh creates |
| Prerequisites (Encryption, ASU_AgentEditor) not set up before CREATE AGENT | Medium | Medium | MDL `create agent` should warn/pre-check that prerequisites are configured |
| Agent document + matching Model/KB/MCP documents out of sync | Medium | Medium | `mxcli check` should validate cross-document references when `--references` is passed |
| Users want third-party LLM providers (OpenRouter, custom OpenAI-compatible) but Agent Editor model documents are Mendix-Cloud-only | High | Low (out of scope) | Document the workarounds (reconfigure OpenAI connector, custom microflow DeployedModel, skip agent documents); keep `create model` syntax open to future `Provider` values |

## References

- Test project: `mx-test-projects/test3-app/test3.mpr` (Mendix 11.9.0)
- **Older agent documents (AgentEditorCommons module)**: `InformationExtractorAgent`, `ProductDescription`, `SummarizationAgent`, `TranslationAgent` — `Excluded: true`, no model reference, no tools/KB
- **New Agent Editor sample documents (Agents module)**:
  - `Agents.Agent007` — fully populated agent with model, MCP tool, KB tool (`mprcontents/e0/72/e072318a-...mxunit`)
  - `Agents.MyFirstModel` — Model document, provider `MxCloudGenAI` (`mprcontents/3a/dd/3addaaa1-...mxunit`)
  - `Agents.Knowledge_base` — Knowledge Base document (`mprcontents/cc/cc/cccc0b5b-...mxunit`)
  - `Agents.Consumed_MCP_service` — Consumed MCP Service document (`mprcontents/47/c9/47c9987a-...mxunit`)
- Agent Editor extension manifest: `.mendix-cache/modules/agenteditor.mxmodule/extensions/agent-editor/manifest.json`
- AgentCommons module: Marketplace v3.1.0 (31 entities, 226 microflows)
- AgentEditorCommons module: Marketplace v1.0.0 (9 entities, 32 microflows, including `ASU_AgentEditor`)
- MCPClient module: Marketplace v3.0.1 (20 entities, 35 microflows)
- GenAICommons module: Marketplace v6.1.0 (34 entities, 112 microflows)
- ConversationalUI module: Marketplace v6.1.0 (17 entities, 152 microflows)
- [Mendix Agent Editor documentation](https://docs.mendix.com/appstore/modules/genai/genai-for-mx/agent-editor/)
