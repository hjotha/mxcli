// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/mpr"
)

func (m *MockBackend) ListNavigationDocuments() ([]*mpr.NavigationDocument, error) {
	if m.ListNavigationDocumentsFunc != nil {
		return m.ListNavigationDocumentsFunc()
	}
	return nil, nil
}

func (m *MockBackend) GetNavigation() (*mpr.NavigationDocument, error) {
	if m.GetNavigationFunc != nil {
		return m.GetNavigationFunc()
	}
	return nil, nil
}

func (m *MockBackend) UpdateNavigationProfile(navDocID model.ID, profileName string, spec mpr.NavigationProfileSpec) error {
	if m.UpdateNavigationProfileFunc != nil {
		return m.UpdateNavigationProfileFunc(navDocID, profileName, spec)
	}
	return nil
}
