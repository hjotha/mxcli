// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/mpr"
)

// ---------------------------------------------------------------------------
// ModuleBackend
// ---------------------------------------------------------------------------

func (m *MockBackend) ListModules() ([]*model.Module, error) {
	if m.ListModulesFunc != nil {
		return m.ListModulesFunc()
	}
	return nil, nil
}

func (m *MockBackend) GetModule(id model.ID) (*model.Module, error) {
	if m.GetModuleFunc != nil {
		return m.GetModuleFunc(id)
	}
	return nil, nil
}

func (m *MockBackend) GetModuleByName(name string) (*model.Module, error) {
	if m.GetModuleByNameFunc != nil {
		return m.GetModuleByNameFunc(name)
	}
	return nil, nil
}

func (m *MockBackend) CreateModule(module *model.Module) error {
	if m.CreateModuleFunc != nil {
		return m.CreateModuleFunc(module)
	}
	return nil
}

func (m *MockBackend) UpdateModule(module *model.Module) error {
	if m.UpdateModuleFunc != nil {
		return m.UpdateModuleFunc(module)
	}
	return nil
}

func (m *MockBackend) DeleteModule(id model.ID) error {
	if m.DeleteModuleFunc != nil {
		return m.DeleteModuleFunc(id)
	}
	return nil
}

func (m *MockBackend) DeleteModuleWithCleanup(id model.ID, moduleName string) error {
	if m.DeleteModuleWithCleanupFunc != nil {
		return m.DeleteModuleWithCleanupFunc(id, moduleName)
	}
	return nil
}

// ---------------------------------------------------------------------------
// FolderBackend
// ---------------------------------------------------------------------------

func (m *MockBackend) ListFolders() ([]*mpr.FolderInfo, error) {
	if m.ListFoldersFunc != nil {
		return m.ListFoldersFunc()
	}
	return nil, nil
}

func (m *MockBackend) CreateFolder(folder *model.Folder) error {
	if m.CreateFolderFunc != nil {
		return m.CreateFolderFunc(folder)
	}
	return nil
}

func (m *MockBackend) DeleteFolder(id model.ID) error {
	if m.DeleteFolderFunc != nil {
		return m.DeleteFolderFunc(id)
	}
	return nil
}

func (m *MockBackend) MoveFolder(id model.ID, newContainerID model.ID) error {
	if m.MoveFolderFunc != nil {
		return m.MoveFolderFunc(id, newContainerID)
	}
	return nil
}
