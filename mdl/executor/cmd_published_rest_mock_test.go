// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/backend/mock"
	"github.com/mendixlabs/mxcli/model"
)

func TestShowPublishedRestServices_Mock(t *testing.T) {
	mod := mkModule("MyModule")
	svc := &model.PublishedRestService{
		BaseElement: model.BaseElement{ID: nextID("prs")},
		ContainerID: mod.ID,
		Name:        "OrderAPI",
		Path:        "/rest/orders/v1",
		Version:     "1.0",
		Resources: []*model.PublishedRestResource{
			{Name: "Orders"},
			{Name: "Items"},
		},
	}
	h := mkHierarchy(mod)
	withContainer(h, svc.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListPublishedRestServicesFunc: func() ([]*model.PublishedRestService, error) {
			return []*model.PublishedRestService{svc}, nil
		},
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, showPublishedRestServices(ctx, ""))

	out := buf.String()
	assertContainsStr(t, out, "QualifiedName")
	assertContainsStr(t, out, "MyModule.OrderAPI")
	assertContainsStr(t, out, "(1 published REST services)")
}

func TestDescribePublishedRestService_Mock(t *testing.T) {
	mod := mkModule("MyModule")
	svc := &model.PublishedRestService{
		BaseElement: model.BaseElement{ID: nextID("prs")},
		ContainerID: mod.ID,
		Name:        "OrderAPI",
		Path:        "/rest/orders/v1",
		Version:     "1.0",
	}
	h := mkHierarchy(mod)
	withContainer(h, svc.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListPublishedRestServicesFunc: func() ([]*model.PublishedRestService, error) {
			return []*model.PublishedRestService{svc}, nil
		},
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, describePublishedRestService(ctx, ast.QualifiedName{Module: "MyModule", Name: "OrderAPI"}))

	out := buf.String()
	assertContainsStr(t, out, "CREATE PUBLISHED REST SERVICE")
	assertContainsStr(t, out, "MyModule.OrderAPI")
}
