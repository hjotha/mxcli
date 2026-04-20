// SPDX-License-Identifier: Apache-2.0

package backend

import (
	"github.com/mendixlabs/mxcli/mdl/types"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/workflows"
)

// WorkflowBackend provides workflow operations.
type WorkflowBackend interface {
	ListWorkflows() ([]*workflows.Workflow, error)
	GetWorkflow(id model.ID) (*workflows.Workflow, error)
	CreateWorkflow(wf *workflows.Workflow) error
	DeleteWorkflow(id model.ID) error
}

// SettingsBackend provides project settings operations.
type SettingsBackend interface {
	GetProjectSettings() (*model.ProjectSettings, error)
	UpdateProjectSettings(ps *model.ProjectSettings) error
}

// ImageBackend provides image collection operations.
type ImageBackend interface {
	ListImageCollections() ([]*types.ImageCollection, error)
	CreateImageCollection(ic *types.ImageCollection) error
	DeleteImageCollection(id string) error
}

// ScheduledEventBackend provides scheduled event operations.
type ScheduledEventBackend interface {
	ListScheduledEvents() ([]*model.ScheduledEvent, error)
	GetScheduledEvent(id model.ID) (*model.ScheduledEvent, error)
}
