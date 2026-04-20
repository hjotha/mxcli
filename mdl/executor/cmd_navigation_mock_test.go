// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/backend/mock"
	"github.com/mendixlabs/mxcli/sdk/mpr"
)

func TestShowNavigation_Mock(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		GetNavigationFunc: func() (*mpr.NavigationDocument, error) {
			return &mpr.NavigationDocument{
				Profiles: []*mpr.NavigationProfile{{
					Name: "Responsive",
					Kind: "Responsive",
					MenuItems: []*mpr.NavMenuItem{
						{Caption: "Home"},
						{Caption: "Admin"},
						{Caption: "Settings"},
					},
				}},
			}, nil
		},
	}
	ctx, buf := newMockCtx(t, withBackend(mb))
	assertNoError(t, showNavigation(ctx))

	out := buf.String()
	assertContainsStr(t, out, "Profile")
	assertContainsStr(t, out, "Responsive")
}

func TestShowNavigationMenu_Mock(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		GetNavigationFunc: func() (*mpr.NavigationDocument, error) {
			return &mpr.NavigationDocument{
				Profiles: []*mpr.NavigationProfile{{
					Name: "Responsive",
					Kind: "Responsive",
					MenuItems: []*mpr.NavMenuItem{
						{Caption: "Dashboard", Page: "MyModule.Dashboard"},
					},
				}},
			}, nil
		},
	}
	ctx, buf := newMockCtx(t, withBackend(mb))
	assertNoError(t, showNavigationMenu(ctx, nil))
	assertContainsStr(t, buf.String(), "Dashboard")
}

func TestShowNavigationHomes_Mock(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		GetNavigationFunc: func() (*mpr.NavigationDocument, error) {
			return &mpr.NavigationDocument{
				Profiles: []*mpr.NavigationProfile{{
					Name:     "Responsive",
					Kind:     "Responsive",
					HomePage: &mpr.NavHomePage{Page: "MyModule.Home"},
				}},
			}, nil
		},
	}
	ctx, buf := newMockCtx(t, withBackend(mb))
	assertNoError(t, showNavigationHomes(ctx))
	assertContainsStr(t, buf.String(), "Default Home:")
}

func TestDescribeNavigation_Mock(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		GetNavigationFunc: func() (*mpr.NavigationDocument, error) {
			return &mpr.NavigationDocument{
				Profiles: []*mpr.NavigationProfile{{
					Name:     "Responsive",
					Kind:     "Responsive",
					HomePage: &mpr.NavHomePage{Page: "MyModule.Home"},
				}},
			}, nil
		},
	}
	ctx, buf := newMockCtx(t, withBackend(mb))
	assertNoError(t, describeNavigation(ctx, ast.QualifiedName{Name: "Responsive"}))
	assertContainsStr(t, buf.String(), "CREATE OR REPLACE NAVIGATION")
}
