// SPDX-License-Identifier: Apache-2.0

package security

import "testing"

func TestPasswordPolicy_ValidatePassword(t *testing.T) {
	policy := &PasswordPolicy{
		MinimumLength:    8,
		RequireDigit:     true,
		RequireMixedCase: true,
		RequireSymbol:    true,
	}

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid", "Passw0rd!", false},
		{"too short", "Pa0!", true},
		{"no digit", "Password!", true},
		{"no mixed case", "passw0rd!", true},
		{"no symbol", "Passw0rdd", true},
		{"empty", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := policy.ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword(%q) error = %v, wantErr %v", tt.password, err, tt.wantErr)
			}
		})
	}
}

func TestPasswordPolicy_ValidatePassword_NilPolicy(t *testing.T) {
	var policy *PasswordPolicy
	if err := policy.ValidatePassword("anything"); err != nil {
		t.Errorf("nil policy should accept any password, got: %v", err)
	}
}

func TestPasswordPolicy_ValidatePassword_ZeroPolicy(t *testing.T) {
	policy := &PasswordPolicy{}
	if err := policy.ValidatePassword("x"); err != nil {
		t.Errorf("zero policy should accept any password, got: %v", err)
	}
}
