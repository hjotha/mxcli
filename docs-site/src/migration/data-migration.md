# Data Migration

mxcli can connect directly to external databases to explore schemas, import data, and generate database connector code. This is useful for migrating reference data, seeding test data, or setting up ongoing integrations with legacy systems.

## Key Skills

| Skill | Purpose |
|-------|---------|
| `demo-data` | Mendix ID system, association storage, direct PostgreSQL insertion |
| `database-connections` | External database connectivity from microflows (Database Connector module) |
| `patterns-data-processing` | Loop patterns, batch processing, list operations |

## Connect to the Source Database

mxcli supports PostgreSQL, Oracle, and SQL Server:

```sql
-- Connect to the legacy database
SQL CONNECT postgres 'host=legacy-db port=5432 dbname=crm user=readonly password=...' AS legacy;

-- Explore the schema
SQL legacy SHOW TABLES;
SQL legacy DESCRIBE customers;
```

## Import Data

The `IMPORT FROM` statement reads from the source database and inserts into the Mendix application's PostgreSQL database:

```sql
-- Import customers
IMPORT FROM legacy
  QUERY 'SELECT name, email, phone, active FROM customers'
  INTO CRM.Customer
  MAP (name AS Name, email AS Email, phone AS Phone, active AS IsActive)
  BATCH 500;

-- Import with association linking
IMPORT FROM legacy
  QUERY 'SELECT order_number, order_date, total, customer_email FROM orders'
  INTO Sales.Order
  MAP (order_number AS OrderNumber, order_date AS OrderDate, total AS TotalAmount)
  LINK (customer_email TO Sales.Order_Customer ON Email)
  BATCH 500;
```

The import pipeline handles:

- **Mendix ID generation** -- creates valid 64-bit object IDs
- **Batch insertion** -- respects PostgreSQL parameter limits
- **Association linking** -- looks up related entities by attribute value
- **Optimistic locking** -- sets `MxObjectVersion` if the entity uses it

## Generate Database Connectors

For ongoing integration (not one-time import), generate Database Connector entities and microflows:

```sql
-- Auto-generate non-persistent entities and query microflows
SQL legacy GENERATE CONNECTOR INTO Integration
  TABLES (customers, orders, products);
```

This creates:

- Non-persistent entities with mapped attributes
- Constants for connection strings (JDBC URL, username, password)
- `DATABASE CONNECTION` definitions with query microflows

## Credential Management

Database credentials should never be hardcoded. mxcli supports:

- **Environment variables** -- set `MXCLI_SQL_<ALIAS>_DSN`
- **YAML config file** -- `~/.mxcli/sql.yaml` with per-alias DSN entries
- **Credential isolation** -- credentials are never exposed to the AI agent or logged
