// SPDX-License-Identifier: Apache-2.0

package executor

import "testing"

func TestIsBuiltinJavaActionReference(t *testing.T) {
	tests := []struct {
		ref  string
		want bool
	}{
		{ref: "System.VerifyPassword", want: true},
		{ref: "System.VerifyPassword.Extra", want: true},
		{ref: "Administration.VerifyPassword", want: false},
		{ref: "VerifyPassword", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			got := isBuiltinJavaActionReference(tt.ref)
			if got != tt.want {
				t.Fatalf("isBuiltinJavaActionReference(%q) = %v, want %v", tt.ref, got, tt.want)
			}
		})
	}
}
