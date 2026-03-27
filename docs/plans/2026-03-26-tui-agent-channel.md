# TUI Agent Communication Channel — Design Rationale

## Goal

Add a Unix socket-based agent communication channel to the mxcli TUI, allowing Claude (or other agents) to send MDL commands and receive structured results while humans observe execution in real-time.

## Architecture

```
Agent (Claude)  ──JSON──>  Unix Socket  ──tea.Msg──>  Bubbletea Event Loop
                                                          │
                                                     ExecView / ConfirmView
                                                          │
                <──JSON──  Response Channel  <──result──  App.Update
```

- **Transport:** Unix domain socket at `/tmp/mxcli-agent.sock`
- **Protocol:** JSON-line (one JSON object per line, newline-delimited)
- **Integration:** Agent messages are converted to `tea.Msg` via `tea.Program.Send()`, reusing existing UI infrastructure
- **Confirmation:** Human-in-the-loop by default; `--agent-auto` flag for unattended operation

## Key Design Decisions

1. **Reuse ExecView path** — Agent exec flows through the same ExecView that humans use, ensuring consistent behavior and making agent actions visible in the TUI.

2. **Human confirmation gate** — By default, agent actions require the user to observe the result and dismiss the overlay. This prevents runaway automation.

3. **Dedicated message types** — Each agent action (exec, check, navigate, delete, create_module) has its own `tea.Msg` type rather than simulating keystrokes, making the protocol robust against UI changes.

4. **Socket permissions** — Socket is created with `0600` permissions to prevent unauthorized access on shared systems.

5. **Timeout** — Server-side 120s timeout matches the documented `socat -t 120` client timeout for long-running operations like `mx check`.

## Supported Actions

| Action | Description | Response |
|--------|-------------|----------|
| `exec` | Execute MDL script | Result text + success flag |
| `check` | Run `mx check` validation | Check results with error details |
| `state` | Query current TUI state | JSON with mode, project, selection |
| `navigate` | Jump to a project element | Confirmation |
| `delete` | DROP an element (with confirm) | Result of DROP command |
| `create_module` | Create a new module | Result of CREATE MODULE |

## File Structure

- `tui/agent_msgs.go` — Message types for bubbletea integration
- `tui/agent_protocol.go` — Request/response JSON types, action routing
- `tui/agent_listener.go` — Unix socket listener, connection handling
- `tui/agent_*_test.go` — Protocol, listener, and integration tests

## Usage

```bash
# Start TUI with agent socket
mxcli tui -p app.mpr --agent

# Auto-proceed mode (no human confirmation)
mxcli tui -p app.mpr --agent --agent-auto

# Send command from another terminal
printf '{"id":1,"action":"exec","mdl":"SHOW ENTITIES"}\n' | socat -t 120 - UNIX-CONNECT:/tmp/mxcli-agent.sock
```
