# Proposal: SHOW/DESCRIBE Import Mappings

## Overview

**Document type:** `ImportMappings$ImportMapping`
**Prevalence:** 83 across test projects (15 Enquiries, 35 Evora, 33 Lato)
**Priority:** High — heavily used for REST/JSON integrations

Import Mappings define how incoming JSON/XML data is mapped to Mendix entities. They reference a schema source (JSON Structure, XML Schema, or Message Definition) and specify how each field maps to entity attributes.

## What Already Exists

| Layer | Status | Location |
|-------|--------|----------|
| **Go type** | No | Only in generated metamodel |
| **Parser** | No | — |
| **Reader** | No | — |
| **Generated metamodel** | Yes | `generated/metamodel/types.go` line 2957 |

## BSON Structure (from test projects)

```
ImportMappings$ImportMapping:
  Name: string
  Documentation: string
  Excluded: bool
  ExportLevel: string
  JsonStructure: string (qualified name reference)
  XmlSchema: string (qualified name reference)
  MessageDefinition: string (qualified name reference)
  WsdlFile: string (web service reference)
  ServiceName: string
  OperationName: string
  PublicName: string
  XsdRootElementName: string
  ParameterType: DataTypes$* (polymorphic)
  UseSubtransactionsForMicroflows: bool
  MappingSourceReference: nullable
  Elements: []*ImportMappings$ObjectMappingElement
    - Entity: string (qualified entity name)
    - ExposedName: string
    - JsonPath: string
    - XmlPath: string
    - ObjectHandling: string ("Create", "Find", "Custom", "CallAMicroflow")
    - Association: string
    - Children: [] (recursive, mix of Object and Value elements)
      - ImportMappings$ValueMappingElement:
        - Attribute: string (qualified name)
        - ExposedName: string
        - JsonPath: string
        - Type: DataTypes$* (polymorphic)
        - IsKey: bool
```

## Proposed MDL Syntax

### SHOW IMPORT MAPPINGS

```
SHOW IMPORT MAPPINGS [IN Module]
```

| Qualified Name | Module | Name | Schema Source | Root Entity | Elements |
|----------------|--------|------|--------------|-------------|----------|

Where "Schema Source" shows the referenced JSON Structure, XML Schema, or Message Definition.

### DESCRIBE IMPORT MAPPING

```
DESCRIBE IMPORT MAPPING Module.Name
```

Output format:

```
/**
 * Maps customer API response to Customer entity
 */
IMPORT MAPPING MyModule.ImportCustomer
  FROM JSON STRUCTURE MyModule.CustomerResponse
{
  root -> MyModule.Customer (Create)
    id -> Id (Integer, KEY)
    name -> Name (String)
    email -> Email (String)
    addresses -> MyModule.Address (Create) VIA MyModule.Customer_Address
      street -> Street (String)
      city -> City (String)
};
/
```

For XML/WSDL-based mappings:

```
IMPORT MAPPING MyModule.ImportOrder
  FROM XML SCHEMA MyModule.OrderSchema
  ROOT ELEMENT 'Order'
{
  Order -> MyModule.Order (Find)
    OrderId -> OrderId (Integer, KEY)
    LineItems -> MyModule.LineItem (Create) VIA MyModule.Order_LineItem
      Product -> ProductName (String)
      Quantity -> Quantity (Integer)
};
/
```

## Implementation Steps

### 1. Add Model Types (model/types.go)

```go
type ImportMapping struct {
    ContainerID   model.ID
    Name          string
    Documentation string
    Excluded      bool
    ExportLevel   string
    // Schema source (one of these is set)
    JsonStructure     string // qualified name
    XmlSchema         string
    MessageDefinition string
    WsdlFile          string
    // Mapping tree
    Elements []ImportMappingElement
}

type ImportMappingElement struct {
    Kind          string // "Object" or "Value"
    Entity        string // qualified name (for objects)
    Attribute     string // qualified name (for values)
    ExposedName   string
    JsonPath      string
    ObjectHandling string // "Create", "Find", "Custom"
    Association   string
    IsKey         bool
    DataType      string
    Children      []ImportMappingElement
}
```

### 2. Add Parser (sdk/mpr/parser_import_mapping.go)

Parse `ImportMappings$ImportMapping` BSON. Recursively parse element tree (mix of `ObjectMappingElement` and `ValueMappingElement`).

### 3. Add Reader

```go
func (r *Reader) ListImportMappings() ([]*model.ImportMapping, error)
```

### 4. Add AST, Grammar, Visitor, Executor

Standard pattern. Grammar tokens: `IMPORT` (already exists), `MAPPING`, `MAPPINGS`.

### 5. Add Autocomplete

```go
func (e *Executor) GetImportMappingNames(moduleFilter string) []string
```

## Dependencies

- Depends on JSON Structures proposal (for resolving schema references in DESCRIBE output)
- Can be implemented independently (just show qualified name references)

## Testing

- Create `mdl-examples/doctype-tests/18-mapping-examples.mdl`
- Verify against all 3 test projects
