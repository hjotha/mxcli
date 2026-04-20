// SPDX-License-Identifier: Apache-2.0

package backend

import (
	"github.com/mendixlabs/mxcli/mdl/types"
	"github.com/mendixlabs/mxcli/model"
)

// ConnectionBackend manages the lifecycle of a backend connection.
type ConnectionBackend interface {
	// Connect opens a connection to the project at path.
	Connect(path string) error
	// Disconnect closes the connection, finalizing any pending work.
	Disconnect() error
	// Commit flushes any pending writes. Implementations that auto-commit
	// (e.g. MprBackend) may treat this as a no-op.
	Commit() error
	// IsConnected reports whether the backend has an active connection.
	IsConnected() bool
	// Path returns the path of the connected project, or "" if not connected.
	Path() string
	// Version returns the MPR format version.
	Version() types.MPRVersion
	// ProjectVersion returns the Mendix project version.
	ProjectVersion() *types.ProjectVersion
	// GetMendixVersion returns the Mendix version string.
	GetMendixVersion() (string, error)
}

// ModuleBackend provides module-level operations.
type ModuleBackend interface {
	ListModules() ([]*model.Module, error)
	GetModule(id model.ID) (*model.Module, error)
	GetModuleByName(name string) (*model.Module, error)
	CreateModule(module *model.Module) error
	UpdateModule(module *model.Module) error
	DeleteModule(id model.ID) error
	DeleteModuleWithCleanup(id model.ID, moduleName string) error
}

// FolderBackend provides folder operations.
type FolderBackend interface {
	ListFolders() ([]*types.FolderInfo, error)
	CreateFolder(folder *model.Folder) error
	DeleteFolder(id model.ID) error
	MoveFolder(id model.ID, newContainerID model.ID) error
}
