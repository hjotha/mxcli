# SHOW/DESCRIBE Proposals for Missing Document Types

These proposals add read-only SHOW and DESCRIBE commands for all Mendix document types found in real-world projects that currently lack MDL support.

## Test Project Analysis

Three real-world Mendix projects were analyzed to determine prevalence:

| Project | Modules | Microflows | Pages | Total Documents |
|---------|---------|------------|-------|-----------------|
| EnquiriesManagement (Mx 11.6) | 28 | 925 | 172 | 2,666 |
| Evora-FactoryManagement | 39 | 1,594 | 184 | 4,059 |
| LatoProductInventory | 30 | 917 | 84 | 2,713 |

## Proposals by Priority

### Tier 1 — High Priority (heavily used, present in all projects)

| Proposal | Document Type | Total Count | Complexity | Reader Exists |
|----------|--------------|-------------|------------|---------------|
| [JavaScript Actions](show-describe-javascript-actions.md) | `JavaScriptActions$JavaScriptAction` | 283 | Medium | Yes |
| [Building Blocks](show-describe-building-blocks.md) | `Forms$BuildingBlock` | 233 | Medium | Yes |
| [Page Templates](show-describe-page-templates.md) | `Forms$PageTemplate` | 215 | Medium | Yes |
| [Nanoflow DESCRIBE](show-describe-nanoflows.md) | `Microflows$Nanoflow` | 227 | **Very Low** | Yes (full) |
| [JSON Structures](show-describe-json-structures.md) | `JsonStructures$JsonStructure` | 96 | Medium | No |
| [Import Mappings](show-describe-import-mappings.md) | `ImportMappings$ImportMapping` | 83 | High | No |
| [Export Mappings](show-describe-export-mappings.md) | `ExportMappings$ExportMapping` | 67 | High | No |
| [Published REST Services](show-describe-published-rest-services.md) | `Rest$PublishedRestService` | 16 | Medium | Yes |

### Tier 2 — Medium Priority (common, useful for project understanding)

| Proposal | Document Type | Total Count | Complexity | Reader Exists |
|----------|--------------|-------------|------------|---------------|
| [Module Settings](show-describe-module-settings.md) | `Projects$ModuleSettings` | 97 | Low | No |
| [Image Collections](show-describe-image-collections.md) | `Images$ImageCollection` | 65 | Low | Yes |
| [Rules](show-describe-rules.md) | `Microflows$Rule` | 49 | Low (reuse MF) | No |
| [Message Definitions](show-describe-message-definitions.md) | `MessageDefinitions$MessageDefinitionCollection` | 28 | Medium | No |
| [Scheduled Events](show-describe-scheduled-events.md) | `ScheduledEvents$ScheduledEvent` | 19 | Low | Yes |
| [Consumed REST Services](show-describe-consumed-rest-services.md) | `Rest$ConsumedRestService` | 2 | Medium | No |

### Tier 3 — Low Priority (small counts, simple types)

| Proposal | Document Type | Total Count | Complexity | Reader Exists |
|----------|--------------|-------------|------------|---------------|
| [Regular Expressions](show-describe-regular-expressions.md) | `RegularExpressions$RegularExpression` | 13 | **Very Low** | No |
| [Custom Icon Collections](show-describe-custom-icon-collections.md) | `CustomIcons$CustomIconCollection` | 8 | Low | No |
| [Menu Documents](show-describe-menu-documents.md) | `Menus$MenuDocument` | 6 | Medium | No |
| [Queues](show-describe-queues.md) | `Queues$Queue` | 5 | **Very Low** | No |

## Implementation Order Recommendation

**Phase 1 — Quick wins (reader already exists, minimal new code):**
1. Nanoflow DESCRIBE — ~30 lines, pure wiring
2. Scheduled Events — reader + parser exist, just add executor
3. Published REST Services — reader exists, enhance parser + add executor
4. JavaScript Actions — reader exists, enhance parser + add executor
5. Image Collections — reader exists, enhance parser + add executor
6. Building Blocks — reader exists, enhance parser + add executor
7. Page Templates — reader exists, enhance parser + add executor

**Phase 2 — Simple new types (flat structure, easy parsing):**
8. Regular Expressions — 5 flat fields
9. Queues — 6 fields with one nested object
10. Module Settings — simple metadata
11. Custom Icon Collections — flat list of icons

**Phase 3 — Complex new types (recursive structures, polymorphic types):**
12. JSON Structures — recursive element tree
13. Rules — reuse microflow parsing
14. Message Definitions — recursive entity exposure tree
15. Import Mappings — recursive object/value mapping tree
16. Export Mappings — share implementation with Import Mappings
17. Menu Documents — recursive menu items with polymorphic actions
18. Consumed REST Services — polymorphic operations

## Shared Implementation Patterns

All proposals follow the established pattern from existing SHOW/DESCRIBE commands:
- AST constants in `mdl/ast/ast_query.go`
- Grammar rules in `MDLLexer.g4` / `MDLParser.g4`
- Visitor mapping in `mdl/visitor/`
- Executor handlers in `mdl/executor/cmd_*.go`
- Autocomplete in `mdl/executor/autocomplete.go`
- Dispatcher wiring in `mdl/executor/executor.go`

See the pattern analysis in any proposal for the full template.

## Grammar Token Summary

New tokens needed across all proposals:

```antlr
// Already exist or likely exist:
IMPORT, EXPORT, REST, SERVICE, SERVICES, MODULE, IMAGE, ICON

// New tokens needed:
JAVASCRIPT    // for JAVASCRIPT ACTION
BUILDING      // for BUILDING BLOCK
BLOCK / BLOCKS
TEMPLATE / TEMPLATES
STRUCTURE / STRUCTURES
MAPPING / MAPPINGS
SCHEDULED     // for SCHEDULED EVENT
EVENT / EVENTS
REGULAR       // for REGULAR EXPRESSION
EXPRESSION / EXPRESSIONS
QUEUE / QUEUES
MENU / MENUS
COLLECTION / COLLECTIONS
CLIENT        // for REST CLIENT (consumed)
```
