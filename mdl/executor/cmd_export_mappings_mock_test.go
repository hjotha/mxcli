// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/backend/mock"
	"github.com/mendixlabs/mxcli/model"
)

func TestShowExportMappings_Mock(t *testing.T) {
	mod := mkModule("Integration")
	em := &model.ExportMapping{
		BaseElement: model.BaseElement{ID: nextID("em")},
		ContainerID: mod.ID,
		Name:        "ExportOrders",
	}

	h := mkHierarchy(mod)
	withContainer(h, em.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc:        func() bool { return true },
		ListExportMappingsFunc: func() ([]*model.ExportMapping, error) { return []*model.ExportMapping{em}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, listExportMappings(ctx, ""))

	out := buf.String()
	assertContainsStr(t, out, "Export Mapping")
	assertContainsStr(t, out, "Integration.ExportOrders")
}
