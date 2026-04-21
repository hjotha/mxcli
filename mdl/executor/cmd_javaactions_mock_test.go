// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/backend/mock"
	"github.com/mendixlabs/mxcli/mdl/types"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/javaactions"
)

func TestShowJavaActions_Mock(t *testing.T) {
	mod := mkModule("MyModule")
	ja := &types.JavaAction{
		BaseElement: model.BaseElement{ID: nextID("ja")},
		ContainerID: mod.ID,
		Name:        "DoSomething",
	}

	h := mkHierarchy(mod)
	withContainer(h, ja.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc:     func() bool { return true },
		ListJavaActionsFunc: func() ([]*types.JavaAction, error) { return []*types.JavaAction{ja}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, listJavaActions(ctx, ""))

	out := buf.String()
	assertContainsStr(t, out, "Qualified Name")
	assertContainsStr(t, out, "MyModule.DoSomething")
}

func TestDescribeJavaAction_Mock(t *testing.T) {
	mod := mkModule("MyModule")

	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ReadJavaActionByNameFunc: func(qn string) (*javaactions.JavaAction, error) {
			return &javaactions.JavaAction{
				BaseElement: model.BaseElement{ID: nextID("ja")},
				ContainerID: mod.ID,
				Name:        "DoSomething",
			}, nil
		},
	}

	ctx, buf := newMockCtx(t, withBackend(mb))
	assertNoError(t, describeJavaAction(ctx, ast.QualifiedName{Module: "MyModule", Name: "DoSomething"}))

	out := buf.String()
	assertContainsStr(t, out, "create java action")
}
