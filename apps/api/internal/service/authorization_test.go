package service

import (
	"testing"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/domain"
)

func TestRolePolicy(t *testing.T) {
	if !RoleAllowed(domain.RoleSuperAdmin, domain.RoleSuperAdmin) || RoleAllowed(domain.RoleBPH, domain.RoleSuperAdmin) {
		t.Fatal("role allow-list policy returned an unexpected result")
	}
}

func TestPermissionSummaryKeepsExclusiveSystemActionsForSuperAdmin(t *testing.T) {
	for _, role := range []domain.Role{domain.RoleBPH, domain.RoleKepala} {
		for _, permission := range PermissionsForRole(role) {
			if permission == "users.manage" || permission == "activity_logs.read" {
				t.Fatalf("role %s received super-admin permission %q", role, permission)
			}
		}
	}
	for _, permission := range PermissionsForRole(domain.RoleSuperAdmin) {
		if permission == "users.manage" || permission == "activity_logs.read" {
			return
		}
	}
	t.Fatal("SUPER_ADMIN permission summary omitted exclusive system actions")
}

func TestDivisionScopePolicy(t *testing.T) {
	own, other := int64(4), int64(5)
	if !CanAccessDivision(domain.RoleSuperAdmin, nil, other) || !CanAccessDivision(domain.RoleBPH, nil, other) {
		t.Fatal("cross-division operational roles must pass division scope")
	}
	if !CanAccessDivision(domain.RoleKepala, &own, own) || CanAccessDivision(domain.RoleKepala, &own, other) || CanAccessDivision(domain.RoleKepala, nil, own) {
		t.Fatal("KEPALA_BIDANG division scope policy is incorrect")
	}
}

func TestAuthorizationStateMustHaveValidCurrentRoleAndDivision(t *testing.T) {
	division := int64(4)
	for _, test := range []struct {
		user  domain.User
		valid bool
	}{
		{domain.User{Role: domain.RoleSuperAdmin, Status: domain.StatusActive}, true},
		{domain.User{Role: domain.RoleBPH, Status: domain.StatusActive}, true},
		{domain.User{Role: domain.RoleKepala, DivisionID: &division, Status: domain.StatusActive}, true},
		{domain.User{Role: domain.RoleKepala, Status: domain.StatusActive}, false},
		{domain.User{Role: domain.RoleBPH, DivisionID: &division, Status: domain.StatusActive}, false},
		{domain.User{Role: domain.RoleBPH, Status: domain.StatusInactive}, false},
	} {
		if got := HasValidAuthorizationState(test.user); got != test.valid {
			t.Errorf("HasValidAuthorizationState(%+v) = %v, want %v", test.user, got, test.valid)
		}
	}
}
