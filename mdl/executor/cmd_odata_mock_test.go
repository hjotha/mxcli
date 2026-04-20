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
	assertNoError(t, showODataClients(ctx, ""))

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
	assertNoError(t, showODataServices(ctx, ""))

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
	assertContainsStr(t, out, "CREATE ODATA CLIENT")
	assertContainsStr(t, out, "MyModule.PetStoreClient")
	assertContainsStr(t, out, "https://example.com/$metadata")
	assertContainsStr(t, out, "2.0")
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
	assertContainsStr(t, out, "CREATE ODATA SERVICE")
	assertContainsStr(t, out, "MyModule.CatalogService")
}
