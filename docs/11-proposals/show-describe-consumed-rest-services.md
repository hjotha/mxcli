# Proposal: SHOW/DESCRIBE Consumed REST Services

## Overview

**Document type:** `Rest$ConsumedRestService`
**Prevalence:** 2 in Evora project (not found in Enquiries or Lato using this type name)
**Priority:** Medium — newer Mendix feature, growing in adoption

Consumed REST Services define external REST API connections. Each service has a base URL, authentication scheme, and one or more operations with HTTP methods, paths, headers, and response handling.

Note: This is different from "Consumed OData Services" (which already have SHOW/DESCRIBE support). Consumed REST Services are a Mendix 10+ feature for calling arbitrary REST APIs.

## What Already Exists

| Layer | Status | Location |
|-------|--------|----------|
| **Go type** | No | — |
| **Parser** | No | — |
| **Reader** | No | — |
| **Generated metamodel** | Yes | Full type definitions exist |

## BSON Structure (from test projects)

```
Rest$ConsumedRestService:
  Name: string
  Documentation: string
  Excluded: bool
  ExportLevel: string
  BaseUrl: Rest$ValueTemplate
    Value: string (e.g., "https://api.example.com/v1")
  BaseUrlParameter: nullable
  AuthenticationScheme: nullable (polymorphic)
  OpenApiFile: nullable
  Operations: []*Rest$RestOperation
    - Name: string
    - Method: polymorphic (Rest$RestOperationMethodWithBody / Rest$RestOperationMethodWithoutBody)
      - HttpMethod: string ("GET", "POST", "PUT", "DELETE", "PATCH")
      - Body: nullable (import mapping reference)
    - Path: Rest$ValueTemplate
    - Headers: []*Rest$HeaderWithValueTemplate
      - Name: string
      - Value: Rest$ValueTemplate
    - QueryParameters: []*Rest$QueryParameter
    - Parameters: []*Rest$RestOperationParameter
    - ResponseHandling: object (export mapping, result handling)
    - Tags: []string
    - Timeout: nullable
```

## Proposed MDL Syntax

### SHOW REST CLIENTS

```
SHOW REST CLIENTS [IN Module]
```

| Qualified Name | Module | Name | Base URL | Operations |
|----------------|--------|------|----------|------------|

### DESCRIBE REST CLIENT

```
DESCRIBE REST CLIENT Module.Name
```

Output format:

```
/**
 * External customer management API
 */
REST CLIENT MyModule.CustomerAPI
  BASE URL 'https://api.example.com/v1'
  AUTHENTICATION Basic
{
  OPERATION GetCustomers
    GET /customers
    HEADERS
      Accept: 'application/json'
      X-API-Key: '{$ApiKey}'
    RESPONSE EXPORT MAPPING MyModule.ExportCustomerList;

  OPERATION CreateCustomer
    POST /customers
    HEADERS
      Content-Type: 'application/json'
    BODY IMPORT MAPPING MyModule.ImportCustomer
    RESPONSE EXPORT MAPPING MyModule.ExportCustomer;
};
/
```

## Implementation Steps

### 1. Add Model Types (model/types.go)

```go
type ConsumedRestService struct {
    ContainerID    model.ID
    Name           string
    Documentation  string
    Excluded       bool
    ExportLevel    string
    BaseUrl        string
    Authentication string
    Operations     []*RestClientOperation
}

type RestClientOperation struct {
    Name       string
    HttpMethod string
    Path       string
    Headers    map[string]string
    Body       string // import mapping reference
    Response   string // export mapping reference
}
```

### 2. Add Parser (sdk/mpr/parser_rest.go)

Extend existing REST parser file with `parseConsumedRestService()`.

### 3. Add Reader

```go
func (r *Reader) ListConsumedRestServices() ([]*model.ConsumedRestService, error)
```

### 4. Add AST, Grammar, Visitor, Executor

Grammar: `SHOW REST CLIENTS` / `DESCRIBE REST CLIENT`.

Note: "REST CLIENT" to mirror "ODATA CLIENT" naming convention.

## Complexity

**Medium** — polymorphic method types, value templates with parameter interpolation, nested operation structure.

## Testing

- Verify against Evora project
