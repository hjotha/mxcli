// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/backend/mock"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/domainmodel"
)

func TestShowAssociations_Mock(t *testing.T) {
	mod := mkModule("MyModule")
	ent1 := mkEntity(mod.ID, "Order")
	ent2 := mkEntity(mod.ID, "Customer")
	assoc := mkAssociation(mod.ID, "Order_Customer", ent1.ID, ent2.ID)

	dm := &domainmodel.DomainModel{
		BaseElement:  model.BaseElement{ID: nextID("dm")},
		ContainerID:  mod.ID,
		Entities:     []*domainmodel.Entity{ent1, ent2},
		Associations: []*domainmodel.Association{assoc},
	}

	mb := &mock.MockBackend{
		IsConnectedFunc:      func() bool { return true },
		ListModulesFunc:      func() ([]*model.Module, error) { return []*model.Module{mod}, nil },
		ListDomainModelsFunc: func() ([]*domainmodel.DomainModel, error) { return []*domainmodel.DomainModel{dm}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb))
	assertNoError(t, listAssociations(ctx, ""))

	out := buf.String()
	assertContainsStr(t, out, "MyModule.Order_Customer")
	assertContainsStr(t, out, "MyModule.Order")
	assertContainsStr(t, out, "MyModule.Customer")
	assertContainsStr(t, out, "Reference")
	assertContainsStr(t, out, "(1 associations)")
}

func TestShowAssociations_Mock_FilterByModule(t *testing.T) {
	mod1 := mkModule("Sales")
	mod2 := mkModule("HR")
	ent1 := mkEntity(mod1.ID, "Order")
	ent2 := mkEntity(mod1.ID, "Product")
	ent3 := mkEntity(mod2.ID, "Employee")
	ent4 := mkEntity(mod2.ID, "Department")

	dm1 := &domainmodel.DomainModel{
		BaseElement:  model.BaseElement{ID: nextID("dm")},
		ContainerID:  mod1.ID,
		Entities:     []*domainmodel.Entity{ent1, ent2},
		Associations: []*domainmodel.Association{mkAssociation(mod1.ID, "Order_Product", ent1.ID, ent2.ID)},
	}
	dm2 := &domainmodel.DomainModel{
		BaseElement:  model.BaseElement{ID: nextID("dm")},
		ContainerID:  mod2.ID,
		Entities:     []*domainmodel.Entity{ent3, ent4},
		Associations: []*domainmodel.Association{mkAssociation(mod2.ID, "Employee_Dept", ent3.ID, ent4.ID)},
	}

	mb := &mock.MockBackend{
		IsConnectedFunc:      func() bool { return true },
		ListModulesFunc:      func() ([]*model.Module, error) { return []*model.Module{mod1, mod2}, nil },
		ListDomainModelsFunc: func() ([]*domainmodel.DomainModel, error) { return []*domainmodel.DomainModel{dm1, dm2}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb))
	assertNoError(t, listAssociations(ctx, "HR"))

	out := buf.String()
	assertNotContainsStr(t, out, "Sales.Order_Product")
	assertContainsStr(t, out, "HR.Employee_Dept")
	assertContainsStr(t, out, "(1 associations)")
}
