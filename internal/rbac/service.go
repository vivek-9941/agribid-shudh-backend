package rbac

import (
	"context"

	"github.com/agribid/agribid-shudh-backend/internal/auth"
)

// Service provides RBAC business logic.
type Service struct {
	repo Repository
}

// NewService creates a new RBAC service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// HasPermission checks whether the user (via their roles) has the given
// resource+action permission. Admin role has all permissions.
func (s *Service) HasPermission(ctx context.Context, userCtx *auth.UserContext, resource, action string) bool {
	if userCtx == nil {
		return false
	}

	// Admin bypasses all checks.
	if userCtx.IsAdmin {
		return true
	}

	for _, roleCode := range userCtx.Roles {
		perms, err := s.repo.GetPermissionsForRole(ctx, roleCode)
		if err != nil {
			continue
		}
		for _, p := range perms {
			if p.Resource == resource && (p.Action == action || p.Action == "*") {
				return true
			}
		}
	}

	return false
}
