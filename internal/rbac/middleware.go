package rbac

import (
	"net/http"

	"github.com/agribid/agribid-shudh-backend/internal/auth"
	apperrors "github.com/agribid/agribid-shudh-backend/internal/errors"
	"github.com/agribid/agribid-shudh-backend/internal/response"
)

// RequirePermission returns middleware that checks if the user has the specified permission.
func RequirePermission(svc *Service, resource, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userCtx, ok := r.Context().Value("user").(*auth.UserContext)
			if !ok || userCtx == nil {
				response.Error(w, apperrors.Unauthorized("authentication required"))
				return
			}

			if !svc.HasPermission(r.Context(), userCtx, resource, action) {
				response.Error(w, apperrors.Forbidden("insufficient permissions"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
