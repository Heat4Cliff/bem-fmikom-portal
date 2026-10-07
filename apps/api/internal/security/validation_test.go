package security

import (
	"testing"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/domain"
)

func TestRoleDivisionValidation(t *testing.T) {
	division := int64(3)
	tests := []struct {
		role     domain.Role
		division *int64
		valid    bool
	}{
		{domain.RoleSuperAdmin, nil, true}, {domain.RoleSuperAdmin, &division, false},
		{domain.RoleBPH, nil, true}, {domain.RoleBPH, &division, false},
		{domain.RoleKepala, &division, true}, {domain.RoleKepala, nil, false},
		{domain.Role("UNKNOWN"), nil, false},
	}
	for _, test := range tests {
		if err := ValidateRoleDivision(test.role, test.division); (err == nil) != test.valid {
			t.Errorf("ValidateRoleDivision(%s, %v) error = %v", test.role, test.division, err)
		}
	}
}
