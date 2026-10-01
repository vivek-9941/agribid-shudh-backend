package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/agribid/agribid-shudh-backend/internal/auth"
	apperrors "github.com/agribid/agribid-shudh-backend/internal/errors"
	"github.com/agribid/agribid-shudh-backend/internal/response"
	"github.com/google/uuid"
)

type ctxKeyUserContext struct{}

// Auth returns middleware that validates JWT Bearer tokens and injects UserContext into context.
func Auth(jwtService *auth.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(w, apperrors.Unauthorized("missing authorization header"))
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				response.Error(w, apperrors.Unauthorized("invalid authorization header format"))
				return
			}

			claims, err := jwtService.VerifyAccessToken(parts[1])
			if err != nil {
				response.Error(w, apperrors.Unauthorized("invalid or expired token"))
				return
			}

			userID, err := uuid.Parse(claims.Subject)
			if err != nil {
				response.Error(w, apperrors.Unauthorized("invalid token subject"))
				return
			}

			partnerID := uuid.Nil
			if claims.PartnerID != "" {
				if pid, err := uuid.Parse(claims.PartnerID); err == nil {
					partnerID = pid
				}
			}

			isAdmin := false
			for _, role := range claims.Roles {
				if role == "ADMIN" {
					isAdmin = true
					break
				}
			}

			uc := &auth.UserContext{
				UserID:    userID,
				PartnerID: partnerID,
				Roles:     claims.Roles,
				IsAdmin:   isAdmin,
			}

			ctx := context.WithValue(r.Context(), ctxKeyUserContext{}, uc)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserContext extracts the UserContext from context.
// Returns nil if not present (e.g., unauthenticated request).
func GetUserContext(ctx context.Context) *auth.UserContext {
	if uc, ok := ctx.Value(ctxKeyUserContext{}).(*auth.UserContext); ok {
		return uc
	}
	return nil
}
