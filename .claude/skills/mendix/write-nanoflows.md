# Mendix Nanoflow Skill

This skill provides guidance for writing Mendix nanoflows in MDL syntax. Nanoflows share syntax with microflows but execute client-side with restricted capabilities.

## When to Use This Skill

Use this skill when:
- Writing CREATE NANOFLOW statements
- Debugging nanoflow validation errors
- Understanding nanoflow restrictions vs microflows

## Key Differences from Microflows

| Aspect | Microflow | Nanoflow |
|--------|-----------|----------|
| **Execution** | Server-side | Client-side (browser/mobile) |
| **Database access** | Full | No direct access |
| **Transactions** | Supported | Not supported |
| **Java actions** | Supported | Not supported |
| **JavaScript actions** | Not supported | Supported |
| **File downloads** | Supported | Not supported |
| **Error handling** | Full `ON ERROR` blocks | Limited |
| **Offline** | Not available | Available |

## Nanoflow Structure

```mdl
/**
 * Nanoflow description
 *
 * @param $Parameter1 Description
 * @returns Description of return value
 */
CREATE [OR REPLACE] NANOFLOW Module.NAV_Name (
  $Parameter1: type
)
RETURNS ReturnType AS $Result
FOLDER 'FolderPath'
BEGIN
  -- Nanoflow logic here
  RETURN $Result;
END;
```

## Naming Convention

Nanoflow names use the `NAV_` prefix by convention:
- `NAV_ValidateCart` — client-side validation
- `NAV_ShowDetails` — page navigation
- `NAV_ToggleFilter` — UI state toggle

## Supported Activities

### Object Operations (in-memory only)
```mdl
$Item = CREATE Sales.CartItem (Quantity = 1);
CHANGE $Item (Quantity = $Item/Quantity + 1);
```

### Calling Other Flows
```mdl
$Result = CALL NANOFLOW Sales.NAV_ValidateCart (Cart = $Cart);
$ServerResult = CALL MICROFLOW Sales.ACT_SubmitOrder (Order = $Order);
$JsResult = CALL JAVASCRIPT ACTION MyModule.MyJsAction (Param = $Value);
```

### UI Activities
```mdl
SHOW PAGE Sales.CartDetail ($Cart = $Cart);
CLOSE PAGE;
VALIDATION FEEDBACK $Item/Quantity MESSAGE 'Quantity must be at least 1';
```

### Logging and Variables
```mdl
LOG INFO 'Cart updated with ' + toString($ItemCount) + ' items';
DECLARE $IsValid Boolean = true;
SET $IsValid = false;
```

### Control Flow
```mdl
IF $Cart/ItemCount = 0 THEN
  VALIDATION FEEDBACK $Cart/ItemCount MESSAGE 'Cart is empty';
  RETURN false;
ELSE
  SHOW PAGE Sales.Checkout ($Cart = $Cart);
  RETURN true;
END IF;
```

## Disallowed Activities

These will produce validation errors:
- `RETRIEVE ... FROM Module.Entity WHERE ...` (database retrieval)
- `COMMIT`
- `DELETE`
- `ROLLBACK`
- `CALL JAVA ACTION`
- `EXECUTE DATABASE QUERY`
- `DOWNLOAD FILE`
- REST calls (`CALL REST SERVICE`, `SEND REST REQUEST`)
- Import/export mapping
- JSON transformation
- All workflow actions

## Return Type Restrictions

Binary return type is NOT allowed in nanoflows.

## Security (GRANT/REVOKE)

```mdl
GRANT EXECUTE ON NANOFLOW Shop.NAV_Filter TO Shop.User, Shop.Admin;
REVOKE EXECUTE ON NANOFLOW Shop.NAV_Filter FROM Shop.User;
```

## Management Commands

```mdl
SHOW NANOFLOWS
SHOW NANOFLOWS IN MyModule
DESCRIBE NANOFLOW MyModule.NAV_ShowDetails
DROP NANOFLOW MyModule.NAV_ShowDetails;
MOVE NANOFLOW Sales.NAV_OpenCart TO FOLDER 'UI/Navigation';
```

## Common Mistakes

1. **Using database operations** — Nanoflows cannot access the database directly. Use CALL MICROFLOW for server operations.
2. **Using Java actions** — Use CALL JAVASCRIPT ACTION instead.
3. **Expecting transactions** — Nanoflows have no rollback. Design for idempotency.
4. **File operations** — DOWNLOAD FILE is server-only.
5. **Binary return types** — Not supported in nanoflows.
6. **Full error handling** — `ON ERROR { ... }` blocks are limited in nanoflows.

## Validation Checklist

- [ ] No database operations (RETRIEVE with WHERE, COMMIT, DELETE, ROLLBACK)
- [ ] No Java action calls
- [ ] No REST calls or external action calls
- [ ] No file download operations
- [ ] No workflow actions
- [ ] No binary return type
- [ ] Parameters and return types are nanoflow-compatible
- [ ] JavaDoc documentation present
- [ ] NAV_ naming prefix used
