// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"testing"
	"time"
)

func TestConfiguredExecuteTimeoutUsesDurationEnv(t *testing.T) {
	t.Setenv("MXCLI_EXEC_TIMEOUT", "12m")

	if got := configuredExecuteTimeout(); got != 12*time.Minute {
		t.Fatalf("configured timeout = %v, want 12m", got)
	}
}

func TestConfiguredExecuteTimeoutUsesSecondEnv(t *testing.T) {
	t.Setenv("MXCLI_EXEC_TIMEOUT", "900")

	if got := configuredExecuteTimeout(); got != 15*time.Minute {
		t.Fatalf("configured timeout = %v, want 15m", got)
	}
}

func TestConfiguredExecuteTimeoutFallsBackForInvalidEnv(t *testing.T) {
	t.Setenv("MXCLI_EXEC_TIMEOUT", "invalid")

	if got := configuredExecuteTimeout(); got != defaultExecuteTimeout {
		t.Fatalf("configured timeout = %v, want default %v", got, defaultExecuteTimeout)
	}
}
