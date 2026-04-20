// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/backend/mock"
	"github.com/mendixlabs/mxcli/sdk/mpr/version"
)

func TestShowVersion_Mock(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		ProjectVersionFunc: func() *version.ProjectVersion {
			return &version.ProjectVersion{
				ProductVersion: "10.18.0",
				BuildVersion:   "10.18.0.12345",
				FormatVersion:  2,
				SchemaHash:     "abc123def456",
			}
		},
	}

	ctx, buf := newMockCtx(t, withBackend(mb))
	assertNoError(t, showVersion(ctx))

	out := buf.String()
	assertContainsStr(t, out, "Mendix Version")
	assertContainsStr(t, out, "10.18.0")
	assertContainsStr(t, out, "Build Version")
	assertContainsStr(t, out, "MPR Format")
	assertContainsStr(t, out, "Schema Hash")
	assertContainsStr(t, out, "abc123def456")
}
