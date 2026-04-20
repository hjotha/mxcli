// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/backend/mock"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/mpr"
)

func TestShowJsonStructures_Mock(t *testing.T) {
	mod := mkModule("API")
	js := &mpr.JsonStructure{
		BaseElement: model.BaseElement{ID: nextID("js")},
		ContainerID: mod.ID,
		Name:        "OrderSchema",
	}

	h := mkHierarchy(mod)
	withContainer(h, js.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc:        func() bool { return true },
		ListJsonStructuresFunc: func() ([]*mpr.JsonStructure, error) { return []*mpr.JsonStructure{js}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, showJsonStructures(ctx, ""))

	out := buf.String()
	assertContainsStr(t, out, "JSON Structure")
	assertContainsStr(t, out, "API.OrderSchema")
}
