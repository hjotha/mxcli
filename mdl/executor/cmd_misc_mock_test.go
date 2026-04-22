// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/backend/mock"
	"github.com/mendixlabs/mxcli/mdl/types"
)

func TestShowVersion_Mock(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ProjectVersionFunc: func() *types.ProjectVersion {
			return &types.ProjectVersion{
				ProductVersion: "10.18.0",
				BuildVersion:   "10.18.0.12345",
				FormatVersion:  2,
				SchemaHash:     "abc123def456",
			}
		},
	}

	ctx, buf := newMockCtx(t, withBackend(mb))
	assertNoError(t, listVersion(ctx))

	out := buf.String()
	assertContainsStr(t, out, "Mendix Version")
	assertContainsStr(t, out, "10.18.0")
	assertContainsStr(t, out, "Build Version")
	assertContainsStr(t, out, "MPR Format")
	assertContainsStr(t, out, "Schema Hash")
	assertContainsStr(t, out, "abc123def456")
}

func TestShowVersion_NotConnected(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return false },
	}
	ctx, _ := newMockCtx(t, withBackend(mb))
	assertError(t, listVersion(ctx))
}

func TestShowVersion_NoSchemaHash(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ProjectVersionFunc: func() *types.ProjectVersion {
			return &types.ProjectVersion{
				ProductVersion: "9.24.0",
				BuildVersion:   "9.24.0.5678",
				FormatVersion:  1,
				SchemaHash:     "",
			}
		},
	}
	ctx, buf := newMockCtx(t, withBackend(mb))
	assertNoError(t, listVersion(ctx))

	out := buf.String()
	assertContainsStr(t, out, "9.24.0")
	assertNotContainsStr(t, out, "Schema Hash")
}

func TestHelp_Mock(t *testing.T) {
	ctx, buf := newMockCtx(t)
	assertNoError(t, execHelp(ctx, &ast.HelpStmt{}))

	out := buf.String()
	assertContainsStr(t, out, "MDL Commands")
	assertContainsStr(t, out, "connect local")
}
