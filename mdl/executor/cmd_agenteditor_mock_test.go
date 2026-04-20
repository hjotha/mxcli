// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/backend/mock"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/agenteditor"
)

func TestShowAgentEditorModels_Mock(t *testing.T) {
	mod := mkModule("M")
	m1 := &agenteditor.Model{
		BaseElement: model.BaseElement{ID: nextID("aem")},
		ContainerID: mod.ID,
		Name:        "GPT4",
		Provider:    "MxCloudGenAI",
		DisplayName: "GPT-4 Turbo",
		Key:         &agenteditor.ConstantRef{QualifiedName: "M.APIKey"},
	}

	h := mkHierarchy(mod)
	withContainer(h, m1.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc:           func() bool { return true },
		ListAgentEditorModelsFunc: func() ([]*agenteditor.Model, error) { return []*agenteditor.Model{m1}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, showAgentEditorModels(ctx, ""))

	out := buf.String()
	assertContainsStr(t, out, "Qualified Name")
	assertContainsStr(t, out, "Module")
	assertContainsStr(t, out, "Provider")
	assertContainsStr(t, out, "Key Constant")
	assertContainsStr(t, out, "Display Name")
	assertContainsStr(t, out, "M.GPT4")
	assertContainsStr(t, out, "MxCloudGenAI")
	assertContainsStr(t, out, "M.APIKey")
	assertContainsStr(t, out, "GPT-4 Turbo")
}

func TestDescribeAgentEditorModel_Mock(t *testing.T) {
	mod := mkModule("M")
	m1 := &agenteditor.Model{
		BaseElement: model.BaseElement{ID: nextID("aem")},
		ContainerID: mod.ID,
		Name:        "GPT4",
		Provider:    "MxCloudGenAI",
		Key:         &agenteditor.ConstantRef{QualifiedName: "M.APIKey"},
	}

	h := mkHierarchy(mod)
	withContainer(h, m1.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc:           func() bool { return true },
		ListAgentEditorModelsFunc: func() ([]*agenteditor.Model, error) { return []*agenteditor.Model{m1}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, describeAgentEditorModel(ctx, ast.QualifiedName{Module: "M", Name: "GPT4"}))

	out := buf.String()
	assertContainsStr(t, out, "CREATE MODEL")
	assertContainsStr(t, out, "Provider")
	assertContainsStr(t, out, "Key")
}

func TestShowAgentEditorAgents_Mock(t *testing.T) {
	mod := mkModule("M")
	a1 := &agenteditor.Agent{
		BaseElement: model.BaseElement{ID: nextID("aea")},
		ContainerID: mod.ID,
		Name:        "MyAgent",
		UsageType:   "Chat",
		Model:       &agenteditor.DocRef{QualifiedName: "M.GPT4"},
		Tools:       []agenteditor.AgentTool{{ID: "t1", Enabled: true}},
		KBTools:     []agenteditor.AgentKBTool{{ID: "kb1", Enabled: true}},
	}

	h := mkHierarchy(mod)
	withContainer(h, a1.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc:           func() bool { return true },
		ListAgentEditorAgentsFunc: func() ([]*agenteditor.Agent, error) { return []*agenteditor.Agent{a1}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, showAgentEditorAgents(ctx, ""))

	out := buf.String()
	assertContainsStr(t, out, "Qualified Name")
	assertContainsStr(t, out, "Usage")
	assertContainsStr(t, out, "Model")
	assertContainsStr(t, out, "Tools")
	assertContainsStr(t, out, "KBs")
	assertContainsStr(t, out, "M.MyAgent")
	assertContainsStr(t, out, "Chat")
	assertContainsStr(t, out, "M.GPT4")
}

func TestDescribeAgentEditorAgent_Mock(t *testing.T) {
	mod := mkModule("M")
	a1 := &agenteditor.Agent{
		BaseElement: model.BaseElement{ID: nextID("aea")},
		ContainerID: mod.ID,
		Name:        "MyAgent",
		UsageType:   "Chat",
		Model:       &agenteditor.DocRef{QualifiedName: "M.GPT4"},
	}

	h := mkHierarchy(mod)
	withContainer(h, a1.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc:           func() bool { return true },
		ListAgentEditorAgentsFunc: func() ([]*agenteditor.Agent, error) { return []*agenteditor.Agent{a1}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, describeAgentEditorAgent(ctx, ast.QualifiedName{Module: "M", Name: "MyAgent"}))

	out := buf.String()
	assertContainsStr(t, out, "CREATE AGENT")
	assertContainsStr(t, out, "UsageType")
	assertContainsStr(t, out, "Model")
}

func TestShowAgentEditorKnowledgeBases_Mock(t *testing.T) {
	mod := mkModule("M")
	kb := &agenteditor.KnowledgeBase{
		BaseElement:      model.BaseElement{ID: nextID("aekb")},
		ContainerID:      mod.ID,
		Name:             "MyKB",
		Provider:         "MxCloudGenAI",
		Key:              &agenteditor.ConstantRef{QualifiedName: "M.KBKey"},
		ModelDisplayName: "text-embedding-ada-002",
	}

	h := mkHierarchy(mod)
	withContainer(h, kb.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc:                   func() bool { return true },
		ListAgentEditorKnowledgeBasesFunc: func() ([]*agenteditor.KnowledgeBase, error) { return []*agenteditor.KnowledgeBase{kb}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, showAgentEditorKnowledgeBases(ctx, ""))

	out := buf.String()
	assertContainsStr(t, out, "Qualified Name")
	assertContainsStr(t, out, "Provider")
	assertContainsStr(t, out, "Key Constant")
	assertContainsStr(t, out, "Embedding Model")
	assertContainsStr(t, out, "M.MyKB")
	assertContainsStr(t, out, "MxCloudGenAI")
	assertContainsStr(t, out, "M.KBKey")
	assertContainsStr(t, out, "text-embedding-ada-002")
}

func TestDescribeAgentEditorKnowledgeBase_Mock(t *testing.T) {
	mod := mkModule("M")
	kb := &agenteditor.KnowledgeBase{
		BaseElement: model.BaseElement{ID: nextID("aekb")},
		ContainerID: mod.ID,
		Name:        "MyKB",
		Provider:    "MxCloudGenAI",
		Key:         &agenteditor.ConstantRef{QualifiedName: "M.KBKey"},
	}

	h := mkHierarchy(mod)
	withContainer(h, kb.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc:                   func() bool { return true },
		ListAgentEditorKnowledgeBasesFunc: func() ([]*agenteditor.KnowledgeBase, error) { return []*agenteditor.KnowledgeBase{kb}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, describeAgentEditorKnowledgeBase(ctx, ast.QualifiedName{Module: "M", Name: "MyKB"}))

	out := buf.String()
	assertContainsStr(t, out, "CREATE KNOWLEDGE BASE")
	assertContainsStr(t, out, "Provider")
}

func TestShowAgentEditorConsumedMCPServices_Mock(t *testing.T) {
	mod := mkModule("M")
	svc := &agenteditor.ConsumedMCPService{
		BaseElement:              model.BaseElement{ID: nextID("aemcp")},
		ContainerID:              mod.ID,
		Name:                     "MySvc",
		ProtocolVersion:          "2025-03-26",
		Version:                  "1.0.0",
		ConnectionTimeoutSeconds: 30,
	}

	h := mkHierarchy(mod)
	withContainer(h, svc.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc:                        func() bool { return true },
		ListAgentEditorConsumedMCPServicesFunc: func() ([]*agenteditor.ConsumedMCPService, error) { return []*agenteditor.ConsumedMCPService{svc}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, showAgentEditorConsumedMCPServices(ctx, ""))

	out := buf.String()
	assertContainsStr(t, out, "Qualified Name")
	assertContainsStr(t, out, "Protocol")
	assertContainsStr(t, out, "Version")
	assertContainsStr(t, out, "Timeout")
	assertContainsStr(t, out, "M.MySvc")
	assertContainsStr(t, out, "2025-03-26")
	assertContainsStr(t, out, "1.0.0")
}

func TestDescribeAgentEditorConsumedMCPService_Mock(t *testing.T) {
	mod := mkModule("M")
	svc := &agenteditor.ConsumedMCPService{
		BaseElement:              model.BaseElement{ID: nextID("aemcp")},
		ContainerID:              mod.ID,
		Name:                     "MySvc",
		ProtocolVersion:          "2025-03-26",
		Version:                  "1.0.0",
		ConnectionTimeoutSeconds: 30,
	}

	h := mkHierarchy(mod)
	withContainer(h, svc.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc:                        func() bool { return true },
		ListAgentEditorConsumedMCPServicesFunc: func() ([]*agenteditor.ConsumedMCPService, error) { return []*agenteditor.ConsumedMCPService{svc}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, describeAgentEditorConsumedMCPService(ctx, ast.QualifiedName{Module: "M", Name: "MySvc"}))

	out := buf.String()
	assertContainsStr(t, out, "CREATE CONSUMED MCP SERVICE")
	assertContainsStr(t, out, "ProtocolVersion")
}
