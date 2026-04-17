// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"github.com/mendixlabs/mxcli/sdk/mpr"
	"github.com/mendixlabs/mxcli/sdk/mpr/version"
)

func (m *MockBackend) Connect(path string) error {
	if m.ConnectFunc != nil {
		return m.ConnectFunc(path)
	}
	return nil
}

func (m *MockBackend) Disconnect() error {
	if m.DisconnectFunc != nil {
		return m.DisconnectFunc()
	}
	return nil
}

func (m *MockBackend) Commit() error {
	if m.CommitFunc != nil {
		return m.CommitFunc()
	}
	return nil
}

func (m *MockBackend) IsConnected() bool {
	if m.IsConnectedFunc != nil {
		return m.IsConnectedFunc()
	}
	return false
}

func (m *MockBackend) Path() string {
	if m.PathFunc != nil {
		return m.PathFunc()
	}
	return ""
}

func (m *MockBackend) Version() mpr.MPRVersion {
	if m.VersionFunc != nil {
		return m.VersionFunc()
	}
	var zero mpr.MPRVersion
	return zero
}

func (m *MockBackend) ProjectVersion() *version.ProjectVersion {
	if m.ProjectVersionFunc != nil {
		return m.ProjectVersionFunc()
	}
	return nil
}

func (m *MockBackend) GetMendixVersion() (string, error) {
	if m.GetMendixVersionFunc != nil {
		return m.GetMendixVersionFunc()
	}
	return "", nil
}
