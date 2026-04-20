// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/backend/mock"
	"github.com/mendixlabs/mxcli/model"
)

func TestShowImportMappings_Mock(t *testing.T) {
	mod := mkModule("Integration")
	im := &model.ImportMapping{
		BaseElement: model.BaseElement{ID: nextID("im")},
		ContainerID: mod.ID,
		Name:        "ImportOrders",
	}

	h := mkHierarchy(mod)
	withContainer(h, im.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc:        func() bool { return true },
		ListImportMappingsFunc: func() ([]*model.ImportMapping, error) { return []*model.ImportMapping{im}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, showImportMappings(ctx, ""))

	out := buf.String()
	assertContainsStr(t, out, "Import Mapping")
	assertContainsStr(t, out, "Integration.ImportOrders")
}
