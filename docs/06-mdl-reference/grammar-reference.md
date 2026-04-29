# MDL Grammar Reference

> Auto-generated from ANTLR4 grammar on 2026-01-16

This document provides a complete reference for the MDL (Mendix Definition Language) syntax.
Each grammar rule is documented with its syntax, description, and examples.

## Table of Contents

- [Statements](#statements)
  - [program](#program)
  - [statement](#statement)
  - [createEntityStatement](#createentitystatement)
  - [createMicroflowStatement](#createmicroflowstatement)
  - [createPageStatement](#createpagestatement)
- [Entity Definitions](#entity-definitions)
  - [attributeDefinition](#attributedefinition)
- [Microflow Statements](#microflow-statements)
- [Page Definitions](#page-definitions)
- [OQL Queries](#oql-queries)
  - [oqlQuery](#oqlquery)
- [Expressions](#expressions)
  - [literal](#literal)
- [Other Rules](#other-rules)
  - [dataType](#datatype)
  - [qualifiedName](#qualifiedname)
  - [docComment](#doccomment)
  - [annotation](#annotation)
  - [keyword](#keyword)

## Statements

### program

Entry point: a program is a sequence of statements

**Syntax:**

```ebnf
program
    : statement* EOF
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "statement*" as s1
    [*] --> s1
    state "EOF" as s2
    s1 --> s2
    s2 --> [*]
```

---

### statement

A statement can be DDL, DQL, or utility

**Syntax:**

```ebnf
statement
    : docComment? (ddlStatement | dqlStatement | utilityStatement) SEMICOLON? SLASH?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "docComment?" as s1
    [*] --> s1
    state "(ddlStatement | dqlStatement | utilityStatement)" as s2
    s1 --> s2
    state "SEMICOLON?" as s3
    s2 --> s3
    state "SLASH?" as s4
    s3 --> s4
    s4 --> [*]
```

---

### ddlStatement

**Syntax:**

```ebnf
ddlStatement
    : createStatement
    | | alterStatement
    | | dropStatement
    | | renameStatement
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "createStatement" as s1
    [*] --> s1
    s1 --> [*]
    state "alterStatement" as s2
    [*] --> s2
    s2 --> [*]
    state "dropStatement" as s3
    [*] --> s3
    s3 --> [*]
    state "renameStatement" as s4
    [*] --> s4
    s4 --> [*]
```

---

### createStatement

**Syntax:**

```ebnf
createStatement
    : docComment? annotation*
    | create (or (modify | replace))?
    | ( createEntityStatement
    | | createAssociationStatement
    | | createModuleStatement
    | | createMicroflowStatement
    | | createNanoflowStatement
    | | createPageStatement
    | | createSnippetStatement
    | | createEnumerationStatement
    | | createValidationRuleStatement
    | | createNotebookStatement
    | | createDatabaseConnectionStatement
    | | createConstantStatement
    | | createRestClientStatement
    | | createIndexStatement
    | )
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "docComment?" as s1
    [*] --> s1
    state "annotation*" as s2
    s1 --> s2
    s2 --> [*]
    state "create" as s3
    [*] --> s3
    state "(or (modify | replace))?" as s4
    s3 --> s4
    s4 --> [*]
    state "( createEntityStatement" as s5
    [*] --> s5
    s5 --> [*]
    state "createAssociationStatement" as s6
    [*] --> s6
    s6 --> [*]
    state "createModuleStatement" as s7
    [*] --> s7
    s7 --> [*]
    state "createMicroflowStatement" as s8
    [*] --> s8
    s8 --> [*]
    state "createPageStatement" as s9
    [*] --> s9
    s9 --> [*]
    state "createSnippetStatement" as s10
    [*] --> s10
    s10 --> [*]
    state "createEnumerationStatement" as s11
    [*] --> s11
    s11 --> [*]
    state "createValidationRuleStatement" as s12
    [*] --> s12
    s12 --> [*]
    state "createNotebookStatement" as s13
    [*] --> s13
    s13 --> [*]
    state "createDatabaseConnectionStatement" as s14
    [*] --> s14
    s14 --> [*]
    state "createConstantStatement" as s15
    [*] --> s15
    s15 --> [*]
    state "createRestClientStatement" as s16
    [*] --> s16
    s16 --> [*]
    state "createIndexStatement" as s17
    [*] --> s17
    s17 --> [*]
    state ")" as s18
    [*] --> s18
    s18 --> [*]
```

---

### alterStatement

**Syntax:**

```ebnf
alterStatement
    : alter entity qualifiedName alterEntityAction+
    | | alter association qualifiedName alterAssociationAction+
    | | alter enumeration qualifiedName alterEnumerationAction+
    | | alter notebook qualifiedName alterNotebookAction+
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "alter" as s1
    [*] --> s1
    state "entity" as s2
    s1 --> s2
    state "qualifiedName" as s3
    s2 --> s3
    state "alterEntityAction+" as s4
    s3 --> s4
    s4 --> [*]
    state "alter" as s5
    [*] --> s5
    state "association" as s6
    s5 --> s6
    state "qualifiedName" as s7
    s6 --> s7
    state "alterAssociationAction+" as s8
    s7 --> s8
    s8 --> [*]
    state "alter" as s9
    [*] --> s9
    state "enumeration" as s10
    s9 --> s10
    state "qualifiedName" as s11
    s10 --> s11
    state "alterEnumerationAction+" as s12
    s11 --> s12
    s12 --> [*]
    state "alter" as s13
    [*] --> s13
    state "notebook" as s14
    s13 --> s14
    state "qualifiedName" as s15
    s14 --> s15
    state "alterNotebookAction+" as s16
    s15 --> s16
    s16 --> [*]
```

---

### dropStatement

**Syntax:**

```ebnf
dropStatement
    : drop entity qualifiedName
    | | drop association qualifiedName
    | | drop enumeration qualifiedName
    | | drop microflow qualifiedName
    | | drop nanoflow qualifiedName
    | | drop page qualifiedName
    | | drop snippet qualifiedName
    | | drop module qualifiedName
    | | drop notebook qualifiedName
    | | drop index qualifiedName on qualifiedName
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "drop" as s1
    [*] --> s1
    state "entity" as s2
    s1 --> s2
    state "qualifiedName" as s3
    s2 --> s3
    s3 --> [*]
    state "drop" as s4
    [*] --> s4
    state "association" as s5
    s4 --> s5
    state "qualifiedName" as s6
    s5 --> s6
    s6 --> [*]
    state "drop" as s7
    [*] --> s7
    state "enumeration" as s8
    s7 --> s8
    state "qualifiedName" as s9
    s8 --> s9
    s9 --> [*]
    state "drop" as s10
    [*] --> s10
    state "microflow" as s11
    s10 --> s11
    state "qualifiedName" as s12
    s11 --> s12
    s12 --> [*]
    state "drop" as s13
    [*] --> s13
    state "nanoflow" as s14
    s13 --> s14
    state "qualifiedName" as s15
    s14 --> s15
    s15 --> [*]
    state "drop" as s16
    [*] --> s16
    state "page" as s17
    s16 --> s17
    state "qualifiedName" as s18
    s17 --> s18
    s18 --> [*]
    state "drop" as s19
    [*] --> s19
    state "snippet" as s20
    s19 --> s20
    state "qualifiedName" as s21
    s20 --> s21
    s21 --> [*]
    state "drop" as s22
    [*] --> s22
    state "module" as s23
    s22 --> s23
    state "qualifiedName" as s24
    s23 --> s24
    s24 --> [*]
    state "drop" as s25
    [*] --> s25
    state "notebook" as s26
    s25 --> s26
    state "qualifiedName" as s27
    s26 --> s27
    s27 --> [*]
    state "drop" as s28
    [*] --> s28
    state "index" as s29
    s28 --> s29
    state "qualifiedName" as s30
    s29 --> s30
    state "on" as s31
    s30 --> s31
    state "qualifiedName" as s32
    s31 --> s32
    s32 --> [*]
```

---

### renameStatement

**Syntax:**

```ebnf
renameStatement
    : rename entity qualifiedName to IDENTIFIER
    | | rename module IDENTIFIER to IDENTIFIER
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "rename" as s1
    [*] --> s1
    state "entity" as s2
    s1 --> s2
    state "qualifiedName" as s3
    s2 --> s3
    state "to" as s4
    s3 --> s4
    state "IDENTIFIER" as s5
    s4 --> s5
    s5 --> [*]
    state "rename" as s6
    [*] --> s6
    state "module" as s7
    s6 --> s7
    state "IDENTIFIER" as s8
    s7 --> s8
    state "to" as s9
    s8 --> s9
    state "IDENTIFIER" as s10
    s9 --> s10
    s10 --> [*]
```

---

### createEntityStatement

Creates a new entity in the domain model.  Entities can be persistent (stored in database), non-persistent (in-memory only), view (based on OQL query), or external (from external data source).

**Syntax:**

```ebnf
createEntityStatement
    : persistent entity qualifiedName entityBody?
    | | NON_PERSISTENT entity qualifiedName entityBody?
    | | view entity qualifiedName entityBody? as LPAREN? oqlQuery RPAREN?  // Parentheses optional
    | | external entity qualifiedName entityBody?
    | | entity qualifiedName entityBody?  // default to persistent
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "persistent" as s1
    [*] --> s1
    state "entity" as s2
    s1 --> s2
    state "qualifiedName" as s3
    s2 --> s3
    state "entityBody?" as s4
    s3 --> s4
    s4 --> [*]
    state "NON_PERSISTENT" as s5
    [*] --> s5
    state "entity" as s6
    s5 --> s6
    state "qualifiedName" as s7
    s6 --> s7
    state "entityBody?" as s8
    s7 --> s8
    s8 --> [*]
    state "view" as s9
    [*] --> s9
    state "entity" as s10
    s9 --> s10
    state "qualifiedName" as s11
    s10 --> s11
    state "entityBody?" as s12
    s11 --> s12
    state "as" as s13
    s12 --> s13
    state "LPAREN?" as s14
    s13 --> s14
    state "oqlQuery" as s15
    s14 --> s15
    state "RPAREN?" as s16
    s15 --> s16
    s16 --> [*]
    state "external" as s17
    [*] --> s17
    state "entity" as s18
    s17 --> s18
    state "qualifiedName" as s19
    s18 --> s19
    state "entityBody?" as s20
    s19 --> s20
    s20 --> [*]
    state "entity" as s21
    [*] --> s21
    state "qualifiedName" as s22
    s21 --> s22
    state "entityBody?" as s23
    s22 --> s23
    s23 --> [*]
```

**Examples:**

*Persistent entity with attributes:*

```sql
create persistent entity MyModule.Customer (
  Name: string(100) not null,
  Email: string(200) unique,
  Age: integer,
  Active: boolean default true
);
```

*Non-persistent entity for search filters:*

```sql
create non-persistent entity MyModule.SearchFilter (
query: string,
MaxResults: integer default 100,
IncludeArchived: boolean default false
);
```

*View entity with OQL query:*

```sql
create view entity MyModule.ActiveCustomers (
CustomerId: integer,
CustomerName: string(100)
) as
select c.Id as CustomerId, c.Name as CustomerName
from MyModule.Customer as c
where c.Active = true;
```

*Entity with index:*

```sql
create persistent entity MyModule.Order (
OrderNumber: string(50) not null,
CustomerRef: MyModule.Customer
)
index (OrderNumber);
```

**See also:** [attributeDefinition for attribute syntax](#attributedefinition-for-attribute-syntax), [dataType for supported data types](#datatype-for-supported-data-types), [oqlQuery for view entity queries](#oqlquery-for-view-entity-queries)

---

### createAssociationStatement

**Syntax:**

```ebnf
createAssociationStatement
    : association qualifiedName
    | from qualifiedName
    | to qualifiedName
    | associationOptions?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "association" as s1
    [*] --> s1
    state "qualifiedName" as s2
    s1 --> s2
    s2 --> [*]
    state "from" as s3
    [*] --> s3
    state "qualifiedName" as s4
    s3 --> s4
    s4 --> [*]
    state "to" as s5
    [*] --> s5
    state "qualifiedName" as s6
    s5 --> s6
    s6 --> [*]
    state "associationOptions?" as s7
    [*] --> s7
    s7 --> [*]
```

---

### createModuleStatement

**Syntax:**

```ebnf
createModuleStatement
    : module IDENTIFIER moduleOptions?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "module" as s1
    [*] --> s1
    state "IDENTIFIER" as s2
    s1 --> s2
    state "moduleOptions?" as s3
    s2 --> s3
    s3 --> [*]
```

---

### createEnumerationStatement

**Syntax:**

```ebnf
createEnumerationStatement
    : enumeration qualifiedName
    | LPAREN enumerationValueList RPAREN
    | enumerationOptions?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "enumeration" as s1
    [*] --> s1
    state "qualifiedName" as s2
    s1 --> s2
    s2 --> [*]
    state "LPAREN" as s3
    [*] --> s3
    state "enumerationValueList" as s4
    s3 --> s4
    state "RPAREN" as s5
    s4 --> s5
    s5 --> [*]
    state "enumerationOptions?" as s6
    [*] --> s6
    s6 --> [*]
```

---

### createValidationRuleStatement

**Syntax:**

```ebnf
createValidationRuleStatement
    : validation rule qualifiedName
    | for qualifiedName
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "validation" as s1
    [*] --> s1
    state "rule" as s2
    s1 --> s2
    state "qualifiedName" as s3
    s2 --> s3
    s3 --> [*]
    state "for" as s4
    [*] --> s4
    state "qualifiedName" as s5
    s4 --> s5
    s5 --> [*]
```

---

### createMicroflowStatement

Creates a new microflow with parameters, return type, and activity body.  Microflows are server-side logic that can include database operations, integrations, and complex business rules.

**Syntax:**

```ebnf
createMicroflowStatement
    : microflow qualifiedName
    | LPAREN microflowParameterList? RPAREN
    | microflowReturnType?
    | microflowOptions?
    | begin microflowBody end SEMICOLON? SLASH?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "microflow" as s1
    [*] --> s1
    state "qualifiedName" as s2
    s1 --> s2
    s2 --> [*]
    state "LPAREN" as s3
    [*] --> s3
    state "microflowParameterList?" as s4
    s3 --> s4
    state "RPAREN" as s5
    s4 --> s5
    s5 --> [*]
    state "microflowReturnType?" as s6
    [*] --> s6
    s6 --> [*]
    state "microflowOptions?" as s7
    [*] --> s7
    s7 --> [*]
    state "begin" as s8
    [*] --> s8
    state "microflowBody" as s9
    s8 --> s9
    state "end" as s10
    s9 --> s10
    state "SEMICOLON?" as s11
    s10 --> s11
    state "SLASH?" as s12
    s11 --> s12
    s12 --> [*]
```

**Examples:**

*Simple microflow returning a string:*

```sql
create microflow MyModule.GetGreeting ($Name: string) returns string
begin
return 'Hello, ' + $Name + '!';
end;
```

*Microflow with entity parameter and database operations:*

```sql
create microflow MyModule.DeactivateCustomer ($Customer: MyModule.Customer) returns boolean
begin
$Customer.Active = false;
commit $Customer;
return true;
end;
```

*Microflow with RETRIEVE and iteration:*

```sql
create microflow MyModule.CountActiveOrders () returns integer
begin
declare $Orders list of MyModule.Order;
$Orders = retrieve MyModule.Order where Active = true;
return length($Orders);
end;
```

*Microflow calling another microflow:*

```sql
create microflow MyModule.ProcessOrder ($Order: MyModule.Order) returns boolean
begin
$Result = call microflow MyModule.ValidateOrder (Order = $Order);
if $Result then
commit $Order;
return true;
end if;
return false;
end;
```

**See also:** [microflowBody for available activities](#microflowbody-for-available-activities), [microflowParameter for parameter syntax](#microflowparameter-for-parameter-syntax)

---

### microflowStatement

**Syntax:**

```ebnf
microflowStatement
    : declareStatement SEMICOLON?
    | | setStatement SEMICOLON?
    | | createObjectStatement SEMICOLON?
    | | changeObjectStatement SEMICOLON?
    | | commitStatement SEMICOLON?
    | | deleteObjectStatement SEMICOLON?
    | | retrieveStatement SEMICOLON?
    | | ifStatement SEMICOLON?
    | | loopStatement SEMICOLON?
    | | whileStatement SEMICOLON?
    | | continueStatement SEMICOLON?
    | | breakStatement SEMICOLON?
    | | returnStatement SEMICOLON?
    | | logStatement SEMICOLON?
    | | callMicroflowStatement SEMICOLON?
    | | callJavaActionStatement SEMICOLON?
    | | showPageStatement SEMICOLON?
    | | closePageStatement SEMICOLON?
    | | throwStatement SEMICOLON?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "declareStatement" as s1
    [*] --> s1
    state "SEMICOLON?" as s2
    s1 --> s2
    s2 --> [*]
    state "setStatement" as s3
    [*] --> s3
    state "SEMICOLON?" as s4
    s3 --> s4
    s4 --> [*]
    state "createObjectStatement" as s5
    [*] --> s5
    state "SEMICOLON?" as s6
    s5 --> s6
    s6 --> [*]
    state "changeObjectStatement" as s7
    [*] --> s7
    state "SEMICOLON?" as s8
    s7 --> s8
    s8 --> [*]
    state "commitStatement" as s9
    [*] --> s9
    state "SEMICOLON?" as s10
    s9 --> s10
    s10 --> [*]
    state "deleteObjectStatement" as s11
    [*] --> s11
    state "SEMICOLON?" as s12
    s11 --> s12
    s12 --> [*]
    state "retrieveStatement" as s13
    [*] --> s13
    state "SEMICOLON?" as s14
    s13 --> s14
    s14 --> [*]
    state "ifStatement" as s15
    [*] --> s15
    state "SEMICOLON?" as s16
    s15 --> s16
    s16 --> [*]
    state "loopStatement" as s17
    [*] --> s17
    state "SEMICOLON?" as s18
    s17 --> s18
    s18 --> [*]
    state "whileStatement" as s19
    [*] --> s19
    state "SEMICOLON?" as s20
    s19 --> s20
    s20 --> [*]
    state "continueStatement" as s21
    [*] --> s21
    state "SEMICOLON?" as s22
    s21 --> s22
    s22 --> [*]
    state "breakStatement" as s23
    [*] --> s23
    state "SEMICOLON?" as s24
    s23 --> s24
    s24 --> [*]
    state "returnStatement" as s25
    [*] --> s25
    state "SEMICOLON?" as s26
    s25 --> s26
    s26 --> [*]
    state "logStatement" as s27
    [*] --> s27
    state "SEMICOLON?" as s28
    s27 --> s28
    s28 --> [*]
    state "callMicroflowStatement" as s29
    [*] --> s29
    state "SEMICOLON?" as s30
    s29 --> s30
    s30 --> [*]
    state "callJavaActionStatement" as s31
    [*] --> s31
    state "SEMICOLON?" as s32
    s31 --> s32
    s32 --> [*]
    state "showPageStatement" as s33
    [*] --> s33
    state "SEMICOLON?" as s34
    s33 --> s34
    s34 --> [*]
    state "closePageStatement" as s35
    [*] --> s35
    state "SEMICOLON?" as s36
    s35 --> s36
    s36 --> [*]
    state "throwStatement" as s37
    [*] --> s37
    state "SEMICOLON?" as s38
    s37 --> s38
    s38 --> [*]
```

---

### declareStatement

**Syntax:**

```ebnf
declareStatement
    : declare VARIABLE dataType (equals expression)?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "declare" as s1
    [*] --> s1
    state "VARIABLE" as s2
    s1 --> s2
    state "dataType" as s3
    s2 --> s3
    state "(equals expression)?" as s4
    s3 --> s4
    s4 --> [*]
```

---

### setStatement

**Syntax:**

```ebnf
setStatement
    : set (VARIABLE | attributePath) equals expression
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "set" as s1
    [*] --> s1
    state "(VARIABLE | attributePath)" as s2
    s1 --> s2
    state "equals" as s3
    s2 --> s3
    state "expression" as s4
    s3 --> s4
    s4 --> [*]
```

---

### createObjectStatement

**Syntax:**

```ebnf
createObjectStatement
    : create VARIABLE as dataType (set memberAssignmentList)?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "create" as s1
    [*] --> s1
    state "VARIABLE" as s2
    s1 --> s2
    state "as" as s3
    s2 --> s3
    state "dataType" as s4
    s3 --> s4
    state "(set memberAssignmentList)?" as s5
    s4 --> s5
    s5 --> [*]
```

---

### changeObjectStatement

**Syntax:**

```ebnf
changeObjectStatement
    : change VARIABLE set memberAssignmentList
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "change" as s1
    [*] --> s1
    state "VARIABLE" as s2
    s1 --> s2
    state "set" as s3
    s2 --> s3
    state "memberAssignmentList" as s4
    s3 --> s4
    s4 --> [*]
```

---

### commitStatement

**Syntax:**

```ebnf
commitStatement
    : commit VARIABLE (with events)? refresh?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "commit" as s1
    [*] --> s1
    state "VARIABLE" as s2
    s1 --> s2
    state "(with events)?" as s3
    s2 --> s3
    s3 --> [*]
```

---

### deleteObjectStatement

**Syntax:**

```ebnf
deleteObjectStatement
    : delete VARIABLE
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "delete" as s1
    [*] --> s1
    state "VARIABLE" as s2
    s1 --> s2
    s2 --> [*]
```

---

### retrieveStatement

**Syntax:**

```ebnf
retrieveStatement
    : retrieve VARIABLE from retrieveSource
    | (where expression)?
    | (offset NUMBER_LITERAL)?
    | (limit NUMBER_LITERAL)?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "retrieve" as s1
    [*] --> s1
    state "VARIABLE" as s2
    s1 --> s2
    state "from" as s3
    s2 --> s3
    state "retrieveSource" as s4
    s3 --> s4
    s4 --> [*]
    state "(where expression)?" as s5
    [*] --> s5
    s5 --> [*]
    state "(offset NUMBER_LITERAL)?" as s6
    [*] --> s6
    s6 --> [*]
    state "(limit NUMBER_LITERAL)?" as s7
    [*] --> s7
    s7 --> [*]
```

---

### ifStatement

**Syntax:**

```ebnf
ifStatement
    : if expression then microflowBody
    | (elsif expression then microflowBody)*
    | (else microflowBody)?
    | end if
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "if" as s1
    [*] --> s1
    state "expression" as s2
    s1 --> s2
    state "then" as s3
    s2 --> s3
    state "microflowBody" as s4
    s3 --> s4
    s4 --> [*]
    state "(elsif expression then microflowBody)*" as s5
    [*] --> s5
    s5 --> [*]
    state "(else microflowBody)?" as s6
    [*] --> s6
    s6 --> [*]
    state "end" as s7
    [*] --> s7
    state "if" as s8
    s7 --> s8
    s8 --> [*]
```

---

### loopStatement

**Syntax:**

```ebnf
loopStatement
    : loop VARIABLE in (VARIABLE | attributePath)
    | begin microflowBody end loop
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "loop" as s1
    [*] --> s1
    state "VARIABLE" as s2
    s1 --> s2
    state "in" as s3
    s2 --> s3
    state "(VARIABLE | attributePath)" as s4
    s3 --> s4
    s4 --> [*]
    state "begin" as s5
    [*] --> s5
    state "microflowBody" as s6
    s5 --> s6
    state "end" as s7
    s6 --> s7
    state "loop" as s8
    s7 --> s8
    s8 --> [*]
```

---

### whileStatement

**Syntax:**

```ebnf
whileStatement
    : while expression
    | begin? microflowBody end while?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "while" as s1
    [*] --> s1
    state "expression" as s2
    s1 --> s2
    s2 --> [*]
    state "begin?" as s3
    [*] --> s3
    state "microflowBody" as s4
    s3 --> s4
    state "end" as s5
    s4 --> s5
    state "while?" as s6
    s5 --> s6
    s6 --> [*]
```

---

### continueStatement

**Syntax:**

```ebnf
continueStatement
    : continue
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "continue" as s1
    [*] --> s1
    s1 --> [*]
```

---

### breakStatement

**Syntax:**

```ebnf
breakStatement
    : break
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "break" as s1
    [*] --> s1
    s1 --> [*]
```

---

### returnStatement

**Syntax:**

```ebnf
returnStatement
    : return expression?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "return" as s1
    [*] --> s1
    state "expression?" as s2
    s1 --> s2
    s2 --> [*]
```

---

### logStatement

**Syntax:**

```ebnf
logStatement
    : log logLevel? (node STRING_LITERAL)? expression logTemplateParams?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "log" as s1
    [*] --> s1
    state "logLevel?" as s2
    s1 --> s2
    state "(node STRING_LITERAL)?" as s3
    s2 --> s3
    state "expression" as s4
    s3 --> s4
    state "logTemplateParams?" as s5
    s4 --> s5
    s5 --> [*]
```

---

### callMicroflowStatement

**Syntax:**

```ebnf
callMicroflowStatement
    : (VARIABLE equals)? call microflow qualifiedName LPAREN callArgumentList? RPAREN
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "(VARIABLE equals)?" as s1
    [*] --> s1
    state "call" as s2
    s1 --> s2
    state "microflow" as s3
    s2 --> s3
    state "qualifiedName" as s4
    s3 --> s4
    state "LPAREN" as s5
    s4 --> s5
    state "callArgumentList?" as s6
    s5 --> s6
    state "RPAREN" as s7
    s6 --> s7
    s7 --> [*]
```

---

### callJavaActionStatement

**Syntax:**

```ebnf
callJavaActionStatement
    : (VARIABLE equals)? call java action qualifiedName LPAREN callArgumentList? RPAREN
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "(VARIABLE equals)?" as s1
    [*] --> s1
    state "call" as s2
    s1 --> s2
    state "java" as s3
    s2 --> s3
    state "action" as s4
    s3 --> s4
    state "qualifiedName" as s5
    s4 --> s5
    state "LPAREN" as s6
    s5 --> s6
    state "callArgumentList?" as s7
    s6 --> s7
    state "RPAREN" as s8
    s7 --> s8
    s8 --> [*]
```

---

### showPageStatement

**Syntax:**

```ebnf
showPageStatement
    : show page qualifiedName (LPAREN showPageArgList? RPAREN)? (for VARIABLE)? (with memberAssignmentList)?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "show" as s1
    [*] --> s1
    state "page" as s2
    s1 --> s2
    state "qualifiedName" as s3
    s2 --> s3
    state "(LPAREN showPageArgList? RPAREN)?" as s4
    s3 --> s4
    state "(for VARIABLE)?" as s5
    s4 --> s5
    state "(with memberAssignmentList)?" as s6
    s5 --> s6
    s6 --> [*]
```

---

### closePageStatement

**Syntax:**

```ebnf
closePageStatement
    : close page
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "close" as s1
    [*] --> s1
    state "page" as s2
    s1 --> s2
    s2 --> [*]
```

---

### throwStatement

**Syntax:**

```ebnf
throwStatement
    : throw expression
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "throw" as s1
    [*] --> s1
    state "expression" as s2
    s1 --> s2
    s2 --> [*]
```

---

### createPageStatement

Creates a new page with layout, parameters, and widget content.  Pages define the user interface with widgets arranged in a layout structure.

**Syntax:**

```ebnf
createPageStatement
    : page qualifiedName
    | ( LPAREN pageParameterList? RPAREN    // Parenthesized params: page Name($p: type)
    | | pageParameterList                   // Inline params: page Name\n$p: type
    | )?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "page" as s1
    [*] --> s1
    state "qualifiedName" as s2
    s1 --> s2
    s2 --> [*]
    state "( LPAREN pageParameterList? RPAREN" as s3
    [*] --> s3
    s3 --> [*]
    state "pageParameterList" as s4
    [*] --> s4
    s4 --> [*]
    state ")?" as s5
    [*] --> s5
    s5 --> [*]
```

**Examples:**

*Simple page with text:*

```sql
create page MyModule.HomePage ()
title 'Welcome'
layout Atlas_Core.Atlas_Default
begin
layoutgrid begin
row begin
column 12 begin
dynamictext (content 'Hello, World!', rendermode 'H1')
end
end
end
end;
```

*Page with parameter and data view:*

```sql
create page MyModule.CustomerDetails ($Customer: MyModule.Customer)
title 'Customer Details'
layout Atlas_Core.Atlas_Default
begin
dataview dvCustomer datasource $Customer begin
textbox (attribute Name, label 'Name'),
textbox (attribute Email, label 'Email')
end
end;
```

*Page with action button:*

```sql
create page MyModule.OrderForm ($Order: MyModule.Order)
title 'New Order'
layout Atlas_Core.Atlas_Default
begin
actionbutton btnSave 'Save Order'
action call microflow MyModule.SaveOrder
style primary
end;
```

**See also:** [pageBody for widget definitions](#pagebody-for-widget-definitions), [widgetDefinition for available widgets](#widgetdefinition-for-available-widgets)

---

### createSnippetStatement

**Syntax:**

```ebnf
createSnippetStatement
    : snippet qualifiedName
    | ( LPAREN snippetParameterList? RPAREN    // Parenthesized params: snippet Name($p: type)
    | | snippetParameterList                   // Inline params: snippet Name\n$p: type
    | )?
    | snippetOptions?
    | begin (pageWidget SEMICOLON?)* end
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "snippet" as s1
    [*] --> s1
    state "qualifiedName" as s2
    s1 --> s2
    s2 --> [*]
    state "( LPAREN snippetParameterList? RPAREN" as s3
    [*] --> s3
    s3 --> [*]
    state "snippetParameterList" as s4
    [*] --> s4
    s4 --> [*]
    state ")?" as s5
    [*] --> s5
    s5 --> [*]
    state "snippetOptions?" as s6
    [*] --> s6
    s6 --> [*]
    state "begin" as s7
    [*] --> s7
    state "(pageWidget SEMICOLON?)*" as s8
    s7 --> s8
    state "end" as s9
    s8 --> s9
    s9 --> [*]
```

---

### createNotebookStatement

**Syntax:**

```ebnf
createNotebookStatement
    : notebook qualifiedName
    | notebookOptions?
    | begin notebookPage* end
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "notebook" as s1
    [*] --> s1
    state "qualifiedName" as s2
    s1 --> s2
    s2 --> [*]
    state "notebookOptions?" as s3
    [*] --> s3
    s3 --> [*]
    state "begin" as s4
    [*] --> s4
    state "notebookPage*" as s5
    s4 --> s5
    state "end" as s6
    s5 --> s6
    s6 --> [*]
```

---

### createDatabaseConnectionStatement

**Syntax:**

```ebnf
createDatabaseConnectionStatement
    : database connection IDENTIFIER
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "database" as s1
    [*] --> s1
    state "connection" as s2
    s1 --> s2
    state "IDENTIFIER" as s3
    s2 --> s3
    s3 --> [*]
```

---

### createConstantStatement

**Syntax:**

```ebnf
createConstantStatement
    : constant qualifiedName
    | type dataType
    | default literal
    | constantOptions?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "constant" as s1
    [*] --> s1
    state "qualifiedName" as s2
    s1 --> s2
    s2 --> [*]
    state "type" as s3
    [*] --> s3
    state "dataType" as s4
    s3 --> s4
    s4 --> [*]
    state "default" as s5
    [*] --> s5
    state "literal" as s6
    s5 --> s6
    s6 --> [*]
    state "constantOptions?" as s7
    [*] --> s7
    s7 --> [*]
```

---

### createRestClientStatement

**Syntax:**

```ebnf
createRestClientStatement
    : rest client qualifiedName
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "rest" as s1
    [*] --> s1
    state "client" as s2
    s1 --> s2
    state "qualifiedName" as s3
    s2 --> s3
    s3 --> [*]
```

---

### createIndexStatement

**Syntax:**

```ebnf
createIndexStatement
    : index IDENTIFIER on qualifiedName LPAREN indexAttributeList RPAREN
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "index" as s1
    [*] --> s1
    state "IDENTIFIER" as s2
    s1 --> s2
    state "on" as s3
    s2 --> s3
    state "qualifiedName" as s4
    s3 --> s4
    state "LPAREN" as s5
    s4 --> s5
    state "indexAttributeList" as s6
    s5 --> s6
    state "RPAREN" as s7
    s6 --> s7
    s7 --> [*]
```

---

### dqlStatement

**Syntax:**

```ebnf
dqlStatement
    : showStatement
    | | describeStatement
    | | catalogSelectQuery
    | | oqlQuery
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "showStatement" as s1
    [*] --> s1
    s1 --> [*]
    state "describeStatement" as s2
    [*] --> s2
    s2 --> [*]
    state "catalogSelectQuery" as s3
    [*] --> s3
    s3 --> [*]
    state "oqlQuery" as s4
    [*] --> s4
    s4 --> [*]
```

---

### showStatement

**Syntax:**

```ebnf
showStatement
    : show modules
    | | show entities (in (qualifiedName | IDENTIFIER))?
    | | show associations (in (qualifiedName | IDENTIFIER))?
    | | show microflows (in (qualifiedName | IDENTIFIER))?
    | | show pages (in (qualifiedName | IDENTIFIER))?
    | | show snippets (in (qualifiedName | IDENTIFIER))?
    | | show enumerations (in (qualifiedName | IDENTIFIER))?
    | | show constants (in (qualifiedName | IDENTIFIER))?
    | | show layouts (in (qualifiedName | IDENTIFIER))?
    | | show notebooks (in (qualifiedName | IDENTIFIER))?
    | | show java actions (in (qualifiedName | IDENTIFIER))?
    | | show entity qualifiedName
    | | show association qualifiedName
    | | show page qualifiedName
    | | show connections
    | | show status
    | | show version
    | | show catalog IDENTIFIER  // show catalog tables, etc.
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "show" as s1
    [*] --> s1
    state "modules" as s2
    s1 --> s2
    s2 --> [*]
    state "show" as s3
    [*] --> s3
    state "entities" as s4
    s3 --> s4
    state "(in (qualifiedName | IDENTIFIER))?" as s5
    s4 --> s5
    s5 --> [*]
    state "show" as s6
    [*] --> s6
    state "associations" as s7
    s6 --> s7
    state "(in (qualifiedName | IDENTIFIER))?" as s8
    s7 --> s8
    s8 --> [*]
    state "show" as s9
    [*] --> s9
    state "microflows" as s10
    s9 --> s10
    state "(in (qualifiedName | IDENTIFIER))?" as s11
    s10 --> s11
    s11 --> [*]
    state "show" as s12
    [*] --> s12
    state "pages" as s13
    s12 --> s13
    state "(in (qualifiedName | IDENTIFIER))?" as s14
    s13 --> s14
    s14 --> [*]
    state "show" as s15
    [*] --> s15
    state "snippets" as s16
    s15 --> s16
    state "(in (qualifiedName | IDENTIFIER))?" as s17
    s16 --> s17
    s17 --> [*]
    state "show" as s18
    [*] --> s18
    state "enumerations" as s19
    s18 --> s19
    state "(in (qualifiedName | IDENTIFIER))?" as s20
    s19 --> s20
    s20 --> [*]
    state "show" as s21
    [*] --> s21
    state "constants" as s22
    s21 --> s22
    state "(in (qualifiedName | IDENTIFIER))?" as s23
    s22 --> s23
    s23 --> [*]
    state "show" as s24
    [*] --> s24
    state "layouts" as s25
    s24 --> s25
    state "(in (qualifiedName | IDENTIFIER))?" as s26
    s25 --> s26
    s26 --> [*]
    state "show" as s27
    [*] --> s27
    state "notebooks" as s28
    s27 --> s28
    state "(in (qualifiedName | IDENTIFIER))?" as s29
    s28 --> s29
    s29 --> [*]
    state "show" as s30
    [*] --> s30
    state "java" as s31
    s30 --> s31
    state "actions" as s32
    s31 --> s32
    state "(in (qualifiedName | IDENTIFIER))?" as s33
    s32 --> s33
    s33 --> [*]
    state "show" as s34
    [*] --> s34
    state "entity" as s35
    s34 --> s35
    state "qualifiedName" as s36
    s35 --> s36
    s36 --> [*]
    state "show" as s37
    [*] --> s37
    state "association" as s38
    s37 --> s38
    state "qualifiedName" as s39
    s38 --> s39
    s39 --> [*]
    state "show" as s40
    [*] --> s40
    state "page" as s41
    s40 --> s41
    state "qualifiedName" as s42
    s41 --> s42
    s42 --> [*]
    state "show" as s43
    [*] --> s43
    state "connections" as s44
    s43 --> s44
    s44 --> [*]
    state "show" as s45
    [*] --> s45
    state "status" as s46
    s45 --> s46
    s46 --> [*]
    state "show" as s47
    [*] --> s47
    state "version" as s48
    s47 --> s48
    s48 --> [*]
    state "show" as s49
    [*] --> s49
    state "catalog" as s50
    s49 --> s50
    state "IDENTIFIER" as s51
    s50 --> s51
    s51 --> [*]
```

---

### describeStatement

**Syntax:**

```ebnf
describeStatement
    : describe entity qualifiedName
    | | describe association qualifiedName
    | | describe microflow qualifiedName
    | | describe nanoflow qualifiedName
    | | describe page qualifiedName
    | | describe snippet qualifiedName
    | | describe enumeration qualifiedName
    | | describe module IDENTIFIER (with all)?  // describe module Name [with all] - optionally include all objects
    | | describe catalog DOT IDENTIFIER  // describe CATALOG.ENTITIES
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "describe" as s1
    [*] --> s1
    state "entity" as s2
    s1 --> s2
    state "qualifiedName" as s3
    s2 --> s3
    s3 --> [*]
    state "describe" as s4
    [*] --> s4
    state "association" as s5
    s4 --> s5
    state "qualifiedName" as s6
    s5 --> s6
    s6 --> [*]
    state "describe" as s7
    [*] --> s7
    state "microflow" as s8
    s7 --> s8
    state "qualifiedName" as s9
    s8 --> s9
    s9 --> [*]
    state "describe" as s10
    [*] --> s10
    state "nanoflow" as s11
    s10 --> s11
    state "qualifiedName" as s12
    s11 --> s12
    s12 --> [*]
    state "describe" as s13
    [*] --> s13
    state "page" as s14
    s13 --> s14
    state "qualifiedName" as s15
    s14 --> s15
    s15 --> [*]
    state "describe" as s16
    [*] --> s16
    state "snippet" as s17
    s16 --> s17
    state "qualifiedName" as s18
    s17 --> s18
    s18 --> [*]
    state "describe" as s19
    [*] --> s19
    state "enumeration" as s20
    s19 --> s20
    state "qualifiedName" as s21
    s20 --> s21
    s21 --> [*]
    state "describe" as s22
    [*] --> s22
    state "module" as s23
    s22 --> s23
    state "IDENTIFIER" as s24
    s23 --> s24
    state "(with all)?" as s25
    s24 --> s25
    s25 --> [*]
    state "describe" as s26
    [*] --> s26
    state "catalog" as s27
    s26 --> s27
    state "DOT" as s28
    s27 --> s28
    state "IDENTIFIER" as s29
    s28 --> s29
    s29 --> [*]
```

---

### utilityStatement

**Syntax:**

```ebnf
utilityStatement
    : connectStatement
    | | disconnectStatement
    | | commitChangesStatement
    | | updateStatement
    | | checkStatement
    | | buildStatement
    | | executeScriptStatement
    | | executeRuntimeStatement
    | | lintStatement
    | | useSessionStatement
    | | introspectApiStatement
    | | debugStatement
    | | helpStatement
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "connectStatement" as s1
    [*] --> s1
    s1 --> [*]
    state "disconnectStatement" as s2
    [*] --> s2
    s2 --> [*]
    state "commitChangesStatement" as s3
    [*] --> s3
    s3 --> [*]
    state "updateStatement" as s4
    [*] --> s4
    s4 --> [*]
    state "checkStatement" as s5
    [*] --> s5
    s5 --> [*]
    state "buildStatement" as s6
    [*] --> s6
    s6 --> [*]
    state "executeScriptStatement" as s7
    [*] --> s7
    s7 --> [*]
    state "executeRuntimeStatement" as s8
    [*] --> s8
    s8 --> [*]
    state "lintStatement" as s9
    [*] --> s9
    s9 --> [*]
    state "useSessionStatement" as s10
    [*] --> s10
    s10 --> [*]
    state "introspectApiStatement" as s11
    [*] --> s11
    s11 --> [*]
    state "debugStatement" as s12
    [*] --> s12
    s12 --> [*]
    state "helpStatement" as s13
    [*] --> s13
    s13 --> [*]
```

---

### connectStatement

**Syntax:**

```ebnf
connectStatement
    : connect to project STRING_LITERAL (branch STRING_LITERAL)? token STRING_LITERAL
    | | connect local STRING_LITERAL
    | | connect runtime host STRING_LITERAL port NUMBER_LITERAL (token STRING_LITERAL)?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "connect" as s1
    [*] --> s1
    state "to" as s2
    s1 --> s2
    state "project" as s3
    s2 --> s3
    state "STRING_LITERAL" as s4
    s3 --> s4
    state "(branch STRING_LITERAL)?" as s5
    s4 --> s5
    state "token" as s6
    s5 --> s6
    state "STRING_LITERAL" as s7
    s6 --> s7
    s7 --> [*]
    state "connect" as s8
    [*] --> s8
    state "local" as s9
    s8 --> s9
    state "STRING_LITERAL" as s10
    s9 --> s10
    s10 --> [*]
    state "connect" as s11
    [*] --> s11
    state "runtime" as s12
    s11 --> s12
    state "host" as s13
    s12 --> s13
    state "STRING_LITERAL" as s14
    s13 --> s14
    state "port" as s15
    s14 --> s15
    state "NUMBER_LITERAL" as s16
    s15 --> s16
    state "(token STRING_LITERAL)?" as s17
    s16 --> s17
    s17 --> [*]
```

---

### disconnectStatement

**Syntax:**

```ebnf
disconnectStatement
    : disconnect
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "disconnect" as s1
    [*] --> s1
    s1 --> [*]
```

---

### commitChangesStatement

**Syntax:**

```ebnf
commitChangesStatement
    : commit (message STRING_LITERAL)?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "commit" as s1
    [*] --> s1
    state "(message STRING_LITERAL)?" as s2
    s1 --> s2
    s2 --> [*]
```

---

### updateStatement

**Syntax:**

```ebnf
updateStatement
    : update
    | | refresh catalog full?
    | | refresh
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "update" as s1
    [*] --> s1
    s1 --> [*]
    state "refresh" as s2
    [*] --> s2
    state "catalog" as s3
    s2 --> s3
    state "full?" as s4
    s3 --> s4
    s4 --> [*]
    state "refresh" as s5
    [*] --> s5
    s5 --> [*]
```

---

### checkStatement

**Syntax:**

```ebnf
checkStatement
    : check
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "check" as s1
    [*] --> s1
    s1 --> [*]
```

---

### buildStatement

**Syntax:**

```ebnf
buildStatement
    : build
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "build" as s1
    [*] --> s1
    s1 --> [*]
```

---

### executeScriptStatement

**Syntax:**

```ebnf
executeScriptStatement
    : execute script STRING_LITERAL
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "execute" as s1
    [*] --> s1
    state "script" as s2
    s1 --> s2
    state "STRING_LITERAL" as s3
    s2 --> s3
    s3 --> [*]
```

---

### executeRuntimeStatement

**Syntax:**

```ebnf
executeRuntimeStatement
    : execute runtime STRING_LITERAL
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "execute" as s1
    [*] --> s1
    state "runtime" as s2
    s1 --> s2
    state "STRING_LITERAL" as s3
    s2 --> s3
    s3 --> [*]
```

---

### lintStatement

**Syntax:**

```ebnf
lintStatement
    : lint STRING_LITERAL?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "lint" as s1
    [*] --> s1
    state "STRING_LITERAL?" as s2
    s1 --> s2
    s2 --> [*]
```

---

### useSessionStatement

**Syntax:**

```ebnf
useSessionStatement
    : use sessionIdList
    | | use all
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "use" as s1
    [*] --> s1
    state "sessionIdList" as s2
    s1 --> s2
    s2 --> [*]
    state "use" as s3
    [*] --> s3
    state "all" as s4
    s3 --> s4
    s4 --> [*]
```

---

### introspectApiStatement

**Syntax:**

```ebnf
introspectApiStatement
    : introspect api
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "introspect" as s1
    [*] --> s1
    state "api" as s2
    s1 --> s2
    s2 --> [*]
```

---

### debugStatement

**Syntax:**

```ebnf
debugStatement
    : debug STRING_LITERAL
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "debug" as s1
    [*] --> s1
    state "STRING_LITERAL" as s2
    s1 --> s2
    s2 --> [*]
```

---

### helpStatement

**Syntax:**

```ebnf
helpStatement
    : IDENTIFIER  // HELP command
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "IDENTIFIER" as s1
    [*] --> s1
    s1 --> [*]
```

---

## Entity Definitions

### entityBody

**Syntax:**

```ebnf
entityBody
    : LPAREN attributeDefinitionList? RPAREN entityOptions?
    | | entityOptions
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "LPAREN" as s1
    [*] --> s1
    state "attributeDefinitionList?" as s2
    s1 --> s2
    state "RPAREN" as s3
    s2 --> s3
    state "entityOptions?" as s4
    s3 --> s4
    s4 --> [*]
    state "entityOptions" as s5
    [*] --> s5
    s5 --> [*]
```

---

### entityOptions

**Syntax:**

```ebnf
entityOptions
    : entityOption (COMMA? entityOption)*  // Allow optional commas between options
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "entityOption" as s1
    [*] --> s1
    state "(COMMA? entityOption)*" as s2
    s1 --> s2
    s2 --> [*]
```

---

### entityOption

**Syntax:**

```ebnf
entityOption
    : extends qualifiedName
    | | generalization qualifiedName
    | | comment STRING_LITERAL
    | | index indexDefinition
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "extends" as s1
    [*] --> s1
    state "qualifiedName" as s2
    s1 --> s2
    s2 --> [*]
    state "generalization" as s3
    [*] --> s3
    state "qualifiedName" as s4
    s3 --> s4
    s4 --> [*]
    state "comment" as s5
    [*] --> s5
    state "STRING_LITERAL" as s6
    s5 --> s6
    s6 --> [*]
    state "index" as s7
    [*] --> s7
    state "indexDefinition" as s8
    s7 --> s8
    s8 --> [*]
```

---

### attributeDefinitionList

**Syntax:**

```ebnf
attributeDefinitionList
    : attributeDefinition (COMMA attributeDefinition)*
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "attributeDefinition" as s1
    [*] --> s1
    state "(COMMA attributeDefinition)*" as s2
    s1 --> s2
    s2 --> [*]
```

---

### attributeDefinition

Defines an attribute within an entity.  Attributes have a name, data type, and optional constraints like NOT NULL, UNIQUE, or DEFAULT. Documentation comments can be added above the attribute.

**Syntax:**

```ebnf
attributeDefinition
    : docComment? annotation* attributename COLON dataType attributeConstraint*
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "docComment?" as s1
    [*] --> s1
    state "annotation*" as s2
    s1 --> s2
    state "attributename" as s3
    s2 --> s3
    state "COLON" as s4
    s3 --> s4
    state "dataType" as s5
    s4 --> s5
    state "attributeConstraint*" as s6
    s5 --> s6
    s6 --> [*]
```

**Examples:**

*Simple attributes:*

```sql
Name: string(100),
Age: integer,
Active: boolean
```

*Attributes with constraints:*

```sql
Code: string(50) not null,
Email: string(200) unique,
status: enum MyModule.Status default 'Active'
```

*Attribute with custom error messages:*

```sql
Name: string(100) not null error 'Name is required',
Code: string(50) unique error 'Code must be unique'
```

*Documented attribute:*

```sql
/** The customer's primary email address
```

---

### attributeName

**Syntax:**

```ebnf
attributename
    : IDENTIFIER
    | | status | type | value | index         // Common keywords used as attribute names
    | | username | password                   // user-related keywords
    | | count | sum | avg | min | max         // Aggregate function names as attributes
    | | action | message                      // Common entity attribute names
    | | owner | reference | cascade           // association keywords that might be attribute names
    | | success | error | warning | info | debug | critical  // log/status keywords
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "IDENTIFIER" as s1
    [*] --> s1
    s1 --> [*]
    state "status" as s2
    [*] --> s2
    state "|" as s3
    s2 --> s3
    state "type" as s4
    s3 --> s4
    state "|" as s5
    s4 --> s5
    state "value" as s6
    s5 --> s6
    state "|" as s7
    s6 --> s7
    state "index" as s8
    s7 --> s8
    s8 --> [*]
    state "username" as s9
    [*] --> s9
    state "|" as s10
    s9 --> s10
    state "password" as s11
    s10 --> s11
    s11 --> [*]
    state "count" as s12
    [*] --> s12
    state "|" as s13
    s12 --> s13
    state "sum" as s14
    s13 --> s14
    state "|" as s15
    s14 --> s15
    state "avg" as s16
    s15 --> s16
    state "|" as s17
    s16 --> s17
    state "min" as s18
    s17 --> s18
    state "|" as s19
    s18 --> s19
    state "max" as s20
    s19 --> s20
    s20 --> [*]
    state "action" as s21
    [*] --> s21
    state "|" as s22
    s21 --> s22
    state "message" as s23
    s22 --> s23
    s23 --> [*]
    state "owner" as s24
    [*] --> s24
    state "|" as s25
    s24 --> s25
    state "reference" as s26
    s25 --> s26
    state "|" as s27
    s26 --> s27
    state "cascade" as s28
    s27 --> s28
    s28 --> [*]
    state "success" as s29
    [*] --> s29
    state "|" as s30
    s29 --> s30
    state "error" as s31
    s30 --> s31
    state "|" as s32
    s31 --> s32
    state "warning" as s33
    s32 --> s33
    state "|" as s34
    s33 --> s34
    state "info" as s35
    s34 --> s35
    state "|" as s36
    s35 --> s36
    state "debug" as s37
    s36 --> s37
    state "|" as s38
    s37 --> s38
    state "critical" as s39
    s38 --> s39
    s39 --> [*]
```

---

### attributeConstraint

**Syntax:**

```ebnf
attributeConstraint
    : NOT_NULL (error STRING_LITERAL)?
    | | not null (error STRING_LITERAL)?
    | | unique (error STRING_LITERAL)?
    | | default (literal | expression)
    | | required (error STRING_LITERAL)?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "NOT_NULL" as s1
    [*] --> s1
    state "(error STRING_LITERAL)?" as s2
    s1 --> s2
    s2 --> [*]
    state "not" as s3
    [*] --> s3
    state "null" as s4
    s3 --> s4
    state "(error STRING_LITERAL)?" as s5
    s4 --> s5
    s5 --> [*]
    state "unique" as s6
    [*] --> s6
    state "(error STRING_LITERAL)?" as s7
    s6 --> s7
    s7 --> [*]
    state "default" as s8
    [*] --> s8
    state "(literal | expression)" as s9
    s8 --> s9
    s9 --> [*]
    state "required" as s10
    [*] --> s10
    state "(error STRING_LITERAL)?" as s11
    s10 --> s11
    s11 --> [*]
```

---

### indexAttributeList

**Syntax:**

```ebnf
indexAttributeList
    : indexAttribute (COMMA indexAttribute)*
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "indexAttribute" as s1
    [*] --> s1
    state "(COMMA indexAttribute)*" as s2
    s1 --> s2
    s2 --> [*]
```

---

### indexAttribute

**Syntax:**

```ebnf
indexAttribute
    : indexColumnName (asc | desc)?  // column name with optional sort order
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "indexColumnName" as s1
    [*] --> s1
    state "(asc | desc)?" as s2
    s1 --> s2
    s2 --> [*]
```

---

### associationOptions

**Syntax:**

```ebnf
associationOptions
    : associationOption+
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "associationOption+" as s1
    [*] --> s1
    s1 --> [*]
```

---

### associationOption

**Syntax:**

```ebnf
associationOption
    : type (reference | reference_set)
    | | owner (default | both)
    | | delete_behavior deleteBehavior
    | | comment STRING_LITERAL
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "type" as s1
    [*] --> s1
    state "(reference | reference_set)" as s2
    s1 --> s2
    s2 --> [*]
    state "owner" as s3
    [*] --> s3
    state "(default | both)" as s4
    s3 --> s4
    s4 --> [*]
    state "delete_behavior" as s5
    [*] --> s5
    state "deleteBehavior" as s6
    s5 --> s6
    s6 --> [*]
    state "comment" as s7
    [*] --> s7
    state "STRING_LITERAL" as s8
    s7 --> s8
    s8 --> [*]
```

---

### alterEntityAction

**Syntax:**

```ebnf
alterEntityAction
    : add attribute attributeDefinition
    | | add column attributeDefinition
    | | rename attribute IDENTIFIER to IDENTIFIER
    | | rename column IDENTIFIER to IDENTIFIER
    | | modify attribute IDENTIFIER ':'? dataType attributeConstraint*
    | | modify column IDENTIFIER ':'? dataType attributeConstraint*
    | | drop attribute IDENTIFIER
    | | drop column IDENTIFIER
    | | set documentation STRING_LITERAL
    | | set comment STRING_LITERAL
    | | set store owner
    | | set position LPAREN NUMBER_LITERAL COMMA NUMBER_LITERAL RPAREN
    | | add index indexDefinition
    | | drop index IDENTIFIER
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "add" as s1
    [*] --> s1
    state "attribute" as s2
    s1 --> s2
    state "attributeDefinition" as s3
    s2 --> s3
    s3 --> [*]
    state "add" as s4
    [*] --> s4
    state "column" as s5
    s4 --> s5
    state "attributeDefinition" as s6
    s5 --> s6
    s6 --> [*]
    state "rename" as s7
    [*] --> s7
    state "attribute" as s8
    s7 --> s8
    state "IDENTIFIER" as s9
    s8 --> s9
    state "to" as s10
    s9 --> s10
    state "IDENTIFIER" as s11
    s10 --> s11
    s11 --> [*]
    state "rename" as s12
    [*] --> s12
    state "column" as s13
    s12 --> s13
    state "IDENTIFIER" as s14
    s13 --> s14
    state "to" as s15
    s14 --> s15
    state "IDENTIFIER" as s16
    s15 --> s16
    s16 --> [*]
    state "modify" as s17
    [*] --> s17
    state "attribute" as s18
    s17 --> s18
    state "IDENTIFIER" as s19
    s18 --> s19
    state "dataType" as s20
    s19 --> s20
    state "attributeConstraint*" as s21
    s20 --> s21
    s21 --> [*]
    state "modify" as s22
    [*] --> s22
    state "column" as s23
    s22 --> s23
    state "IDENTIFIER" as s24
    s23 --> s24
    state "dataType" as s25
    s24 --> s25
    state "attributeConstraint*" as s26
    s25 --> s26
    s26 --> [*]
    state "drop" as s27
    [*] --> s27
    state "attribute" as s28
    s27 --> s28
    state "IDENTIFIER" as s29
    s28 --> s29
    s29 --> [*]
    state "drop" as s30
    [*] --> s30
    state "column" as s31
    s30 --> s31
    state "IDENTIFIER" as s32
    s31 --> s32
    s32 --> [*]
    state "set" as s33
    [*] --> s33
    state "documentation" as s34
    s33 --> s34
    state "STRING_LITERAL" as s35
    s34 --> s35
    s35 --> [*]
    state "set" as s36
    [*] --> s36
    state "comment" as s37
    s36 --> s37
    state "STRING_LITERAL" as s38
    s37 --> s38
    s38 --> [*]
    state "add" as s39
    [*] --> s39
    state "index" as s40
    s39 --> s40
    state "indexDefinition" as s41
    s40 --> s41
    s41 --> [*]
    state "drop" as s42
    [*] --> s42
    state "index" as s43
    s42 --> s43
    state "IDENTIFIER" as s44
    s43 --> s44
    s44 --> [*]
```

---

### alterAssociationAction

**Syntax:**

```ebnf
alterAssociationAction
    : set delete_behavior deleteBehavior
    | | set owner (default | both)
    | | set comment STRING_LITERAL
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "set" as s1
    [*] --> s1
    state "delete_behavior" as s2
    s1 --> s2
    state "deleteBehavior" as s3
    s2 --> s3
    s3 --> [*]
    state "set" as s4
    [*] --> s4
    state "owner" as s5
    s4 --> s5
    state "(default | both)" as s6
    s5 --> s6
    s6 --> [*]
    state "set" as s7
    [*] --> s7
    state "comment" as s8
    s7 --> s8
    state "STRING_LITERAL" as s9
    s8 --> s9
    s9 --> [*]
```

---

### attributeReference

**Syntax:**

```ebnf
attributeReference
    : IDENTIFIER (SLASH IDENTIFIER)*
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "IDENTIFIER" as s1
    [*] --> s1
    state "(SLASH IDENTIFIER)*" as s2
    s1 --> s2
    s2 --> [*]
```

---

### attributeReferenceList

**Syntax:**

```ebnf
attributeReferenceList
    : attributeReference (COMMA attributeReference)*
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "attributeReference" as s1
    [*] --> s1
    state "(COMMA attributeReference)*" as s2
    s1 --> s2
    s2 --> [*]
```

---

### attributePath

**Syntax:**

```ebnf
attributePath
    : VARIABLE (SLASH (IDENTIFIER | qualifiedName))+
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "VARIABLE" as s1
    [*] --> s1
    state "(SLASH (IDENTIFIER | qualifiedName))+" as s2
    s1 --> s2
    s2 --> [*]
```

---

### memberAttributeName

**Syntax:**

```ebnf
memberAttributeName
    : IDENTIFIER
    | | status | type | value | index
    | | username | password
    | | action | message
    | | owner | reference | cascade
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "IDENTIFIER" as s1
    [*] --> s1
    s1 --> [*]
    state "status" as s2
    [*] --> s2
    state "|" as s3
    s2 --> s3
    state "type" as s4
    s3 --> s4
    state "|" as s5
    s4 --> s5
    state "value" as s6
    s5 --> s6
    state "|" as s7
    s6 --> s7
    state "index" as s8
    s7 --> s8
    s8 --> [*]
    state "username" as s9
    [*] --> s9
    state "|" as s10
    s9 --> s10
    state "password" as s11
    s10 --> s11
    s11 --> [*]
    state "action" as s12
    [*] --> s12
    state "|" as s13
    s12 --> s13
    state "message" as s14
    s13 --> s14
    s14 --> [*]
    state "owner" as s15
    [*] --> s15
    state "|" as s16
    s15 --> s16
    state "reference" as s17
    s16 --> s17
    state "|" as s18
    s17 --> s18
    state "cascade" as s19
    s18 --> s19
    s19 --> [*]
```

---

### attributeClause

**Syntax:**

```ebnf
attributeClause
    : attribute (STRING_LITERAL | qualifiedName)
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "attribute" as s1
    [*] --> s1
    state "(STRING_LITERAL | qualifiedName)" as s2
    s1 --> s2
    s2 --> [*]
```

---

## Microflow Statements

### alterEnumerationAction

**Syntax:**

```ebnf
alterEnumerationAction
    : add value IDENTIFIER (caption STRING_LITERAL)?
    | | rename value IDENTIFIER to IDENTIFIER
    | | drop value IDENTIFIER
    | | set comment STRING_LITERAL
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "add" as s1
    [*] --> s1
    state "value" as s2
    s1 --> s2
    state "IDENTIFIER" as s3
    s2 --> s3
    state "(caption STRING_LITERAL)?" as s4
    s3 --> s4
    s4 --> [*]
    state "rename" as s5
    [*] --> s5
    state "value" as s6
    s5 --> s6
    state "IDENTIFIER" as s7
    s6 --> s7
    state "to" as s8
    s7 --> s8
    state "IDENTIFIER" as s9
    s8 --> s9
    s9 --> [*]
    state "drop" as s10
    [*] --> s10
    state "value" as s11
    s10 --> s11
    state "IDENTIFIER" as s12
    s11 --> s12
    s12 --> [*]
    state "set" as s13
    [*] --> s13
    state "comment" as s14
    s13 --> s14
    state "STRING_LITERAL" as s15
    s14 --> s15
    s15 --> [*]
```

---

### alterNotebookAction

**Syntax:**

```ebnf
alterNotebookAction
    : add page qualifiedName (position NUMBER_LITERAL)?
    | | drop page qualifiedName
    | | set comment STRING_LITERAL
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "add" as s1
    [*] --> s1
    state "page" as s2
    s1 --> s2
    state "qualifiedName" as s3
    s2 --> s3
    state "(position NUMBER_LITERAL)?" as s4
    s3 --> s4
    s4 --> [*]
    state "drop" as s5
    [*] --> s5
    state "page" as s6
    s5 --> s6
    state "qualifiedName" as s7
    s6 --> s7
    s7 --> [*]
    state "set" as s8
    [*] --> s8
    state "comment" as s9
    s8 --> s9
    state "STRING_LITERAL" as s10
    s9 --> s10
    s10 --> [*]
```

---

### microflowParameterList

**Syntax:**

```ebnf
microflowParameterList
    : microflowParameter (COMMA microflowParameter)*
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "microflowParameter" as s1
    [*] --> s1
    state "(COMMA microflowParameter)*" as s2
    s1 --> s2
    s2 --> [*]
```

---

### microflowParameter

**Syntax:**

```ebnf
microflowParameter
    : (IDENTIFIER | VARIABLE) COLON dataType
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "(IDENTIFIER | VARIABLE)" as s1
    [*] --> s1
    state "COLON" as s2
    s1 --> s2
    state "dataType" as s3
    s2 --> s3
    s3 --> [*]
```

---

### microflowReturnType

**Syntax:**

```ebnf
microflowReturnType
    : returns dataType (as VARIABLE)?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "returns" as s1
    [*] --> s1
    state "dataType" as s2
    s1 --> s2
    state "(as VARIABLE)?" as s3
    s2 --> s3
    s3 --> [*]
```

---

### microflowOptions

**Syntax:**

```ebnf
microflowOptions
    : microflowOption+
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "microflowOption+" as s1
    [*] --> s1
    s1 --> [*]
```

---

### microflowOption

**Syntax:**

```ebnf
microflowOption
    : folder STRING_LITERAL
    | | comment STRING_LITERAL
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "folder" as s1
    [*] --> s1
    state "STRING_LITERAL" as s2
    s1 --> s2
    s2 --> [*]
    state "comment" as s3
    [*] --> s3
    state "STRING_LITERAL" as s4
    s3 --> s4
    s4 --> [*]
```

---

### microflowBody

**Syntax:**

```ebnf
microflowBody
    : microflowStatement*
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "microflowStatement*" as s1
    [*] --> s1
    s1 --> [*]
```

---

### actionButtonWidget

**Syntax:**

```ebnf
actionButtonWidget
    : actionbutton IDENTIFIER? STRING_LITERAL?
    | templateParams?
    | buttonAction?
    | (style buttonstyle)?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "actionbutton" as s1
    [*] --> s1
    state "IDENTIFIER?" as s2
    s1 --> s2
    state "STRING_LITERAL?" as s3
    s2 --> s3
    s3 --> [*]
    state "templateParams?" as s4
    [*] --> s4
    s4 --> [*]
    state "buttonAction?" as s5
    [*] --> s5
    s5 --> [*]
    state "(style buttonstyle)?" as s6
    [*] --> s6
    s6 --> [*]
```

---

### buttonAction

**Syntax:**

```ebnf
buttonAction
    : action actionType (STRING_LITERAL | qualifiedName)?
    | secondaryAction?
    | (passing LPAREN passArgList RPAREN)?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "action" as s1
    [*] --> s1
    state "actionType" as s2
    s1 --> s2
    state "(STRING_LITERAL | qualifiedName)?" as s3
    s2 --> s3
    s3 --> [*]
    state "secondaryAction?" as s4
    [*] --> s4
    s4 --> [*]
    state "(passing LPAREN passArgList RPAREN)?" as s5
    [*] --> s5
    s5 --> [*]
```

---

### secondaryAction

**Syntax:**

```ebnf
secondaryAction
    : close_page booleanLiteral?                    // close_page true
    | | show_page (STRING_LITERAL | qualifiedName)    // show_page 'Module.Page'
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "close_page" as s1
    [*] --> s1
    state "booleanLiteral?" as s2
    s1 --> s2
    s2 --> [*]
    state "show_page" as s3
    [*] --> s3
    state "(STRING_LITERAL | qualifiedName)" as s4
    s3 --> s4
    s4 --> [*]
```

---

### actionType

**Syntax:**

```ebnf
actionType
    : save_changes
    | | cancel_changes
    | | close_page
    | | show_page
    | | delete_action
    | | create_object qualifiedName?
    | | call_microflow
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "save_changes" as s1
    [*] --> s1
    s1 --> [*]
    state "cancel_changes" as s2
    [*] --> s2
    s2 --> [*]
    state "close_page" as s3
    [*] --> s3
    s3 --> [*]
    state "show_page" as s4
    [*] --> s4
    s4 --> [*]
    state "delete_action" as s5
    [*] --> s5
    s5 --> [*]
    state "create_object" as s6
    [*] --> s6
    state "qualifiedName?" as s7
    s6 --> s7
    s7 --> [*]
    state "call_microflow" as s8
    [*] --> s8
    s8 --> [*]
```

---

## Page Definitions

### showPageArgList

**Syntax:**

```ebnf
showPageArgList
    : showPageArg (COMMA showPageArg)*
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "showPageArg" as s1
    [*] --> s1
    state "(COMMA showPageArg)*" as s2
    s1 --> s2
    s2 --> [*]
```

---

### showPageArg

**Syntax:**

```ebnf
showPageArg
    : VARIABLE equals (VARIABLE | expression)
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "VARIABLE" as s1
    [*] --> s1
    state "equals" as s2
    s1 --> s2
    state "(VARIABLE | expression)" as s3
    s2 --> s3
    s3 --> [*]
```

---

### pageOptions

**Syntax:**

```ebnf
pageOptions
    : (begin pageBody end)?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "(begin pageBody end)?" as s1
    [*] --> s1
    s1 --> [*]
```

---

### pageParameterList

**Syntax:**

```ebnf
pageParameterList
    : pageParameter (COMMA pageParameter)*
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "pageParameter" as s1
    [*] --> s1
    state "(COMMA pageParameter)*" as s2
    s1 --> s2
    s2 --> [*]
```

---

### pageParameter

**Syntax:**

```ebnf
pageParameter
    : (IDENTIFIER | VARIABLE) COLON dataType
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "(IDENTIFIER | VARIABLE)" as s1
    [*] --> s1
    state "COLON" as s2
    s1 --> s2
    state "dataType" as s3
    s2 --> s3
    s3 --> [*]
```

---

### pageOptions

**Syntax:**

```ebnf
pageOptions
    : pageOption*
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "pageOption*" as s1
    [*] --> s1
    s1 --> [*]
```

---

### pageOption

**Syntax:**

```ebnf
pageOption
    : title STRING_LITERAL
    | | layout (qualifiedName | STRING_LITERAL)
    | | url STRING_LITERAL
    | | folder STRING_LITERAL
    | | class STRING_LITERAL
    | | comment STRING_LITERAL
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "title" as s1
    [*] --> s1
    state "STRING_LITERAL" as s2
    s1 --> s2
    s2 --> [*]
    state "layout" as s3
    [*] --> s3
    state "(qualifiedName | STRING_LITERAL)" as s4
    s3 --> s4
    s4 --> [*]
    state "url" as s5
    [*] --> s5
    state "STRING_LITERAL" as s6
    s5 --> s6
    s6 --> [*]
    state "folder" as s7
    [*] --> s7
    state "STRING_LITERAL" as s8
    s7 --> s8
    s8 --> [*]
    state "class" as s9
    [*] --> s9
    state "STRING_LITERAL" as s10
    s9 --> s10
    s10 --> [*]
    state "comment" as s11
    [*] --> s11
    state "STRING_LITERAL" as s12
    s11 --> s12
    s12 --> [*]
```

---

### pageBody

**Syntax:**

```ebnf
pageBody
    : (placeholderBlock | pageWidget SEMICOLON?)*
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "(placeholderBlock | pageWidget SEMICOLON?)*" as s1
    [*] --> s1
    s1 --> [*]
```

---

### pageWidget

**Syntax:**

```ebnf
pageWidget
    : layoutGridWidget
    | | dataGridWidget
    | | dataViewWidget
    | | listViewWidget
    | | galleryWidget
    | | containerWidget
    | | actionButtonWidget
    | | linkButtonWidget
    | | titleWidget
    | | dynamicTextWidget
    | | staticTextWidget
    | | snippetCallWidget
    | | textBoxWidget
    | | textAreaWidget
    | | datePickerWidget
    | | radioButtonsWidget
    | | dropDownWidget
    | | comboBoxWidget
    | | checkBoxWidget
    | | referenceSelectorWidget
    | | customWidgetWidget
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "layoutGridWidget" as s1
    [*] --> s1
    s1 --> [*]
    state "dataGridWidget" as s2
    [*] --> s2
    s2 --> [*]
    state "dataViewWidget" as s3
    [*] --> s3
    s3 --> [*]
    state "listViewWidget" as s4
    [*] --> s4
    s4 --> [*]
    state "galleryWidget" as s5
    [*] --> s5
    s5 --> [*]
    state "containerWidget" as s6
    [*] --> s6
    s6 --> [*]
    state "actionButtonWidget" as s7
    [*] --> s7
    s7 --> [*]
    state "linkButtonWidget" as s8
    [*] --> s8
    s8 --> [*]
    state "titleWidget" as s9
    [*] --> s9
    s9 --> [*]
    state "dynamicTextWidget" as s10
    [*] --> s10
    s10 --> [*]
    state "staticTextWidget" as s11
    [*] --> s11
    s11 --> [*]
    state "snippetCallWidget" as s12
    [*] --> s12
    s12 --> [*]
    state "textBoxWidget" as s13
    [*] --> s13
    s13 --> [*]
    state "textAreaWidget" as s14
    [*] --> s14
    s14 --> [*]
    state "datePickerWidget" as s15
    [*] --> s15
    s15 --> [*]
    state "radioButtonsWidget" as s16
    [*] --> s16
    s16 --> [*]
    state "dropDownWidget" as s17
    [*] --> s17
    s17 --> [*]
    state "comboBoxWidget" as s18
    [*] --> s18
    s18 --> [*]
    state "checkBoxWidget" as s19
    [*] --> s19
    s19 --> [*]
    state "referenceSelectorWidget" as s20
    [*] --> s20
    s20 --> [*]
    state "customWidgetWidget" as s21
    [*] --> s21
    s21 --> [*]
```

---

### layoutGridWidget

**Syntax:**

```ebnf
layoutGridWidget
    : layoutgrid IDENTIFIER? widgetOptions?
    | ( begin layoutGridRow* end
    | | layoutGridRow+
    | )?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "layoutgrid" as s1
    [*] --> s1
    state "IDENTIFIER?" as s2
    s1 --> s2
    state "widgetOptions?" as s3
    s2 --> s3
    s3 --> [*]
    state "( begin layoutGridRow* end" as s4
    [*] --> s4
    s4 --> [*]
    state "layoutGridRow+" as s5
    [*] --> s5
    s5 --> [*]
    state ")?" as s6
    [*] --> s6
    s6 --> [*]
```

---

### layoutGridRow

**Syntax:**

```ebnf
layoutGridRow
    : row widgetOptions?
    | ( begin layoutGridColumn* end
    | | layoutGridColumn+
    | )
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "row" as s1
    [*] --> s1
    state "widgetOptions?" as s2
    s1 --> s2
    s2 --> [*]
    state "( begin layoutGridColumn* end" as s3
    [*] --> s3
    s3 --> [*]
    state "layoutGridColumn+" as s4
    [*] --> s4
    s4 --> [*]
    state ")" as s5
    [*] --> s5
    s5 --> [*]
```

---

### layoutGridColumn

**Syntax:**

```ebnf
layoutGridColumn
    : column widgetOptions?
    | ( begin (pageWidget SEMICOLON?)* end
    | | (pageWidget SEMICOLON?)*
    | )
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "column" as s1
    [*] --> s1
    state "widgetOptions?" as s2
    s1 --> s2
    s2 --> [*]
    state "( begin (pageWidget SEMICOLON?)* end" as s3
    [*] --> s3
    s3 --> [*]
    state "(pageWidget SEMICOLON?)*" as s4
    [*] --> s4
    s4 --> [*]
    state ")" as s5
    [*] --> s5
    s5 --> [*]
```

---

### dataGridWidget

**Syntax:**

```ebnf
dataGridWidget
    : datagrid qualifiedName? widgetOptions?
    | dataGridSource?
    | (begin dataGridContent end)?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "datagrid" as s1
    [*] --> s1
    state "qualifiedName?" as s2
    s1 --> s2
    state "widgetOptions?" as s3
    s2 --> s3
    s3 --> [*]
    state "dataGridSource?" as s4
    [*] --> s4
    s4 --> [*]
    state "(begin dataGridContent end)?" as s5
    [*] --> s5
    s5 --> [*]
```

---

### dataViewWidget

**Syntax:**

```ebnf
dataViewWidget
    : dataview qualifiedName? widgetOptions?
    | dataSourceClause?
    | (begin (pageWidget SEMICOLON?)* dataViewFooter? end)?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "dataview" as s1
    [*] --> s1
    state "qualifiedName?" as s2
    s1 --> s2
    state "widgetOptions?" as s3
    s2 --> s3
    s3 --> [*]
    state "dataSourceClause?" as s4
    [*] --> s4
    s4 --> [*]
    state "(begin (pageWidget SEMICOLON?)* dataViewFooter? end)?" as s5
    [*] --> s5
    s5 --> [*]
```

---

### listViewWidget

**Syntax:**

```ebnf
listViewWidget
    : listview qualifiedName? widgetOptions?
    | (begin pageWidget* end)?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "listview" as s1
    [*] --> s1
    state "qualifiedName?" as s2
    s1 --> s2
    state "widgetOptions?" as s3
    s2 --> s3
    s3 --> [*]
    state "(begin pageWidget* end)?" as s4
    [*] --> s4
    s4 --> [*]
```

---

### galleryWidget

**Syntax:**

```ebnf
galleryWidget
    : gallery qualifiedName? widgetOptions?
    | gallerySource?
    | (selection (single | multiple))?
    | (DESKTOPCOLUMNS NUMBER)?
    | (TABLETCOLUMNS NUMBER)?
    | (PHONECOLUMNS NUMBER)?
    | (begin galleryContent end)?
```

**Responsive column properties:**

| Property | Description | Default |
|----------|-------------|---------|
| `DesktopColumns` | Grid columns on desktop | `1` |
| `TabletColumns` | Grid columns on tablet | `1` |
| `PhoneColumns` | Grid columns on phone | `1` |

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "gallery" as s1
    [*] --> s1
    state "qualifiedName?" as s2
    s1 --> s2
    state "widgetOptions?" as s3
    s2 --> s3
    s3 --> [*]
    state "gallerySource?" as s4
    [*] --> s4
    s4 --> [*]
    state "(selection (single | multiple))?" as s5
    [*] --> s5
    s5 --> [*]
    state "(begin galleryContent end)?" as s6
    [*] --> s6
    s6 --> [*]
```

---

### containerWidget

**Syntax:**

```ebnf
containerWidget
    : container widgetOptions?
    | (begin pageWidget* end)?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "container" as s1
    [*] --> s1
    state "widgetOptions?" as s2
    s1 --> s2
    s2 --> [*]
    state "(begin pageWidget* end)?" as s3
    [*] --> s3
    s3 --> [*]
```

---

### linkButtonWidget

**Syntax:**

```ebnf
linkButtonWidget
    : linkbutton widgetOptions?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "linkbutton" as s1
    [*] --> s1
    state "widgetOptions?" as s2
    s1 --> s2
    s2 --> [*]
```

---

### titleWidget

**Syntax:**

```ebnf
titleWidget
    : title STRING_LITERAL
    | | title IDENTIFIER widgetOptions?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "title" as s1
    [*] --> s1
    state "STRING_LITERAL" as s2
    s1 --> s2
    s2 --> [*]
    state "title" as s3
    [*] --> s3
    state "IDENTIFIER" as s4
    s3 --> s4
    state "widgetOptions?" as s5
    s4 --> s5
    s5 --> [*]
```

---

### dynamicTextWidget

**Syntax:**

```ebnf
dynamicTextWidget
    : dynamictext IDENTIFIER? widgetOptions?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "dynamictext" as s1
    [*] --> s1
    state "IDENTIFIER?" as s2
    s1 --> s2
    state "widgetOptions?" as s3
    s2 --> s3
    s3 --> [*]
```

---

### staticTextWidget

**Syntax:**

```ebnf
staticTextWidget
    : statictext STRING_LITERAL widgetOptions?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "statictext" as s1
    [*] --> s1
    state "STRING_LITERAL" as s2
    s1 --> s2
    state "widgetOptions?" as s3
    s2 --> s3
    s3 --> [*]
```

---

### snippetCallWidget

**Syntax:**

```ebnf
snippetCallWidget
    : snippetcall IDENTIFIER? (qualifiedName | STRING_LITERAL)
    | (passing LPAREN passArgList RPAREN)?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "snippetcall" as s1
    [*] --> s1
    state "IDENTIFIER?" as s2
    s1 --> s2
    state "(qualifiedName | STRING_LITERAL)" as s3
    s2 --> s3
    s3 --> [*]
    state "(passing LPAREN passArgList RPAREN)?" as s4
    [*] --> s4
    s4 --> [*]
```

---

### textBoxWidget

**Syntax:**

```ebnf
textBoxWidget
    : textbox IDENTIFIER? STRING_LITERAL? attributeClause? widgetOptions?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "textbox" as s1
    [*] --> s1
    state "IDENTIFIER?" as s2
    s1 --> s2
    state "STRING_LITERAL?" as s3
    s2 --> s3
    state "attributeClause?" as s4
    s3 --> s4
    state "widgetOptions?" as s5
    s4 --> s5
    s5 --> [*]
```

---

### textAreaWidget

**Syntax:**

```ebnf
textAreaWidget
    : textarea IDENTIFIER? STRING_LITERAL? attributeClause? widgetOptions?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "textarea" as s1
    [*] --> s1
    state "IDENTIFIER?" as s2
    s1 --> s2
    state "STRING_LITERAL?" as s3
    s2 --> s3
    state "attributeClause?" as s4
    s3 --> s4
    state "widgetOptions?" as s5
    s4 --> s5
    s5 --> [*]
```

---

### datePickerWidget

**Syntax:**

```ebnf
datePickerWidget
    : datepicker IDENTIFIER? STRING_LITERAL? attributeClause? widgetOptions?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "datepicker" as s1
    [*] --> s1
    state "IDENTIFIER?" as s2
    s1 --> s2
    state "STRING_LITERAL?" as s3
    s2 --> s3
    state "attributeClause?" as s4
    s3 --> s4
    state "widgetOptions?" as s5
    s4 --> s5
    s5 --> [*]
```

---

### radioButtonsWidget

**Syntax:**

```ebnf
radioButtonsWidget
    : radiobuttons IDENTIFIER? STRING_LITERAL? attributeClause? widgetOptions?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "radiobuttons" as s1
    [*] --> s1
    state "IDENTIFIER?" as s2
    s1 --> s2
    state "STRING_LITERAL?" as s3
    s2 --> s3
    state "attributeClause?" as s4
    s3 --> s4
    state "widgetOptions?" as s5
    s4 --> s5
    s5 --> [*]
```

---

### dropDownWidget

**Syntax:**

```ebnf
dropDownWidget
    : dropdown IDENTIFIER? STRING_LITERAL? attributeClause? widgetOptions?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "dropdown" as s1
    [*] --> s1
    state "IDENTIFIER?" as s2
    s1 --> s2
    state "STRING_LITERAL?" as s3
    s2 --> s3
    state "attributeClause?" as s4
    s3 --> s4
    state "widgetOptions?" as s5
    s4 --> s5
    s5 --> [*]
```

---

### comboBoxWidget

**Syntax:**

```ebnf
comboBoxWidget
    : combobox IDENTIFIER? STRING_LITERAL? attributeClause? widgetOptions?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "combobox" as s1
    [*] --> s1
    state "IDENTIFIER?" as s2
    s1 --> s2
    state "STRING_LITERAL?" as s3
    s2 --> s3
    state "attributeClause?" as s4
    s3 --> s4
    state "widgetOptions?" as s5
    s4 --> s5
    s5 --> [*]
```

---

### checkBoxWidget

**Syntax:**

```ebnf
checkBoxWidget
    : checkbox IDENTIFIER? STRING_LITERAL? attributeClause? widgetOptions?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "checkbox" as s1
    [*] --> s1
    state "IDENTIFIER?" as s2
    s1 --> s2
    state "STRING_LITERAL?" as s3
    s2 --> s3
    state "attributeClause?" as s4
    s3 --> s4
    state "widgetOptions?" as s5
    s4 --> s5
    s5 --> [*]
```

---

### referenceSelectorWidget

**Syntax:**

```ebnf
referenceSelectorWidget
    : referenceselector IDENTIFIER? STRING_LITERAL? attributeClause? widgetOptions?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "referenceselector" as s1
    [*] --> s1
    state "IDENTIFIER?" as s2
    s1 --> s2
    state "STRING_LITERAL?" as s3
    s2 --> s3
    state "attributeClause?" as s4
    s3 --> s4
    state "widgetOptions?" as s5
    s4 --> s5
    s5 --> [*]
```

---

### customWidgetWidget

**Syntax:**

```ebnf
customWidgetWidget
    : customwidget STRING_LITERAL widgetOptions?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "customwidget" as s1
    [*] --> s1
    state "STRING_LITERAL" as s2
    s1 --> s2
    state "widgetOptions?" as s3
    s2 --> s3
    s3 --> [*]
```

---

### widgetOptions

**Syntax:**

```ebnf
widgetOptions
    : LPAREN widgetOptionList RPAREN
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "LPAREN" as s1
    [*] --> s1
    state "widgetOptionList" as s2
    s1 --> s2
    state "RPAREN" as s3
    s2 --> s3
    s3 --> [*]
```

---

### widgetOptionList

**Syntax:**

```ebnf
widgetOptionList
    : widgetOption (COMMA widgetOption)*
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "widgetOption" as s1
    [*] --> s1
    state "(COMMA widgetOption)*" as s2
    s1 --> s2
    s2 --> [*]
```

---

### widgetOption

**Syntax:**

```ebnf
widgetOption
    : datasource COLON? (qualifiedName | expression)
    | | content COLON? (STRING_LITERAL | expression) templateParams?
    | | caption COLON? STRING_LITERAL templateParams?
    | | label COLON? STRING_LITERAL
    | | tooltip COLON? STRING_LITERAL
    | | class COLON? STRING_LITERAL
    | | style COLON? STRING_LITERAL
    | | width COLON? (NUMBER_LITERAL | STRING_LITERAL)
    | | height COLON? (NUMBER_LITERAL | STRING_LITERAL)
    | | editable COLON? booleanLiteral
    | | readonly COLON? booleanLiteral
    | | visible COLON? (booleanLiteral | expression)
    | | required COLON? booleanLiteral
    | | onclick COLON? (qualifiedName | expression)
    | | onchange COLON? (qualifiedName | expression)
    | | selection COLON? (single | multiple | none)
    | | rendermode COLON? (IDENTIFIER | STRING_LITERAL)
    | | icon COLON? STRING_LITERAL
    | | tabindex COLON? NUMBER_LITERAL
    | | parameters arrayLiteral                              // deprecated: use with clause
    | | IDENTIFIER COLON? NUMBER_LITERAL
    | | IDENTIFIER COLON? STRING_LITERAL
    | | IDENTIFIER COLON? booleanLiteral
    | | IDENTIFIER COLON? (IDENTIFIER | HYPHENATED_ID)
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "datasource" as s1
    [*] --> s1
    state "COLON?" as s2
    s1 --> s2
    state "(qualifiedName | expression)" as s3
    s2 --> s3
    s3 --> [*]
    state "content" as s4
    [*] --> s4
    state "COLON?" as s5
    s4 --> s5
    state "(STRING_LITERAL | expression)" as s6
    s5 --> s6
    state "templateParams?" as s7
    s6 --> s7
    s7 --> [*]
    state "caption" as s8
    [*] --> s8
    state "COLON?" as s9
    s8 --> s9
    state "STRING_LITERAL" as s10
    s9 --> s10
    state "templateParams?" as s11
    s10 --> s11
    s11 --> [*]
    state "label" as s12
    [*] --> s12
    state "COLON?" as s13
    s12 --> s13
    state "STRING_LITERAL" as s14
    s13 --> s14
    s14 --> [*]
    state "tooltip" as s15
    [*] --> s15
    state "COLON?" as s16
    s15 --> s16
    state "STRING_LITERAL" as s17
    s16 --> s17
    s17 --> [*]
    state "class" as s18
    [*] --> s18
    state "COLON?" as s19
    s18 --> s19
    state "STRING_LITERAL" as s20
    s19 --> s20
    s20 --> [*]
    state "style" as s21
    [*] --> s21
    state "COLON?" as s22
    s21 --> s22
    state "STRING_LITERAL" as s23
    s22 --> s23
    s23 --> [*]
    state "width" as s24
    [*] --> s24
    state "COLON?" as s25
    s24 --> s25
    state "(NUMBER_LITERAL | STRING_LITERAL)" as s26
    s25 --> s26
    s26 --> [*]
    state "height" as s27
    [*] --> s27
    state "COLON?" as s28
    s27 --> s28
    state "(NUMBER_LITERAL | STRING_LITERAL)" as s29
    s28 --> s29
    s29 --> [*]
    state "editable" as s30
    [*] --> s30
    state "COLON?" as s31
    s30 --> s31
    state "booleanLiteral" as s32
    s31 --> s32
    s32 --> [*]
    state "readonly" as s33
    [*] --> s33
    state "COLON?" as s34
    s33 --> s34
    state "booleanLiteral" as s35
    s34 --> s35
    s35 --> [*]
    state "visible" as s36
    [*] --> s36
    state "COLON?" as s37
    s36 --> s37
    state "(booleanLiteral | expression)" as s38
    s37 --> s38
    s38 --> [*]
    state "required" as s39
    [*] --> s39
    state "COLON?" as s40
    s39 --> s40
    state "booleanLiteral" as s41
    s40 --> s41
    s41 --> [*]
    state "onclick" as s42
    [*] --> s42
    state "COLON?" as s43
    s42 --> s43
    state "(qualifiedName | expression)" as s44
    s43 --> s44
    s44 --> [*]
    state "onchange" as s45
    [*] --> s45
    state "COLON?" as s46
    s45 --> s46
    state "(qualifiedName | expression)" as s47
    s46 --> s47
    s47 --> [*]
    state "selection" as s48
    [*] --> s48
    state "COLON?" as s49
    s48 --> s49
    state "(single | multiple | none)" as s50
    s49 --> s50
    s50 --> [*]
    state "rendermode" as s51
    [*] --> s51
    state "COLON?" as s52
    s51 --> s52
    state "(IDENTIFIER | STRING_LITERAL)" as s53
    s52 --> s53
    s53 --> [*]
    state "icon" as s54
    [*] --> s54
    state "COLON?" as s55
    s54 --> s55
    state "STRING_LITERAL" as s56
    s55 --> s56
    s56 --> [*]
    state "tabindex" as s57
    [*] --> s57
    state "COLON?" as s58
    s57 --> s58
    state "NUMBER_LITERAL" as s59
    s58 --> s59
    s59 --> [*]
    state "parameters" as s60
    [*] --> s60
    state "arrayLiteral" as s61
    s60 --> s61
    s61 --> [*]
    state "IDENTIFIER" as s62
    [*] --> s62
    state "COLON?" as s63
    s62 --> s63
    state "NUMBER_LITERAL" as s64
    s63 --> s64
    s64 --> [*]
    state "IDENTIFIER" as s65
    [*] --> s65
    state "COLON?" as s66
    s65 --> s66
    state "STRING_LITERAL" as s67
    s66 --> s67
    s67 --> [*]
    state "IDENTIFIER" as s68
    [*] --> s68
    state "COLON?" as s69
    s68 --> s69
    state "booleanLiteral" as s70
    s69 --> s70
    s70 --> [*]
    state "IDENTIFIER" as s71
    [*] --> s71
    state "COLON?" as s72
    s71 --> s72
    state "(IDENTIFIER | HYPHENATED_ID)" as s73
    s72 --> s73
    s73 --> [*]
```

---

### notebookPage

**Syntax:**

```ebnf
notebookPage
    : page qualifiedName (caption STRING_LITERAL)?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "page" as s1
    [*] --> s1
    state "qualifiedName" as s2
    s1 --> s2
    state "(caption STRING_LITERAL)?" as s3
    s2 --> s3
    s3 --> [*]
```

---

## OQL Queries

### catalogSelectQuery

**Syntax:**

```ebnf
catalogSelectQuery
    : select selectList
    | from catalog DOT catalogTableName
    | (where expression)?
    | (ORDER_BY orderByList)?
    | (limit NUMBER_LITERAL)?
    | (offset NUMBER_LITERAL)?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "select" as s1
    [*] --> s1
    state "selectList" as s2
    s1 --> s2
    s2 --> [*]
    state "from" as s3
    [*] --> s3
    state "catalog" as s4
    s3 --> s4
    state "DOT" as s5
    s4 --> s5
    state "catalogTableName" as s6
    s5 --> s6
    s6 --> [*]
    state "(where expression)?" as s7
    [*] --> s7
    s7 --> [*]
    state "(ORDER_BY orderByList)?" as s8
    [*] --> s8
    s8 --> [*]
    state "(limit NUMBER_LITERAL)?" as s9
    [*] --> s9
    s9 --> [*]
    state "(offset NUMBER_LITERAL)?" as s10
    [*] --> s10
    s10 --> [*]
```

---

### oqlQuery

OQL (Object Query Language) query for retrieving data.  OQL is similar to SQL but operates on Mendix entities and supports associations, aggregations, and subqueries.

**Syntax:**

```ebnf
oqlQuery
    : selectClause fromClause? whereClause? groupByClause? havingClause?
    | orderByClause? limitOffsetClause?
    | | fromClause whereClause? groupByClause? havingClause?
    | selectClause orderByClause? limitOffsetClause?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "selectClause" as s1
    [*] --> s1
    state "fromClause?" as s2
    s1 --> s2
    state "whereClause?" as s3
    s2 --> s3
    state "groupByClause?" as s4
    s3 --> s4
    state "havingClause?" as s5
    s4 --> s5
    s5 --> [*]
    state "orderByClause?" as s6
    [*] --> s6
    state "limitOffsetClause?" as s7
    s6 --> s7
    s7 --> [*]
    state "fromClause" as s8
    [*] --> s8
    state "whereClause?" as s9
    s8 --> s9
    state "groupByClause?" as s10
    s9 --> s10
    state "havingClause?" as s11
    s10 --> s11
    s11 --> [*]
    state "selectClause" as s12
    [*] --> s12
    state "orderByClause?" as s13
    s12 --> s13
    state "limitOffsetClause?" as s14
    s13 --> s14
    s14 --> [*]
```

**Examples:**

*Simple SELECT query:*

```sql
select Name, Email from MyModule.Customer
```

*Query with WHERE clause:*

```sql
select c.Name, c.Email
from MyModule.Customer as c
where c.Active = true and c.Age > 18
```

*Query with JOIN via association:*

```sql
select o.OrderNumber, c.Name as CustomerName
from MyModule.Order as o
inner join o/MyModule.Order_Customer/MyModule.Customer as c
where o.Status = 'Completed'
```

*Aggregation query:*

```sql
select c.Country, count(*) as CustomerCount, avg(c.Age) as AvgAge
from MyModule.Customer as c
GROUP by c.Country
having count(*) > 10
ORDER by CustomerCount desc
```

*Subquery:*

```sql
select p.Name, p.Price
from MyModule.Product as p
where p.Price > (select avg(p2.Price) from MyModule.Product as p2)
```

**See also:** [createEntityStatement for using OQL in VIEW entities](#createentitystatement-for-using-oql-in-view-entities), [retrieveStatement for using OQL in microflows](#retrievestatement-for-using-oql-in-microflows)

---

### selectClause

**Syntax:**

```ebnf
selectClause
    : select (distinct | all)? selectList
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "select" as s1
    [*] --> s1
    state "(distinct | all)?" as s2
    s1 --> s2
    state "selectList" as s3
    s2 --> s3
    s3 --> [*]
```

---

### selectList

**Syntax:**

```ebnf
selectList
    : STAR
    | | selectItem (COMMA selectItem)*
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "STAR" as s1
    [*] --> s1
    s1 --> [*]
    state "selectItem" as s2
    [*] --> s2
    state "(COMMA selectItem)*" as s3
    s2 --> s3
    s3 --> [*]
```

---

### selectItem

**Syntax:**

```ebnf
selectItem
    : expression (as selectAlias)?
    | | aggregateFunction (as selectAlias)?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "expression" as s1
    [*] --> s1
    state "(as selectAlias)?" as s2
    s1 --> s2
    s2 --> [*]
    state "aggregateFunction" as s3
    [*] --> s3
    state "(as selectAlias)?" as s4
    s3 --> s4
    s4 --> [*]
```

---

### selectAlias

**Syntax:**

```ebnf
selectAlias
    : IDENTIFIER
    | | status | type | value | index
    | | username | password
    | | action | message
    | | owner | reference | cascade
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "IDENTIFIER" as s1
    [*] --> s1
    s1 --> [*]
    state "status" as s2
    [*] --> s2
    state "|" as s3
    s2 --> s3
    state "type" as s4
    s3 --> s4
    state "|" as s5
    s4 --> s5
    state "value" as s6
    s5 --> s6
    state "|" as s7
    s6 --> s7
    state "index" as s8
    s7 --> s8
    s8 --> [*]
    state "username" as s9
    [*] --> s9
    state "|" as s10
    s9 --> s10
    state "password" as s11
    s10 --> s11
    s11 --> [*]
    state "action" as s12
    [*] --> s12
    state "|" as s13
    s12 --> s13
    state "message" as s14
    s13 --> s14
    s14 --> [*]
    state "owner" as s15
    [*] --> s15
    state "|" as s16
    s15 --> s16
    state "reference" as s17
    s16 --> s17
    state "|" as s18
    s17 --> s18
    state "cascade" as s19
    s18 --> s19
    s19 --> [*]
```

---

### fromClause

**Syntax:**

```ebnf
fromClause
    : from tableReference (joinClause)*
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "from" as s1
    [*] --> s1
    state "tableReference" as s2
    s1 --> s2
    state "(joinClause)*" as s3
    s2 --> s3
    s3 --> [*]
```

---

### whereClause

**Syntax:**

```ebnf
whereClause
    : where expression
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "where" as s1
    [*] --> s1
    state "expression" as s2
    s1 --> s2
    s2 --> [*]
```

---

## Expressions

### expression

**Syntax:**

```ebnf
expression
    : orExpression
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "orExpression" as s1
    [*] --> s1
    s1 --> [*]
```

---

### orExpression

**Syntax:**

```ebnf
orExpression
    : andExpression (or andExpression)*
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "andExpression" as s1
    [*] --> s1
    state "(or andExpression)*" as s2
    s1 --> s2
    s2 --> [*]
```

---

### andExpression

**Syntax:**

```ebnf
andExpression
    : notExpression (and notExpression)*
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "notExpression" as s1
    [*] --> s1
    state "(and notExpression)*" as s2
    s1 --> s2
    s2 --> [*]
```

---

### notExpression

**Syntax:**

```ebnf
notExpression
    : not? comparisonExpression
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "not?" as s1
    [*] --> s1
    state "comparisonExpression" as s2
    s1 --> s2
    s2 --> [*]
```

---

### comparisonExpression

**Syntax:**

```ebnf
comparisonExpression
    : additiveExpression
    | ( comparisonOperator additiveExpression
    | | IS_NULL
    | | IS_NOT_NULL
    | | in LPAREN (oqlQuery | expressionList) RPAREN
    | | not? between additiveExpression and additiveExpression
    | | not? like additiveExpression
    | )?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "additiveExpression" as s1
    [*] --> s1
    s1 --> [*]
    state "( comparisonOperator additiveExpression" as s2
    [*] --> s2
    s2 --> [*]
    state "IS_NULL" as s3
    [*] --> s3
    s3 --> [*]
    state "IS_NOT_NULL" as s4
    [*] --> s4
    s4 --> [*]
    state "in" as s5
    [*] --> s5
    state "LPAREN" as s6
    s5 --> s6
    state "(oqlQuery | expressionList)" as s7
    s6 --> s7
    state "RPAREN" as s8
    s7 --> s8
    s8 --> [*]
    state "not?" as s9
    [*] --> s9
    state "between" as s10
    s9 --> s10
    state "additiveExpression" as s11
    s10 --> s11
    state "and" as s12
    s11 --> s12
    state "additiveExpression" as s13
    s12 --> s13
    s13 --> [*]
    state "not?" as s14
    [*] --> s14
    state "like" as s15
    s14 --> s15
    state "additiveExpression" as s16
    s15 --> s16
    s16 --> [*]
    state ")?" as s17
    [*] --> s17
    s17 --> [*]
```

---

### comparisonOperator

**Syntax:**

```ebnf
comparisonOperator
    : equals
    | | NOT_EQUALS
    | | LESS_THAN
    | | LESS_THAN_OR_EQUAL
    | | GREATER_THAN
    | | GREATER_THAN_OR_EQUAL
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "equals" as s1
    [*] --> s1
    s1 --> [*]
    state "NOT_EQUALS" as s2
    [*] --> s2
    s2 --> [*]
    state "LESS_THAN" as s3
    [*] --> s3
    s3 --> [*]
    state "LESS_THAN_OR_EQUAL" as s4
    [*] --> s4
    s4 --> [*]
    state "GREATER_THAN" as s5
    [*] --> s5
    s5 --> [*]
    state "GREATER_THAN_OR_EQUAL" as s6
    [*] --> s6
    s6 --> [*]
```

---

### additiveExpression

**Syntax:**

```ebnf
additiveExpression
    : multiplicativeExpression ((PLUS | MINUS) multiplicativeExpression)*
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "multiplicativeExpression" as s1
    [*] --> s1
    state "((PLUS | MINUS) multiplicativeExpression)*" as s2
    s1 --> s2
    s2 --> [*]
```

---

### multiplicativeExpression

**Syntax:**

```ebnf
multiplicativeExpression
    : unaryExpression ((STAR | SLASH | COLON | PERCENT | mod | div) unaryExpression)*  // COLON is OQL division
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "unaryExpression" as s1
    [*] --> s1
    state "((STAR | SLASH | COLON | PERCENT | mod | div) unaryExpression)*" as s2
    s1 --> s2
    s2 --> [*]
```

---

### unaryExpression

**Syntax:**

```ebnf
unaryExpression
    : (PLUS | MINUS)? primaryExpression
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "(PLUS | MINUS)?" as s1
    [*] --> s1
    state "primaryExpression" as s2
    s1 --> s2
    s2 --> [*]
```

---

### primaryExpression

**Syntax:**

```ebnf
primaryExpression
    : LPAREN expression RPAREN
    | | LPAREN oqlQuery RPAREN          // Scalar subquery
    | | caseExpression
    | | aggregateFunction
    | | functionCall
    | | atomicExpression
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "LPAREN" as s1
    [*] --> s1
    state "expression" as s2
    s1 --> s2
    state "RPAREN" as s3
    s2 --> s3
    s3 --> [*]
    state "LPAREN" as s4
    [*] --> s4
    state "oqlQuery" as s5
    s4 --> s5
    state "RPAREN" as s6
    s5 --> s6
    s6 --> [*]
    state "caseExpression" as s7
    [*] --> s7
    s7 --> [*]
    state "aggregateFunction" as s8
    [*] --> s8
    s8 --> [*]
    state "functionCall" as s9
    [*] --> s9
    s9 --> [*]
    state "atomicExpression" as s10
    [*] --> s10
    s10 --> [*]
```

---

### caseExpression

**Syntax:**

```ebnf
caseExpression
    : case
    | (when expression then expression)+
    | (else expression)?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "case" as s1
    [*] --> s1
    s1 --> [*]
    state "(when expression then expression)+" as s2
    [*] --> s2
    s2 --> [*]
    state "(else expression)?" as s3
    [*] --> s3
    s3 --> [*]
```

---

### atomicExpression

**Syntax:**

```ebnf
atomicExpression
    : literal
    | | VARIABLE (DOT attributename)*    // $Var or $Widget.Attribute (data source ref)
    | | qualifiedName
    | | IDENTIFIER
    | | MENDIX_TOKEN
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "literal" as s1
    [*] --> s1
    s1 --> [*]
    state "VARIABLE" as s2
    [*] --> s2
    state "(DOT attributename)*" as s3
    s2 --> s3
    s3 --> [*]
    state "qualifiedName" as s4
    [*] --> s4
    s4 --> [*]
    state "IDENTIFIER" as s5
    [*] --> s5
    s5 --> [*]
    state "MENDIX_TOKEN" as s6
    [*] --> s6
    s6 --> [*]
```

---

### expressionList

**Syntax:**

```ebnf
expressionList
    : expression (COMMA expression)*
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "expression" as s1
    [*] --> s1
    state "(COMMA expression)*" as s2
    s1 --> s2
    s2 --> [*]
```

---

### literal

Literal values

**Syntax:**

```ebnf
literal
    : STRING_LITERAL
    | | NUMBER_LITERAL
    | | booleanLiteral
    | | null
    | | empty
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "STRING_LITERAL" as s1
    [*] --> s1
    s1 --> [*]
    state "NUMBER_LITERAL" as s2
    [*] --> s2
    s2 --> [*]
    state "booleanLiteral" as s3
    [*] --> s3
    s3 --> [*]
    state "null" as s4
    [*] --> s4
    s4 --> [*]
    state "empty" as s5
    [*] --> s5
    s5 --> [*]
```

---

### arrayLiteral

**Syntax:**

```ebnf
arrayLiteral
    : LBRACKET (literal (COMMA literal)*)? RBRACKET
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "LBRACKET" as s1
    [*] --> s1
    state "(literal (COMMA literal)*)?" as s2
    s1 --> s2
    state "RBRACKET" as s3
    s2 --> s3
    s3 --> [*]
```

---

### booleanLiteral

**Syntax:**

```ebnf
booleanLiteral
    : true
    | | false
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "true" as s1
    [*] --> s1
    s1 --> [*]
    state "false" as s2
    [*] --> s2
    s2 --> [*]
```

---

## Other Rules

### dataType

Specifies the data type for an attribute.  MDL supports all Mendix primitive types, enumerations, and entity references.

**Syntax:**

```ebnf
dataType
    : STRING_TYPE (LPAREN NUMBER_LITERAL RPAREN)?
    | | INTEGER_TYPE
    | | LONG_TYPE
    | | DECIMAL_TYPE
    | | BOOLEAN_TYPE
    | | DATETIME_TYPE
    | | DATE_TYPE
    | | AUTONUMBER_TYPE
    | | BINARY_TYPE
    | | HASHEDSTRING_TYPE
    | | CURRENCY_TYPE
    | | FLOAT_TYPE
    | | ENUM_TYPE qualifiedName
    | | enumeration LPAREN qualifiedName RPAREN  // enumeration(Module.Enum) syntax
    | | LIST_OF qualifiedName
    | | qualifiedName  // entity reference type
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "STRING_TYPE" as s1
    [*] --> s1
    state "(LPAREN NUMBER_LITERAL RPAREN)?" as s2
    s1 --> s2
    s2 --> [*]
    state "INTEGER_TYPE" as s3
    [*] --> s3
    s3 --> [*]
    state "LONG_TYPE" as s4
    [*] --> s4
    s4 --> [*]
    state "DECIMAL_TYPE" as s5
    [*] --> s5
    s5 --> [*]
    state "BOOLEAN_TYPE" as s6
    [*] --> s6
    s6 --> [*]
    state "DATETIME_TYPE" as s7
    [*] --> s7
    s7 --> [*]
    state "DATE_TYPE" as s8
    [*] --> s8
    s8 --> [*]
    state "AUTONUMBER_TYPE" as s9
    [*] --> s9
    s9 --> [*]
    state "BINARY_TYPE" as s10
    [*] --> s10
    s10 --> [*]
    state "HASHEDSTRING_TYPE" as s11
    [*] --> s11
    s11 --> [*]
    state "CURRENCY_TYPE" as s12
    [*] --> s12
    s12 --> [*]
    state "FLOAT_TYPE" as s13
    [*] --> s13
    s13 --> [*]
    state "ENUM_TYPE" as s14
    [*] --> s14
    state "qualifiedName" as s15
    s14 --> s15
    s15 --> [*]
    state "enumeration" as s16
    [*] --> s16
    state "LPAREN" as s17
    s16 --> s17
    state "qualifiedName" as s18
    s17 --> s18
    state "RPAREN" as s19
    s18 --> s19
    s19 --> [*]
    state "LIST_OF" as s20
    [*] --> s20
    state "qualifiedName" as s21
    s20 --> s21
    s21 --> [*]
    state "qualifiedName" as s22
    [*] --> s22
    s22 --> [*]
```

**Examples:**

*Primitive types:*

```sql
Name: string(200),       -- String with max length 200
Age: integer,            -- 32-bit integer
Total: decimal,          -- Fixed-point decimal
Active: boolean,         -- true/false
created: datetime,       -- Date and time
BirthDate: date,         -- Date only
Counter: autonumber,     -- Auto-incrementing number
data: binary,            -- Binary data (files)
password: hashedstring   -- Securely hashed string
```

*Enumeration types:*

```sql
status: enum MyModule.OrderStatus,
Priority: enumeration(MyModule.Priority)
```

*Entity references:*

```sql
Customer: MyModule.Customer,       -- Single reference
Items: list of MyModule.OrderItem  -- List of references
```

---

### indexDefinition

**Syntax:**

```ebnf
indexDefinition
    : IDENTIFIER? LPAREN indexAttributeList RPAREN
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "IDENTIFIER?" as s1
    [*] --> s1
    state "LPAREN" as s2
    s1 --> s2
    state "indexAttributeList" as s3
    s2 --> s3
    state "RPAREN" as s4
    s3 --> s4
    s4 --> [*]
```

---

### indexColumnName

**Syntax:**

```ebnf
indexColumnName
    : IDENTIFIER
    | | status | type | value | index
    | | username | password
    | | action | message
    | | owner | reference | cascade
    | | success | error | warning | info | debug | critical
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "IDENTIFIER" as s1
    [*] --> s1
    s1 --> [*]
    state "status" as s2
    [*] --> s2
    state "|" as s3
    s2 --> s3
    state "type" as s4
    s3 --> s4
    state "|" as s5
    s4 --> s5
    state "value" as s6
    s5 --> s6
    state "|" as s7
    s6 --> s7
    state "index" as s8
    s7 --> s8
    s8 --> [*]
    state "username" as s9
    [*] --> s9
    state "|" as s10
    s9 --> s10
    state "password" as s11
    s10 --> s11
    s11 --> [*]
    state "action" as s12
    [*] --> s12
    state "|" as s13
    s12 --> s13
    state "message" as s14
    s13 --> s14
    s14 --> [*]
    state "owner" as s15
    [*] --> s15
    state "|" as s16
    s15 --> s16
    state "reference" as s17
    s16 --> s17
    state "|" as s18
    s17 --> s18
    state "cascade" as s19
    s18 --> s19
    s19 --> [*]
    state "success" as s20
    [*] --> s20
    state "|" as s21
    s20 --> s21
    state "error" as s22
    s21 --> s22
    state "|" as s23
    s22 --> s23
    state "warning" as s24
    s23 --> s24
    state "|" as s25
    s24 --> s25
    state "info" as s26
    s25 --> s26
    state "|" as s27
    s26 --> s27
    state "debug" as s28
    s27 --> s28
    state "|" as s29
    s28 --> s29
    state "critical" as s30
    s29 --> s30
    s30 --> [*]
```

---

### deleteBehavior

**Syntax:**

```ebnf
deleteBehavior
    : DELETE_AND_REFERENCES
    | | DELETE_BUT_KEEP_REFERENCES
    | | DELETE_IF_NO_REFERENCES
    | | cascade
    | | prevent
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "DELETE_AND_REFERENCES" as s1
    [*] --> s1
    s1 --> [*]
    state "DELETE_BUT_KEEP_REFERENCES" as s2
    [*] --> s2
    s2 --> [*]
    state "DELETE_IF_NO_REFERENCES" as s3
    [*] --> s3
    s3 --> [*]
    state "cascade" as s4
    [*] --> s4
    s4 --> [*]
    state "prevent" as s5
    [*] --> s5
    s5 --> [*]
```

---

### moduleOptions

**Syntax:**

```ebnf
moduleOptions
    : moduleOption+
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "moduleOption+" as s1
    [*] --> s1
    s1 --> [*]
```

---

### moduleOption

**Syntax:**

```ebnf
moduleOption
    : comment STRING_LITERAL
    | | folder STRING_LITERAL
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "comment" as s1
    [*] --> s1
    state "STRING_LITERAL" as s2
    s1 --> s2
    s2 --> [*]
    state "folder" as s3
    [*] --> s3
    state "STRING_LITERAL" as s4
    s3 --> s4
    s4 --> [*]
```

---

### enumerationValueList

**Syntax:**

```ebnf
enumerationValueList
    : enumerationValue (COMMA enumerationValue)*
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "enumerationValue" as s1
    [*] --> s1
    state "(COMMA enumerationValue)*" as s2
    s1 --> s2
    s2 --> [*]
```

---

### enumerationValue

**Syntax:**

```ebnf
enumerationValue
    : docComment? enumValueName (caption? STRING_LITERAL)?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "docComment?" as s1
    [*] --> s1
    state "enumValueName" as s2
    s1 --> s2
    state "(caption? STRING_LITERAL)?" as s3
    s2 --> s3
    s3 --> [*]
```

---

### enumValueName

**Syntax:**

```ebnf
enumValueName
    : IDENTIFIER
    | | service | services  // Common keywords that might be used as enum values
    | | status | type | value | index
    | | critical | success | error | warning | info | debug  // log level keywords
    | | message | action | username | password
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "IDENTIFIER" as s1
    [*] --> s1
    s1 --> [*]
    state "service" as s2
    [*] --> s2
    state "|" as s3
    s2 --> s3
    state "services" as s4
    s3 --> s4
    s4 --> [*]
    state "status" as s5
    [*] --> s5
    state "|" as s6
    s5 --> s6
    state "type" as s7
    s6 --> s7
    state "|" as s8
    s7 --> s8
    state "value" as s9
    s8 --> s9
    state "|" as s10
    s9 --> s10
    state "index" as s11
    s10 --> s11
    s11 --> [*]
    state "critical" as s12
    [*] --> s12
    state "|" as s13
    s12 --> s13
    state "success" as s14
    s13 --> s14
    state "|" as s15
    s14 --> s15
    state "error" as s16
    s15 --> s16
    state "|" as s17
    s16 --> s17
    state "warning" as s18
    s17 --> s18
    state "|" as s19
    s18 --> s19
    state "info" as s20
    s19 --> s20
    state "|" as s21
    s20 --> s21
    state "debug" as s22
    s21 --> s22
    s22 --> [*]
    state "message" as s23
    [*] --> s23
    state "|" as s24
    s23 --> s24
    state "action" as s25
    s24 --> s25
    state "|" as s26
    s25 --> s26
    state "username" as s27
    s26 --> s27
    state "|" as s28
    s27 --> s28
    state "password" as s29
    s28 --> s29
    s29 --> [*]
```

---

### enumerationOptions

**Syntax:**

```ebnf
enumerationOptions
    : enumerationOption+
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "enumerationOption+" as s1
    [*] --> s1
    s1 --> [*]
```

---

### enumerationOption

**Syntax:**

```ebnf
enumerationOption
    : comment STRING_LITERAL
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "comment" as s1
    [*] --> s1
    state "STRING_LITERAL" as s2
    s1 --> s2
    s2 --> [*]
```

---

### validationRuleBody

**Syntax:**

```ebnf
validationRuleBody
    : expression expression feedback STRING_LITERAL
    | | required attributeReference feedback STRING_LITERAL
    | | unique attributeReferenceList feedback STRING_LITERAL
    | | range attributeReference rangeConstraint feedback STRING_LITERAL
    | | regex attributeReference STRING_LITERAL feedback STRING_LITERAL
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "expression" as s1
    [*] --> s1
    state "expression" as s2
    s1 --> s2
    state "feedback" as s3
    s2 --> s3
    state "STRING_LITERAL" as s4
    s3 --> s4
    s4 --> [*]
    state "required" as s5
    [*] --> s5
    state "attributeReference" as s6
    s5 --> s6
    state "feedback" as s7
    s6 --> s7
    state "STRING_LITERAL" as s8
    s7 --> s8
    s8 --> [*]
    state "unique" as s9
    [*] --> s9
    state "attributeReferenceList" as s10
    s9 --> s10
    state "feedback" as s11
    s10 --> s11
    state "STRING_LITERAL" as s12
    s11 --> s12
    s12 --> [*]
    state "range" as s13
    [*] --> s13
    state "attributeReference" as s14
    s13 --> s14
    state "rangeConstraint" as s15
    s14 --> s15
    state "feedback" as s16
    s15 --> s16
    state "STRING_LITERAL" as s17
    s16 --> s17
    s17 --> [*]
    state "regex" as s18
    [*] --> s18
    state "attributeReference" as s19
    s18 --> s19
    state "STRING_LITERAL" as s20
    s19 --> s20
    state "feedback" as s21
    s20 --> s21
    state "STRING_LITERAL" as s22
    s21 --> s22
    s22 --> [*]
```

---

### rangeConstraint

**Syntax:**

```ebnf
rangeConstraint
    : between literal and literal
    | | LESS_THAN literal
    | | LESS_THAN_OR_EQUAL literal
    | | GREATER_THAN literal
    | | GREATER_THAN_OR_EQUAL literal
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "between" as s1
    [*] --> s1
    state "literal" as s2
    s1 --> s2
    state "and" as s3
    s2 --> s3
    state "literal" as s4
    s3 --> s4
    s4 --> [*]
    state "LESS_THAN" as s5
    [*] --> s5
    state "literal" as s6
    s5 --> s6
    s6 --> [*]
    state "LESS_THAN_OR_EQUAL" as s7
    [*] --> s7
    state "literal" as s8
    s7 --> s8
    s8 --> [*]
    state "GREATER_THAN" as s9
    [*] --> s9
    state "literal" as s10
    s9 --> s10
    s10 --> [*]
    state "GREATER_THAN_OR_EQUAL" as s11
    [*] --> s11
    state "literal" as s12
    s11 --> s12
    s12 --> [*]
```

---

### retrieveSource

**Syntax:**

```ebnf
retrieveSource
    : qualifiedName
    | | LPAREN oqlQuery RPAREN
    | | database STRING_LITERAL
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "qualifiedName" as s1
    [*] --> s1
    s1 --> [*]
    state "LPAREN" as s2
    [*] --> s2
    state "oqlQuery" as s3
    s2 --> s3
    state "RPAREN" as s4
    s3 --> s4
    s4 --> [*]
    state "database" as s5
    [*] --> s5
    state "STRING_LITERAL" as s6
    s5 --> s6
    s6 --> [*]
```

---

### logLevel

**Syntax:**

```ebnf
logLevel
    : info
    | | warning
    | | error
    | | debug
    | | trace
    | | critical
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "info" as s1
    [*] --> s1
    s1 --> [*]
    state "warning" as s2
    [*] --> s2
    s2 --> [*]
    state "error" as s3
    [*] --> s3
    s3 --> [*]
    state "debug" as s4
    [*] --> s4
    s4 --> [*]
    state "trace" as s5
    [*] --> s5
    s5 --> [*]
    state "critical" as s6
    [*] --> s6
    s6 --> [*]
```

---

### templateParams

**Syntax:**

```ebnf
templateParams
    : with LPAREN templateParam (COMMA templateParam)* RPAREN    // with ({1} = $var)
    | | parameters arrayLiteral                                     // parameters ['val'] (deprecated)
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "with" as s1
    [*] --> s1
    state "LPAREN" as s2
    s1 --> s2
    state "templateParam" as s3
    s2 --> s3
    state "(COMMA templateParam)*" as s4
    s3 --> s4
    state "RPAREN" as s5
    s4 --> s5
    s5 --> [*]
    state "parameters" as s6
    [*] --> s6
    state "arrayLiteral" as s7
    s6 --> s7
    s7 --> [*]
```

---

### templateParam

**Syntax:**

```ebnf
templateParam
    : LBRACE NUMBER_LITERAL RBRACE equals expression
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "LBRACE" as s1
    [*] --> s1
    state "NUMBER_LITERAL" as s2
    s1 --> s2
    state "RBRACE" as s3
    s2 --> s3
    state "equals" as s4
    s3 --> s4
    state "expression" as s5
    s4 --> s5
    s5 --> [*]
```

---

### logTemplateParams

**Syntax:**

```ebnf
logTemplateParams
    : templateParams;
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "templateParams;" as s1
    [*] --> s1
    s1 --> [*]
```

---

### logTemplateParam

**Syntax:**

```ebnf
logTemplateParam
    : templateParam;
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "templateParam;" as s1
    [*] --> s1
    s1 --> [*]
```

---

### callArgumentList

**Syntax:**

```ebnf
callArgumentList
    : callArgument (COMMA callArgument)*
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "callArgument" as s1
    [*] --> s1
    state "(COMMA callArgument)*" as s2
    s1 --> s2
    s2 --> [*]
```

---

### callArgument

**Syntax:**

```ebnf
callArgument
    : (VARIABLE | IDENTIFIER) equals expression
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "(VARIABLE | IDENTIFIER)" as s1
    [*] --> s1
    state "equals" as s2
    s1 --> s2
    state "expression" as s3
    s2 --> s3
    s3 --> [*]
```

---

### memberAssignmentList

**Syntax:**

```ebnf
memberAssignmentList
    : memberAssignment (COMMA memberAssignment)*
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "memberAssignment" as s1
    [*] --> s1
    state "(COMMA memberAssignment)*" as s2
    s1 --> s2
    s2 --> [*]
```

---

### memberAssignment

**Syntax:**

```ebnf
memberAssignment
    : memberAttributeName equals expression
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "memberAttributeName" as s1
    [*] --> s1
    state "equals" as s2
    s1 --> s2
    state "expression" as s3
    s2 --> s3
    s3 --> [*]
```

---

### changeList

**Syntax:**

```ebnf
changeList
    : changeItem (COMMA changeItem)*
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "changeItem" as s1
    [*] --> s1
    state "(COMMA changeItem)*" as s2
    s1 --> s2
    s2 --> [*]
```

---

### changeItem

**Syntax:**

```ebnf
changeItem
    : IDENTIFIER equals expression
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "IDENTIFIER" as s1
    [*] --> s1
    state "equals" as s2
    s1 --> s2
    state "expression" as s3
    s2 --> s3
    s3 --> [*]
```

---

### placeholderBlock

**Syntax:**

```ebnf
placeholderBlock
    : placeholder (IDENTIFIER | STRING_LITERAL) begin (pageWidget SEMICOLON?)* end
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "placeholder" as s1
    [*] --> s1
    state "(IDENTIFIER | STRING_LITERAL)" as s2
    s1 --> s2
    state "begin" as s3
    s2 --> s3
    state "(pageWidget SEMICOLON?)*" as s4
    s3 --> s4
    state "end" as s5
    s4 --> s5
    s5 --> [*]
```

---

### dataGridSource

**Syntax:**

```ebnf
dataGridSource
    : SOURCE_KW database qualifiedName
    | (where expression)?
    | (SORT_BY IDENTIFIER (asc | desc)?)?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "SOURCE_KW" as s1
    [*] --> s1
    state "database" as s2
    s1 --> s2
    state "qualifiedName" as s3
    s2 --> s3
    s3 --> [*]
    state "(where expression)?" as s4
    [*] --> s4
    s4 --> [*]
    state "(SORT_BY IDENTIFIER (asc | desc)?)?" as s5
    [*] --> s5
    s5 --> [*]
```

---

### dataGridContent

**Syntax:**

```ebnf
dataGridContent
    : (dataGridHeader | dataGridColumn SEMICOLON? | controlbar | searchbar)*
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "(dataGridHeader | dataGridColumn SEMICOLON? | controlbar | searchbar)*" as s1
    [*] --> s1
    s1 --> [*]
```

---

### dataGridHeader

**Syntax:**

```ebnf
dataGridHeader
    : header (pageWidget SEMICOLON?)*
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "header" as s1
    [*] --> s1
    state "(pageWidget SEMICOLON?)*" as s2
    s1 --> s2
    s2 --> [*]
```

---

### dataGridColumn

**Syntax:**

```ebnf
dataGridColumn
    : column IDENTIFIER (as STRING_LITERAL)? widgetOptions?
    | | column STRING_LITERAL (begin (pageWidget SEMICOLON?)* end)?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "column" as s1
    [*] --> s1
    state "IDENTIFIER" as s2
    s1 --> s2
    state "(as STRING_LITERAL)?" as s3
    s2 --> s3
    state "widgetOptions?" as s4
    s3 --> s4
    s4 --> [*]
    state "column" as s5
    [*] --> s5
    state "STRING_LITERAL" as s6
    s5 --> s6
    state "(begin (pageWidget SEMICOLON?)* end)?" as s7
    s6 --> s7
    s7 --> [*]
```

---

### controlBar

**Syntax:**

```ebnf
controlbar
    : controlbar begin (actionButtonWidget | linkButtonWidget)* end
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "controlbar" as s1
    [*] --> s1
    state "begin" as s2
    s1 --> s2
    state "(actionButtonWidget | linkButtonWidget)*" as s3
    s2 --> s3
    state "end" as s4
    s3 --> s4
    s4 --> [*]
```

---

### searchBar

**Syntax:**

```ebnf
searchbar
    : searchbar begin searchField* end
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "searchbar" as s1
    [*] --> s1
    state "begin" as s2
    s1 --> s2
    state "searchField*" as s3
    s2 --> s3
    state "end" as s4
    s3 --> s4
    s4 --> [*]
```

---

### searchField

**Syntax:**

```ebnf
searchField
    : textfilter IDENTIFIER attributeClause? widgetOptions?
    | | dropdown IDENTIFIER attributeClause? widgetOptions?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "textfilter" as s1
    [*] --> s1
    state "IDENTIFIER" as s2
    s1 --> s2
    state "attributeClause?" as s3
    s2 --> s3
    state "widgetOptions?" as s4
    s3 --> s4
    s4 --> [*]
    state "dropdown" as s5
    [*] --> s5
    state "IDENTIFIER" as s6
    s5 --> s6
    state "attributeClause?" as s7
    s6 --> s7
    state "widgetOptions?" as s8
    s7 --> s8
    s8 --> [*]
```

---

### dataSourceClause

**Syntax:**

```ebnf
dataSourceClause
    : datasource (VARIABLE | microflow STRING_LITERAL | selection IDENTIFIER)
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "datasource" as s1
    [*] --> s1
    state "(VARIABLE | microflow STRING_LITERAL | selection IDENTIFIER)" as s2
    s1 --> s2
    s2 --> [*]
```

---

### dataViewFooter

**Syntax:**

```ebnf
dataViewFooter
    : footer
    | ( begin (pageWidget SEMICOLON?)* end
    | | (pageWidget SEMICOLON?)+
    | )
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "footer" as s1
    [*] --> s1
    s1 --> [*]
    state "( begin (pageWidget SEMICOLON?)* end" as s2
    [*] --> s2
    s2 --> [*]
    state "(pageWidget SEMICOLON?)+" as s3
    [*] --> s3
    s3 --> [*]
    state ")" as s4
    [*] --> s4
    s4 --> [*]
```

---

### gallerySource

**Syntax:**

```ebnf
gallerySource
    : SOURCE_KW database qualifiedName
    | (SORT_BY IDENTIFIER (asc | desc)?)?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "SOURCE_KW" as s1
    [*] --> s1
    state "database" as s2
    s1 --> s2
    state "qualifiedName" as s3
    s2 --> s3
    s3 --> [*]
    state "(SORT_BY IDENTIFIER (asc | desc)?)?" as s4
    [*] --> s4
    s4 --> [*]
```

---

### galleryContent

**Syntax:**

```ebnf
galleryContent
    : (galleryFilter)?
    | (template (pageWidget SEMICOLON?)*)?
    | (pageWidget SEMICOLON?)*
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "(galleryFilter)?" as s1
    [*] --> s1
    s1 --> [*]
    state "(template (pageWidget SEMICOLON?)*)?" as s2
    [*] --> s2
    s2 --> [*]
    state "(pageWidget SEMICOLON?)*" as s3
    [*] --> s3
    s3 --> [*]
```

---

### galleryFilter

**Syntax:**

```ebnf
galleryFilter
    : filter (searchField SEMICOLON?)*
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "filter" as s1
    [*] --> s1
    state "(searchField SEMICOLON?)*" as s2
    s1 --> s2
    s2 --> [*]
```

---

### passArgList

**Syntax:**

```ebnf
passArgList
    : passArg (COMMA passArg)*
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "passArg" as s1
    [*] --> s1
    state "(COMMA passArg)*" as s2
    s1 --> s2
    s2 --> [*]
```

---

### passArg

**Syntax:**

```ebnf
passArg
    : VARIABLE equals (VARIABLE | expression)
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "VARIABLE" as s1
    [*] --> s1
    state "equals" as s2
    s1 --> s2
    state "(VARIABLE | expression)" as s3
    s2 --> s3
    s3 --> [*]
```

---

### buttonStyle

**Syntax:**

```ebnf
buttonstyle
    : primary
    | | default
    | | success
    | | danger
    | | WARNING_STYLE
    | | INFO_STYLE
    | | IDENTIFIER
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "primary" as s1
    [*] --> s1
    s1 --> [*]
    state "default" as s2
    [*] --> s2
    s2 --> [*]
    state "success" as s3
    [*] --> s3
    s3 --> [*]
    state "danger" as s4
    [*] --> s4
    s4 --> [*]
    state "WARNING_STYLE" as s5
    [*] --> s5
    s5 --> [*]
    state "INFO_STYLE" as s6
    [*] --> s6
    s6 --> [*]
    state "IDENTIFIER" as s7
    [*] --> s7
    s7 --> [*]
```

---

### snippetParameterList

**Syntax:**

```ebnf
snippetParameterList
    : snippetParameter (COMMA snippetParameter)*
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "snippetParameter" as s1
    [*] --> s1
    state "(COMMA snippetParameter)*" as s2
    s1 --> s2
    s2 --> [*]
```

---

### snippetParameter

**Syntax:**

```ebnf
snippetParameter
    : (IDENTIFIER | VARIABLE) COLON dataType
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "(IDENTIFIER | VARIABLE)" as s1
    [*] --> s1
    state "COLON" as s2
    s1 --> s2
    state "dataType" as s3
    s2 --> s3
    s3 --> [*]
```

---

### snippetOptions

**Syntax:**

```ebnf
snippetOptions
    : snippetOption+
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "snippetOption+" as s1
    [*] --> s1
    s1 --> [*]
```

---

### snippetOption

**Syntax:**

```ebnf
snippetOption
    : folder STRING_LITERAL
    | | comment STRING_LITERAL
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "folder" as s1
    [*] --> s1
    state "STRING_LITERAL" as s2
    s1 --> s2
    s2 --> [*]
    state "comment" as s3
    [*] --> s3
    state "STRING_LITERAL" as s4
    s3 --> s4
    s4 --> [*]
```

---

### notebookOptions

**Syntax:**

```ebnf
notebookOptions
    : notebookOption+
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "notebookOption+" as s1
    [*] --> s1
    s1 --> [*]
```

---

### notebookOption

**Syntax:**

```ebnf
notebookOption
    : comment STRING_LITERAL
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "comment" as s1
    [*] --> s1
    state "STRING_LITERAL" as s2
    s1 --> s2
    s2 --> [*]
```

---

### databaseConnectionOptions

**Syntax:**

```ebnf
databaseConnectionOptions
    : databaseConnectionOption+
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "databaseConnectionOption+" as s1
    [*] --> s1
    s1 --> [*]
```

---

### databaseConnectionOption

**Syntax:**

```ebnf
databaseConnectionOption
    : type IDENTIFIER
    | | host STRING_LITERAL
    | | port NUMBER_LITERAL
    | | database STRING_LITERAL
    | | username STRING_LITERAL
    | | password STRING_LITERAL
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "type" as s1
    [*] --> s1
    state "IDENTIFIER" as s2
    s1 --> s2
    s2 --> [*]
    state "host" as s3
    [*] --> s3
    state "STRING_LITERAL" as s4
    s3 --> s4
    s4 --> [*]
    state "port" as s5
    [*] --> s5
    state "NUMBER_LITERAL" as s6
    s5 --> s6
    s6 --> [*]
    state "database" as s7
    [*] --> s7
    state "STRING_LITERAL" as s8
    s7 --> s8
    s8 --> [*]
    state "username" as s9
    [*] --> s9
    state "STRING_LITERAL" as s10
    s9 --> s10
    s10 --> [*]
    state "password" as s11
    [*] --> s11
    state "STRING_LITERAL" as s12
    s11 --> s12
    s12 --> [*]
```

---

### constantOptions

**Syntax:**

```ebnf
constantOptions
    : constantOption+
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "constantOption+" as s1
    [*] --> s1
    s1 --> [*]
```

---

### constantOption

**Syntax:**

```ebnf
constantOption
    : comment STRING_LITERAL
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "comment" as s1
    [*] --> s1
    state "STRING_LITERAL" as s2
    s1 --> s2
    s2 --> [*]
```

---

### restClientOptions

**Syntax:**

```ebnf
restClientOptions
    : begin restOperation* end
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "begin" as s1
    [*] --> s1
    state "restOperation*" as s2
    s1 --> s2
    state "end" as s3
    s2 --> s3
    s3 --> [*]
```

---

### restClientOptions

**Syntax:**

```ebnf
restClientOptions
    : restClientOption+
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "restClientOption+" as s1
    [*] --> s1
    s1 --> [*]
```

---

### restClientOption

**Syntax:**

```ebnf
restClientOption
    : base url STRING_LITERAL
    | | timeout NUMBER_LITERAL
    | | authentication restAuthentication
    | | comment STRING_LITERAL
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "base" as s1
    [*] --> s1
    state "url" as s2
    s1 --> s2
    state "STRING_LITERAL" as s3
    s2 --> s3
    s3 --> [*]
    state "timeout" as s4
    [*] --> s4
    state "NUMBER_LITERAL" as s5
    s4 --> s5
    s5 --> [*]
    state "authentication" as s6
    [*] --> s6
    state "restAuthentication" as s7
    s6 --> s7
    s7 --> [*]
    state "comment" as s8
    [*] --> s8
    state "STRING_LITERAL" as s9
    s8 --> s9
    s9 --> [*]
```

---

### restAuthentication

**Syntax:**

```ebnf
restAuthentication
    : basic username STRING_LITERAL password STRING_LITERAL
    | | oauth STRING_LITERAL
    | | none
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "basic" as s1
    [*] --> s1
    state "username" as s2
    s1 --> s2
    state "STRING_LITERAL" as s3
    s2 --> s3
    state "password" as s4
    s3 --> s4
    state "STRING_LITERAL" as s5
    s4 --> s5
    s5 --> [*]
    state "oauth" as s6
    [*] --> s6
    state "STRING_LITERAL" as s7
    s6 --> s7
    s7 --> [*]
    state "none" as s8
    [*] --> s8
    s8 --> [*]
```

---

### restOperation

**Syntax:**

```ebnf
restOperation
    : operation IDENTIFIER
    | method restMethod
    | path STRING_LITERAL
    | restOperationOptions?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "operation" as s1
    [*] --> s1
    state "IDENTIFIER" as s2
    s1 --> s2
    s2 --> [*]
    state "method" as s3
    [*] --> s3
    state "restMethod" as s4
    s3 --> s4
    s4 --> [*]
    state "path" as s5
    [*] --> s5
    state "STRING_LITERAL" as s6
    s5 --> s6
    s6 --> [*]
    state "restOperationOptions?" as s7
    [*] --> s7
    s7 --> [*]
```

---

### restMethod

**Syntax:**

```ebnf
restMethod
    : get | post | put | patch | delete
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "get" as s1
    [*] --> s1
    state "|" as s2
    s1 --> s2
    state "post" as s3
    s2 --> s3
    state "|" as s4
    s3 --> s4
    state "put" as s5
    s4 --> s5
    state "|" as s6
    s5 --> s6
    state "patch" as s7
    s6 --> s7
    state "|" as s8
    s7 --> s8
    state "delete" as s9
    s8 --> s9
    s9 --> [*]
```

---

### restOperationOptions

**Syntax:**

```ebnf
restOperationOptions
    : restOperationOption+
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "restOperationOption+" as s1
    [*] --> s1
    s1 --> [*]
```

---

### restOperationOption

**Syntax:**

```ebnf
restOperationOption
    : body STRING_LITERAL
    | | response restResponse
    | | parameter restParameter
    | | timeout NUMBER_LITERAL
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "body" as s1
    [*] --> s1
    state "STRING_LITERAL" as s2
    s1 --> s2
    s2 --> [*]
    state "response" as s3
    [*] --> s3
    state "restResponse" as s4
    s3 --> s4
    s4 --> [*]
    state "parameter" as s5
    [*] --> s5
    state "restParameter" as s6
    s5 --> s6
    s6 --> [*]
    state "timeout" as s7
    [*] --> s7
    state "NUMBER_LITERAL" as s8
    s7 --> s8
    s8 --> [*]
```

---

### restResponse

**Syntax:**

```ebnf
restResponse
    : status NUMBER_LITERAL dataType
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "status" as s1
    [*] --> s1
    state "NUMBER_LITERAL" as s2
    s1 --> s2
    state "dataType" as s3
    s2 --> s3
    s3 --> [*]
```

---

### restParameter

**Syntax:**

```ebnf
restParameter
    : IDENTIFIER COLON dataType (in (path | query | body | header))?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "IDENTIFIER" as s1
    [*] --> s1
    state "COLON" as s2
    s1 --> s2
    state "dataType" as s3
    s2 --> s3
    state "(in (path | query | body | header))?" as s4
    s3 --> s4
    s4 --> [*]
```

---

### catalogTableName

**Syntax:**

```ebnf
catalogTableName
    : modules
    | | entities
    | | microflows
    | | pages
    | | snippets
    | | enumerations
    | | widgets
    | | IDENTIFIER  // for tables like nanoflows, activities, xpath_expressions, objects, projects, snapshots
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "modules" as s1
    [*] --> s1
    s1 --> [*]
    state "entities" as s2
    [*] --> s2
    s2 --> [*]
    state "microflows" as s3
    [*] --> s3
    s3 --> [*]
    state "pages" as s4
    [*] --> s4
    s4 --> [*]
    state "snippets" as s5
    [*] --> s5
    s5 --> [*]
    state "enumerations" as s6
    [*] --> s6
    s6 --> [*]
    state "widgets" as s7
    [*] --> s7
    s7 --> [*]
    state "IDENTIFIER" as s8
    [*] --> s8
    s8 --> [*]
```

---

### tableReference

**Syntax:**

```ebnf
tableReference
    : qualifiedName (as? IDENTIFIER)?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "qualifiedName" as s1
    [*] --> s1
    state "(as? IDENTIFIER)?" as s2
    s1 --> s2
    s2 --> [*]
```

---

### joinClause

**Syntax:**

```ebnf
joinClause
    : joinType? join tableReference (on expression)?
    | | joinType? join IDENTIFIER SLASH qualifiedName SLASH qualifiedName (as IDENTIFIER)?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "joinType?" as s1
    [*] --> s1
    state "join" as s2
    s1 --> s2
    state "tableReference" as s3
    s2 --> s3
    state "(on expression)?" as s4
    s3 --> s4
    s4 --> [*]
    state "joinType?" as s5
    [*] --> s5
    state "join" as s6
    s5 --> s6
    state "IDENTIFIER" as s7
    s6 --> s7
    state "SLASH" as s8
    s7 --> s8
    state "qualifiedName" as s9
    s8 --> s9
    state "SLASH" as s10
    s9 --> s10
    state "qualifiedName" as s11
    s10 --> s11
    state "(as IDENTIFIER)?" as s12
    s11 --> s12
    s12 --> [*]
```

---

### joinType

**Syntax:**

```ebnf
joinType
    : left outer?
    | | right outer?
    | | inner
    | | full outer?
    | | cross
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "left" as s1
    [*] --> s1
    state "outer?" as s2
    s1 --> s2
    s2 --> [*]
    state "right" as s3
    [*] --> s3
    state "outer?" as s4
    s3 --> s4
    s4 --> [*]
    state "inner" as s5
    [*] --> s5
    s5 --> [*]
    state "full" as s6
    [*] --> s6
    state "outer?" as s7
    s6 --> s7
    s7 --> [*]
    state "cross" as s8
    [*] --> s8
    s8 --> [*]
```

---

### groupByClause

**Syntax:**

```ebnf
groupByClause
    : GROUP_BY expressionList
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "GROUP_BY" as s1
    [*] --> s1
    state "expressionList" as s2
    s1 --> s2
    s2 --> [*]
```

---

### havingClause

**Syntax:**

```ebnf
havingClause
    : having expression
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "having" as s1
    [*] --> s1
    state "expression" as s2
    s1 --> s2
    s2 --> [*]
```

---

### orderByClause

**Syntax:**

```ebnf
orderByClause
    : ORDER_BY orderByList
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "ORDER_BY" as s1
    [*] --> s1
    state "orderByList" as s2
    s1 --> s2
    s2 --> [*]
```

---

### orderByList

**Syntax:**

```ebnf
orderByList
    : orderByItem (COMMA orderByItem)*
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "orderByItem" as s1
    [*] --> s1
    state "(COMMA orderByItem)*" as s2
    s1 --> s2
    s2 --> [*]
```

---

### orderByItem

**Syntax:**

```ebnf
orderByItem
    : expression (asc | desc)?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "expression" as s1
    [*] --> s1
    state "(asc | desc)?" as s2
    s1 --> s2
    s2 --> [*]
```

---

### limitOffsetClause

**Syntax:**

```ebnf
limitOffsetClause
    : limit NUMBER_LITERAL (offset NUMBER_LITERAL)?
    | | offset NUMBER_LITERAL (limit NUMBER_LITERAL)?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "limit" as s1
    [*] --> s1
    state "NUMBER_LITERAL" as s2
    s1 --> s2
    state "(offset NUMBER_LITERAL)?" as s3
    s2 --> s3
    s3 --> [*]
    state "offset" as s4
    [*] --> s4
    state "NUMBER_LITERAL" as s5
    s4 --> s5
    state "(limit NUMBER_LITERAL)?" as s6
    s5 --> s6
    s6 --> [*]
```

---

### sessionIdList

**Syntax:**

```ebnf
sessionIdList
    : sessionId (COMMA sessionId)*
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "sessionId" as s1
    [*] --> s1
    state "(COMMA sessionId)*" as s2
    s1 --> s2
    s2 --> [*]
```

---

### sessionId

**Syntax:**

```ebnf
sessionId
    : IDENTIFIER
    | | HYPHENATED_ID
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "IDENTIFIER" as s1
    [*] --> s1
    s1 --> [*]
    state "HYPHENATED_ID" as s2
    [*] --> s2
    s2 --> [*]
```

---

### aggregateFunction

**Syntax:**

```ebnf
aggregateFunction
    : (count | sum | avg | min | max) LPAREN (distinct? expression | STAR) RPAREN
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "(count | sum | avg | min | max)" as s1
    [*] --> s1
    state "LPAREN" as s2
    s1 --> s2
    state "(distinct? expression | STAR)" as s3
    s2 --> s3
    state "RPAREN" as s4
    s3 --> s4
    s4 --> [*]
```

---

### functionCall

**Syntax:**

```ebnf
functionCall
    : (IDENTIFIER | HYPHENATED_ID) LPAREN argumentList? RPAREN
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "(IDENTIFIER | HYPHENATED_ID)" as s1
    [*] --> s1
    state "LPAREN" as s2
    s1 --> s2
    state "argumentList?" as s3
    s2 --> s3
    state "RPAREN" as s4
    s3 --> s4
    s4 --> [*]
```

---

### argumentList

**Syntax:**

```ebnf
argumentList
    : expression (COMMA expression)*
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "expression" as s1
    [*] --> s1
    state "(COMMA expression)*" as s2
    s1 --> s2
    s2 --> [*]
```

---

### qualifiedName

Qualified name: Module.Entity or Module.Entity.Attribute

**Syntax:**

```ebnf
qualifiedName
    : (IDENTIFIER | keyword) (DOT (IDENTIFIER | keyword))*
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "(IDENTIFIER | keyword)" as s1
    [*] --> s1
    state "(DOT (IDENTIFIER | keyword))*" as s2
    s1 --> s2
    s2 --> [*]
```

---

### docComment

Documentation comment

**Syntax:**

```ebnf
docComment
    : DOC_COMMENT
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "DOC_COMMENT" as s1
    [*] --> s1
    s1 --> [*]
```

---

### annotation

Annotation: @Name or @Name(params)

**Syntax:**

```ebnf
annotation
    : AT annotationName (LPAREN annotationParams RPAREN)?
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "AT" as s1
    [*] --> s1
    state "annotationName" as s2
    s1 --> s2
    state "(LPAREN annotationParams RPAREN)?" as s3
    s2 --> s3
    s3 --> [*]
```

---

### annotationName

**Syntax:**

```ebnf
annotationName
    : IDENTIFIER
    | | position
    | | comment
    | | icon
    | | folder
    | | required
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "IDENTIFIER" as s1
    [*] --> s1
    s1 --> [*]
    state "position" as s2
    [*] --> s2
    s2 --> [*]
    state "comment" as s3
    [*] --> s3
    s3 --> [*]
    state "icon" as s4
    [*] --> s4
    s4 --> [*]
    state "folder" as s5
    [*] --> s5
    s5 --> [*]
    state "required" as s6
    [*] --> s6
    s6 --> [*]
```

---

### annotationParams

**Syntax:**

```ebnf
annotationParams
    : annotationParam (COMMA annotationParam)*
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "annotationParam" as s1
    [*] --> s1
    state "(COMMA annotationParam)*" as s2
    s1 --> s2
    s2 --> [*]
```

---

### annotationParam

**Syntax:**

```ebnf
annotationParam
    : IDENTIFIER COLON annotationValue   // Named parameter
    | | annotationValue                     // Positional parameter
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "IDENTIFIER" as s1
    [*] --> s1
    state "COLON" as s2
    s1 --> s2
    state "annotationValue" as s3
    s2 --> s3
    s3 --> [*]
    state "annotationValue" as s4
    [*] --> s4
    s4 --> [*]
```

---

### annotationValue

**Syntax:**

```ebnf
annotationValue
    : literal
    | | expression
    | | qualifiedName
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "literal" as s1
    [*] --> s1
    s1 --> [*]
    state "expression" as s2
    [*] --> s2
    s2 --> [*]
    state "qualifiedName" as s3
    [*] --> s3
    s3 --> [*]
```

---

### keyword

Keywords that can be used as identifiers in certain contexts

**Syntax:**

```ebnf
keyword
    : create | alter | drop | rename | entity | persistent | view | module
    | | association | microflow | page | snippet | enumeration
    | | STRING_TYPE | INTEGER_TYPE | LONG_TYPE | DECIMAL_TYPE | BOOLEAN_TYPE
    | | DATETIME_TYPE | DATE_TYPE | AUTONUMBER_TYPE | BINARY_TYPE
    | | select | from | where | join | left | right | inner | outer
    | | ORDER_BY | GROUP_BY | having | limit | offset | as | on
    | | and | or | not | null | in | like | between | true | false
    | | count | sum | avg | min | max | distinct | all
    | | begin | end | if | else | elsif | then | while | loop
    | | declare | set | change | retrieve | delete | commit | return
    | | call | log | with | for | to | of | type | value
    | | show | describe | connect | disconnect | use | status
    | | title | layout | caption | label | width | height | style | class
    | | datasource | editable | visible | required | default | unique
    | | index | owner | reference | cascade | both | single | multiple | none
    | | critical | success | error | warning | info | debug
    | | message | action | username | password
```

**Railroad Diagram:**

```mermaid
stateDiagram-v2
    direction LR
    state "create" as s1
    [*] --> s1
    state "|" as s2
    s1 --> s2
    state "alter" as s3
    s2 --> s3
    state "|" as s4
    s3 --> s4
    state "drop" as s5
    s4 --> s5
    state "|" as s6
    s5 --> s6
    state "rename" as s7
    s6 --> s7
    state "|" as s8
    s7 --> s8
    state "entity" as s9
    s8 --> s9
    state "|" as s10
    s9 --> s10
    state "persistent" as s11
    s10 --> s11
    state "|" as s12
    s11 --> s12
    state "view" as s13
    s12 --> s13
    state "|" as s14
    s13 --> s14
    state "module" as s15
    s14 --> s15
    s15 --> [*]
    state "association" as s16
    [*] --> s16
    state "|" as s17
    s16 --> s17
    state "microflow" as s18
    s17 --> s18
    state "|" as s19
    s18 --> s19
    state "page" as s20
    s19 --> s20
    state "|" as s21
    s20 --> s21
    state "snippet" as s22
    s21 --> s22
    state "|" as s23
    s22 --> s23
    state "enumeration" as s24
    s23 --> s24
    s24 --> [*]
    state "STRING_TYPE" as s25
    [*] --> s25
    state "|" as s26
    s25 --> s26
    state "INTEGER_TYPE" as s27
    s26 --> s27
    state "|" as s28
    s27 --> s28
    state "LONG_TYPE" as s29
    s28 --> s29
    state "|" as s30
    s29 --> s30
    state "DECIMAL_TYPE" as s31
    s30 --> s31
    state "|" as s32
    s31 --> s32
    state "BOOLEAN_TYPE" as s33
    s32 --> s33
    s33 --> [*]
    state "DATETIME_TYPE" as s34
    [*] --> s34
    state "|" as s35
    s34 --> s35
    state "DATE_TYPE" as s36
    s35 --> s36
    state "|" as s37
    s36 --> s37
    state "AUTONUMBER_TYPE" as s38
    s37 --> s38
    state "|" as s39
    s38 --> s39
    state "BINARY_TYPE" as s40
    s39 --> s40
    s40 --> [*]
    state "select" as s41
    [*] --> s41
    state "|" as s42
    s41 --> s42
    state "from" as s43
    s42 --> s43
    state "|" as s44
    s43 --> s44
    state "where" as s45
    s44 --> s45
    state "|" as s46
    s45 --> s46
    state "join" as s47
    s46 --> s47
    state "|" as s48
    s47 --> s48
    state "left" as s49
    s48 --> s49
    state "|" as s50
    s49 --> s50
    state "right" as s51
    s50 --> s51
    state "|" as s52
    s51 --> s52
    state "inner" as s53
    s52 --> s53
    state "|" as s54
    s53 --> s54
    state "outer" as s55
    s54 --> s55
    s55 --> [*]
    state "ORDER_BY" as s56
    [*] --> s56
    state "|" as s57
    s56 --> s57
    state "GROUP_BY" as s58
    s57 --> s58
    state "|" as s59
    s58 --> s59
    state "having" as s60
    s59 --> s60
    state "|" as s61
    s60 --> s61
    state "limit" as s62
    s61 --> s62
    state "|" as s63
    s62 --> s63
    state "offset" as s64
    s63 --> s64
    state "|" as s65
    s64 --> s65
    state "as" as s66
    s65 --> s66
    state "|" as s67
    s66 --> s67
    state "on" as s68
    s67 --> s68
    s68 --> [*]
    state "and" as s69
    [*] --> s69
    state "|" as s70
    s69 --> s70
    state "or" as s71
    s70 --> s71
    state "|" as s72
    s71 --> s72
    state "not" as s73
    s72 --> s73
    state "|" as s74
    s73 --> s74
    state "null" as s75
    s74 --> s75
    state "|" as s76
    s75 --> s76
    state "in" as s77
    s76 --> s77
    state "|" as s78
    s77 --> s78
    state "like" as s79
    s78 --> s79
    state "|" as s80
    s79 --> s80
    state "between" as s81
    s80 --> s81
    state "|" as s82
    s81 --> s82
    state "true" as s83
    s82 --> s83
    state "|" as s84
    s83 --> s84
    state "false" as s85
    s84 --> s85
    s85 --> [*]
    state "count" as s86
    [*] --> s86
    state "|" as s87
    s86 --> s87
    state "sum" as s88
    s87 --> s88
    state "|" as s89
    s88 --> s89
    state "avg" as s90
    s89 --> s90
    state "|" as s91
    s90 --> s91
    state "min" as s92
    s91 --> s92
    state "|" as s93
    s92 --> s93
    state "max" as s94
    s93 --> s94
    state "|" as s95
    s94 --> s95
    state "distinct" as s96
    s95 --> s96
    state "|" as s97
    s96 --> s97
    state "all" as s98
    s97 --> s98
    s98 --> [*]
    state "begin" as s99
    [*] --> s99
    state "|" as s100
    s99 --> s100
    state "end" as s101
    s100 --> s101
    state "|" as s102
    s101 --> s102
    state "if" as s103
    s102 --> s103
    state "|" as s104
    s103 --> s104
    state "else" as s105
    s104 --> s105
    state "|" as s106
    s105 --> s106
    state "elsif" as s107
    s106 --> s107
    state "|" as s108
    s107 --> s108
    state "then" as s109
    s108 --> s109
    state "|" as s110
    s109 --> s110
    state "while" as s111
    s110 --> s111
    state "|" as s112
    s111 --> s112
    state "loop" as s113
    s112 --> s113
    s113 --> [*]
    state "declare" as s114
    [*] --> s114
    state "|" as s115
    s114 --> s115
    state "set" as s116
    s115 --> s116
    state "|" as s117
    s116 --> s117
    state "change" as s118
    s117 --> s118
    state "|" as s119
    s118 --> s119
    state "retrieve" as s120
    s119 --> s120
    state "|" as s121
    s120 --> s121
    state "delete" as s122
    s121 --> s122
    state "|" as s123
    s122 --> s123
    state "commit" as s124
    s123 --> s124
    state "|" as s125
    s124 --> s125
    state "return" as s126
    s125 --> s126
    s126 --> [*]
    state "call" as s127
    [*] --> s127
    state "|" as s128
    s127 --> s128
    state "log" as s129
    s128 --> s129
    state "|" as s130
    s129 --> s130
    state "with" as s131
    s130 --> s131
    state "|" as s132
    s131 --> s132
    state "for" as s133
    s132 --> s133
    state "|" as s134
    s133 --> s134
    state "to" as s135
    s134 --> s135
    state "|" as s136
    s135 --> s136
    state "of" as s137
    s136 --> s137
    state "|" as s138
    s137 --> s138
    state "type" as s139
    s138 --> s139
    state "|" as s140
    s139 --> s140
    state "value" as s141
    s140 --> s141
    s141 --> [*]
    state "show" as s142
    [*] --> s142
    state "|" as s143
    s142 --> s143
    state "describe" as s144
    s143 --> s144
    state "|" as s145
    s144 --> s145
    state "connect" as s146
    s145 --> s146
    state "|" as s147
    s146 --> s147
    state "disconnect" as s148
    s147 --> s148
    state "|" as s149
    s148 --> s149
    state "use" as s150
    s149 --> s150
    state "|" as s151
    s150 --> s151
    state "status" as s152
    s151 --> s152
    s152 --> [*]
    state "title" as s153
    [*] --> s153
    state "|" as s154
    s153 --> s154
    state "layout" as s155
    s154 --> s155
    state "|" as s156
    s155 --> s156
    state "caption" as s157
    s156 --> s157
    state "|" as s158
    s157 --> s158
    state "label" as s159
    s158 --> s159
    state "|" as s160
    s159 --> s160
    state "width" as s161
    s160 --> s161
    state "|" as s162
    s161 --> s162
    state "height" as s163
    s162 --> s163
    state "|" as s164
    s163 --> s164
    state "style" as s165
    s164 --> s165
    state "|" as s166
    s165 --> s166
    state "class" as s167
    s166 --> s167
    s167 --> [*]
    state "datasource" as s168
    [*] --> s168
    state "|" as s169
    s168 --> s169
    state "editable" as s170
    s169 --> s170
    state "|" as s171
    s170 --> s171
    state "visible" as s172
    s171 --> s172
    state "|" as s173
    s172 --> s173
    state "required" as s174
    s173 --> s174
    state "|" as s175
    s174 --> s175
    state "default" as s176
    s175 --> s176
    state "|" as s177
    s176 --> s177
    state "unique" as s178
    s177 --> s178
    s178 --> [*]
    state "index" as s179
    [*] --> s179
    state "|" as s180
    s179 --> s180
    state "owner" as s181
    s180 --> s181
    state "|" as s182
    s181 --> s182
    state "reference" as s183
    s182 --> s183
    state "|" as s184
    s183 --> s184
    state "cascade" as s185
    s184 --> s185
    state "|" as s186
    s185 --> s186
    state "both" as s187
    s186 --> s187
    state "|" as s188
    s187 --> s188
    state "single" as s189
    s188 --> s189
    state "|" as s190
    s189 --> s190
    state "multiple" as s191
    s190 --> s191
    state "|" as s192
    s191 --> s192
    state "none" as s193
    s192 --> s193
    s193 --> [*]
    state "critical" as s194
    [*] --> s194
    state "|" as s195
    s194 --> s195
    state "success" as s196
    s195 --> s196
    state "|" as s197
    s196 --> s197
    state "error" as s198
    s197 --> s198
    state "|" as s199
    s198 --> s199
    state "warning" as s200
    s199 --> s200
    state "|" as s201
    s200 --> s201
    state "info" as s202
    s201 --> s202
    state "|" as s203
    s202 --> s203
    state "debug" as s204
    s203 --> s204
    s204 --> [*]
    state "message" as s205
    [*] --> s205
    state "|" as s206
    s205 --> s206
    state "action" as s207
    s206 --> s207
    state "|" as s208
    s207 --> s208
    state "username" as s209
    s208 --> s209
    state "|" as s210
    s209 --> s210
    state "password" as s211
    s210 --> s211
    s211 --> [*]
```

---

