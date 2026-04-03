# GRANT

## Synopsis

```sql
-- Entity access
GRANT module.Role ON module.Entity ( rights ) [ WHERE 'xpath' ]

-- Microflow access
GRANT EXECUTE ON MICROFLOW module.Name TO module.Role [, ...]

-- Page access
GRANT VIEW ON PAGE module.Name TO module.Role [, ...]

-- Nanoflow access
GRANT EXECUTE ON NANOFLOW module.Name TO module.Role [, ...]
```

## Description

Grants access rights to module roles. There are four forms of the GRANT statement, each controlling a different kind of access.

### Entity Access

The entity access form creates or updates an access rule on an entity for a given module role. **GRANT is additive**: if the role already has an access rule on the entity, the new rights are merged with existing ones. Existing permissions are never downgraded by a GRANT.

Entity access rules control:
- **CREATE** -- whether the role can create new instances
- **DELETE** -- whether the role can delete instances
- **READ** -- which attributes the role can read
- **WRITE** -- which attributes the role can modify

For READ and WRITE, use `*` to include all members (attributes and associations), or specify a parenthesized list of specific attribute names.

The optional `WHERE` clause accepts an XPath expression that restricts which objects the role can access. The XPath is enclosed in single quotes. Use doubled single quotes to escape single quotes inside the expression.

### Microflow Access

The microflow access form grants execute permission on a microflow to one or more module roles. This controls whether the role can trigger the microflow (from pages, other microflows, or REST/web service calls).

### Page Access

The page access form grants view permission on a page to one or more module roles. This controls whether the role can open the page.

### Nanoflow Access

The nanoflow access form grants execute permission on a nanoflow to one or more module roles.

## Parameters

`module.Role`
:   The module role receiving the access grant. Must be a qualified name (`Module.RoleName`).

`module.Entity`
:   The target entity for entity access rules.

`rights`
:   A comma-separated list of access rights for entity access. Valid values:
    - `CREATE` -- allow creating instances
    - `DELETE` -- allow deleting instances
    - `READ *` -- read all members
    - `READ (Attr1, Attr2, ...)` -- read specific attributes
    - `WRITE *` -- write all members
    - `WRITE (Attr1, Attr2, ...)` -- write specific attributes

`WHERE 'xpath'`
:   Optional XPath constraint for entity access. Restricts which objects the rule applies to.

`module.Name`
:   The target microflow, nanoflow, or page.

`TO module.Role [, ...]`
:   One or more module roles receiving the access (for microflow, page, and nanoflow forms).

## Examples

Grant full entity access:

```sql
GRANT Shop.Admin ON Shop.Customer (CREATE, DELETE, READ *, WRITE *);
```

Grant read-only access:

```sql
GRANT Shop.Viewer ON Shop.Customer (READ *);
```

Grant selective attribute access:

```sql
GRANT Shop.User ON Shop.Customer (READ (Name, Email), WRITE (Email));
```

Grant entity access with an XPath constraint:

```sql
GRANT Shop.User ON Shop.Order (READ *, WRITE *)
    WHERE '[Status = ''Open'']';
```

Grant microflow execution to multiple roles:

```sql
GRANT EXECUTE ON MICROFLOW Shop.ACT_Order_Process TO Shop.User, Shop.Admin;
```

Grant page visibility:

```sql
GRANT VIEW ON PAGE Shop.Order_Overview TO Shop.User, Shop.Admin;
```

Grant nanoflow execution:

```sql
GRANT EXECUTE ON NANOFLOW Shop.NAV_ValidateInput TO Shop.User;
```

Additive grant -- add new attribute access without removing existing:

```sql
-- Viewer already has READ (Name, Email)
GRANT Shop.Viewer ON Shop.Customer (READ (Phone));
-- Result: READ (Name, Email, Phone)
```

## See Also

[REVOKE](revoke.md), [CREATE MODULE ROLE](create-module-role.md), [CREATE USER ROLE](create-user-role.md)
