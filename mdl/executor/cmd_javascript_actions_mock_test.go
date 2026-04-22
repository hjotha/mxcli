// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"fmt"
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/backend/mock"
	"github.com/mendixlabs/mxcli/mdl/types"
	"github.com/mendixlabs/mxcli/model"
)

func TestShowJavaScriptActions_Mock(t *testing.T) {
	mod := mkModule("WebMod")
	jsa := &types.JavaScriptAction{
		BaseElement: model.BaseElement{ID: nextID("jsa")},
		ContainerID: mod.ID,
		Name:        "ShowAlert",
		Platform:    "Web",
	}

	h := mkHierarchy(mod)
	withContainer(h, jsa.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc:           func() bool { return true },
		ListJavaScriptActionsFunc: func() ([]*types.JavaScriptAction, error) { return []*types.JavaScriptAction{jsa}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, listJavaScriptActions(ctx, ""))

	out := buf.String()
	assertContainsStr(t, out, "Qualified Name")
	assertContainsStr(t, out, "WebMod.ShowAlert")
}

func TestDescribeJavaScriptAction_Mock(t *testing.T) {
	mod := mkModule("WebMod")

	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ReadJavaScriptActionByNameFunc: func(qn string) (*types.JavaScriptAction, error) {
			return &types.JavaScriptAction{
				BaseElement: model.BaseElement{ID: nextID("jsa")},
				ContainerID: mod.ID,
				Name:        "ShowAlert",
				Platform:    "Web",
			}, nil
		},
	}

	ctx, buf := newMockCtx(t, withBackend(mb))
	assertNoError(t, describeJavaScriptAction(ctx, ast.QualifiedName{Module: "WebMod", Name: "ShowAlert"}))

	out := buf.String()
	assertContainsStr(t, out, "create javascript action")
}

func TestDescribeJavaScriptAction_NotFound(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ReadJavaScriptActionByNameFunc: func(qn string) (*types.JavaScriptAction, error) {
			return nil, fmt.Errorf("not found: %s", qn)
		},
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, describeJavaScriptAction(ctx, ast.QualifiedName{Module: "X", Name: "NoSuch"}))
}

func TestShowJavaScriptActions_FilterByModule(t *testing.T) {
	mod := mkModule("WebMod")
	jsa := &types.JavaScriptAction{
		BaseElement: model.BaseElement{ID: nextID("jsa")},
		ContainerID: mod.ID,
		Name:        "ShowAlert",
		Platform:    "Web",
	}

	h := mkHierarchy(mod)
	withContainer(h, jsa.ContainerID, mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc:           func() bool { return true },
		ListJavaScriptActionsFunc: func() ([]*types.JavaScriptAction, error) { return []*types.JavaScriptAction{jsa}, nil },
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, listJavaScriptActions(ctx, "WebMod"))
	assertContainsStr(t, buf.String(), "WebMod.ShowAlert")
}
