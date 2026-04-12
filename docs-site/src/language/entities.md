# Entities

Entities are the primary data structures in a Mendix domain model. Each entity corresponds to a database table (for persistent entities) or an in-memory object (for non-persistent entities).

## Entity Types

| Type | MDL Keyword | Description |
|------|-------------|-------------|
| Persistent | `PERSISTENT` | Stored in the database with a corresponding table |
| Non-Persistent | `NON-PERSISTENT` | In-memory only, scoped to the user session |
| View | `VIEW` | Based on an OQL query, read-only |
| External | `EXTERNAL` | From an external data source (OData, etc.) |

## CREATE ENTITY

```sql
[/** <documentation> */]
[@Position(<x>, <y>)]
CREATE [OR MODIFY] <entity-type> ENTITY <Module>.<Name> (
  <attribute-definitions>
)
[INDEX (<column-list>)]
[ON BEFORE|AFTER CREATE|COMMIT|DELETE|ROLLBACK CALL <Module>.<Microflow> [RAISE ERROR]]
[;|/]
```

### Persistent Entity

The most common type. Data is stored in the application database:

```sql
/** Customer master data */
@Position(100, 200)
CREATE PERSISTENT ENTITY Sales.Customer (
  /** Auto-incrementing unique identifier */
  CustomerId: AutoNumber NOT NULL UNIQUE DEFAULT 1,
  /** Full legal name of the customer */
  Name: String(200) NOT NULL ERROR 'Name is required',
  /** Primary contact email address */
  Email: String(200) UNIQUE ERROR 'Email must be unique',
  /** Current account balance */
  Balance: Decimal DEFAULT 0,
  /** Whether the account is active */
  IsActive: Boolean DEFAULT TRUE,
  /** Timestamp of account creation */
  CreatedDate: DateTime,
  /** Current lifecycle status */
  Status: Enumeration(Sales.CustomerStatus) DEFAULT 'Active'
)
INDEX (Name)
INDEX (Email);
```

### Non-Persistent Entity

Used for helper objects, filter parameters, and UI state that does not need database storage:

```sql
CREATE NON-PERSISTENT ENTITY Sales.CustomerFilter (
  SearchName: String(200),
  MinBalance: Decimal,
  MaxBalance: Decimal
);
```

### View Entity

Defined by an OQL query. View entities are read-only:

```sql
CREATE VIEW ENTITY Reports.CustomerSummary (
  CustomerName: String,
  TotalOrders: Integer,
  TotalAmount: Decimal
) AS
  SELECT
    c.Name AS CustomerName,
    COUNT(o.OrderId) AS TotalOrders,
    SUM(o.Amount) AS TotalAmount
  FROM Sales.Customer c
  LEFT JOIN Sales.Order o ON o.Customer = c
  GROUP BY c.Name;
```

## CREATE OR MODIFY

Creates the entity if it does not exist, or updates it if it does. New attributes are added; existing attributes are preserved:

```sql
CREATE OR MODIFY PERSISTENT ENTITY Sales.Customer (
  CustomerId: AutoNumber NOT NULL UNIQUE,
  Name: String(200) NOT NULL,
  Email: String(200),
  Phone: String(50)  -- new attribute added on modify
);
```

## System Attributes (Auditing)

Persistent entities can track who created/modified objects and when, using built-in system attributes. Add these clauses after the closing parenthesis:

```sql
CREATE PERSISTENT ENTITY Sales.Order (
  OrderNumber: AutoNumber,
  TotalAmount: Decimal
)
STORE OWNER
STORE CHANGED BY
STORE CREATED DATE
STORE CHANGED DATE;
```

| Clause | System Attribute | Type | Set When |
|--------|-----------------|------|----------|
| `STORE OWNER` | `System.owner` | Association → System.User | Object created |
| `STORE CHANGED BY` | `System.changedBy` | Association → System.User | Every commit |
| `STORE CREATED DATE` | `CreatedDate` | DateTime | Object created |
| `STORE CHANGED DATE` | `ChangedDate` | DateTime | Every commit |

Toggle on existing entities with ALTER ENTITY:

```sql
ALTER ENTITY Sales.Order SET STORE OWNER;
ALTER ENTITY Sales.Order DROP STORE CHANGED DATE;
```

## Annotations

### @Position

Controls where the entity appears in the domain model diagram:

```sql
@Position(100, 200)
CREATE PERSISTENT ENTITY Sales.Customer ( ... );
```

### Documentation

A `/** ... */` comment before the entity or before an individual attribute becomes its documentation in Studio Pro:

```sql
/** Customer master data.
 *  Stores both active and inactive customers.
 */
@Position(100, 200)
CREATE PERSISTENT ENTITY Sales.Customer (
  /** Auto-incrementing unique identifier */
  CustomerId: AutoNumber NOT NULL UNIQUE DEFAULT 1,

  /** Full legal name of the customer */
  Name: String(200) NOT NULL ERROR 'Name is required',

  /** Primary contact email address */
  Email: String(200) UNIQUE ERROR 'Email must be unique',

  /** Current account balance in the base currency */
  Balance: Decimal DEFAULT 0,

  /** Whether the customer account is active */
  IsActive: Boolean DEFAULT TRUE,

  /** Timestamp of account creation */
  CreatedDate: DateTime,

  /** Current lifecycle status */
  Status: Enumeration(Sales.CustomerStatus) DEFAULT 'Active'
)
INDEX (Name)
INDEX (Email);
```

Attribute-level documentation appears in Studio Pro when hovering over the attribute in the domain model.

## Entity Event Handlers

Persistent entities can have microflow event handlers that run before or after Create, Commit, Delete, or Rollback operations. The optional `RAISE ERROR` clause makes the handler act as a validation microflow — if it returns false, the operation is aborted.

```sql
CREATE PERSISTENT ENTITY Sales.Order (
  Total: Decimal,
  Status: String(50)
)
ON BEFORE COMMIT CALL Sales.ACT_ValidateOrder RAISE ERROR
ON AFTER CREATE CALL Sales.ACT_InitDefaults;
```

Event handlers can also be added or removed via `ALTER ENTITY`:

```sql
-- Add a handler to an existing entity
ALTER ENTITY Sales.Order
  ADD EVENT HANDLER ON BEFORE DELETE CALL Sales.ACT_CheckCanDelete RAISE ERROR;

-- Remove a handler
ALTER ENTITY Sales.Order
  DROP EVENT HANDLER ON BEFORE COMMIT;
```

**Moments**: `BEFORE`, `AFTER`
**Events**: `CREATE`, `COMMIT`, `DELETE`, `ROLLBACK`

Each (Moment, Event) combination supports one handler per entity. The microflow must exist in the project (validated at execution time).

## DROP ENTITY

Removes an entity from the domain model:

```sql
DROP ENTITY Sales.Customer;
```

## See Also

- [Attributes](./attributes.md) -- attribute definitions within entities
- [Indexes](./indexes.md) -- adding indexes to entities
- [Generalization](./generalization.md) -- entity inheritance with EXTENDS
- [ALTER ENTITY](./alter-entity.md) -- modifying existing entities
- [Associations](./associations.md) -- relationships between entities
