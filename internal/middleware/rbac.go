package middleware

import (
	"net/http"

	apperrors "github.com/agribid/agribid-shudh-backend/internal/errors"
	"github.com/agribid/agribid-shudh-backend/internal/rbac"
	"github.com/agribid/agribid-shudh-backend/internal/response"
)

// RBAC returns middleware that checks if the authenticated user has the required
// resource+action permission. Returns 403 on denial.
func RBAC(rbacService *rbac.Service, resource, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userCtx := GetUserContext(r.Context())
			if userCtx == nil {
				response.Error(w, apperrors.Unauthorized("authentication required"))
				return
			}

			if !rbacService.HasPermission(r.Context(), userCtx, resource, action) {
				response.Error(w, apperrors.Forbidden("you do not have permission to perform this action"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
