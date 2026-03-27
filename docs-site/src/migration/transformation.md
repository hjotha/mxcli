# Phase 2: Transformation Plan

With the assessment complete, the agent creates a detailed transformation plan that maps every source element to its Mendix equivalent.

## Module Structure

Plan the Mendix module structure. A common pattern is to consolidate many small source modules into fewer Mendix modules:

```
Source Application           Mendix
─────────────────           ─────────
CRM_Core, CRM_UI    ──▶    CRM
Cases_Core, Cases_UI ──▶    Cases
Auth_Core, Auth_SSO  ──▶    Administration
Shared_Utils         ──▶    Commons
```

## Transformation Mapping

For each element in the assessment, the agent documents the target Mendix implementation:

| Source Element | Source Location | Mendix Target | MDL Statement |
|---------------|----------------|---------------|---------------|
| Customer table | `schema.sql` | `CRM.Customer` entity | `CREATE PERSISTENT ENTITY` |
| OrderStatus enum | `enums.py` | `Sales.OrderStatus` enumeration | `CREATE ENUMERATION` |
| calculateTotal() | `OrderService.java` | `Sales.ACT_Order_CalculateTotal` microflow | `CREATE MICROFLOW` |
| Customer list page | `customers.html` | `CRM.Customer_Overview` page | `CREATE PAGE` |
| Admin role | `security.xml` | `Administration.Admin` module role | `GRANT` statements |

## Prioritization

Order the migration work to maximize early value and minimize dependency issues:

1. **Enumerations** -- no dependencies, used by entities
2. **Domain model** -- entities, attributes, associations
3. **Security** -- module roles, user roles, access rules
4. **Core business logic** -- validation microflows, calculation microflows
5. **Pages** -- overview pages, edit forms, dashboards
6. **Integrations** -- REST clients, OData services, file handling
7. **Navigation** -- menu structure, home pages
8. **Advanced features** -- scheduled events, workflows, business events

This order ensures that each layer can reference the elements created before it.
