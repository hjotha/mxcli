// SPDX-License-Identifier: Apache-2.0

package backend

import (
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/mpr"
)

// NavigationBackend provides navigation document operations.
type NavigationBackend interface {
	ListNavigationDocuments() ([]*mpr.NavigationDocument, error)
	GetNavigation() (*mpr.NavigationDocument, error)
	UpdateNavigationProfile(navDocID model.ID, profileName string, spec mpr.NavigationProfileSpec) error
}
