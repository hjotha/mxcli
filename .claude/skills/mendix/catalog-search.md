# Catalog Search

Search and discover services registered in Mendix Catalog programmatically.

## Authentication Required

Catalog search requires a Personal Access Token (PAT):

```bash
# One-time setup
mxcli auth login
```

Create a PAT at: https://user-settings.mendix.com/ (Developer Settings → Personal Access Tokens)

## Basic Usage

```bash
# Search for services
mxcli catalog search "customer"

# Filter by service type
mxcli catalog search "order" --service-type OData
mxcli catalog search "api" --service-type REST

# Production endpoints only
mxcli catalog search "inventory" --production-only

# Services you own
mxcli catalog search "sales" --owned-only

# JSON output for scripting
mxcli catalog search "data" --json | jq '.[] | {name, uuid, type}'
```

## Output Formats

**Table (default):**
```
NAME                   TYPE    VERSION    APPLICATION           ENVIRONMENT   PROD  UUID
CustomerService        OData   1.2.0      CRM Application       Production    Yes   a7f3c2d1
OrderAPI               REST    2.0.1      E-commerce Platform   Acceptance    No    b8e4d3e2
InventorySync          SOAP    1.0.0      Warehouse System      Test          No    c9f5e4f3

Total: 42 results (showing 1-3)
```

- **NAME**: Service name (truncated if > 22 chars)
- **TYPE**: OData, REST, SOAP
- **VERSION**: Service version
- **APPLICATION**: Hosting application name
- **ENVIRONMENT**: Production, Acceptance, Test
- **PROD**: "Yes" if production, blank otherwise
- **UUID**: First 8 characters (full UUID in JSON mode)

**JSON mode:**
```bash
mxcli catalog search "customer" --json
```

Returns full endpoint details including:
- Complete UUIDs
- Descriptions
- Security classification
- Last updated timestamp
- Entity and action metadata (for OData)

## Pagination

```bash
# First 10 results
mxcli catalog search "api" --limit 10

# Next 10 results
mxcli catalog search "api" --limit 10 --offset 10

# Maximum 100 per request
mxcli catalog search "service" --limit 100
```

## Common Use Cases

**Find production OData services:**
```bash
mxcli catalog search "customer" --service-type OData --production-only
```

**Get UUIDs for automation:**
```bash
mxcli catalog search "order" --json | jq -r '.[] | .uuid'
```

**Generate service inventory report:**
```bash
mxcli catalog search "api" --json | \
  jq -r '.[] | "\(.name) (\(.serviceType)) - \(.application.name)"'
```

**Filter by multiple criteria:**
```bash
mxcli catalog search "data" \
  --service-type OData \
  --production-only \
  --limit 50
```

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--profile` | string | "default" | Auth profile name |
| `--service-type` | string | (all) | Filter by OData, REST, or SOAP |
| `--production-only` | bool | false | Show only production endpoints |
| `--owned-only` | bool | false | Show only owned services |
| `--limit` | int | 20 | Results per page (max 100) |
| `--offset` | int | 0 | Pagination offset |
| `--json` | bool | false | Output as JSON array |

## Error Handling

**No credential:**
```
Error: no credential found. Run: mxcli auth login
```

**Authentication failed:**
```
Error: authentication failed. Run: mxcli auth login
```

Solution: Log in with a valid PAT.

**Network errors:**
Catalog API requires internet connectivity. Check network connection and firewall settings.

## Future Features

Phase 2 (not yet implemented):
- `mxcli catalog show <uuid>` - Display detailed endpoint metadata
- `mxcli catalog create-odata-client <uuid>` - Generate OData client from Catalog entry
- Interactive search UI with arrow-key navigation

See GitHub issue #213 for architecture discussion.

## Related

- Platform authentication: `.claude/skills/mendix/platform-auth.md`
- OData client creation: `.claude/skills/mendix/odata-data-sharing.md`
