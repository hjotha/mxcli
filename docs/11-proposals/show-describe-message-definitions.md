# Proposal: SHOW/DESCRIBE Message Definitions

## Overview

**Document type:** `MessageDefinitions$MessageDefinitionCollection`
**Prevalence:** 28 across test projects (7 Enquiries, 11 Evora, 10 Lato)
**Priority:** Medium — used for service contracts, business events, and mappings

Message Definition Collections define entity-based message schemas for integrations. Each collection contains one or more Message Definitions, each exposing an entity's attributes and associations as a structured message contract. They are referenced by Published/Consumed REST services, OData services, and Business Events.

## What Already Exists

| Layer | Status | Location |
|-------|--------|----------|
| **Go type** | No | Only in generated metamodel |
| **Parser** | No | — |
| **Reader** | No | — |
| **Generated metamodel** | Yes | `generated/metamodel/types.go` line 3457 |

## BSON Structure (from test projects)

```
MessageDefinitions$MessageDefinitionCollection:
  Name: string
  Documentation: string
  Excluded: bool
  ExportLevel: string
  MessageDefinitions: []*MessageDefinitions$EntityMessageDefinition
    - Name: string
    - Documentation: string
    - ExposedEntity: MessageDefinitions$ExposedEntity
      - Entity: string (qualified entity name)
      - ExposedName: string
      - Children: [] (recursive)
        - MessageDefinitions$ExposedAttribute:
          - Attribute: string (qualified name)
          - ExposedName: string
          - PrimitiveType: string
        - MessageDefinitions$ExposedAssociation:
          - Association: string (qualified name)
          - Entity: string (qualified entity name)
          - ExposedName: string
          - Children: [] (recursive — nested entity exposure)
```

## Proposed MDL Syntax

### SHOW MESSAGE DEFINITIONS

```
SHOW MESSAGE DEFINITIONS [IN Module]
```

| Qualified Name | Module | Name | Messages | Entities |
|----------------|--------|------|----------|----------|

Where "Messages" is the count of message definitions in the collection, and "Entities" lists the exposed entity names.

### DESCRIBE MESSAGE DEFINITION

```
DESCRIBE MESSAGE DEFINITION Module.Name
```

Output format:

```
MESSAGE DEFINITION COLLECTION MyModule.CustomerMessages
{
  MESSAGE CustomerMessage
    ENTITY MyModule.Customer AS 'Customer'
    {
      Id AS 'id': Integer
      Name AS 'name': String
      Email AS 'email': String
      ASSOCIATION MyModule.Customer_Address AS 'addresses'
        ENTITY MyModule.Address AS 'Address'
        {
          Street AS 'street': String
          City AS 'city': String
        }
    };

  MESSAGE OrderMessage
    ENTITY MyModule.Order AS 'Order'
    {
      OrderNumber AS 'orderNumber': Integer
      TotalAmount AS 'totalAmount': Decimal
    };
};
/
```

## Implementation Steps

### 1. Add Model Types (model/types.go)

```go
type MessageDefinitionCollection struct {
    ContainerID model.ID
    Name        string
    Documentation string
    Excluded    bool
    ExportLevel string
    Definitions []*MessageDefinition
}

type MessageDefinition struct {
    Name          string
    Documentation string
    Entity        string // qualified entity name
    ExposedName   string
    Attributes    []*ExposedAttribute
    Associations  []*ExposedAssociation
}

type ExposedAttribute struct {
    Attribute   string // qualified name
    ExposedName string
    Type        string
}

type ExposedAssociation struct {
    Association string
    Entity      string
    ExposedName string
    Attributes  []*ExposedAttribute
    Associations []*ExposedAssociation // recursive
}
```

### 2. Add Parser (sdk/mpr/parser_message_definitions.go)

Parse `MessageDefinitions$MessageDefinitionCollection` BSON. Recursively parse the exposed entity tree with its attributes and associations.

### 3. Add Reader

```go
func (r *Reader) ListMessageDefinitions() ([]*model.MessageDefinitionCollection, error)
```

### 4. Add AST, Grammar, Visitor, Executor

Grammar tokens: `MESSAGE` (may already exist for business events), `DEFINITION`, `DEFINITIONS`.

### 5. Add Autocomplete

```go
func (e *Executor) GetMessageDefinitionNames(moduleFilter string) []string
```

## Testing

- Verify against all 3 test projects
