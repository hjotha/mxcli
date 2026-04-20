// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/backend/mock"
	"github.com/mendixlabs/mxcli/model"
)

func TestShowSettings_Mock(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		GetProjectSettingsFunc: func() (*model.ProjectSettings, error) {
			return &model.ProjectSettings{
				Model: &model.ModelSettings{
					HashAlgorithm: "BCrypt",
					JavaVersion:   "17",
				},
			}, nil
		},
	}
	ctx, buf := newMockCtx(t, withBackend(mb))
	assertNoError(t, showSettings(ctx))

	out := buf.String()
	assertContainsStr(t, out, "Section")
	assertContainsStr(t, out, "Key Values")
}

func TestDescribeSettings_Mock(t *testing.T) {
	mb := &mock.MockBackend{
		IsConnectedFunc: func() bool { return true },
		GetProjectSettingsFunc: func() (*model.ProjectSettings, error) {
			return &model.ProjectSettings{
				Model: &model.ModelSettings{
					HashAlgorithm: "BCrypt",
					JavaVersion:   "17",
					RoundingMode:  "HalfUp",
				},
			}, nil
		},
	}
	ctx, buf := newMockCtx(t, withBackend(mb))
	assertNoError(t, describeSettings(ctx))
	assertContainsStr(t, buf.String(), "ALTER SETTINGS")
}
