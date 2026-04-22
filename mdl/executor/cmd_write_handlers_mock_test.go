// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/backend/mock"
	"github.com/mendixlabs/mxcli/mdl/types"
	"github.com/mendixlabs/mxcli/model"
	"github.com/mendixlabs/mxcli/sdk/domainmodel"
	"github.com/mendixlabs/mxcli/sdk/microflows"
	"github.com/mendixlabs/mxcli/sdk/pages"
)

func TestExecCreateModule_Mock(t *testing.T) {
	called := false
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListModulesFunc: func() ([]*model.Module, error) {
			return nil, nil // no existing modules
		},
		CreateModuleFunc: func(m *model.Module) error {
			called = true
			return nil
		},
	}

	ctx, buf := newMockCtx(t, withBackend(mb))
	err := execCreateModule(ctx, &ast.CreateModuleStmt{Name: "NewModule"})
	assertNoError(t, err)
	assertContainsStr(t, buf.String(), "Created module: NewModule")
	if !called {
		t.Fatal("CreateModuleFunc was not called")
	}
}

func TestExecDropEnumeration_Mock(t *testing.T) {
	mod := mkModule("MyModule")
	enum := mkEnumeration(mod.ID, "Status", "Active", "Inactive")

	called := false
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListEnumerationsFunc: func() ([]*model.Enumeration, error) {
			return []*model.Enumeration{enum}, nil
		},
		ListModulesFunc: func() ([]*model.Module, error) {
			return []*model.Module{mod}, nil
		},
		DeleteEnumerationFunc: func(id model.ID) error {
			called = true
			return nil
		},
	}

	ctx, buf := newMockCtx(t, withBackend(mb))
	err := execDropEnumeration(ctx, &ast.DropEnumerationStmt{
		Name: ast.QualifiedName{Module: "MyModule", Name: "Status"},
	})
	assertNoError(t, err)
	assertContainsStr(t, buf.String(), "Dropped enumeration:")
	if !called {
		t.Fatal("DeleteEnumerationFunc was not called")
	}
}

func TestExecCreateEnumeration_Mock(t *testing.T) {
	mod := mkModule("MyModule")

	h := mkHierarchy(mod)

	called := false
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListModulesFunc: func() ([]*model.Module, error) {
			return []*model.Module{mod}, nil
		},
		ListEnumerationsFunc: func() ([]*model.Enumeration, error) {
			return nil, nil // no duplicates
		},
		CreateEnumerationFunc: func(e *model.Enumeration) error {
			called = true
			return nil
		},
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	err := execCreateEnumeration(ctx, &ast.CreateEnumerationStmt{
		Name:   ast.QualifiedName{Module: "MyModule", Name: "Color"},
		Values: []ast.EnumValue{{Name: "Red", Caption: "Red"}, {Name: "Blue", Caption: "Blue"}},
	})
	assertNoError(t, err)
	assertContainsStr(t, buf.String(), "Created enumeration:")
	if !called {
		t.Fatal("CreateEnumerationFunc was not called")
	}
}

func TestExecDropEntity_Mock(t *testing.T) {
	mod := mkModule("MyModule")
	ent := mkEntity(mod.ID, "Customer")
	dm := mkDomainModel(mod.ID, ent)

	called := false
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListModulesFunc: func() ([]*model.Module, error) {
			return []*model.Module{mod}, nil
		},
		GetDomainModelFunc: func(moduleID model.ID) (*domainmodel.DomainModel, error) {
			return dm, nil
		},
		DeleteEntityFunc: func(domainModelID model.ID, entityID model.ID) error {
			called = true
			return nil
		},
	}

	ctx, buf := newMockCtx(t, withBackend(mb))
	err := execDropEntity(ctx, &ast.DropEntityStmt{
		Name: ast.QualifiedName{Module: "MyModule", Name: "Customer"},
	})
	assertNoError(t, err)
	assertContainsStr(t, buf.String(), "Dropped entity:")
	if !called {
		t.Fatal("DeleteEntityFunc was not called")
	}
}

func TestExecDropMicroflow_Mock(t *testing.T) {
	mod := mkModule("MyModule")
	mf := mkMicroflow(mod.ID, "DoSomething")

	h := mkHierarchy(mod)
	withContainer(h, mf.ContainerID, mod.ID)

	called := false
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListMicroflowsFunc: func() ([]*microflows.Microflow, error) {
			return []*microflows.Microflow{mf}, nil
		},
		DeleteMicroflowFunc: func(id model.ID) error {
			called = true
			return nil
		},
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	err := execDropMicroflow(ctx, &ast.DropMicroflowStmt{
		Name: ast.QualifiedName{Module: "MyModule", Name: "DoSomething"},
	})
	assertNoError(t, err)
	assertContainsStr(t, buf.String(), "Dropped microflow:")
	if !called {
		t.Fatal("DeleteMicroflowFunc was not called")
	}
}

func TestExecDropPage_Mock(t *testing.T) {
	mod := mkModule("MyModule")
	pg := mkPage(mod.ID, "HomePage")

	h := mkHierarchy(mod)
	withContainer(h, pg.ContainerID, mod.ID)

	called := false
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListPagesFunc: func() ([]*pages.Page, error) {
			return []*pages.Page{pg}, nil
		},
		DeletePageFunc: func(id model.ID) error {
			called = true
			return nil
		},
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	err := execDropPage(ctx, &ast.DropPageStmt{
		Name: ast.QualifiedName{Module: "MyModule", Name: "HomePage"},
	})
	assertNoError(t, err)
	assertContainsStr(t, buf.String(), "Dropped page")
	if !called {
		t.Fatal("DeletePageFunc was not called")
	}
}

func TestExecDropSnippet_Mock(t *testing.T) {
	mod := mkModule("MyModule")
	snp := mkSnippet(mod.ID, "HeaderSnippet")

	h := mkHierarchy(mod)
	withContainer(h, snp.ContainerID, mod.ID)

	called := false
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListSnippetsFunc: func() ([]*pages.Snippet, error) {
			return []*pages.Snippet{snp}, nil
		},
		DeleteSnippetFunc: func(id model.ID) error {
			called = true
			return nil
		},
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	err := execDropSnippet(ctx, &ast.DropSnippetStmt{
		Name: ast.QualifiedName{Module: "MyModule", Name: "HeaderSnippet"},
	})
	assertNoError(t, err)
	assertContainsStr(t, buf.String(), "Dropped snippet")
	if !called {
		t.Fatal("DeleteSnippetFunc was not called")
	}
}

func TestExecDropAssociation_Mock(t *testing.T) {
	mod := mkModule("MyModule")
	ent1 := mkEntity(mod.ID, "Order")
	ent2 := mkEntity(mod.ID, "Customer")
	assoc := mkAssociation(mod.ID, "Order_Customer", ent1.ID, ent2.ID)

	dm := mkDomainModel(mod.ID, ent1, ent2)
	dm.Associations = []*domainmodel.Association{assoc}

	called := false
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListModulesFunc: func() ([]*model.Module, error) {
			return []*model.Module{mod}, nil
		},
		GetDomainModelFunc: func(moduleID model.ID) (*domainmodel.DomainModel, error) {
			return dm, nil
		},
		DeleteAssociationFunc: func(domainModelID model.ID, assocID model.ID) error {
			called = true
			return nil
		},
	}

	ctx, buf := newMockCtx(t, withBackend(mb))
	err := execDropAssociation(ctx, &ast.DropAssociationStmt{
		Name: ast.QualifiedName{Module: "MyModule", Name: "Order_Customer"},
	})
	assertNoError(t, err)
	assertContainsStr(t, buf.String(), "Dropped association:")
	if !called {
		t.Fatal("DeleteAssociationFunc was not called")
	}
}

func TestExecDropJavaAction_Mock(t *testing.T) {
	mod := mkModule("MyModule")
	jaID := nextID("ja")
	ja := &types.JavaAction{
		BaseElement: model.BaseElement{ID: jaID},
		ContainerID: mod.ID,
		Name:        "MyAction",
	}

	h := mkHierarchy(mod)

	called := false
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListJavaActionsFunc: func() ([]*types.JavaAction, error) {
			return []*types.JavaAction{ja}, nil
		},
		DeleteJavaActionFunc: func(id model.ID) error {
			called = true
			return nil
		},
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	err := execDropJavaAction(ctx, &ast.DropJavaActionStmt{
		Name: ast.QualifiedName{Module: "MyModule", Name: "MyAction"},
	})
	assertNoError(t, err)
	assertContainsStr(t, buf.String(), "Dropped java action:")
	if !called {
		t.Fatal("DeleteJavaActionFunc was not called")
	}
}

func TestExecDropFolder_Mock(t *testing.T) {
	mod := mkModule("MyModule")
	folderID := nextID("folder")
	folder := &types.FolderInfo{
		ID:          folderID,
		ContainerID: mod.ID,
		Name:        "Resources",
	}

	h := mkHierarchy(mod)
	withContainer(h, folderID, mod.ID)

	called := false
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListModulesFunc: func() ([]*model.Module, error) {
			return []*model.Module{mod}, nil
		},
		ListFoldersFunc: func() ([]*types.FolderInfo, error) {
			return []*types.FolderInfo{folder}, nil
		},
		DeleteFolderFunc: func(id model.ID) error {
			called = true
			return nil
		},
	}

	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	err := execDropFolder(ctx, &ast.DropFolderStmt{
		FolderPath: "Resources",
		Module:     "MyModule",
	})
	assertNoError(t, err)
	assertContainsStr(t, buf.String(), "Dropped folder:")
	if !called {
		t.Fatal("DeleteFolderFunc was not called")
	}
}

func TestExecCreateModule_Mock_AlreadyExists(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListModulesFunc: func() ([]*model.Module, error) {
			return []*model.Module{{Name: "Existing"}}, nil
		},
	}
	ctx, buf := newMockCtx(t, withBackend(mb))
	assertNoError(t, execCreateModule(ctx, &ast.CreateModuleStmt{Name: "Existing"}))
	assertContainsStr(t, buf.String(), "already exists")
}

func TestExecDropEnumeration_Mock_NotFound(t *testing.T) {
	mod := mkModule("MyModule")
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListEnumerationsFunc: func() ([]*model.Enumeration, error) {
			return nil, nil
		},
		ListModulesFunc: func() ([]*model.Module, error) {
			return []*model.Module{mod}, nil
		},
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, execDropEnumeration(ctx, &ast.DropEnumerationStmt{
		Name: ast.QualifiedName{Module: "MyModule", Name: "NonExistent"},
	}))
}

func TestExecDropEntity_Mock_NotFound(t *testing.T) {
	mod := mkModule("MyModule")
	dm := mkDomainModel(mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListModulesFunc: func() ([]*model.Module, error) {
			return []*model.Module{mod}, nil
		},
		GetDomainModelFunc: func(moduleID model.ID) (*domainmodel.DomainModel, error) {
			return dm, nil
		},
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, execDropEntity(ctx, &ast.DropEntityStmt{
		Name: ast.QualifiedName{Module: "MyModule", Name: "NonExistent"},
	}))
}

func TestExecDropMicroflow_Mock_NotFound(t *testing.T) {
	mod := mkModule("MyModule")
	h := mkHierarchy(mod)

	mb := &mock.MockBackend{
		IsConnectedFunc:    func() bool { return true },
		ListMicroflowsFunc: func() ([]*microflows.Microflow, error) { return nil, nil },
	}
	ctx, _ := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertError(t, execDropMicroflow(ctx, &ast.DropMicroflowStmt{
		Name: ast.QualifiedName{Module: "MyModule", Name: "NonExistent"},
	}))
}

func TestExecDropPage_Mock_NotFound(t *testing.T) {
	mod := mkModule("MyModule")
	h := mkHierarchy(mod)

	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListPagesFunc:   func() ([]*pages.Page, error) { return nil, nil },
	}
	ctx, _ := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertError(t, execDropPage(ctx, &ast.DropPageStmt{
		Name: ast.QualifiedName{Module: "MyModule", Name: "NonExistent"},
	}))
}

func TestExecDropSnippet_Mock_NotFound(t *testing.T) {
	mod := mkModule("MyModule")
	h := mkHierarchy(mod)

	mb := &mock.MockBackend{
		IsConnectedFunc:  func() bool { return true },
		ListSnippetsFunc: func() ([]*pages.Snippet, error) { return nil, nil },
	}
	ctx, _ := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertError(t, execDropSnippet(ctx, &ast.DropSnippetStmt{
		Name: ast.QualifiedName{Module: "MyModule", Name: "NonExistent"},
	}))
}

func TestExecDropAssociation_Mock_NotFound(t *testing.T) {
	mod := mkModule("MyModule")
	dm := mkDomainModel(mod.ID)

	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListModulesFunc: func() ([]*model.Module, error) {
			return []*model.Module{mod}, nil
		},
		GetDomainModelFunc: func(moduleID model.ID) (*domainmodel.DomainModel, error) {
			return dm, nil
		},
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, execDropAssociation(ctx, &ast.DropAssociationStmt{
		Name: ast.QualifiedName{Module: "MyModule", Name: "NonExistent"},
	}))
}
