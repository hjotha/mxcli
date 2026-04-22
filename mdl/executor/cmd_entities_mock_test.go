// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"fmt"
	"testing"

	"github.com/mendixlabs/mxcli/mdl/backend/mock"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/domainmodel"
)

func TestShowEntities_Mock(t *testing.T) {
	mod := mkModule("MyModule")
	ent1 := mkEntity(mod.ID, "Customer")
	ent2 := mkEntity(mod.ID, "Order")

	dm := mkDomainModel(mod.ID, ent1, ent2)

	mb := &mock.MockBackend{
		IsConnectedFunc:      func() bool { return true },
		ListModulesFunc:      func() ([]*model.Module, error) { return []*model.Module{mod}, nil },
		ListDomainModelsFunc: func() ([]*domainmodel.DomainModel, error) { return []*domainmodel.DomainModel{dm}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb))
	assertNoError(t, listEntities(ctx, ""))

	out := buf.String()
	assertContainsStr(t, out, "MyModule.Customer")
	assertContainsStr(t, out, "MyModule.Order")
	assertContainsStr(t, out, "Persistent")
	assertContainsStr(t, out, "(2 entities)")
}

func TestShowEntities_Mock_FilterByModule(t *testing.T) {
	mod1 := mkModule("Sales")
	mod2 := mkModule("HR")
	ent1 := mkEntity(mod1.ID, "Product")
	ent2 := mkEntity(mod2.ID, "Employee")

	dm1 := mkDomainModel(mod1.ID, ent1)
	dm2 := mkDomainModel(mod2.ID, ent2)

	mb := &mock.MockBackend{
		IsConnectedFunc:      func() bool { return true },
		ListModulesFunc:      func() ([]*model.Module, error) { return []*model.Module{mod1, mod2}, nil },
		ListDomainModelsFunc: func() ([]*domainmodel.DomainModel, error) { return []*domainmodel.DomainModel{dm1, dm2}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb))
	assertNoError(t, listEntities(ctx, "HR"))

	out := buf.String()
	assertNotContainsStr(t, out, "Sales.Product")
	assertContainsStr(t, out, "HR.Employee")
	assertContainsStr(t, out, "(1 entities)")
}

// NOTE: listEntities has no Connected() guard — it calls backend directly.

func TestShowEntities_BackendError_Modules(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListModulesFunc: func() ([]*model.Module, error) { return nil, fmt.Errorf("not connected") },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, listEntities(ctx, ""))
}

func TestShowEntities_BackendError(t *testing.T) {
	mod := mkModule("Sales")
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListModulesFunc: func() ([]*model.Module, error) { return []*model.Module{mod}, nil },
		ListDomainModelsFunc: func() ([]*domainmodel.DomainModel, error) {
			return nil, fmt.Errorf("backend down")
		},
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, listEntities(ctx, ""))
}

func TestShowEntities_JSON(t *testing.T) {
	mod := mkModule("App")
	ent := mkEntity(mod.ID, "Item")
	dm := mkDomainModel(mod.ID, ent)

	mb := &mock.MockBackend{
		IsConnectedFunc:      func() bool { return true },
		ListModulesFunc:      func() ([]*model.Module, error) { return []*model.Module{mod}, nil },
		ListDomainModelsFunc: func() ([]*domainmodel.DomainModel, error) { return []*domainmodel.DomainModel{dm}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withFormat(FormatJSON))
	assertNoError(t, listEntities(ctx, ""))
	assertValidJSON(t, buf.String())
	assertContainsStr(t, buf.String(), "Item")
}
