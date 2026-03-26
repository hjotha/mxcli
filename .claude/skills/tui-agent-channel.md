---
name: tui-agent-channel
description: Use when Claude needs to send MDL commands to a running mxcli TUI, execute operations with human oversight, or automate TUI interactions via the agent socket. Also use when setting up Claude-TUI integration for supervised Mendix project modifications.
---

# TUI Agent Channel

Send MDL commands to a running mxcli TUI over a Unix socket. **All agent actions simulate normal user operations** — every command is visible in the TUI, understandable by the human, and interruptible.

## Core Principle

**AI automation = simulated user operations.** No special privileges, no bypassing the UI. Agent actions route through the same views (ExecView, ConfirmView, InputView) that a human uses.

| Agent action | Maps to human action |
|-------------|---------------------|
| `exec` | Human presses `x` → ExecView with MDL → Ctrl+E to execute |
| `list` | Same as `exec` with `SHOW ...` command |
| `describe` | Same as `exec` with `DESCRIBE ...` command |
| `delete` | Human presses `D` → ConfirmView → `y` to confirm |
| `create_module` | Human presses `C` → InputView → Enter to submit |
| `navigate` | Human presses `Space` → jumps to element |
| `state` | Read-only query (no UI side effect) |
| `format` | Pure text computation (no project interaction) |

## Setup

### 1. Start TUI with agent socket

```bash
# Human confirmation mode (default) — user must press Ctrl+E / y / Enter / q for each operation
mxcli tui -p app.mpr --agent-socket /tmp/mxcli-agent.sock

# Auto-proceed mode — views auto-execute but remain visible for review
mxcli tui -p app.mpr --agent-socket /tmp/mxcli-agent.sock --agent-auto-proceed
```

### 2. Send commands via socat

```bash
# All UI-visible actions need socat -t timeout (they go through views)
printf '{"id":1,"action":"exec","mdl":"SHOW ENTITIES"}\n' | socat -t 30 - UNIX-CONNECT:/tmp/mxcli-agent.sock

# check uses Docker mx check — allow longer timeout
printf '{"id":2,"action":"check"}\n' | socat -t 120 - UNIX-CONNECT:/tmp/mxcli-agent.sock

# state and format are instant (no UI round-trip)
echo '{"id":3,"action":"state"}' | socat - UNIX-CONNECT:/tmp/mxcli-agent.sock
```

## Protocol

JSON-line protocol over Unix socket. One request per line, one JSON response per line.

### Request

```json
{"id": 1, "action": "exec", "mdl": "CREATE ENTITY Mod.E (Name: String(100));"}
```

| Field | Required by | Description |
|-------|-------------|-------------|
| `id` | all | Unique request ID (nonzero integer) |
| `action` | all | Action name (see Actions table) |
| `mdl` | `exec`, `format` | MDL statement(s) or text |
| `target` | `navigate`, `delete`, `describe`, `list` | Element reference (see Target Format) |
| `name` | `create_module` | Module name to create |

### Target Format

Targets use `type:qualifiedName` format:

| Example | Meaning |
|---------|---------|
| `entity:Module.Entity` | A specific entity |
| `microflow:Module.MF_Name` | A specific microflow |
| `module:MyModule` | A module |
| `entities` | All entities (for `list`) |
| `entities:MyModule` | Entities in a module (for `list`) |

Supported types for `delete`/`describe`: `entity`, `association`, `enumeration`, `constant`, `microflow`, `page`, `snippet`, `workflow`, `imagecollection`, `javaaction`, `module`

Supported types for `list`: plural forms (`entities`, `microflows`, `pages`, `modules`, etc.)

### Response

```json
{"id": 1, "ok": true, "result": "Created entity: Mod.E", "mode": "overlay:exec-result", "changes": [{"action":"created","target":"entity: Mod.E"}]}
```

| Field | Description |
|-------|-------------|
| `id` | Echoed request ID |
| `ok` | `true` if operation succeeded |
| `result` | Output text (same as TUI overlay content) |
| `error` | Error message (when `ok` is `false`) |
| `mode` | TUI state after operation |
| `changes` | Structured changes array (write operations only, when applicable) |

### State Response

The `state` action returns a structured JSON object:

```json
{
  "mode": "Browse",
  "project": "/path/to/app.mpr",
  "selectedNode": {"type": "entity", "qualifiedName": "MyModule.Customer"},
  "previewMode": "MDL",
  "checkErrors": 0,
  "checkRunning": false
}
```

## Actions

| Action | UI View | Visible | Example |
|--------|---------|---------|---------|
| `exec` | ExecView → Overlay | yes | `{"id":1,"action":"exec","mdl":"CREATE ENTITY M.E (X: String(100));"}` |
| `list` | ExecView → Overlay | yes | `{"id":2,"action":"list","target":"entities:MyModule"}` |
| `describe` | ExecView → Overlay | yes | `{"id":3,"action":"describe","target":"entity:M.E"}` |
| `check` | Status bar → Overlay | yes | `{"id":4,"action":"check"}` |
| `delete` | ConfirmView → Overlay | yes | `{"id":5,"action":"delete","target":"entity:M.E"}` |
| `create_module` | InputView → Overlay | yes | `{"id":6,"action":"create_module","name":"NewModule"}` |
| `navigate` | Browser (miller columns) | yes | `{"id":7,"action":"navigate","target":"entity:M.E"}` |
| `state` | — | no | `{"id":8,"action":"state"}` |
| `format` | — | no | `{"id":9,"action":"format","mdl":"create entity m.e(x:string(100));"}` |

## Human Confirmation Flow

### Without `--agent-auto-proceed` (default)

1. Agent sends command via socket
2. TUI pushes the appropriate view (ExecView/ConfirmView/InputView)
3. **Human must act**: press Ctrl+E to execute, `y` to confirm delete, Enter to submit, or Esc to cancel
4. Result displayed in overlay — human presses `q`/`Esc` to dismiss
5. Response sent back to agent

**Cancellation**: If human presses Esc in any view, agent receives `{"ok":false,"error":"cancelled by user"}`.

### With `--agent-auto-proceed`

1. Agent sends command via socket
2. TUI pushes the view and **auto-triggers** the action (Ctrl+E / y / Enter)
3. Result displayed in overlay — response sent immediately to agent
4. Human can still review the overlay and press `q`/`Esc` to dismiss at leisure

The status bar shows `⚡agent` badge while an agent operation is in progress.

## Typical Claude Workflow

```bash
# 1. Check current state
echo '{"id":1,"action":"state"}' | socat - UNIX-CONNECT:/tmp/mxcli-agent.sock

# 2. List entities in a module (human sees ExecView + Overlay)
printf '{"id":2,"action":"list","target":"entities:MyModule"}\n' | socat -t 30 - UNIX-CONNECT:/tmp/mxcli-agent.sock

# 3. Describe an entity (human sees ExecView + Overlay)
printf '{"id":3,"action":"describe","target":"entity:MyModule.Customer"}\n' | socat -t 30 - UNIX-CONNECT:/tmp/mxcli-agent.sock

# 4. Create a new module (human sees InputView + Overlay)
printf '{"id":4,"action":"create_module","name":"Orders"}\n' | socat -t 30 - UNIX-CONNECT:/tmp/mxcli-agent.sock

# 5. Execute MDL (human sees ExecView + Overlay)
printf '{"id":5,"action":"exec","mdl":"CREATE ENTITY Orders.Order (OrderNo: String(20), Total: Decimal);"}\n' | socat -t 30 - UNIX-CONNECT:/tmp/mxcli-agent.sock

# 6. Delete an entity (human sees ConfirmView + Overlay)
printf '{"id":6,"action":"delete","target":"entity:Orders.OldEntity"}\n' | socat -t 30 - UNIX-CONNECT:/tmp/mxcli-agent.sock

# 7. Format MDL text (instant, no UI)
echo '{"id":7,"action":"format","mdl":"create entity m.e(x:string(100));"}' | socat - UNIX-CONNECT:/tmp/mxcli-agent.sock

# 8. Navigate to see result (human sees browser jump)
echo '{"id":8,"action":"navigate","target":"entity:Orders.Order"}' | socat - UNIX-CONNECT:/tmp/mxcli-agent.sock

# 9. Verify with check (human sees status bar + Overlay)
printf '{"id":9,"action":"check"}\n' | socat -t 120 - UNIX-CONNECT:/tmp/mxcli-agent.sock
```

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Missing newline after JSON | Append `\n` — protocol is line-delimited |
| `id: 0` | ID must be nonzero |
| `exec` without `mdl` field | Always include MDL text for exec action |
| Socket not found | Ensure TUI is running with `--agent-socket` |
| Response never arrives | In confirmation mode, human must act in the view (Ctrl+E / y / Enter / q) |
| `echo` pipe closes before response | Use `printf '...\n' \| socat -t 30` for all UI-visible actions |
| Overlay stays open after auto-proceed | By design — human can still review; press `q`/`Esc` to dismiss |
| Sending navigate while overlay is open | Close overlay first (`Esc`/`q`) — navigate only works in Browse mode |
| `delete` target missing colon | Use `entity:Module.Entity` format, not just `Module.Entity` |
| `list` using singular type | Use plural: `entities` not `entity`, `microflows` not `microflow` |
| `list`/`describe` timeout | These now go through ExecView — use `socat -t 30`, not bare `echo` |

## Architecture Notes

**No special privileges**: All agent actions (except `state` and `format`) go through the same bubbletea views that humans use. `list` and `describe` are converted to `SHOW`/`DESCRIBE` MDL commands and routed through `AgentExecMsg` → ExecView. `delete` goes through ConfirmView. `create_module` goes through InputView.

**agentExecContext**: Tracks agent-initiated UI operations. When an agent action pushes a view (ExecView/ConfirmView/InputView), the response channel is stored in `agentExecContext`. When the view completes (`execShowResultMsg`), the response is sent back. If the user cancels (`PopViewMsg` with Esc), a rejection response is sent.

**bubbletea model copy**: bubbletea copies the model at `tea.NewProgram()` time. The `agentAutoProceed` flag is set via `SetAgentAutoProceed()` before `NewProgram`. The `agentListener` (set after) is only used for lifecycle management.

**check action uses Docker `mx check`**: Triggers `runMxCheck()` which uses Docker-based `mx check` (the Mendix project checker), not `mxcli check` (MDL syntax checker). Result returned via `MxCheckResultMsg` and forwarded through `agentCheckCh`.

**Status bar badge**: `⚡agent` badge shown in the status bar while `agentExecCtx` is non-nil (agent operation in progress).

**Structured changes**: Write operations (`exec`, `delete`, `create_module`) include a `changes` array extracted by regex-matching output lines like "Created entity: Mod.E".

## Key Files

- `cmd/mxcli/tui/agent_protocol.go` — Request/Response types, `parseTarget()`, `buildListCmd()`, `buildAgentDescribeCmd()`
- `cmd/mxcli/tui/agent_msgs.go` — tea.Msg types for agent actions
- `cmd/mxcli/tui/agent_listener.go` — Unix socket server, sync handler (format only), async dispatch for all other actions
- `cmd/mxcli/tui/app.go` — Update handlers, `agentExecContext`, `agentBuildState()`, `agentParseChanges()`
- `cmd/mxcli/tui/execview.go` — ExecView (used by exec/list/describe)
- `cmd/mxcli/tui/confirmview.go` — ConfirmView (used by delete), `buildDropCmd()`
- `cmd/mxcli/tui/inputview.go` — InputView (used by create_module)
- `cmd/mxcli/cmd_tui.go` — CLI flags (`--agent-socket`, `--agent-auto-proceed`)
