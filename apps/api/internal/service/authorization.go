package service

import (
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/domain"
)

func RoleAllowed(role domain.Role, allowed ...domain.Role) bool {
	for _, candidate := range allowed {
		if role == candidate {
			return true
		}
	}
	return false
}

func CanAccessDivision(role domain.Role, ownDivisionID *int64, resourceDivisionID int64) bool {
	switch role {
	case domain.RoleSuperAdmin, domain.RoleBPH:
		return true
	case domain.RoleKepala:
		return ownDivisionID != nil && *ownDivisionID > 0 && *ownDivisionID == resourceDivisionID
	default:
		return false
	}
}

func HasValidAuthorizationState(user domain.User) bool {
	if user.Status != domain.StatusActive {
		return false
	}
	switch user.Role {
	case domain.RoleSuperAdmin, domain.RoleBPH:
		return user.DivisionID == nil
	case domain.RoleKepala:
		return user.DivisionID != nil && *user.DivisionID > 0
	default:
		return false
	}
}

func PermissionsForRole(role domain.Role) []string {
	switch role {
	case domain.RoleSuperAdmin:
		return []string{"dashboard.read", "users.manage", "periods.manage", "divisions.manage", "members.manage", "content.manage_all", "documents.manage_all", "registrations.manage_all", "aspirations.read_all", "landing_settings.manage", "activity_logs.read"}
	case domain.RoleBPH:
		return []string{"dashboard.read", "periods.read", "divisions.manage", "members.manage", "content.manage_all", "documents.manage_all", "registrations.manage_all", "aspirations.read_all", "landing_settings.manage"}
	case domain.RoleKepala:
		return []string{"dashboard.read", "periods.read", "divisions.read_own", "members.read_own", "articles.manage_own", "gallery.manage_own", "announcements.read", "documents.manage_own", "registrations.manage_own", "aspirations.read_routed"}
	default:
		return []string{}
	}
}
