// SPDX-License-Identifier: Apache-2.0

package backend

import (
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/agenteditor"
	"github.com/mendixlabs/mxcli/sdk/mpr"
)

// RenameBackend provides cross-cutting rename and reference-update operations.
type RenameBackend interface {
	UpdateQualifiedNameInAllUnits(oldName, newName string) (int, error)
	RenameReferences(oldName, newName string, dryRun bool) ([]mpr.RenameHit, error)
	RenameDocumentByName(moduleName, oldName, newName string) error
}

// RawUnitBackend provides low-level unit access for operations that
// manipulate raw BSON (e.g. widget patching, alter page/workflow).
type RawUnitBackend interface {
	GetRawUnit(id model.ID) (map[string]any, error)
	GetRawUnitBytes(id model.ID) ([]byte, error)
	ListRawUnitsByType(typePrefix string) ([]*mpr.RawUnit, error)
	ListRawUnits(objectType string) ([]*mpr.RawUnitInfo, error)
	GetRawUnitByName(objectType, qualifiedName string) (*mpr.RawUnitInfo, error)
	GetRawMicroflowByName(qualifiedName string) ([]byte, error)
	UpdateRawUnit(unitID string, contents []byte) error
}

// MetadataBackend provides project-level metadata and introspection.
type MetadataBackend interface {
	ListAllUnitIDs() ([]string, error)
	ListUnits() ([]*mpr.UnitInfo, error)
	GetUnitTypes() (map[string]int, error)
	GetProjectRootID() (string, error)
	ContentsDir() string
	ExportJSON() ([]byte, error)
	InvalidateCache()
}

// WidgetBackend provides widget introspection operations.
type WidgetBackend interface {
	FindCustomWidgetType(widgetID string) (*mpr.RawCustomWidgetType, error)
	FindAllCustomWidgetTypes(widgetID string) ([]*mpr.RawCustomWidgetType, error)
}

// AgentEditorBackend provides agent editor document operations.
type AgentEditorBackend interface {
	ListAgentEditorModels() ([]*agenteditor.Model, error)
	ListAgentEditorKnowledgeBases() ([]*agenteditor.KnowledgeBase, error)
	ListAgentEditorConsumedMCPServices() ([]*agenteditor.ConsumedMCPService, error)
	ListAgentEditorAgents() ([]*agenteditor.Agent, error)
}
