// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/backend/mock"
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
	assertNoError(t, showProjectSecurity(ctx))

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
	assertNoError(t, showModuleRoles(ctx, ""))

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
	assertNoError(t, showUserRoles(ctx))

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
	assertNoError(t, showDemoUsers(ctx))

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
	assertNoError(t, showDemoUsers(ctx))
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
	assertContainsStr(t, buf.String(), "CREATE MODULE ROLE")
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
	assertContainsStr(t, buf.String(), "CREATE USER ROLE")
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
	assertContainsStr(t, buf.String(), "CREATE DEMO USER")
}
