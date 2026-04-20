// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/backend/mock"
	"github.com/mendixlabs/mxcli/sdk/pages"
)

func TestShowPages_Mock(t *testing.T) {
	mod := mkModule("MyModule")
	pg := mkPage(mod.ID, "Home")

	h := mkHierarchy(mod)
	withContainer(h, pg.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListPagesFunc:   func() ([]*pages.Page, error) { return []*pages.Page{pg}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, showPages(ctx, ""))

	out := buf.String()
	assertContainsStr(t, out, "MyModule.Home")
	assertContainsStr(t, out, "(1 pages)")
}

func TestShowPages_Mock_FilterByModule(t *testing.T) {
	mod1 := mkModule("Sales")
	mod2 := mkModule("HR")
	pg1 := mkPage(mod1.ID, "OrderList")
	pg2 := mkPage(mod2.ID, "EmployeeList")

	h := mkHierarchy(mod1, mod2)
	withContainer(h, pg1.ContainerID, mod1.ID)
	withContainer(h, pg2.ContainerID, mod2.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListPagesFunc:   func() ([]*pages.Page, error) { return []*pages.Page{pg1, pg2}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, showPages(ctx, "HR"))

	out := buf.String()
	assertNotContainsStr(t, out, "Sales.OrderList")
	assertContainsStr(t, out, "HR.EmployeeList")
}

func TestShowSnippets_Mock(t *testing.T) {
	mod := mkModule("MyModule")
	snp := mkSnippet(mod.ID, "Header")

	h := mkHierarchy(mod)
	withContainer(h, snp.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc:  func() bool { return true },
		ListSnippetsFunc: func() ([]*pages.Snippet, error) { return []*pages.Snippet{snp}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, showSnippets(ctx, ""))

	out := buf.String()
	assertContainsStr(t, out, "MyModule.Header")
	assertContainsStr(t, out, "(1 snippets)")
}

func TestShowLayouts_Mock(t *testing.T) {
	mod := mkModule("MyModule")
	lay := mkLayout(mod.ID, "Atlas_Default")

	h := mkHierarchy(mod)
	withContainer(h, lay.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListLayoutsFunc: func() ([]*pages.Layout, error) { return []*pages.Layout{lay}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, showLayouts(ctx, ""))

	out := buf.String()
	assertContainsStr(t, out, "MyModule.Atlas_Default")
	assertContainsStr(t, out, "(1 layouts)")
}
