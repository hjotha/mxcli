// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"fmt"
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
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

func TestShowExportMappings_FilterByModule(t *testing.T) {
	mod1 := mkModule("Integration")
	mod2 := mkModule("Other")
	em1 := &model.ExportMapping{
		BaseElement: model.BaseElement{ID: nextID("em")},
		ContainerID: mod1.ID,
		Name:        "ExportOrders",
	}
	em2 := &model.ExportMapping{
		BaseElement: model.BaseElement{ID: nextID("em")},
		ContainerID: mod2.ID,
		Name:        "ExportOther",
	}

	h := mkHierarchy(mod1, mod2)
	withContainer(h, em1.ContainerID, mod1.ID)
	withContainer(h, em2.ContainerID, mod2.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc:        func() bool { return true },
		ListExportMappingsFunc: func() ([]*model.ExportMapping, error) { return []*model.ExportMapping{em1, em2}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, listExportMappings(ctx, "Integration"))

	out := buf.String()
	assertContainsStr(t, out, "Integration.ExportOrders")
	assertNotContainsStr(t, out, "Other.ExportOther")
}

func TestDescribeExportMapping_Mock(t *testing.T) {
	mod := mkModule("Integration")
	em := &model.ExportMapping{
		BaseElement: model.BaseElement{ID: nextID("em")},
		ContainerID: mod.ID,
		Name:        "ExportOrders",
	}

	h := mkHierarchy(mod)
	withContainer(h, em.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		GetExportMappingByQualifiedNameFunc: func(moduleName, name string) (*model.ExportMapping, error) {
			return em, nil
		},
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, describeExportMapping(ctx, ast.QualifiedName{Module: "Integration", Name: "ExportOrders"}))
	assertContainsStr(t, buf.String(), "create export mapping")
}

func TestDescribeExportMapping_NotFound(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		GetExportMappingByQualifiedNameFunc: func(moduleName, name string) (*model.ExportMapping, error) {
			return nil, fmt.Errorf("export mapping not found: %s.%s", moduleName, name)
		},
	}

	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, describeExportMapping(ctx, ast.QualifiedName{Module: "Integration", Name: "NoSuch"}))
}
