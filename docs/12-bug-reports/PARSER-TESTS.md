# MDL Parser Test Cases

Test cases for `mxcli check` syntax validation. Each test case includes MDL input and expected result.

## Test Format

```
Test ID: unique identifier
Category: Feature category
description: What is being tested
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
description: declare integer variable with initial value
expected: PASS
```

```mdl
create microflow Test.IntegerDecl()
returns boolean
begin
    declare $x integer = 0;
    return true;
end;
/
```

### TEST-VAR-002: String declaration with initialization

```yaml
id: TEST-VAR-002
category: variables/primitives
description: declare string variable with initial value
expected: PASS
```

```mdl
create microflow Test.StringDecl()
returns boolean
begin
    declare $s string = 'hello';
    return true;
end;
/
```

### TEST-VAR-003: Boolean declaration with initialization

```yaml
id: TEST-VAR-003
category: variables/primitives
description: declare boolean variable with initial value
expected: PASS
```

```mdl
create microflow Test.BooleanDecl()
returns boolean
begin
    declare $b boolean = true;
    return true;
end;
/
```

### TEST-VAR-004: Decimal declaration with initialization

```yaml
id: TEST-VAR-004
category: variables/primitives
description: declare decimal variable with initial value
expected: PASS
```

```mdl
create microflow Test.DecimalDecl()
returns boolean
begin
    declare $d decimal = 0.0;
    return true;
end;
/
```

### TEST-VAR-005: Multiple primitive declarations

```yaml
id: TEST-VAR-005
category: variables/primitives
description: declare multiple primitive variables in one microflow
expected: PASS
```

```mdl
create microflow Test.MultiplePrimitives()
returns boolean
begin
    declare $x integer = 0;
    declare $s string = 'hello';
    declare $b boolean = true;
    declare $d decimal = 0.0;
    return true;
end;
/
```

---

## Category: List Variable Declarations

### TEST-LIST-001: List declaration with qualified entity name

```yaml
id: TEST-list-001
category: variables/lists
description: declare list variable with Module.Entity syntax
expected: PASS
```

```mdl
create microflow Test.ListDecl()
returns boolean
begin
    declare $Orders list of Test.Order = empty;
    return true;
end;
/
```

### TEST-LIST-002: List declaration with reserved word entity name

```yaml
id: TEST-list-002
category: variables/lists
description: declare list with entity name "item" (reserved word)
expected: FAIL
notes: "item" is a reserved word in the parser, causing syntax error
```

```mdl
create microflow Test.ListDeclItem()
returns boolean
begin
    declare $Items list of Test.Item = empty;
    return true;
end;
/
```

### TEST-LIST-003: List declaration with compound entity name containing reserved word

```yaml
id: TEST-list-003
category: variables/lists
description: declare list with entity name "OrderItem" (contains reserved word but is valid)
expected: PASS
```

```mdl
create microflow Test.ListDeclOrderItem()
returns boolean
begin
    declare $Items list of Test.OrderItem = empty;
    return true;
end;
/
```

### TEST-LIST-004: List declaration with "LineItem" entity name

```yaml
id: TEST-list-004
category: variables/lists
description: declare list with entity name "LineItem"
expected: PASS
```

```mdl
create microflow Test.ListDeclLineItem()
returns boolean
begin
    declare $Items list of Test.LineItem = empty;
    return true;
end;
/
```

---

## Category: Entity Variable Declarations

### TEST-ENTITY-001: Entity declaration with AS and qualified name

```yaml
id: TEST-entity-001
category: variables/entities
description: declare entity variable using as keyword with Module.Entity
expected: FAIL
notes: Parser does not accept qualified name after as keyword
```

```mdl
create microflow Test.EntityDecl()
returns boolean
begin
    declare $Order as Test.Order;
    return true;
end;
/
```

### TEST-ENTITY-002: Entity declaration with AS and simple name

```yaml
id: TEST-entity-002
category: variables/entities
description: declare entity variable using as keyword with simple name (no module)
expected: FAIL
notes: Parser does not accept any identifier after as keyword
```

```mdl
create microflow Test.EntityDeclSimple()
returns boolean
begin
    declare $Order as Order;
    return true;
end;
/
```

---

## Category: Entity Creation

### TEST-CREATE-001: Create entity with assignment

```yaml
id: TEST-create-001
category: statements/create
description: create entity and assign to variable
expected: FAIL
notes: entity creation syntax not supported in parser
```

```mdl
create microflow Test.CreateEntity()
returns boolean
begin
    declare $Order as Test.Order;
    $Order = create Test.Order (
        Name = 'test'
    );
    return true;
end;
/
```

---

## Category: RETRIEVE Statement

### TEST-RETRIEVE-001: Basic RETRIEVE from entity

```yaml
id: TEST-retrieve-001
category: statements/retrieve
description: retrieve all objects of an entity type
expected: PASS
```

```mdl
create microflow Test.RetrieveBasic()
returns boolean
begin
    declare $Orders list of Test.Order = empty;
    retrieve $Orders from Test.Order;
    return true;
end;
/
```

### TEST-RETRIEVE-002: RETRIEVE with WHERE clause using association

```yaml
id: TEST-retrieve-002
category: statements/retrieve
description: retrieve with association filter
expected: PASS
```

```mdl
create microflow Test.RetrieveWhere($Customer: Test.Customer)
returns boolean
begin
    declare $Orders list of Test.Order = empty;
    retrieve $Orders from Test.Order
        where Test.Order_Customer = $Customer;
    return true;
end;
/
```

---

## Category: LOOP Statement

### TEST-LOOP-001: Basic LOOP over list

```yaml
id: TEST-loop-001
category: statements/loop
description: Iterate over a list with loop
expected: PASS
```

```mdl
create microflow Test.LoopBasic()
returns boolean
begin
    declare $Orders list of Test.Order = empty;
    loop $Order in $Orders
    begin
        declare $x integer = 1;
    end loop;
    return true;
end;
/
```

### TEST-LOOP-002: LOOP with SET inside

```yaml
id: TEST-loop-002
category: statements/loop
description: Iterate and accumulate with set
expected: PASS
```

```mdl
create microflow Test.LoopWithSet()
returns boolean
begin
    declare $Total decimal = 0;
    declare $Orders list of Test.Order = empty;

    retrieve $Orders from Test.Order;

    loop $Order in $Orders
    begin
        set $Total = $Total + $Order/Amount;
    end loop;

    return true;
end;
/
```

### TEST-LOOP-003: LOOP with CHANGE and COMMIT inside

```yaml
id: TEST-loop-003
category: statements/loop
description: Iterate and modify objects
expected: PASS
```

```mdl
create microflow Test.LoopWithChange($Customer: Test.Customer)
returns boolean
begin
    declare $Orders list of Test.Order = empty;

    retrieve $Orders from Test.Order
        where Test.Order_Customer = $Customer;

    loop $Order in $Orders
    begin
        change $Order (status = 'Processed');
        commit $Order;
    end loop;

    return true;
end;
/
```

---

## Category: CHANGE Statement

### TEST-CHANGE-001: Basic CHANGE with single attribute

```yaml
id: TEST-change-001
category: statements/change
description: change single attribute on entity
expected: PASS
```

```mdl
create microflow Test.ChangeBasic($Order: Test.Order)
returns boolean
begin
    change $Order (status = 'Complete');
    return true;
end;
/
```

### TEST-CHANGE-002: CHANGE with multiple attributes

```yaml
id: TEST-change-002
category: statements/change
description: change multiple attributes on entity
expected: PASS
```

```mdl
create microflow Test.ChangeMultiple($Order: Test.Order)
returns boolean
begin
    change $Order (
        status = 'Complete',
        ProcessedDate = [%CurrentDateTime%]
    );
    return true;
end;
/
```

---

## Category: COMMIT Statement

### TEST-COMMIT-001: Basic COMMIT

```yaml
id: TEST-commit-001
category: statements/commit
description: commit entity to database
expected: PASS
```

```mdl
create microflow Test.CommitBasic($Order: Test.Order)
returns boolean
begin
    commit $Order;
    return true;
end;
/
```

### TEST-COMMIT-002: COMMIT WITH EVENTS

```yaml
id: TEST-commit-002
category: statements/commit
description: commit entity with event handlers
expected: PASS
```

```mdl
create microflow Test.CommitWithEvents($Order: Test.Order)
returns boolean
begin
    commit $Order with events;
    return true;
end;
/
```

---

## Category: DELETE Statement

### TEST-DELETE-001: Basic DELETE

```yaml
id: TEST-delete-001
category: statements/delete
description: delete entity from database
expected: PASS
```

```mdl
create microflow Test.DeleteBasic($Order: Test.Order)
returns boolean
begin
    delete $Order;
    return true;
end;
/
```

---

## Category: ROLLBACK Statement

### TEST-ROLLBACK-001: Basic ROLLBACK

```yaml
id: TEST-rollback-001
category: statements/rollback
description: rollback changes to entity
expected: FAIL
notes: rollback is documented but not implemented in parser
```

```mdl
create microflow Test.RollbackBasic($Order: Test.Order)
returns boolean
begin
    rollback $Order;
    return true;
end;
/
```

---

## Category: Page Navigation

### TEST-PAGE-001: SHOW PAGE with parameter

```yaml
id: TEST-page-001
category: statements/navigation
description: Navigate to page with parameter
expected: PASS
```

```mdl
create microflow Test.ShowPage($Order: Test.Order)
returns boolean
begin
    show page Test.Order_Edit ($Order = $Order);
    return true;
end;
/
```

### TEST-PAGE-002: CLOSE PAGE

```yaml
id: TEST-page-002
category: statements/navigation
description: close current page
expected: PASS
```

```mdl
create microflow Test.ClosePage()
returns boolean
begin
    close page;
    return true;
end;
/
```

---

## Category: Microflow Calls

### TEST-CALL-001: CALL MICROFLOW with parameter

```yaml
id: TEST-call-001
category: statements/call
description: call another microflow with parameter
expected: PASS
```

```mdl
create microflow Test.CallMicroflow($Order: Test.Order)
returns boolean
begin
    declare $Result boolean = false;
    $Result = call microflow Test.ValidateOrder($Order = $Order);
    return $Result;
end;
/
```

---

## Category: Validation Feedback

### TEST-VALIDATION-001: Basic validation feedback

```yaml
id: TEST-validation-001
category: statements/validation
description: show validation feedback on attribute
expected: PASS
```

```mdl
create microflow Test.ValidationFeedback($Order: Test.Order)
returns boolean
begin
    declare $IsValid boolean = true;

    if $Order/Name = empty then
        set $IsValid = false;
        validation feedback $Order/Name message 'Name is required';
    end if;

    return $IsValid;
end;
/
```

---

## Category: Logging

### TEST-LOG-001: LOG INFO

```yaml
id: TEST-log-001
category: statements/log
description: log info message
expected: PASS
```

```mdl
create microflow Test.LogInfo()
returns boolean
begin
    log info node 'TestNode' 'This is an info message';
    return true;
end;
/
```

### TEST-LOG-002: LOG WARNING

```yaml
id: TEST-log-002
category: statements/log
description: log warning message
expected: PASS
```

```mdl
create microflow Test.LogWarning()
returns boolean
begin
    log warning node 'TestNode' 'This is a warning message';
    return true;
end;
/
```

---

## Category: Control Flow

### TEST-IF-001: Basic IF statement

```yaml
id: TEST-if-001
category: control/if
description: Simple if condition
expected: PASS
```

```mdl
create microflow Test.IfBasic($value: integer)
returns boolean
begin
    declare $Result boolean = false;

    if $value > 10 then
        set $Result = true;
    end if;

    return $Result;
end;
/
```

### TEST-IF-002: IF-ELSE statement

```yaml
id: TEST-if-002
category: control/if
description: if with else branch
expected: PASS
```

```mdl
create microflow Test.IfElse($value: integer)
returns boolean
begin
    declare $Result string = '';

    if $value > 10 then
        set $Result = 'high';
    else
        set $Result = 'low';
    end if;

    return true;
end;
/
```

### TEST-IF-003: Nested IF statements

```yaml
id: TEST-if-003
category: control/if
description: Nested if conditions
expected: PASS
```

```mdl
create microflow Test.IfNested($value: integer)
returns string
begin
    declare $Result string = '';

    if $value > 100 then
        set $Result = 'very high';
    else
        if $value > 10 then
            set $Result = 'high';
        else
            set $Result = 'low';
        end if;
    end if;

    return $Result;
end;
/
```

---

## Category: Attribute Access

### TEST-ATTR-001: Read attribute from parameter

```yaml
id: TEST-attr-001
category: expressions/attributes
description: access attribute using slash notation
expected: PASS
```

```mdl
create microflow Test.ReadAttribute($Order: Test.Order)
returns boolean
begin
    declare $Name string = '';
    set $Name = $Order/Name;
    return true;
end;
/
```

### TEST-ATTR-002: Compare attribute to empty

```yaml
id: TEST-attr-002
category: expressions/attributes
description: check if attribute is empty
expected: PASS
```

```mdl
create microflow Test.CheckEmpty($Order: Test.Order)
returns boolean
begin
    if $Order/Name = empty then
        return false;
    end if;
    return true;
end;
/
```

---

## Category: Reserved Words

### TEST-RESERVED-001: Entity name "Item"

```yaml
id: TEST-RESERVED-001
category: reserved-words
description: Using "item" as entity name in qualified reference
expected: FAIL
notes: "item" is a reserved word
```

```mdl
create microflow Test.ReservedItem()
returns boolean
begin
    declare $Items list of Test.Item = empty;
    return true;
end;
/
```

### TEST-RESERVED-002: Entity name "Items" (plural)

```yaml
id: TEST-RESERVED-002
category: reserved-words
description: Using "Items" as entity name (plural of reserved word)
expected: PASS
notes: "Items" is not reserved, only "item"
```

```mdl
create microflow Test.NotReservedItems()
returns boolean
begin
    declare $list list of Test.Items = empty;
    return true;
end;
/
```

### TEST-RESERVED-003: Various common entity names

```yaml
id: TEST-RESERVED-003
category: reserved-words
description: Test common entity names for reserved word conflicts
expected: PASS
notes: Names like Order, Customer, Product, Name, type, value, status should all work
```

```mdl
create microflow Test.CommonNames()
returns boolean
begin
    declare $Orders list of Test.Order = empty;
    declare $Customers list of Test.Customer = empty;
    declare $Products list of Test.Product = empty;
    return true;
end;
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

### 1. Reserved Word: `item`

The word `item` cannot be used as an entity name. Using `Test.Item` in any context causes a parse error.

**Workaround**: Use `OrderItem`, `LineItem`, `ProductItem`, etc.

### 2. DECLARE AS Not Working

The documented syntax `declare $var as Module.Entity;` does not parse. The parser rejects the qualified name after `as`.

**Impact**: Cannot declare entity variables for later assignment.

### 3. CREATE Entity Not Working

The documented syntax `$var = create Module.Entity (...)` does not parse.

**Impact**: Cannot create new entity objects in microflows.

### 4. ROLLBACK Not Implemented

The `rollback $entity;` statement is not recognized by the parser.

**Workaround**: Use `close page;` to discard changes (though this is not semantically equivalent).
