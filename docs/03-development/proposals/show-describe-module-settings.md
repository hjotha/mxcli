# Proposal: SHOW/DESCRIBE Module Settings

## Overview

**Document type:** `Projects$ModuleSettings`
**Prevalence:** 97 across test projects (28 Enquiries, 39 Evora, 30 Lato) — one per module
**Priority:** Medium — every module has one, useful for identifying App Store modules and versions

Module Settings contain metadata about a module: its version, whether it came from the Mendix Marketplace, its protection level, and JAR dependencies. This is separate from Module Security (which is already implemented).

## What Already Exists

| Layer | Status | Location |
|-------|--------|----------|
| **Go type** | No | — |
| **Parser** | No | — |
| **Reader** | No | — |
| **Generated metamodel** | No | Not in generated types (simple project-level type) |

## BSON Structure (from test projects)

```
Projects$ModuleSettings:
  Version: string (e.g., "2.3.0")
  BasedOnVersion: string
  ExportLevel: string ("Source")
  ExtensionName: string
  ProtectedModuleType: string ("AddOn", "None", etc.)
  SolutionIdentifier: string
  JarDependencies: [] (array of JAR refs)
```

Note: There is one `ModuleSettings` per module, stored as a separate unit. The `ModuleImpl` unit (also one per module) stores: `AppStoreGuid`, `AppStoreVersion`, `FromAppStore`, `IsThemeModule`, `Name`.

## Proposed MDL Syntax

### SHOW MODULE SETTINGS

```
SHOW MODULE SETTINGS [IN Module]
```

| Module | Version | Based On | From App Store | App Store Version | Protected | Theme Module |
|--------|---------|----------|----------------|-------------------|-----------|--------------|

This combines data from both `ModuleSettings` and `ModuleImpl` to give a complete picture per module.

### DESCRIBE MODULE SETTINGS

```
DESCRIBE MODULE SETTINGS Module
```

Output format:

```
-- Module Settings: Atlas_Core
MODULE SETTINGS Atlas_Core
  VERSION '3.0.9'
  BASED ON VERSION '3.0.8'
  FROM APP STORE
    GUID 'abc-123'
    APP STORE VERSION '3.0.9'
  PROTECTED TYPE AddOn
  THEME MODULE;
/
```

For user-created modules with minimal settings:

```
-- Module Settings: MyFirstModule
MODULE SETTINGS MyFirstModule
  VERSION ''
  PROTECTED TYPE None;
/
```

## Implementation Steps

### 1. Add Model Types (model/types.go)

```go
type ModuleSettings struct {
    ContainerID       model.ID
    ModuleName        string
    Version           string
    BasedOnVersion    string
    ExportLevel       string
    ProtectedModuleType string
    SolutionIdentifier  string
    ExtensionName     string
    // From ModuleImpl:
    FromAppStore      bool
    AppStoreGuid      string
    AppStoreVersion   string
    IsThemeModule     bool
}
```

### 2. Add Parser (sdk/mpr/parser_misc.go)

Parse both `Projects$ModuleSettings` and correlate with `Projects$ModuleImpl` data (already parsed for modules).

### 3. Add Reader

```go
func (r *Reader) ListModuleSettings() ([]*model.ModuleSettings, error)
```

### 4. Add AST, Grammar, Visitor, Executor

Standard pattern. Reuse existing `MODULE` token; add `SETTINGS` context for `SHOW MODULE SETTINGS`.

Note: `SHOW SETTINGS` already exists for Project Settings. Use `SHOW MODULE SETTINGS` to disambiguate.

## Testing

- Verify against all 3 test projects
- Check that App Store modules are correctly identified
