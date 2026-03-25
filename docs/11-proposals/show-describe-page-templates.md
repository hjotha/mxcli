# Proposal: SHOW/DESCRIBE Page Templates

## Overview

**Document type:** `Forms$PageTemplate`
**Prevalence:** 215 across test projects (81 Enquiries, 63 Evora, 71 Lato)
**Priority:** High — present in every project, define page scaffolding

Page Templates provide the initial structure when creating new pages in Studio Pro. Each template has a type (Standard, Edit, Select), a layout reference, and a widget tree.

## What Already Exists

| Layer | Status | Location |
|-------|--------|----------|
| **Go type** | Yes | `sdk/pages/pages.go` — `PageTemplate{Name, Documentation, DisplayName, LayoutID, PageTemplateType, Widget}` |
| **Parser** | Minimal | `sdk/mpr/parser_misc.go` line 192 — Name + Documentation only |
| **Reader** | Yes | `ListPageTemplates()` in `sdk/mpr/reader_types.go` |
| **AST** | No | — |
| **Executor** | No | — |

## BSON Structure (from test projects)

```
Forms$PageTemplate:
  Name: string
  Documentation: string
  DisplayName: string
  Excluded: bool
  ExportLevel: string
  TemplateCategory: string
  TemplateCategoryWeight: int32
  CanvasWidth: int32
  CanvasHeight: int32
  DocumentationUrl: string
  ImageData: binary (preview thumbnail)
  LayoutCall: Forms$LayoutCall
    Form: string (layout qualified name)
    Arguments: []*LayoutCallArgument
  TemplateType: polymorphic
    Forms$RegularPageTemplateType (no extra fields)
    -- could also be other types
  Appearance: Forms$Appearance
    Class: string
    Style: string
    DesignProperties: map
  Widgets: []*Widget
```

## Proposed MDL Syntax

### SHOW PAGE TEMPLATES

```
SHOW PAGE TEMPLATES [IN Module]
```

| Qualified Name | Module | Name | Display Name | Type | Layout | Category |
|----------------|--------|------|--------------|------|--------|----------|

### DESCRIBE PAGE TEMPLATE

```
DESCRIBE PAGE TEMPLATE Module.Name
```

Output format:

```
/**
 * Standard edit page with save/cancel bar
 */
-- Page Template: MyModule.EditTemplate
-- Display Name: Edit Form
-- Type: Edit
-- Category: Forms
PAGE TEMPLATE MyModule.EditTemplate
  LAYOUT Atlas_Core.Atlas_Default
{
  DATAVIEW
  {
    TEXTBOX $Name;
    ACTIONBUTTON 'Save';
  };
};
/
```

## Implementation Steps

### 1. Enhance Parser

Extend `parsePageTemplate()` to capture:
- `DisplayName`, `TemplateCategory`, `Excluded`, `ExportLevel`
- `LayoutCall` (layout reference)
- `TemplateType` (Standard/Edit/Select)
- Widget tree (reuse `parseWidgets()`)

### 2. Add AST Types

```go
ShowPageTemplates    // in ShowObjectType enum
DescribePageTemplate // in DescribeObjectType enum
```

### 3. Add Grammar Rules

```antlr
TEMPLATE: 'TEMPLATE';
TEMPLATES: 'TEMPLATES';

// SHOW PAGE TEMPLATES [IN module]
// DESCRIBE PAGE TEMPLATE qualifiedName
```

### 4. Add Executor (mdl/executor/cmd_page_templates.go)

- `showPageTemplates(moduleName string)` — table listing
- `describePageTemplate(name QualifiedName)` — MDL output

### 5. Add Autocomplete

```go
func (e *Executor) GetPageTemplateNames(moduleFilter string) []string
```

## Testing

- Create example file or extend page examples
- Verify against all 3 test projects
