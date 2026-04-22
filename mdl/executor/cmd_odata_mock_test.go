// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/backend/mock"
	"github.com/mendixlabs/mxcli/model"
)

func TestShowODataClients_Mock(t *testing.T) {
	mod := mkModule("MyModule")
	svc := &model.ConsumedODataService{
		BaseElement:  model.BaseElement{ID: nextID("cos")},
		ContainerID:  mod.ID,
		Name:         "PetStoreClient",
		MetadataUrl:  "https://example.com/$metadata",
		Version:      "1.0",
		ODataVersion: "4.0",
	}
	h := mkHierarchy(mod)
	withContainer(h, svc.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListConsumedODataServicesFunc: func() ([]*model.ConsumedODataService, error) {
			return []*model.ConsumedODataService{svc}, nil
		},
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, listODataClients(ctx, ""))

	out := buf.String()
	assertContainsStr(t, out, "QualifiedName")
	assertContainsStr(t, out, "MyModule.PetStoreClient")
}

func TestShowODataServices_Mock(t *testing.T) {
	mod := mkModule("MyModule")
	svc := &model.PublishedODataService{
		BaseElement:  model.BaseElement{ID: nextID("pos")},
		ContainerID:  mod.ID,
		Name:         "CatalogService",
		Path:         "/odata/v1",
		Version:      "1.0",
		ODataVersion: "4.0",
	}
	h := mkHierarchy(mod)
	withContainer(h, svc.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListPublishedODataServicesFunc: func() ([]*model.PublishedODataService, error) {
			return []*model.PublishedODataService{svc}, nil
		},
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, listODataServices(ctx, ""))

	out := buf.String()
	assertContainsStr(t, out, "QualifiedName")
	assertContainsStr(t, out, "MyModule.CatalogService")
}

func TestDescribeODataClient_Mock(t *testing.T) {
	mod := mkModule("MyModule")
	svc := &model.ConsumedODataService{
		BaseElement:  model.BaseElement{ID: nextID("cos")},
		ContainerID:  mod.ID,
		Name:         "PetStoreClient",
		MetadataUrl:  "https://example.com/$metadata",
		Version:      "2.0",
		ODataVersion: "4.0",
	}
	h := mkHierarchy(mod)
	withContainer(h, svc.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListConsumedODataServicesFunc: func() ([]*model.ConsumedODataService, error) {
			return []*model.ConsumedODataService{svc}, nil
		},
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, describeODataClient(ctx, ast.QualifiedName{Module: "MyModule", Name: "PetStoreClient"}))

	out := buf.String()
	assertContainsStr(t, out, "create odata client")
	assertContainsStr(t, out, "MyModule.PetStoreClient")
	assertContainsStr(t, out, "https://example.com/$metadata")
	assertContainsStr(t, out, "2.0")
}

func TestDescribeODataClient_NotFound(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListConsumedODataServicesFunc: func() ([]*model.ConsumedODataService, error) {
			return nil, nil
		},
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, describeODataClient(ctx, ast.QualifiedName{Module: "MyModule", Name: "NoSuch"}))
}

func TestShowODataClients_FilterByModule(t *testing.T) {
	mod1 := mkModule("Alpha")
	mod2 := mkModule("Beta")
	svc1 := &model.ConsumedODataService{
		BaseElement: model.BaseElement{ID: nextID("cos")},
		ContainerID: mod1.ID,
		Name:        "AlphaSvc",
	}
	svc2 := &model.ConsumedODataService{
		BaseElement: model.BaseElement{ID: nextID("cos")},
		ContainerID: mod2.ID,
		Name:        "BetaSvc",
	}
	h := mkHierarchy(mod1, mod2)
	withContainer(h, svc1.ContainerID, mod1.ID)
	withContainer(h, svc2.ContainerID, mod2.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListConsumedODataServicesFunc: func() ([]*model.ConsumedODataService, error) {
			return []*model.ConsumedODataService{svc1, svc2}, nil
		},
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, listODataClients(ctx, "Alpha"))

	out := buf.String()
	assertContainsStr(t, out, "Alpha.AlphaSvc")
	assertNotContainsStr(t, out, "Beta.BetaSvc")
}

func TestShowODataServices_FilterByModule(t *testing.T) {
	mod := mkModule("Sales")
	svc := &model.PublishedODataService{
		BaseElement: model.BaseElement{ID: nextID("pos")},
		ContainerID: mod.ID,
		Name:        "SalesSvc",
	}
	h := mkHierarchy(mod)
	withContainer(h, svc.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListPublishedODataServicesFunc: func() ([]*model.PublishedODataService, error) {
			return []*model.PublishedODataService{svc}, nil
		},
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, listODataServices(ctx, "Sales"))
	assertContainsStr(t, buf.String(), "Sales.SalesSvc")
}

func TestDescribeODataService_Mock(t *testing.T) {
	mod := mkModule("MyModule")
	svc := &model.PublishedODataService{
		BaseElement:  model.BaseElement{ID: nextID("pos")},
		ContainerID:  mod.ID,
		Name:         "CatalogService",
		Path:         "/odata/v1",
		Version:      "1.0",
		ODataVersion: "4.0",
		Namespace:    "MyApp",
	}
	h := mkHierarchy(mod)
	withContainer(h, svc.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListPublishedODataServicesFunc: func() ([]*model.PublishedODataService, error) {
			return []*model.PublishedODataService{svc}, nil
		},
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, describeODataService(ctx, ast.QualifiedName{Module: "MyModule", Name: "CatalogService"}))

	out := buf.String()
	assertContainsStr(t, out, "create odata service")
	assertContainsStr(t, out, "MyModule.CatalogService")
}

func TestDescribeODataService_NotFound(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListPublishedODataServicesFunc: func() ([]*model.PublishedODataService, error) {
			return nil, nil
		},
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, describeODataService(ctx, ast.QualifiedName{Module: "X", Name: "NoSuch"}))
}
