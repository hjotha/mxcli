// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/backend/mock"
	"github.com/mendixlabs/mxcli/mdl/types"
	"github.com/mendixlabs/mxcli/model"
)

func TestShowImageCollections_Mock(t *testing.T) {
	mod := mkModule("Icons")
	ic := &types.ImageCollection{
		BaseElement: model.BaseElement{ID: nextID("ic")},
		ContainerID: mod.ID,
		Name:        "AppIcons",
		ExportLevel: "Hidden",
	}

	h := mkHierarchy(mod)
	withContainer(h, ic.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc:          func() bool { return true },
		ListImageCollectionsFunc: func() ([]*types.ImageCollection, error) { return []*types.ImageCollection{ic}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, listImageCollections(ctx, ""))

	out := buf.String()
	assertContainsStr(t, out, "Image Collection")
	assertContainsStr(t, out, "Icons.AppIcons")
}

func TestShowImageCollections_FilterByModule(t *testing.T) {
	mod1 := mkModule("Icons")
	mod2 := mkModule("Other")
	ic1 := &types.ImageCollection{
		BaseElement: model.BaseElement{ID: nextID("ic")},
		ContainerID: mod1.ID,
		Name:        "AppIcons",
		ExportLevel: "Hidden",
	}
	ic2 := &types.ImageCollection{
		BaseElement: model.BaseElement{ID: nextID("ic")},
		ContainerID: mod2.ID,
		Name:        "OtherIcons",
		ExportLevel: "Hidden",
	}

	h := mkHierarchy(mod1, mod2)
	withContainer(h, ic1.ContainerID, mod1.ID)
	withContainer(h, ic2.ContainerID, mod2.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc:          func() bool { return true },
		ListImageCollectionsFunc: func() ([]*types.ImageCollection, error) { return []*types.ImageCollection{ic1, ic2}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, listImageCollections(ctx, "Icons"))

	out := buf.String()
	assertContainsStr(t, out, "Icons.AppIcons")
	assertNotContainsStr(t, out, "Other.OtherIcons")
}

func TestDescribeImageCollection_NotFound(t *testing.T) {
	mod := mkModule("Icons")
	h := mkHierarchy(mod)

	mb := &mock.MockBackend{
		IsConnectedFunc:          func() bool { return true },
		ListImageCollectionsFunc: func() ([]*types.ImageCollection, error) { return nil, nil },
	}

	ctx, _ := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertError(t, describeImageCollection(ctx, ast.QualifiedName{Module: "Icons", Name: "NoSuch"}))
}

func TestDescribeImageCollection_Mock(t *testing.T) {
	mod := mkModule("Icons")
	ic := &types.ImageCollection{
		BaseElement: model.BaseElement{ID: nextID("ic")},
		ContainerID: mod.ID,
		Name:        "AppIcons",
		ExportLevel: "Hidden",
	}

	h := mkHierarchy(mod)
	withContainer(h, ic.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc:          func() bool { return true },
		ListImageCollectionsFunc: func() ([]*types.ImageCollection, error) { return []*types.ImageCollection{ic}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, describeImageCollection(ctx, ast.QualifiedName{Module: "Icons", Name: "AppIcons"}))

	out := buf.String()
	assertContainsStr(t, out, "create or replace image collection")
}
