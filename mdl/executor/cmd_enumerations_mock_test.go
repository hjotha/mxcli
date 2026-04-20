// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/backend/mock"
	"github.com/mendixlabs/mxcli/model"
)

func TestShowEnumerations_Mock(t *testing.T) {
	mod := mkModule("MyModule")
	enum := mkEnumeration(mod.ID, "Color", "Red", "Green", "Blue")

	h := mkHierarchy(mod)
	withContainer(h, enum.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc:      func() bool { return true },
		ListEnumerationsFunc: func() ([]*model.Enumeration, error) { return []*model.Enumeration{enum}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, showEnumerations(ctx, ""))

	out := buf.String()
	assertContainsStr(t, out, "MyModule.Color")
	assertContainsStr(t, out, "| 3")
	assertContainsStr(t, out, "(1 enumerations)")
}

func TestShowEnumerations_Mock_FilterByModule(t *testing.T) {
	mod1 := mkModule("Alpha")
	mod2 := mkModule("Beta")
	e1 := mkEnumeration(mod1.ID, "Color", "Red")
	e2 := mkEnumeration(mod2.ID, "Size", "S", "M")

	h := mkHierarchy(mod1, mod2)
	withContainer(h, e1.ContainerID, mod1.ID)
	withContainer(h, e2.ContainerID, mod2.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc:      func() bool { return true },
		ListEnumerationsFunc: func() ([]*model.Enumeration, error) { return []*model.Enumeration{e1, e2}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, showEnumerations(ctx, "Beta"))

	out := buf.String()
	assertNotContainsStr(t, out, "Alpha.Color")
	assertContainsStr(t, out, "Beta.Size")
	assertContainsStr(t, out, "(1 enumerations)")
}

func TestDescribeEnumeration_Mock(t *testing.T) {
	mod := mkModule("MyModule")
	enum := &model.Enumeration{
		BaseElement: model.BaseElement{ID: nextID("enum")},
		ContainerID: mod.ID,
		Name:        "Status",
		Values: []model.EnumerationValue{
			{BaseElement: model.BaseElement{ID: nextID("ev")}, Name: "Active", Caption: &model.Text{Translations: map[string]string{"en_US": "Active"}}},
			{BaseElement: model.BaseElement{ID: nextID("ev")}, Name: "Inactive", Caption: &model.Text{Translations: map[string]string{"en_US": "Inactive"}}},
		},
	}

	h := mkHierarchy(mod)
	withContainer(h, enum.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc:      func() bool { return true },
		ListEnumerationsFunc: func() ([]*model.Enumeration, error) { return []*model.Enumeration{enum}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, describeEnumeration(ctx, ast.QualifiedName{Module: "MyModule", Name: "Status"}))

	out := buf.String()
	assertContainsStr(t, out, "CREATE OR MODIFY ENUMERATION MyModule.Status")
	assertContainsStr(t, out, "Active")
	assertContainsStr(t, out, "Inactive")
}

func TestDescribeEnumeration_Mock_NotFound(t *testing.T) {
	mod := mkModule("MyModule")
	h := mkHierarchy(mod)

	mb := &mock.MockBackend{
		IsConnectedFunc:      func() bool { return true },
		ListEnumerationsFunc: func() ([]*model.Enumeration, error) { return nil, nil },
	}

	ctx, _ := newMockCtx(t, withBackend(mb), withHierarchy(h))
	err := describeEnumeration(ctx, ast.QualifiedName{Module: "MyModule", Name: "Missing"})
	assertError(t, err)
}
