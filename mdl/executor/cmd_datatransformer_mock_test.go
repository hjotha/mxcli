// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/backend/mock"
	"github.com/mendixlabs/mxcli/model"
)

func TestListDataTransformers_Mock(t *testing.T) {
	mod := mkModule("ETL")
	dt := &model.DataTransformer{
		BaseElement: model.BaseElement{ID: nextID("dt")},
		ContainerID: mod.ID,
		Name:        "TransformOrders",
		SourceType:  "Entity",
	}

	h := mkHierarchy(mod)
	withContainer(h, dt.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc:          func() bool { return true },
		ListDataTransformersFunc: func() ([]*model.DataTransformer, error) { return []*model.DataTransformer{dt}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, listDataTransformers(ctx, ""))

	out := buf.String()
	assertContainsStr(t, out, "Qualified Name")
	assertContainsStr(t, out, "ETL.TransformOrders")
}

func TestListDataTransformers_FilterByModule(t *testing.T) {
	mod1 := mkModule("ETL")
	mod2 := mkModule("Other")
	dt1 := &model.DataTransformer{
		BaseElement: model.BaseElement{ID: nextID("dt")},
		ContainerID: mod1.ID,
		Name:        "TransformOrders",
		SourceType:  "Entity",
	}
	dt2 := &model.DataTransformer{
		BaseElement: model.BaseElement{ID: nextID("dt")},
		ContainerID: mod2.ID,
		Name:        "TransformCustomers",
		SourceType:  "Entity",
	}

	h := mkHierarchy(mod1, mod2)
	withContainer(h, dt1.ContainerID, mod1.ID)
	withContainer(h, dt2.ContainerID, mod2.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc:          func() bool { return true },
		ListDataTransformersFunc: func() ([]*model.DataTransformer, error) { return []*model.DataTransformer{dt1, dt2}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, listDataTransformers(ctx, "ETL"))

	out := buf.String()
	assertContainsStr(t, out, "ETL.TransformOrders")
	assertNotContainsStr(t, out, "Other.TransformCustomers")
}

func TestDescribeDataTransformer_NotFound(t *testing.T) {
	mod := mkModule("ETL")
	h := mkHierarchy(mod)

	mb := &mock.MockBackend{
		IsConnectedFunc:          func() bool { return true },
		ListDataTransformersFunc: func() ([]*model.DataTransformer, error) { return nil, nil },
	}

	ctx, _ := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertError(t, describeDataTransformer(ctx, ast.QualifiedName{Module: "ETL", Name: "NoSuch"}))
}

func TestDescribeDataTransformer_Mock(t *testing.T) {
	mod := mkModule("ETL")
	dt := &model.DataTransformer{
		BaseElement: model.BaseElement{ID: nextID("dt")},
		ContainerID: mod.ID,
		Name:        "TransformOrders",
		SourceType:  "Entity",
	}

	h := mkHierarchy(mod)
	withContainer(h, dt.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc:          func() bool { return true },
		ListDataTransformersFunc: func() ([]*model.DataTransformer, error) { return []*model.DataTransformer{dt}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, describeDataTransformer(ctx, ast.QualifiedName{Module: "ETL", Name: "TransformOrders"}))

	out := buf.String()
	assertContainsStr(t, out, "create data transformer")
}
