// SPDX-License-Identifier: Apache-2.0

package mpr

import (
	"testing"

	"github.com/mendixlabs/mxcli/sdk/javaactions"
)

func TestParseCodeActionParameterType_JavaActionMicroflowParameter(t *testing.T) {
	value := parseCodeActionParameterType(map[string]any{
		"$ID":   "type-1",
		"$Type": "JavaActions$MicroflowJavaActionParameterType",
	})

	if _, ok := value.(*javaactions.MicroflowType); !ok {
		t.Fatalf("value = %T, want *MicroflowType", value)
	}
}
