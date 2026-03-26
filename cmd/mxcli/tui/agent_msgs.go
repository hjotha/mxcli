package tui

import tea "github.com/charmbracelet/bubbletea"

// AgentExecMsg requests MDL execution from an external agent.
type AgentExecMsg struct {
	RequestID  int
	MDL        string
	ResponseCh chan<- AgentResponse
}

// AgentCheckMsg requests a syntax/reference check.
type AgentCheckMsg struct {
	RequestID  int
	ResponseCh chan<- AgentResponse
}

// AgentStateMsg requests current TUI state (active view, project path, etc.).
type AgentStateMsg struct {
	RequestID  int
	ResponseCh chan<- AgentResponse
}

// AgentNavigateMsg requests navigation to a specific element.
type AgentNavigateMsg struct {
	RequestID  int
	Target     string
	ResponseCh chan<- AgentResponse
}

// AgentDeleteMsg requests deletion of an element via DROP command.
type AgentDeleteMsg struct {
	RequestID  int
	Target     string // "entity:Module.Entity"
	ResponseCh chan<- AgentResponse
}

// AgentCreateModuleMsg requests creation of a new module.
type AgentCreateModuleMsg struct {
	RequestID  int
	Name       string
	ResponseCh chan<- AgentResponse
}

// Ensure messages satisfy tea.Msg.
var (
	_ tea.Msg = AgentExecMsg{}
	_ tea.Msg = AgentCheckMsg{}
	_ tea.Msg = AgentStateMsg{}
	_ tea.Msg = AgentNavigateMsg{}
	_ tea.Msg = AgentDeleteMsg{}
	_ tea.Msg = AgentCreateModuleMsg{}
)
