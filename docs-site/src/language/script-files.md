# Script Files

MDL statements can be saved in `.mdl` files and executed as scripts. This is the primary way to automate Mendix model changes.

## File Format

- **Extension:** `.mdl`
- **Encoding:** UTF-8
- **Statement termination:** Semicolons (`;`) or forward slash (`/`) on its own line

A typical script file:

```sql
-- setup_domain_model.mdl
-- Creates the Sales domain model

CREATE ENUMERATION Sales.OrderStatus (
  Draft 'Draft',
  Pending 'Pending',
  Confirmed 'Confirmed',
  Shipped 'Shipped'
);

CREATE PERSISTENT ENTITY Sales.Customer (
  Name: String(200) NOT NULL,
  Email: String(200) UNIQUE
);

CREATE PERSISTENT ENTITY Sales.Order (
  OrderDate: DateTime NOT NULL,
  Status: Enumeration(Sales.OrderStatus) DEFAULT 'Draft'
);

CREATE ASSOCIATION Sales.Order_Customer
  FROM Sales.Customer
  TO Sales.Order
  TYPE Reference;
```

## Execution Modes

mxcli supports several ways to execute MDL. All modes open a single connection to the project — there is no per-command overhead.

### Script File

Use `mxcli exec` to run a `.mdl` file:

```bash
mxcli exec setup_domain_model.mdl -p /path/to/app.mpr
```

### Inline Commands (`-c`)

Pass one or more semicolon-separated commands with the `-c` flag:

```bash
# Single command
mxcli -p app.mpr -c "SHOW ENTITIES"

# Multiple commands in one connection
mxcli -p app.mpr -c "DESCRIBE ENTITY Sales.Customer; DESCRIBE ENTITY Sales.Order; SHOW MICROFLOWS IN Sales"
```

This is the fastest way for AI agents to batch multiple queries — all commands share a single connection.

### Stdin Piping

When stdin is a pipe (not a terminal), mxcli reads commands from it in quiet mode (no banner, no prompts):

```bash
# Pipe from echo
echo "SHOW ENTITIES; SHOW MICROFLOWS" | mxcli -p app.mpr

# Pipe from file
mxcli -p app.mpr < commands.mdl

# Pipe from heredoc
mxcli -p app.mpr <<'EOF'
DESCRIBE ENTITY Sales.Customer;
DESCRIBE ENTITY Sales.Order;
SHOW MICROFLOWS IN Sales;
EOF
```

### From the REPL

Use `EXECUTE SCRIPT` inside an interactive session:

```sql
CONNECT LOCAL '/path/to/app.mpr';
EXECUTE SCRIPT './scripts/setup_domain_model.mdl';
```

## Syntax Checking

You can validate a script without connecting to a project:

```bash
# Syntax only
mxcli check script.mdl

# Syntax + reference validation (requires project)
mxcli check script.mdl -p app.mpr --references
```

Syntax checking catches parse errors, unknown keywords, and common anti-patterns before execution.

## Comments in Scripts

Scripts support the same [comment syntax](./comments.md) as interactive MDL:

```sql
-- Single-line comment
/* Multi-line comment */
/** Documentation comment (attached to next element) */
```
