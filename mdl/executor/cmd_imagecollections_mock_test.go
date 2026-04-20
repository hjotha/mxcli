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
	assertNoError(t, showImageCollections(ctx, ""))

	out := buf.String()
	assertContainsStr(t, out, "Image Collection")
	assertContainsStr(t, out, "Icons.AppIcons")
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
	assertContainsStr(t, out, "CREATE OR REPLACE IMAGE COLLECTION")
}
