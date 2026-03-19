# Proposal: SHOW/DESCRIBE Published REST Services

## Overview

**Document type:** `Rest$PublishedRestService`
**Prevalence:** 16 across test projects (1 Enquiries, 7 Evora, 8 Lato)
**Priority:** High — REST APIs are critical for modern Mendix apps

Published REST Services expose HTTP endpoints backed by microflows. Each service has a base path, version, authentication configuration, and one or more resources with operations (GET, POST, PUT, DELETE, PATCH).

## What Already Exists

| Layer | Status | Location |
|-------|--------|----------|
| **Go type** | Partial | `model/types.go` line 556 — Name, Path, Version, ServiceName, Excluded, Resources |
| **Parser** | Partial | `sdk/mpr/parser_rest.go` — parses basic fields + resources + operations |
| **Reader** | Yes | `ListPublishedRestServices()` in `sdk/mpr/reader_documents.go` |
| **Generated metamodel** | Yes | `generated/metamodel/types.go` line 8177 |
| **AST** | No | — |
| **Executor** | No | — |

## BSON Structure (from test projects)

```
Rest$PublishedRestService:
  Name: string
  Documentation: string
  Excluded: bool
  ExportLevel: string
  Path: string (e.g., "rest/customers/v1")
  Version: string (e.g., "1.0.0")
  ServiceName: string
  AllowedRoles: [] (qualified role names)
  AuthenticationTypes: [] ("Basic", "Custom", "None")
  AuthenticationMicroflow: string (qualified name)
  CorsConfiguration: nullable
  Parameters: []*RestOperationParameter (service-level params)
  Resources: []*Rest$PublishedRestServiceResource
    - Name: string (e.g., "Customers")
    - Documentation: string
    - Operations: []*Rest$PublishedRestServiceOperation
      - HttpMethod: string ("GET", "POST", "PUT", "DELETE", "PATCH")
      - Path: string (e.g., "/{id}")
      - Microflow: string (qualified name)
      - Summary: string
      - Deprecated: bool
      - Commit: string ("Yes", "No")
      - ImportMapping: string (qualified name)
      - ExportMapping: string (qualified name)
      - Parameters: []*RestOperationParameter
```

## Proposed MDL Syntax

### SHOW REST SERVICES

```
SHOW REST SERVICES [IN Module]
```

| Qualified Name | Module | Name | Path | Version | Resources | Operations |
|----------------|--------|------|------|---------|-----------|------------|

### DESCRIBE REST SERVICE

```
DESCRIBE REST SERVICE Module.Name
```

Output format:

```
/**
 * Customer management API
 */
REST SERVICE MyModule.CustomerAPI
  PATH 'rest/customers/v1'
  VERSION '1.0.0'
  AUTHENTICATION Basic
{
  RESOURCE Customers
  {
    GET /
      MICROFLOW MyModule.GetAllCustomers
      EXPORT MAPPING MyModule.ExportCustomerList
      SUMMARY 'List all customers';

    GET /{id}
      MICROFLOW MyModule.GetCustomerById
      EXPORT MAPPING MyModule.ExportCustomer
      SUMMARY 'Get customer by ID';

    POST /
      MICROFLOW MyModule.CreateCustomer
      IMPORT MAPPING MyModule.ImportCustomer
      COMMIT Yes
      SUMMARY 'Create a new customer';

    PUT /{id}
      MICROFLOW MyModule.UpdateCustomer
      IMPORT MAPPING MyModule.ImportCustomer
      COMMIT Yes
      SUMMARY 'Update an existing customer';

    DELETE /{id}
      MICROFLOW MyModule.DeleteCustomer
      SUMMARY 'Delete a customer';
  };
};
/
```

## Implementation Steps

### 1. Enhance Model Type (model/types.go)

The existing `PublishedRestService` struct needs:
- `Documentation`, `AllowedRoles`, `AuthenticationTypes`, `AuthenticationMicroflow`

The existing `RestResource` and `RestOperation` structs need:
- `Summary`, `Deprecated`, `Commit`, `ImportMapping`, `ExportMapping`

### 2. Enhance Parser (sdk/mpr/parser_rest.go)

Extend existing parser to capture all fields listed above.

### 3. Add AST Types

```go
ShowRestServices    // in ShowObjectType enum
DescribeRestService // in DescribeObjectType enum
```

### 4. Add Grammar Rules

```antlr
REST: 'REST';
SERVICE: 'SERVICE';   // may already exist for OData
SERVICES: 'SERVICES'; // may already exist

// SHOW REST SERVICES [IN module]
// DESCRIBE REST SERVICE qualifiedName
```

### 5. Add Executor (mdl/executor/cmd_rest_services.go)

- `showRestServices(moduleName string)` — table listing
- `describeRestService(name QualifiedName)` — MDL output with resources and operations

### 6. Add Autocomplete

```go
func (e *Executor) GetRestServiceNames(moduleFilter string) []string
```

## Testing

- Create `mdl-examples/doctype-tests/19-rest-service-examples.mdl`
- Verify against Lato project (8 REST services — most comprehensive)
