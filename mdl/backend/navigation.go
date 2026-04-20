// SPDX-License-Identifier: Apache-2.0

package backend

import (
	"github.com/mendixlabs/mxcli/mdl/types"
	"github.com/mendixlabs/mxcli/model"
)

// NavigationBackend provides navigation document operations.
type NavigationBackend interface {
	ListNavigationDocuments() ([]*types.NavigationDocument, error)
	GetNavigation() (*types.NavigationDocument, error)
	UpdateNavigationProfile(navDocID model.ID, profileName string, spec types.NavigationProfileSpec) error
}
