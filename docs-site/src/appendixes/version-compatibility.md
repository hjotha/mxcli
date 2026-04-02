# Mendix Version Compatibility

Supported Mendix Studio Pro versions, BSON format differences, and known limitations.

## Supported Versions

mxcli supports Mendix Studio Pro versions **9.x through 11.x**. Development and nightly testing targets three versions:

| Studio Pro Version | MPR Format | Nightly Tested | Status |
|-------------------|------------|----------------|--------|
| 9.x | v1 | No | Read-only support |
| 10.0 -- 10.17 | v1 | No | Supported |
| 10.18 -- 10.23 | v2 | No | Supported |
| 10.24 (LTS) | v2 | **Yes** | Supported |
| 11.0 -- 11.5 | v2 | No | Supported |
| 11.6 | v2 | **Yes** | Primary development target |
| 11.8 | v2 | **Yes** | Latest tested |

## MPR Format Versions

### v1 (Mendix < 10.18)

- Single `.mpr` SQLite database file
- All documents stored as BSON blobs in the `UnitContents` table
- Self-contained -- one file holds the entire project

### v2 (Mendix >= 10.18)

- `.mpr` SQLite file for metadata only
- `mprcontents/` folder with individual `.mxunit` files for each document
- Better suited for Git version control (smaller, per-document diffs)

The library auto-detects the format. No configuration is needed.

## Feature Availability by Version

Not all MDL features work on all Mendix versions. The BSON document structure changes across versions, and some features were introduced in specific releases.

### Core Features (all supported versions)

| Feature | Minimum Version | Notes |
|---------|----------------|-------|
| Domain models (entities, attributes, associations) | 9.0 | Full CRUD support |
| Microflows (60+ activity types) | 9.0 | Including loops, splits, error handling |
| Nanoflows | 9.0 | Client-side flows |
| Pages (50+ widget types) | 9.0 | Built-in widgets |
| Enumerations | 9.0 | CREATE/ALTER/DROP |
| Security (module roles, access rules) | 9.0 | Full support |
| Navigation | 9.0 | Profiles, menus, home pages |
| Workflows | 9.0 | User tasks, decisions, parallel splits |

### Features Requiring Mendix 10.x+

| Feature | Minimum Version | Notes |
|---------|----------------|-------|
| Business events | 10.0 | Event service definitions |
| Pluggable widgets (ComboBox, DataGrid2, Gallery) | 10.0 | Requires widget templates |
| Image collections | 10.0 | CREATE/DROP IMAGE COLLECTION |

### Features Requiring Mendix 11.0+

These features use BSON structures that changed in Mendix 11.0 and are **not compatible with 10.x projects**:

| Feature | Minimum Version | Error on 10.x |
|---------|----------------|---------------|
| View entities (CREATE VIEW ENTITY) | 10.18 | Version-aware serialization (inline OQL on 10.x) |
| Page parameters (Params: { ... }) | 11.0 | InvalidOperationException on 'Variable' property |
| Design properties (Atlas v3) | 11.0 | CE6083: "Design property not supported by your theme" |
| REST client (CREATE REST CLIENT) | 11.0 | BSON format incompatibility |
| Database Connector (EXECUTE DATABASE QUERY) | 11.0 | Module format incompatibility |
| Association storage format | 11.0 | Different BSON encoding for associations |

### Features Requiring Mendix 11.6+

| Feature | Minimum Version | Notes |
|---------|----------------|-------|
| Portable app format | 11.6 | New deployment format |

## BSON Differences Across Versions

The Mendix metamodel evolves across versions. The reflection data shows ~42% type growth from Mendix 9.0 to 11.6. Key structural differences:

### View Entities (10.x vs 11.x)

View entities exist in both 10.18+ and 11.x, but the BSON structure differs. In Mendix 10.x, the `OqlViewEntitySource` object has an `Oql` field that stores the OQL query inline (in addition to the separate `ViewEntitySourceDocument`). Mendix 11.0 removed the inline `Oql` field. The writer detects the project version and includes the inline field for 10.x projects.

### Page Parameters (10.x vs 11.x)

Mendix 11.0 changed how page parameters are stored in BSON. The `Variable` property in page parameter mappings uses a different structure. Writing 11.x-style page parameters to a 10.x project causes an `InvalidOperationException`.

### Design Properties (Atlas v2 vs v3)

Atlas UI v3 (Mendix 11.0+) introduced new design properties like "Card style" and "Disable row wrap" for containers. These don't exist in the Atlas v2 theme bundled with 10.x, causing CE6083.

### Widget Type Definitions

Pluggable widget PropertyTypes change between Mendix versions. A widget template extracted from 11.6 may have more or fewer properties than the widget installed in a 10.24 project. The [widget augmentation system](../internals/widget-templates.md) handles this by reconciling templates against the project's `.mpk` files.

## Widget Template Versions

Embedded widget templates are extracted from Mendix 11.6 and cover:

| Widget | Template | Widget ID |
|--------|----------|-----------|
| Combo box | combobox.json | com.mendix.widget.web.combobox.Combobox |
| Data grid 2 | datagrid.json | com.mendix.widget.web.datagrid.Datagrid |
| Gallery | gallery.json | com.mendix.widget.web.gallery.Gallery |
| Image | image.json | com.mendix.widget.web.image.Image |
| Text filter | datagrid-text-filter.json | com.mendix.widget.web.datagridtextfilter.DatagridTextFilter |
| Number filter | datagrid-number-filter.json | com.mendix.widget.web.datagridnumberfilter.DatagridNumberFilter |
| Date filter | datagrid-date-filter.json | com.mendix.widget.web.datagriddatefilter.DatagridDateFilter |
| Dropdown filter | datagrid-dropdown-filter.json | com.mendix.widget.web.datagriddropdownfilter.DatagridDropdownFilter |

### MPK Augmentation

When opening a project, mxcli checks the project's `widgets/` folder for `.mpk` packages. If a widget's installed version differs from the embedded template, the augmentation system:

1. Extracts the XML property definition from the `.mpk`
2. Adds properties found in `.mpk` but missing from the template
3. Removes properties in the template but not in the `.mpk`
4. Preserves the BSON structure (IDs, nesting, cross-references)

This reduces CE0463 ("widget definition changed") errors from widget version drift. See [Widget Templates](../internals/widget-templates.md) for details.

## Version Gates in MDL Scripts

MDL scripts can use `-- @version:` directives to gate sections by Mendix version. This is used in the doctype integration tests to skip features incompatible with older versions:

```mdl
-- Runs on all versions
CREATE ENTITY MyModule.Customer (Name: String(200));

-- @version: 11.0+
-- Only runs on Mendix 11.0 and later
CREATE VIEW ENTITY MyModule.ActiveCustomers (...) AS
  SELECT c.Name FROM MyModule.Customer AS c WHERE c.IsActive;

-- @version: 10.6..10.24
-- Only runs on Mendix 10.6 through 10.24
CREATE ENTITY MyModule.LegacyConfig (...);

-- @version: any
-- Resets to unconditional (runs on all versions)
CREATE ENTITY MyModule.Universal (...);
```

**Directive formats:**
- `-- @version: 11.0+` -- minimum version (run on 11.0 and later)
- `-- @version: 10.6..10.24` -- version range (inclusive)
- `-- @version: ..10.24` -- maximum version only
- `-- @version: any` -- reset to unconditional

The directive applies to all lines until the next `-- @version:` directive.

## MxBuild Compatibility

The `mx` validation tool must match the project's Mendix version:

```bash
# Auto-download the correct MxBuild version
mxcli setup mxbuild -p app.mpr

# Check the project
~/.mxcli/mxbuild/*/modeler/mx check app.mpr
```

MxBuild is downloaded on demand and cached in `~/.mxcli/mxbuild/{version}/`.

## Platform Support

mxcli runs on:

| Platform | Architecture |
|----------|-------------|
| Linux | amd64, arm64 |
| macOS | amd64, arm64 (Apple Silicon) |
| Windows | amd64, arm64 |

No CGO or C compiler is required -- the binary is fully statically linked using pure Go dependencies.
