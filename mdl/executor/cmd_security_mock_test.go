// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/backend/mock"
	"github.com/mendixlabs/mxcli/sdk/microflows"
	"github.com/mendixlabs/mxcli/sdk/pages"
	"github.com/mendixlabs/mxcli/sdk/security"
)

func TestShowProjectSecurity_Mock(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		GetProjectSecurityFunc: func() (*security.ProjectSecurity, error) {
			return &security.ProjectSecurity{
				SecurityLevel:   "CheckEverything",
				EnableDemoUsers: true,
				AdminUserName:   "MxAdmin",
				UserRoles:       []*security.UserRole{{Name: "Admin"}, {Name: "User"}},
				DemoUsers:       []*security.DemoUser{{UserName: "demo_admin"}},
				PasswordPolicy:  &security.PasswordPolicy{MinimumLength: 8},
			}, nil
		},
	}
	ctx, buf := newMockCtx(t, withBackend(mb))
	assertNoError(t, listProjectSecurity(ctx))

	out := buf.String()
	assertContainsStr(t, out, "Security Level:")
	assertContainsStr(t, out, "MxAdmin")
	assertContainsStr(t, out, "Demo Users Enabled:")
}

func TestShowModuleRoles_Mock(t *testing.T) {
	mod := mkModule("MyModule")
	h := mkHierarchy(mod)

	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListModuleSecurityFunc: func() ([]*security.ModuleSecurity, error) {
			return []*security.ModuleSecurity{{
				ContainerID: mod.ID,
				ModuleRoles: []*security.ModuleRole{
					{Name: "Admin"},
					{Name: "User"},
				},
			}}, nil
		},
	}
	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, listModuleRoles(ctx, ""))

	out := buf.String()
	assertContainsStr(t, out, "Qualified Name")
	assertContainsStr(t, out, "Role")
	assertContainsStr(t, out, "Admin")
	assertContainsStr(t, out, "User")
}

func TestShowUserRoles_Mock(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		GetProjectSecurityFunc: func() (*security.ProjectSecurity, error) {
			return &security.ProjectSecurity{
				UserRoles: []*security.UserRole{
					{Name: "Administrator", ModuleRoles: []string{"MyModule.Admin"}},
					{Name: "NormalUser", ModuleRoles: []string{"MyModule.User"}},
				},
			}, nil
		},
	}
	ctx, buf := newMockCtx(t, withBackend(mb))
	assertNoError(t, listUserRoles(ctx))

	out := buf.String()
	assertContainsStr(t, out, "Name")
	assertContainsStr(t, out, "Module Roles")
	assertContainsStr(t, out, "Administrator")
	assertContainsStr(t, out, "NormalUser")
}

func TestShowDemoUsers_Mock(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		GetProjectSecurityFunc: func() (*security.ProjectSecurity, error) {
			return &security.ProjectSecurity{
				EnableDemoUsers: true,
				DemoUsers: []*security.DemoUser{
					{UserName: "demo_admin", UserRoles: []string{"Administrator"}},
				},
			}, nil
		},
	}
	ctx, buf := newMockCtx(t, withBackend(mb))
	assertNoError(t, listDemoUsers(ctx))

	out := buf.String()
	assertContainsStr(t, out, "User Name")
	assertContainsStr(t, out, "demo_admin")
}

func TestShowDemoUsers_Disabled_Mock(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		GetProjectSecurityFunc: func() (*security.ProjectSecurity, error) {
			return &security.ProjectSecurity{
				EnableDemoUsers: false,
			}, nil
		},
	}
	ctx, buf := newMockCtx(t, withBackend(mb))
	assertNoError(t, listDemoUsers(ctx))
	assertContainsStr(t, buf.String(), "Demo users are disabled.")
}

func TestDescribeModuleRole_Mock(t *testing.T) {
	mod := mkModule("MyModule")
	h := mkHierarchy(mod)

	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListModuleSecurityFunc: func() ([]*security.ModuleSecurity, error) {
			return []*security.ModuleSecurity{{
				ContainerID: mod.ID,
				ModuleRoles: []*security.ModuleRole{{Name: "Admin", Description: "Full access"}},
			}}, nil
		},
		GetProjectSecurityFunc: func() (*security.ProjectSecurity, error) {
			return &security.ProjectSecurity{
				UserRoles: []*security.UserRole{
					{Name: "Administrator", ModuleRoles: []string{"MyModule.Admin"}},
				},
			}, nil
		},
	}
	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, describeModuleRole(ctx, ast.QualifiedName{Module: "MyModule", Name: "Admin"}))
	assertContainsStr(t, buf.String(), "create module role")
}

func TestDescribeUserRole_Mock(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		GetProjectSecurityFunc: func() (*security.ProjectSecurity, error) {
			return &security.ProjectSecurity{
				UserRoles: []*security.UserRole{
					{Name: "Administrator", ModuleRoles: []string{"MyModule.Admin"}},
				},
			}, nil
		},
	}
	ctx, buf := newMockCtx(t, withBackend(mb))
	assertNoError(t, describeUserRole(ctx, ast.QualifiedName{Name: "Administrator"}))
	assertContainsStr(t, buf.String(), "create user role")
}

func TestDescribeDemoUser_Mock(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		GetProjectSecurityFunc: func() (*security.ProjectSecurity, error) {
			return &security.ProjectSecurity{
				EnableDemoUsers: true,
				DemoUsers: []*security.DemoUser{
					{UserName: "demo_admin", UserRoles: []string{"Administrator"}},
				},
			}, nil
		},
	}
	ctx, buf := newMockCtx(t, withBackend(mb))
	assertNoError(t, describeDemoUser(ctx, "demo_admin"))
	assertContainsStr(t, buf.String(), "create demo user")
}

func TestShowModuleRoles_Mock_FilterByModule(t *testing.T) {
	mod1 := mkModule("Sales")
	mod2 := mkModule("HR")
	h := mkHierarchy(mod1, mod2)

	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListModuleSecurityFunc: func() ([]*security.ModuleSecurity, error) {
			return []*security.ModuleSecurity{
				{ContainerID: mod1.ID, ModuleRoles: []*security.ModuleRole{{Name: "Manager"}}},
				{ContainerID: mod2.ID, ModuleRoles: []*security.ModuleRole{{Name: "Employee"}}},
			}, nil
		},
	}
	ctx, buf := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertNoError(t, listModuleRoles(ctx, "HR"))

	out := buf.String()
	assertNotContainsStr(t, out, "Sales")
	assertContainsStr(t, out, "HR")
	assertContainsStr(t, out, "Employee")
}

func TestDescribeModuleRole_Mock_NotFound(t *testing.T) {
	mod := mkModule("MyModule")
	h := mkHierarchy(mod)

	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListModuleSecurityFunc: func() ([]*security.ModuleSecurity, error) {
			return []*security.ModuleSecurity{{
				ContainerID: mod.ID,
				ModuleRoles: []*security.ModuleRole{{Name: "Admin"}},
			}}, nil
		},
	}
	ctx, _ := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertError(t, describeModuleRole(ctx, ast.QualifiedName{Module: "MyModule", Name: "NonExistent"}))
}

func TestDescribeUserRole_Mock_NotFound(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		GetProjectSecurityFunc: func() (*security.ProjectSecurity, error) {
			return &security.ProjectSecurity{
				UserRoles: []*security.UserRole{{Name: "Admin"}},
			}, nil
		},
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, describeUserRole(ctx, ast.QualifiedName{Name: "NonExistent"}))
}

func TestDescribeDemoUser_Mock_NotFound(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		GetProjectSecurityFunc: func() (*security.ProjectSecurity, error) {
			return &security.ProjectSecurity{
				EnableDemoUsers: true,
				DemoUsers:       []*security.DemoUser{{UserName: "demo_admin"}},
			}, nil
		},
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, describeDemoUser(ctx, "nonexistent"))
}

func TestShowAccessOnEntity_Mock_NilName(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, listAccessOnEntity(ctx, nil))
}

func TestShowAccessOnMicroflow_Mock_NotFound(t *testing.T) {
	mod := mkModule("MyModule")
	h := mkHierarchy(mod)

	mb := &mock.MockBackend{
		IsConnectedFunc:    func() bool { return true },
		ListMicroflowsFunc: func() ([]*microflows.Microflow, error) { return nil, nil },
	}
	ctx, _ := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertError(t, listAccessOnMicroflow(ctx, &ast.QualifiedName{Module: "MyModule", Name: "NonExistent"}))
}

func TestShowAccessOnPage_Mock_NotFound(t *testing.T) {
	mod := mkModule("MyModule")
	h := mkHierarchy(mod)

	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ListPagesFunc:   func() ([]*pages.Page, error) { return nil, nil },
	}
	ctx, _ := newMockCtx(t, withBackend(mb), withHierarchy(h))
	assertError(t, listAccessOnPage(ctx, &ast.QualifiedName{Module: "MyModule", Name: "NonExistent"}))
}

func TestShowAccessOnWorkflow_Mock_Unsupported(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, listAccessOnWorkflow(ctx, &ast.QualifiedName{Module: "MyModule", Name: "SomeWorkflow"}))
}
