# MDL Parser Test Cases

Test cases for `mxcli check` syntax validation. Each test case includes MDL input and expected result.

## Test Format

```
Test ID: Unique identifier
Category: Feature category
Description: What is being tested
Expected: PASS | FAIL
Input: MDL code block
Notes: Additional context (optional)
```

---

## Category: Primitive Variable Declarations

### TEST-VAR-001: Integer declaration with initialization

```yaml
id: TEST-VAR-001
category: variables/primitives
description: Declare integer variable with initial value
expected: PASS
```

```mdl
CREATE MICROFLOW Test.IntegerDecl()
RETURNS Boolean
BEGIN
    DECLARE $x Integer = 0;
    RETURN true;
END;
/
```

### TEST-VAR-002: String declaration with initialization

```yaml
id: TEST-VAR-002
category: variables/primitives
description: Declare string variable with initial value
expected: PASS
```

```mdl
CREATE MICROFLOW Test.StringDecl()
RETURNS Boolean
BEGIN
    DECLARE $s String = 'hello';
    RETURN true;
END;
/
```

### TEST-VAR-003: Boolean declaration with initialization

```yaml
id: TEST-VAR-003
category: variables/primitives
description: Declare boolean variable with initial value
expected: PASS
```

```mdl
CREATE MICROFLOW Test.BooleanDecl()
RETURNS Boolean
BEGIN
    DECLARE $b Boolean = true;
    RETURN true;
END;
/
```

### TEST-VAR-004: Decimal declaration with initialization

```yaml
id: TEST-VAR-004
category: variables/primitives
description: Declare decimal variable with initial value
expected: PASS
```

```mdl
CREATE MICROFLOW Test.DecimalDecl()
RETURNS Boolean
BEGIN
    DECLARE $d Decimal = 0.0;
    RETURN true;
END;
/
```

### TEST-VAR-005: Multiple primitive declarations

```yaml
id: TEST-VAR-005
category: variables/primitives
description: Declare multiple primitive variables in one microflow
expected: PASS
```

```mdl
CREATE MICROFLOW Test.MultiplePrimitives()
RETURNS Boolean
BEGIN
    DECLARE $x Integer = 0;
    DECLARE $s String = 'hello';
    DECLARE $b Boolean = true;
    DECLARE $d Decimal = 0.0;
    RETURN true;
END;
/
```

---

## Category: List Variable Declarations

### TEST-LIST-001: List declaration with qualified entity name

```yaml
id: TEST-LIST-001
category: variables/lists
description: Declare list variable with Module.Entity syntax
expected: PASS
```

```mdl
CREATE MICROFLOW Test.ListDecl()
RETURNS Boolean
BEGIN
    DECLARE $Orders List of Test.Order = empty;
    RETURN true;
END;
/
```

### TEST-LIST-002: List declaration with reserved word entity name

```yaml
id: TEST-LIST-002
category: variables/lists
description: Declare list with entity name "Item" (reserved word)
expected: FAIL
notes: "Item" is a reserved word in the parser, causing syntax error
```

```mdl
CREATE MICROFLOW Test.ListDeclItem()
RETURNS Boolean
BEGIN
    DECLARE $Items List of Test.Item = empty;
    RETURN true;
END;
/
```

### TEST-LIST-003: List declaration with compound entity name containing reserved word

```yaml
id: TEST-LIST-003
category: variables/lists
description: Declare list with entity name "OrderItem" (contains reserved word but is valid)
expected: PASS
```

```mdl
CREATE MICROFLOW Test.ListDeclOrderItem()
RETURNS Boolean
BEGIN
    DECLARE $Items List of Test.OrderItem = empty;
    RETURN true;
END;
/
```

### TEST-LIST-004: List declaration with "LineItem" entity name

```yaml
id: TEST-LIST-004
category: variables/lists
description: Declare list with entity name "LineItem"
expected: PASS
```

```mdl
CREATE MICROFLOW Test.ListDeclLineItem()
RETURNS Boolean
BEGIN
    DECLARE $Items List of Test.LineItem = empty;
    RETURN true;
END;
/
```

---

## Category: Entity Variable Declarations

### TEST-ENTITY-001: Entity declaration with AS and qualified name

```yaml
id: TEST-ENTITY-001
category: variables/entities
description: Declare entity variable using AS keyword with Module.Entity
expected: FAIL
notes: Parser does not accept qualified name after AS keyword
```

```mdl
CREATE MICROFLOW Test.EntityDecl()
RETURNS Boolean
BEGIN
    DECLARE $Order AS Test.Order;
    RETURN true;
END;
/
```

### TEST-ENTITY-002: Entity declaration with AS and simple name

```yaml
id: TEST-ENTITY-002
category: variables/entities
description: Declare entity variable using AS keyword with simple name (no module)
expected: FAIL
notes: Parser does not accept any identifier after AS keyword
```

```mdl
CREATE MICROFLOW Test.EntityDeclSimple()
RETURNS Boolean
BEGIN
    DECLARE $Order AS Order;
    RETURN true;
END;
/
```

---

## Category: Entity Creation

### TEST-CREATE-001: Create entity with assignment

```yaml
id: TEST-CREATE-001
category: statements/create
description: Create entity and assign to variable
expected: FAIL
notes: Entity creation syntax not supported in parser
```

```mdl
CREATE MICROFLOW Test.CreateEntity()
RETURNS Boolean
BEGIN
    DECLARE $Order AS Test.Order;
    $Order = CREATE Test.Order (
        Name = 'test'
    );
    RETURN true;
END;
/
```

---

## Category: RETRIEVE Statement

### TEST-RETRIEVE-001: Basic RETRIEVE from entity

```yaml
id: TEST-RETRIEVE-001
category: statements/retrieve
description: Retrieve all objects of an entity type
expected: PASS
```

```mdl
CREATE MICROFLOW Test.RetrieveBasic()
RETURNS Boolean
BEGIN
    DECLARE $Orders List of Test.Order = empty;
    RETRIEVE $Orders FROM Test.Order;
    RETURN true;
END;
/
```

### TEST-RETRIEVE-002: RETRIEVE with WHERE clause using association

```yaml
id: TEST-RETRIEVE-002
category: statements/retrieve
description: Retrieve with association filter
expected: PASS
```

```mdl
CREATE MICROFLOW Test.RetrieveWhere($Customer: Test.Customer)
RETURNS Boolean
BEGIN
    DECLARE $Orders List of Test.Order = empty;
    RETRIEVE $Orders FROM Test.Order
        WHERE Test.Order_Customer = $Customer;
    RETURN true;
END;
/
```

---

## Category: LOOP Statement

### TEST-LOOP-001: Basic LOOP over list

```yaml
id: TEST-LOOP-001
category: statements/loop
description: Iterate over a list with LOOP
expected: PASS
```

```mdl
CREATE MICROFLOW Test.LoopBasic()
RETURNS Boolean
BEGIN
    DECLARE $Orders List of Test.Order = empty;
    LOOP $Order IN $Orders
    BEGIN
        DECLARE $x Integer = 1;
    END LOOP;
    RETURN true;
END;
/
```

### TEST-LOOP-002: LOOP with SET inside

```yaml
id: TEST-LOOP-002
category: statements/loop
description: Iterate and accumulate with SET
expected: PASS
```

```mdl
CREATE MICROFLOW Test.LoopWithSet()
RETURNS Boolean
BEGIN
    DECLARE $Total Decimal = 0;
    DECLARE $Orders List of Test.Order = empty;

    RETRIEVE $Orders FROM Test.Order;

    LOOP $Order IN $Orders
    BEGIN
        SET $Total = $Total + $Order/Amount;
    END LOOP;

    RETURN true;
END;
/
```

### TEST-LOOP-003: LOOP with CHANGE and COMMIT inside

```yaml
id: TEST-LOOP-003
category: statements/loop
description: Iterate and modify objects
expected: PASS
```

```mdl
CREATE MICROFLOW Test.LoopWithChange($Customer: Test.Customer)
RETURNS Boolean
BEGIN
    DECLARE $Orders List of Test.Order = empty;

    RETRIEVE $Orders FROM Test.Order
        WHERE Test.Order_Customer = $Customer;

    LOOP $Order IN $Orders
    BEGIN
        CHANGE $Order (Status = 'Processed');
        COMMIT $Order;
    END LOOP;

    RETURN true;
END;
/
```

---

## Category: CHANGE Statement

### TEST-CHANGE-001: Basic CHANGE with single attribute

```yaml
id: TEST-CHANGE-001
category: statements/change
description: Change single attribute on entity
expected: PASS
```

```mdl
CREATE MICROFLOW Test.ChangeBasic($Order: Test.Order)
RETURNS Boolean
BEGIN
    CHANGE $Order (Status = 'Complete');
    RETURN true;
END;
/
```

### TEST-CHANGE-002: CHANGE with multiple attributes

```yaml
id: TEST-CHANGE-002
category: statements/change
description: Change multiple attributes on entity
expected: PASS
```

```mdl
CREATE MICROFLOW Test.ChangeMultiple($Order: Test.Order)
RETURNS Boolean
BEGIN
    CHANGE $Order (
        Status = 'Complete',
        ProcessedDate = [%CurrentDateTime%]
    );
    RETURN true;
END;
/
```

---

## Category: COMMIT Statement

### TEST-COMMIT-001: Basic COMMIT

```yaml
id: TEST-COMMIT-001
category: statements/commit
description: Commit entity to database
expected: PASS
```

```mdl
CREATE MICROFLOW Test.CommitBasic($Order: Test.Order)
RETURNS Boolean
BEGIN
    COMMIT $Order;
    RETURN true;
END;
/
```

### TEST-COMMIT-002: COMMIT WITH EVENTS

```yaml
id: TEST-COMMIT-002
category: statements/commit
description: Commit entity with event handlers
expected: PASS
```

```mdl
CREATE MICROFLOW Test.CommitWithEvents($Order: Test.Order)
RETURNS Boolean
BEGIN
    COMMIT $Order WITH EVENTS;
    RETURN true;
END;
/
```

---

## Category: DELETE Statement

### TEST-DELETE-001: Basic DELETE

```yaml
id: TEST-DELETE-001
category: statements/delete
description: Delete entity from database
expected: PASS
```

```mdl
CREATE MICROFLOW Test.DeleteBasic($Order: Test.Order)
RETURNS Boolean
BEGIN
    DELETE $Order;
    RETURN true;
END;
/
```

---

## Category: ROLLBACK Statement

### TEST-ROLLBACK-001: Basic ROLLBACK

```yaml
id: TEST-ROLLBACK-001
category: statements/rollback
description: Rollback changes to entity
expected: FAIL
notes: ROLLBACK is documented but not implemented in parser
```

```mdl
CREATE MICROFLOW Test.RollbackBasic($Order: Test.Order)
RETURNS Boolean
BEGIN
    ROLLBACK $Order;
    RETURN true;
END;
/
```

---

## Category: Page Navigation

### TEST-PAGE-001: SHOW PAGE with parameter

```yaml
id: TEST-PAGE-001
category: statements/navigation
description: Navigate to page with parameter
expected: PASS
```

```mdl
CREATE MICROFLOW Test.ShowPage($Order: Test.Order)
RETURNS Boolean
BEGIN
    SHOW PAGE Test.Order_Edit ($Order = $Order);
    RETURN true;
END;
/
```

### TEST-PAGE-002: CLOSE PAGE

```yaml
id: TEST-PAGE-002
category: statements/navigation
description: Close current page
expected: PASS
```

```mdl
CREATE MICROFLOW Test.ClosePage()
RETURNS Boolean
BEGIN
    CLOSE PAGE;
    RETURN true;
END;
/
```

---

## Category: Microflow Calls

### TEST-CALL-001: CALL MICROFLOW with parameter

```yaml
id: TEST-CALL-001
category: statements/call
description: Call another microflow with parameter
expected: PASS
```

```mdl
CREATE MICROFLOW Test.CallMicroflow($Order: Test.Order)
RETURNS Boolean
BEGIN
    DECLARE $Result Boolean = false;
    $Result = CALL MICROFLOW Test.ValidateOrder($Order = $Order);
    RETURN $Result;
END;
/
```

---

## Category: Validation Feedback

### TEST-VALIDATION-001: Basic validation feedback

```yaml
id: TEST-VALIDATION-001
category: statements/validation
description: Show validation feedback on attribute
expected: PASS
```

```mdl
CREATE MICROFLOW Test.ValidationFeedback($Order: Test.Order)
RETURNS Boolean
BEGIN
    DECLARE $IsValid Boolean = true;

    IF $Order/Name = empty THEN
        SET $IsValid = false;
        VALIDATION FEEDBACK $Order/Name MESSAGE 'Name is required';
    END IF;

    RETURN $IsValid;
END;
/
```

---

## Category: Logging

### TEST-LOG-001: LOG INFO

```yaml
id: TEST-LOG-001
category: statements/log
description: Log info message
expected: PASS
```

```mdl
CREATE MICROFLOW Test.LogInfo()
RETURNS Boolean
BEGIN
    LOG INFO NODE 'TestNode' 'This is an info message';
    RETURN true;
END;
/
```

### TEST-LOG-002: LOG WARNING

```yaml
id: TEST-LOG-002
category: statements/log
description: Log warning message
expected: PASS
```

```mdl
CREATE MICROFLOW Test.LogWarning()
RETURNS Boolean
BEGIN
    LOG WARNING NODE 'TestNode' 'This is a warning message';
    RETURN true;
END;
/
```

---

## Category: Control Flow

### TEST-IF-001: Basic IF statement

```yaml
id: TEST-IF-001
category: control/if
description: Simple IF condition
expected: PASS
```

```mdl
CREATE MICROFLOW Test.IfBasic($Value: Integer)
RETURNS Boolean
BEGIN
    DECLARE $Result Boolean = false;

    IF $Value > 10 THEN
        SET $Result = true;
    END IF;

    RETURN $Result;
END;
/
```

### TEST-IF-002: IF-ELSE statement

```yaml
id: TEST-IF-002
category: control/if
description: IF with ELSE branch
expected: PASS
```

```mdl
CREATE MICROFLOW Test.IfElse($Value: Integer)
RETURNS Boolean
BEGIN
    DECLARE $Result String = '';

    IF $Value > 10 THEN
        SET $Result = 'high';
    ELSE
        SET $Result = 'low';
    END IF;

    RETURN true;
END;
/
```

### TEST-IF-003: Nested IF statements

```yaml
id: TEST-IF-003
category: control/if
description: Nested IF conditions
expected: PASS
```

```mdl
CREATE MICROFLOW Test.IfNested($Value: Integer)
RETURNS String
BEGIN
    DECLARE $Result String = '';

    IF $Value > 100 THEN
        SET $Result = 'very high';
    ELSE
        IF $Value > 10 THEN
            SET $Result = 'high';
        ELSE
            SET $Result = 'low';
        END IF;
    END IF;

    RETURN $Result;
END;
/
```

---

## Category: Attribute Access

### TEST-ATTR-001: Read attribute from parameter

```yaml
id: TEST-ATTR-001
category: expressions/attributes
description: Access attribute using slash notation
expected: PASS
```

```mdl
CREATE MICROFLOW Test.ReadAttribute($Order: Test.Order)
RETURNS Boolean
BEGIN
    DECLARE $Name String = '';
    SET $Name = $Order/Name;
    RETURN true;
END;
/
```

### TEST-ATTR-002: Compare attribute to empty

```yaml
id: TEST-ATTR-002
category: expressions/attributes
description: Check if attribute is empty
expected: PASS
```

```mdl
CREATE MICROFLOW Test.CheckEmpty($Order: Test.Order)
RETURNS Boolean
BEGIN
    IF $Order/Name = empty THEN
        RETURN false;
    END IF;
    RETURN true;
END;
/
```

---

## Category: Reserved Words

### TEST-RESERVED-001: Entity name "Item"

```yaml
id: TEST-RESERVED-001
category: reserved-words
description: Using "Item" as entity name in qualified reference
expected: FAIL
notes: "Item" is a reserved word
```

```mdl
CREATE MICROFLOW Test.ReservedItem()
RETURNS Boolean
BEGIN
    DECLARE $Items List of Test.Item = empty;
    RETURN true;
END;
/
```

### TEST-RESERVED-002: Entity name "Items" (plural)

```yaml
id: TEST-RESERVED-002
category: reserved-words
description: Using "Items" as entity name (plural of reserved word)
expected: PASS
notes: "Items" is NOT reserved, only "Item"
```

```mdl
CREATE MICROFLOW Test.NotReservedItems()
RETURNS Boolean
BEGIN
    DECLARE $List List of Test.Items = empty;
    RETURN true;
END;
/
```

### TEST-RESERVED-003: Various common entity names

```yaml
id: TEST-RESERVED-003
category: reserved-words
description: Test common entity names for reserved word conflicts
expected: PASS
notes: Names like Order, Customer, Product, Name, Type, Value, Status should all work
```

```mdl
CREATE MICROFLOW Test.CommonNames()
RETURNS Boolean
BEGIN
    DECLARE $Orders List of Test.Order = empty;
    DECLARE $Customers List of Test.Customer = empty;
    DECLARE $Products List of Test.Product = empty;
    RETURN true;
END;
/
```

---

## Summary Table

| Test ID | Category | Expected | Description |
|---------|----------|----------|-------------|
| TEST-VAR-001 | variables/primitives | PASS | Integer declaration |
| TEST-VAR-002 | variables/primitives | PASS | String declaration |
| TEST-VAR-003 | variables/primitives | PASS | Boolean declaration |
| TEST-VAR-004 | variables/primitives | PASS | Decimal declaration |
| TEST-VAR-005 | variables/primitives | PASS | Multiple primitives |
| TEST-LIST-001 | variables/lists | PASS | List with qualified name |
| TEST-LIST-002 | variables/lists | FAIL | List with "Item" (reserved) |
| TEST-LIST-003 | variables/lists | PASS | List with "OrderItem" |
| TEST-LIST-004 | variables/lists | PASS | List with "LineItem" |
| TEST-ENTITY-001 | variables/entities | FAIL | DECLARE AS with qualified name |
| TEST-ENTITY-002 | variables/entities | FAIL | DECLARE AS with simple name |
| TEST-CREATE-001 | statements/create | FAIL | CREATE entity assignment |
| TEST-RETRIEVE-001 | statements/retrieve | PASS | Basic RETRIEVE |
| TEST-RETRIEVE-002 | statements/retrieve | PASS | RETRIEVE with WHERE |
| TEST-LOOP-001 | statements/loop | PASS | Basic LOOP |
| TEST-LOOP-002 | statements/loop | PASS | LOOP with SET |
| TEST-LOOP-003 | statements/loop | PASS | LOOP with CHANGE/COMMIT |
| TEST-CHANGE-001 | statements/change | PASS | Single attribute CHANGE |
| TEST-CHANGE-002 | statements/change | PASS | Multiple attribute CHANGE |
| TEST-COMMIT-001 | statements/commit | PASS | Basic COMMIT |
| TEST-COMMIT-002 | statements/commit | PASS | COMMIT WITH EVENTS |
| TEST-DELETE-001 | statements/delete | PASS | Basic DELETE |
| TEST-ROLLBACK-001 | statements/rollback | FAIL | ROLLBACK (not implemented) |
| TEST-PAGE-001 | statements/navigation | PASS | SHOW PAGE |
| TEST-PAGE-002 | statements/navigation | PASS | CLOSE PAGE |
| TEST-CALL-001 | statements/call | PASS | CALL MICROFLOW |
| TEST-VALIDATION-001 | statements/validation | PASS | VALIDATION FEEDBACK |
| TEST-LOG-001 | statements/log | PASS | LOG INFO |
| TEST-LOG-002 | statements/log | PASS | LOG WARNING |
| TEST-IF-001 | control/if | PASS | Basic IF |
| TEST-IF-002 | control/if | PASS | IF-ELSE |
| TEST-IF-003 | control/if | PASS | Nested IF |
| TEST-ATTR-001 | expressions/attributes | PASS | Read attribute |
| TEST-ATTR-002 | expressions/attributes | PASS | Compare to empty |
| TEST-RESERVED-001 | reserved-words | FAIL | "Item" is reserved |
| TEST-RESERVED-002 | reserved-words | PASS | "Items" is not reserved |
| TEST-RESERVED-003 | reserved-words | PASS | Common names work |

---

## Known Issues

### 1. Reserved Word: `Item`

The word `Item` cannot be used as an entity name. Using `Test.Item` in any context causes a parse error.

**Workaround**: Use `OrderItem`, `LineItem`, `ProductItem`, etc.

### 2. DECLARE AS Not Working

The documented syntax `DECLARE $var AS Module.Entity;` does not parse. The parser rejects the qualified name after `AS`.

**Impact**: Cannot declare entity variables for later assignment.

### 3. CREATE Entity Not Working

The documented syntax `$var = CREATE Module.Entity (...)` does not parse.

**Impact**: Cannot create new entity objects in microflows.

### 4. ROLLBACK Not Implemented

The `ROLLBACK $Entity;` statement is not recognized by the parser.

**Workaround**: Use `CLOSE PAGE;` to discard changes (though this is not semantically equivalent).
