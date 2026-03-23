# Proposal: `mxcli marketplace` — Download & Manage Marketplace Modules

**Status:** Draft
**Date:** 2026-03-23
**Author:** Generated with Claude Code

## Problem

Mendix projects depend on marketplace modules (e.g., ExternalDatabaseConnector, BusinessEvents) that are not bundled with the Mendix runtime. Currently, these modules must be manually downloaded from the Mendix Marketplace website and imported via Studio Pro or `mx module-import`.

This creates friction in two areas:

1. **CI/CD pipelines** — Integration tests that exercise database connector or business event features require marketplace modules. We currently bundle `.mpk` files in `mx-modules/`, which is not scalable and requires manual updates.
2. **Project setup** — When `mxcli init` creates a project or when MDL scripts reference marketplace modules, there's no way to auto-install dependencies.

## API Discovery

The Mendix Marketplace REST API is at:

```
https://appstore.home.mendix.com/rest/packagesapi/v2/
```

All endpoints return `401 Unauthorized` without authentication. Mendix uses header-based auth on its platform APIs:

| Auth Pattern | Headers | Used By |
|---|---|---|
| API Key | `Mendix-UserName` + `Mendix-ApiKey` | Deploy API, Team Server |
| PAT | `Authorization: MxToken <token>` | Platform APIs |

The exact auth scheme for the marketplace API needs validation with real credentials.

### Discovered Endpoints (to validate)

| Endpoint | Purpose |
|----------|---------|
| `GET /rest/packagesapi/v2/packages/{id}` | Module metadata (name, versions, AppStoreGuid) |
| `GET /rest/packagesapi/v2/packages/{id}/versions` | List available versions |
| `GET /rest/packagesapi/v2/packages/{id}/versions/{ver}/download` | Download .mpk |
| `GET /rest/packagesapi/v2/packages?search=keyword` | Search marketplace |

### Known Component IDs

| Module | Component ID | Used By |
|--------|-------------|---------|
| External Database Connector | 2888 | 05-database-connection-examples.mdl |
| Business Events | (TBD) | 13-business-events-examples.mdl |

## Proposed Commands

### Authentication

```bash
# Interactive login (prompts for email + API key or PAT)
mxcli auth login

# Non-interactive login (for CI)
mxcli auth login --token <PAT>
mxcli auth login --username user@company.com --api-key <KEY>

# Check auth status
mxcli auth status

# Clear stored credentials
mxcli auth logout
```

### Marketplace Operations

```bash
# Search marketplace
mxcli marketplace search "database connector"

# Show module details
mxcli marketplace info 2888

# Install module into project
mxcli marketplace install 2888 -p app.mpr
mxcli marketplace install 2888 --version 6.2.3 -p app.mpr

# List installed marketplace modules
mxcli marketplace list -p app.mpr

# Update all marketplace modules to latest compatible version
mxcli marketplace update -p app.mpr

# Update specific module
mxcli marketplace update 2888 -p app.mpr
```

## Architecture

### File Structure

```
cmd/mxcli/
├── auth.go                 # Login/logout/status commands
├── cmd_marketplace.go      # Search/install/update/list commands

internal/
├── auth/
│   ├── store.go            # Credential storage (~/.mxcli/auth.json)
│   └── client.go           # Authenticated HTTP client factory
│
├── marketplace/
│   ├── api.go              # REST API client for appstore.home.mendix.com
│   ├── search.go           # Search and list components
│   ├── download.go         # Download .mpk files (with progress bar)
│   └── install.go          # Import via mx module-import + version tracking
```

### Credential Storage

Credentials stored at `~/.mxcli/auth.json` with `0600` permissions:

```json
{
  "email": "user@company.com",
  "api_key": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "created_at": "2026-03-23T12:00:00Z"
}
```

Environment variables take precedence for CI:

```bash
export MENDIX_USERNAME=user@company.com
export MENDIX_API_KEY=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
```

### Install Flow

```
mxcli marketplace install 2888 -p app.mpr
  │
  ├─ 1. Load credentials (env vars → ~/.mxcli/auth.json)
  ├─ 2. GET /packages/2888 → module name, available versions
  ├─ 3. Filter versions compatible with project's Mendix version
  ├─ 4. GET /packages/2888/versions/{ver}/download → download .mpk to temp
  ├─ 5. mx module-import /tmp/module.mpk app.mpr
  └─ 6. Log: "Installed ExternalDatabaseConnector v6.2.3"
```

### Integration with Existing Code

The download infrastructure in `cmd/mxcli/docker/download.go` provides a reusable pattern:
- HTTP download with progress bar
- Temporary file handling
- Archive extraction

The `mx module-import MPK_PATH MPR_PATH` command (already available in mxbuild) handles the actual module import.

Module metadata fields already exist in `model.Module`:
- `AppStoreGuid` — Marketplace component identifier
- `AppStoreVersion` — Installed version
- `FromAppStore` — Whether module came from marketplace

## Integration with Tests

### Current approach (Phase 1)

Marketplace `.mpk` files bundled in `mx-modules/`:

```go
// roundtrip_doctype_test.go
var scriptModuleDeps = map[string][]string{
    "05-database-connection-examples.mdl": {"ExternalDatabaseConnector-v6.2.3.mpk"},
    "13-business-events-examples.mdl":     {"BusinessEvents_3.12.0.mpk"},
}
```

### Future approach (Phase 3+)

Tests use `mxcli marketplace install` with fallback to local `.mpk`:

```go
func importMarketplaceModule(t *testing.T, componentID string, mprPath string) {
    // Try local mpk cache first (fast, no auth needed)
    localMPK := findLocalMPK(componentID)
    if localMPK != "" {
        runMxModuleImport(localMPK, mprPath)
        return
    }

    // Fall back to marketplace download (needs auth)
    if !hasMarketplaceAuth() {
        t.Skipf("marketplace module %s not available (no auth)", componentID)
        return
    }
    runMxcliMarketplaceInstall(componentID, mprPath)
}
```

## Implementation Phases

### Phase 1: Bundle .mpk files (current)

- `.mpk` files in `mx-modules/` directory
- Test imports via `mx module-import` before script execution
- **Status: Done**

### Phase 2: Authentication

- `mxcli auth login/logout/status` commands
- Credential storage at `~/.mxcli/auth.json`
- Environment variable support for CI
- Authenticated HTTP client factory

### Phase 3: Install & Download

- `mxcli marketplace install <id> -p app.mpr`
- Validate API endpoints with real credentials
- Download `.mpk` with progress bar
- Import via `mx module-import`
- Version compatibility filtering by project Mendix version

### Phase 4: Search & Management

- `mxcli marketplace search <query>`
- `mxcli marketplace list -p app.mpr`
- `mxcli marketplace update -p app.mpr`
- `mxcli marketplace info <id>`

### Phase 5: MDL Integration

- `INSTALL MODULE <id> [VERSION <version>];` MDL statement
- Auto-dependency resolution in `mxcli exec` when module references are missing

## Open Questions

1. **Which auth scheme does the marketplace API use?** Need to test with real credentials. Mendix platform APIs use either `Mendix-UserName + Mendix-ApiKey` headers or `MxToken` PAT tokens.

2. **Version compatibility filtering** — The API likely returns Mendix version compatibility metadata per module version. Need to confirm the response format.

3. **Protected modules** — Some marketplace modules are add-on/solution modules that `mx module-import` rejects (exit code X11). Need to handle gracefully.

4. **Rate limiting** — Unknown; need to test and add retry logic if needed.

5. **Component ID discovery** — Users may not know the numeric component ID. Search by module name (`mxcli marketplace search "database connector"`) addresses this, but we may also want to support names directly: `mxcli marketplace install DatabaseConnector`.

6. **Cache strategy** — Should downloaded `.mpk` files be cached in `~/.mxcli/marketplace/{id}/{version}/module.mpk` to avoid re-downloading?
