// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/backend/mock"
	"github.com/mendixlabs/mxcli/mdl/types"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/domainmodel"
)

func TestShowModules_Mock(t *testing.T) {
	mod1 := mkModule("MyModule")
	mod2 := mkModule("System")

	// listModules uses ListUnits to count documents per module.
	// Provide a unit belonging to mod1 so the count is non-zero.
	unitID := nextID("unit")
	units := []*types.UnitInfo{{ID: unitID, ContainerID: mod1.ID}}

	// Need a hierarchy for getHierarchy — provide modules + units + folders
	h := mkHierarchy(mod1, mod2)
	withContainer(h, unitID, mod1.ID)

	// Provide one domain model for mod1 with one entity
	ent := mkEntity(mod1.ID, "Customer")
	dm := mkDomainModel(mod1.ID, ent)

	mb := &mock.MockBackend{
		IsConnectedFunc:      func() bool { return true },
		ListModulesFunc:      func() ([]*model.Module, error) { return []*model.Module{mod1, mod2}, nil },
		ListUnitsFunc:        func() ([]*types.UnitInfo, error) { return units, nil },
		ListDomainModelsFunc: func() ([]*domainmodel.DomainModel, error) { return []*domainmodel.DomainModel{dm}, nil },
		// All other list functions return nil (zero counts) via MockBackend defaults.
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, listModules(ctx))

	out := buf.String()
	assertContainsStr(t, out, "MyModule")
	assertContainsStr(t, out, "System")
	assertContainsStr(t, out, "(2 modules)")
}
