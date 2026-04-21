# MDL Data Types

This document describes the data type system in MDL and how types map to different backends.

## Table of Contents

1. [Primitive Types](#primitive-types)
2. [Complex Types](#complex-types)
3. [Type Syntax](#type-syntax)
4. [Default Values](#default-values)
5. [Type Mapping Table](#type-mapping-table)

---

## Primitive Types

### String

Variable-length text data.

**Syntax:**
```sql
string              -- Default length (200)
string(n)           -- Specific length (1-unlimited)
```

**Parameters:**
- `n` - Maximum length in characters. Use `unlimited` for unlimited length.

**Examples:**
```sql
Name: string(200)
description: string(unlimited)
Code: string(10)
```

**Default value format:**
```sql
default 'text value'
default ''
```

---

### Integer

32-bit signed integer.

**Syntax:**
```sql
integer
```

**Range:** -2,147,483,648 to 2,147,483,647

**Examples:**
```sql
Quantity: integer
Age: integer default 0
Priority: integer not null default 1
```

---

### Long

64-bit signed integer.

**Syntax:**
```sql
long
```

**Range:** -9,223,372,036,854,775,808 to 9,223,372,036,854,775,807

**Examples:**
```sql
FileSize: long
TotalCount: long default 0
```

---

### Decimal

High-precision decimal number.

**Syntax:**
```sql
decimal
```

**Precision:** Up to 20 digits total, with configurable decimal places.

**Examples:**
```sql
Price: decimal
Amount: decimal default 0
TaxRate: decimal default 0.21
```

---

### Boolean

True/false value.

**Syntax:**
```sql
boolean default true|false
```

> **Required:** Boolean attributes must have a DEFAULT value. This is enforced by Mendix Studio Pro.

**Examples:**
```sql
IsActive: boolean default true
Enabled: boolean default true
Deleted: boolean default false
```

---

### DateTime

Date and time with timezone awareness.

**Syntax:**
```sql
datetime
```

**Storage:** UTC timestamp with optional localization.

**Examples:**
```sql
CreatedAt: datetime
ModifiedAt: datetime
ScheduledFor: datetime
```

**Note:** DateTime values include both date and time components. For date-only values, still use DateTime but only populate the date portion.

---

### Date

Date only (no time component).

**Syntax:**
```sql
date
```

**Note:** Internally stored as DateTime in Mendix, but UI only shows date.

**Examples:**
```sql
BirthDate: date
ExpiryDate: date
```

---

### AutoNumber

Auto-incrementing integer, typically used for IDs.

**Syntax:**
```sql
autonumber
```

**Examples:**
```sql
OrderId: autonumber not null unique default 1
CustomerId: autonumber
```

**Notes:**
- AutoNumber attributes automatically get the next value on object creation
- The DEFAULT value specifies the starting number
- Typically combined with NOT NULL and UNIQUE constraints

---

### Binary

Binary data (file contents, images, etc.).

**Syntax:**
```sql
binary
```

**Examples:**
```sql
ProfileImage: binary
Document: binary
Thumbnail: binary
```

**Notes:**
- Binary attributes can store files of any type
- File metadata (name, size, mime type) is stored separately by Mendix
- Maximum size is configurable per attribute

---

## Complex Types

### Enumeration

Reference to an enumeration type.

**Syntax:**
```sql
enumeration(<qualified-name>)
```

**Parameters:**
- `<qualified-name>` - The Module.EnumerationName of the enumeration

**Examples:**
```sql
status: enumeration(Sales.OrderStatus)
Priority: enumeration(Core.Priority) default Core.Priority.Normal
type: enumeration(MyModule.ItemType) not null
```

**Default value format:**
```sql
default Module.EnumName.ValueName
-- or legacy string literal form:
default 'ValueName'
```

The default value is the **name** of the enumeration value (not the caption). The fully qualified form `Module.EnumName.ValueName` is preferred as it is explicit and unambiguous.

---

### HashedString

Securely hashed string (for passwords).

**Syntax:**
```sql
hashedstring
```

**Examples:**
```sql
password: hashedstring
```

**Notes:**
- Values are one-way hashed and cannot be retrieved
- Comparison is done by hashing the input and comparing hashes
- Used primarily for password storage

---

## Type Syntax

### Full Attribute Definition

```sql
[/** documentation */]
<name>: <type> [constraints] [default <value>]
```

### Constraints

| Constraint | Syntax | Description |
|------------|--------|-------------|
| Not Null | `not null` | Value is required |
| Not Null with Error | `not null error 'message'` | Required with custom error |
| Unique | `unique` | Value must be unique |
| Unique with Error | `unique error 'message'` | Unique with custom error |

### Constraint Order

Constraints must appear in this order:
1. `not null [error '...']`
2. `unique [error '...']`
3. `default <value>`

**Examples:**
```sql
-- All constraints
Email: string(200) not null error 'Email is required' unique error 'Email already exists' default ''

-- Some constraints
Name: string(200) not null
Code: string(10) unique
count: integer default 0

-- No constraints
description: string(unlimited)
```

---

## Default Values

### Syntax by Type

| Type | Default Syntax | Examples |
|------|----------------|----------|
| String | `default 'value'` | `default ''`, `default 'Unknown'` |
| Integer | `default n` | `default 0`, `default -1` |
| Long | `default n` | `default 0` |
| Decimal | `default n.n` | `default 0`, `default 0.00`, `default 99.99` |
| Boolean | `default true/false` | `default true`, `default false` |
| AutoNumber | `default n` | `default 1` (starting value) |
| Enumeration | `default Module.Enum.Value` | `default Shop.Status.Active`, `default 'Pending'` |

### No Default

Omit the DEFAULT clause for no default value:
```sql
OptionalField: string(200)
```

---

## Type Mapping Table

### MDL to Backend Type Mapping

| MDL Type | BSON $Type | Go Type (SDK) | Model API Type |
|----------|------------|---------------|----------------|
| String | `DomainModels$StringAttributeType` | `*StringAttributeType` | TBD |
| String(n) | `DomainModels$StringAttributeType` + Length | `*StringAttributeType{length: n}` | TBD |
| Integer | `DomainModels$IntegerAttributeType` | `*IntegerAttributeType` | TBD |
| Long | `DomainModels$LongAttributeType` | `*LongAttributeType` | TBD |
| Decimal | `DomainModels$DecimalAttributeType` | `*DecimalAttributeType` | TBD |
| Boolean | `DomainModels$BooleanAttributeType` | `*BooleanAttributeType` | TBD |
| DateTime | `DomainModels$DateTimeAttributeType` | `*DateTimeAttributeType` | TBD |
| Date | `DomainModels$DateTimeAttributeType` | `*DateTimeAttributeType` | TBD |
| AutoNumber | `DomainModels$AutoNumberAttributeType` | `*AutoNumberAttributeType` | TBD |
| Binary | `DomainModels$BinaryAttributeType` | `*BinaryAttributeType` | TBD |
| Enumeration | `DomainModels$EnumerationAttributeType` | `*EnumerationAttributeType` | TBD |
| HashedString | `DomainModels$HashedStringAttributeType` | `*HashedStringAttributeType` | TBD |

### Default Value Mapping

| MDL Default | BSON Structure | Go Structure |
|-------------|----------------|--------------|
| `default 'text'` | `value: {$type: "DomainModels$StoredValue", DefaultValue: "text"}` | `value: &AttributeValue{DefaultValue: "text"}` |
| `default 123` | `value: {$type: "DomainModels$StoredValue", DefaultValue: "123"}` | `value: &AttributeValue{DefaultValue: "123"}` |
| `default true` | `value: {$type: "DomainModels$StoredValue", DefaultValue: "true"}` | `value: &AttributeValue{DefaultValue: "true"}` |
| (calculated) | `value: {$type: "DomainModels$CalculatedValue", microflow: <id>}` | `value: &AttributeValue{type: "CalculatedValue", MicroflowID: id}` |

---

## Type Compatibility

### Implicit Conversions

MDL does not perform implicit type conversions. Types must match exactly.

### Attribute Type Changes

When modifying an entity with `create or modify`, attribute types cannot be changed if:
- The attribute contains data
- The new type is incompatible

Compatible type changes:
- `string(100)` to `string(200)` - Increasing length
- `integer` to `long` - Widening numeric type

Incompatible type changes:
- `string` to `integer`
- `boolean` to `string`
- Any type to `autonumber` (if not empty)

---

## Examples

### Complete Entity with All Types

```sql
/** Example entity demonstrating all data types */
@position(100, 100)
create persistent entity Demo.AllTypes (
  /** Auto-generated ID */
  Id: autonumber not null unique default 1,

  /** Short text field */
  Code: string(10) not null unique,

  /** Standard text field */
  Name: string(200) not null,

  /** Long text field */
  description: string(unlimited),

  /** Integer counter */
  Counter: integer default 0,

  /** Large number */
  BigNumber: long,

  /** Money amount */
  Amount: decimal default 0.00,

  /** Flag field */
  IsActive: boolean default true,

  /** Timestamp */
  CreatedAt: datetime,

  /** Date only */
  BirthDate: date,

  /** File attachment */
  Attachment: binary,

  /** Status from enumeration */
  status: enumeration(Demo.Status) default 'Active'
)
index (Code)
index (Name, CreatedAt desc);
/
```
