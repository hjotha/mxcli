// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/backend/mock"
	"github.com/mendixlabs/mxcli/model"
)

func TestShowRestClients_Mock(t *testing.T) {
	mod := mkModule("MyModule")
	svc := &model.ConsumedRestService{
		BaseElement: model.BaseElement{ID: nextID("crs")},
		ContainerID: mod.ID,
		Name:        "WeatherAPI",
		BaseUrl:     "https://api.weather.com",
	}
	h := mkHierarchy(mod)
	withContainer(h, svc.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListConsumedRestServicesFunc: func() ([]*model.ConsumedRestService, error) {
			return []*model.ConsumedRestService{svc}, nil
		},
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, listRestClients(ctx, ""))

	out := buf.String()
	assertContainsStr(t, out, "QualifiedName")
	assertContainsStr(t, out, "MyModule.WeatherAPI")
}

func TestDescribeRestClient_Mock(t *testing.T) {
	mod := mkModule("MyModule")
	svc := &model.ConsumedRestService{
		BaseElement: model.BaseElement{ID: nextID("crs")},
		ContainerID: mod.ID,
		Name:        "WeatherAPI",
		BaseUrl:     "https://api.weather.com",
	}
	h := mkHierarchy(mod)
	withContainer(h, svc.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListConsumedRestServicesFunc: func() ([]*model.ConsumedRestService, error) {
			return []*model.ConsumedRestService{svc}, nil
		},
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, describeRestClient(ctx, ast.QualifiedName{Module: "MyModule", Name: "WeatherAPI"}))

	out := buf.String()
	assertContainsStr(t, out, "create rest client")
	assertContainsStr(t, out, "MyModule.WeatherAPI")
}
