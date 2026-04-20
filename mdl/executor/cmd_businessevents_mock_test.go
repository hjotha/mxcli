// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/backend/mock"
	"github.com/mendixlabs/mxcli/model"
)

func TestShowBusinessEventServices_Mock(t *testing.T) {
	mod := mkModule("MyModule")
	svc := &model.BusinessEventService{
		BaseElement: model.BaseElement{ID: nextID("bes")},
		ContainerID: mod.ID,
		Name:        "OrderEvents",
	}
	h := mkHierarchy(mod)
	withContainer(h, svc.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListBusinessEventServicesFunc: func() ([]*model.BusinessEventService, error) {
			return []*model.BusinessEventService{svc}, nil
		},
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, showBusinessEventServices(ctx, ""))

	out := buf.String()
	assertContainsStr(t, out, "QualifiedName")
	assertContainsStr(t, out, "MyModule.OrderEvents")
}

func TestShowBusinessEventClients_Mock(t *testing.T) {
	ctx, buf := newMockCtx(t)
	assertNoError(t, showBusinessEventClients(ctx, ""))
	assertContainsStr(t, buf.String(), "not yet implemented")
}

func TestDescribeBusinessEventService_Mock(t *testing.T) {
	mod := mkModule("MyModule")
	svc := &model.BusinessEventService{
		BaseElement: model.BaseElement{ID: nextID("bes")},
		ContainerID: mod.ID,
		Name:        "OrderEvents",
		Definition: &model.BusinessEventDefinition{
			ServiceName:     "com.example.orders",
			EventNamePrefix: "order",
			Channels: []*model.BusinessEventChannel{
				{
					ChannelName: "ch1",
					Messages: []*model.BusinessEventMessage{
						{MessageName: "OrderCreated"},
					},
				},
			},
		},
	}
	h := mkHierarchy(mod)
	withContainer(h, svc.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListBusinessEventServicesFunc: func() ([]*model.BusinessEventService, error) {
			return []*model.BusinessEventService{svc}, nil
		},
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, describeBusinessEventService(ctx, ast.QualifiedName{Module: "MyModule", Name: "OrderEvents"}))

	out := buf.String()
	assertContainsStr(t, out, "CREATE OR REPLACE BUSINESS EVENT SERVICE")
	assertContainsStr(t, out, "MyModule.OrderEvents")
}
