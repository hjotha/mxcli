# Agents

## Overview

Four AI agent document types stored as JSON inside Mendix MPR files:

| Type | CREATE keyword | Notes |
|------|---------------|-------|
| Model | `CREATE MODEL` | GenAI model configuration; required by Agent |
| Knowledge Base | `CREATE KNOWLEDGE BASE` | KB source; referenced by Agent body |
| Consumed MCP Service | `CREATE CONSUMED MCP SERVICE` | MCP tool server; referenced by Agent body |
| Agent | `CREATE AGENT` | Orchestrates model + tools + prompts |

**Requires:** `AgentEditorCommons` marketplace module, Mendix 11.9+.

## Syntax

### Model

```sql
CREATE MODEL Module.MyModel (
  Provider: MxCloudGenAI,   -- default, can omit
  Key: Module.ApiKeyConst   -- must be a String constant
);
```

### Knowledge Base

```sql
CREATE KNOWLEDGE BASE Module.ProductDocs (
  Provider: MxCloudGenAI,
  Key: Module.KBKeyConst
);
```

### Consumed MCP Service

```sql
CREATE CONSUMED MCP SERVICE Module.WebSearch (
  ProtocolVersion: v2025_03_26,
  Version: '1.0',
  ConnectionTimeoutSeconds: 30,
  Documentation: 'Web search MCP server'
);
```

### Agent (full syntax)

```sql
CREATE AGENT Module.MyAgent (
  UsageType: Task,              -- Task | Conversational
  Model: Module.MyModel,        -- must exist
  Description: 'Agent description',
  MaxTokens: 4096,
  Temperature: 0.7,             -- float
  TopP: 0.9,                    -- float
  ToolChoice: Auto,
  Variables: ("Language": EntityAttribute, "Name": String),
  SystemPrompt: $$Multi-line
prompt here.$$,
  UserPrompt: 'Single line prompt.'
)
{
  MCP SERVICE Module.WebSearch {
    Enabled: true
  }

  KNOWLEDGE BASE KBAlias {
    Source: Module.ProductDocs,
    Collection: 'product-docs',
    MaxResults: 5,
    Description: 'Product docs',
    Enabled: true
  }

  TOOL MyMicroflowTool {
    Description: 'Fetch customer data',
    Enabled: true
  }
};
```

## Gotchas

### Dollar-quoting for multi-line prompts
Single-quoted strings cannot span lines. Use `$$...$$` for any SystemPrompt or UserPrompt that contains newlines. DESCRIBE always emits `$$...$$` when the value contains newlines, so DESCRIBE output re-parses cleanly.

### Portal-populated metadata fields
`DisplayName`, `KeyName`, `KeyID`, `Environment`, `ResourceName`, `DeepLinkURL` are populated by the Mendix portal at sync time. Do not set them manually in CREATE statements — they will be overwritten.

### documentId vs qualifiedName
Each document has both a `qualifiedName` (e.g. `Module.MyModel`) and an opaque `documentId` UUID. The UUID is assigned by ASU_AgentEditor at runtime. Only `qualifiedName` is set by CREATE; cross-reference lookups resolve by scanning all documents for a matching name.

### Drop order
Agents reference Model, Knowledge Base, and MCP Service documents. Always drop Agents before dropping their dependencies:
```sql
DROP AGENT Module.MyAgent;
DROP CONSUMED MCP SERVICE Module.WebSearch;
DROP KNOWLEDGE BASE Module.ProductDocs;
DROP MODEL Module.MyModel;
```

### Variables: syntax
- `"Key": EntityAttribute` — binds an attribute from the entity in the agent's context
- `"Key": String` — binds a plain string value
- Keys must be quoted (string literals or quoted identifiers)

### Association between BSON and MDL names
The feature uses `CustomBlobDocument` BSON type with a `Contents` field holding the JSON payload. The `$Type` field is always `"AgentEditorCommons$CustomBlobDocument"`. The document type is identified by the `readableTypeName` inside `Metadata`.

## Common Patterns

### Minimal agent (no tools)
```sql
CREATE MODEL Module.M (Provider: MxCloudGenAI, Key: Module.K);
CREATE AGENT Module.A (
  UsageType: Task,
  Model: Module.M,
  SystemPrompt: 'You are a helpful assistant.',
  UserPrompt: 'Ask me anything.'
);
```

### Check all agent documents in a module
```sql
LIST MODELS IN Module;
LIST KNOWLEDGE BASES IN Module;
LIST CONSUMED MCP SERVICES IN Module;
LIST AGENTS IN Module;
```
