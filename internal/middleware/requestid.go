// Package middleware provides HTTP middleware for the Agribid Shudh API server.
package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type ctxKeyRequestID struct{}

// RequestID injects a UUID into context and the X-Request-ID response header.
// If the incoming request already carries X-Request-ID, that value is reused.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get("X-Request-ID")
		if rid == "" {
			rid = uuid.New().String()
		}

		// Set the header so downstream code (including response helpers) can read it.
		w.Header().Set("X-Request-ID", rid)

		// Inject into context for structured logging.
		ctx := context.WithValue(r.Context(), ctxKeyRequestID{}, rid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID extracts the request ID from the context.
// Returns an empty string if not present.
func GetRequestID(ctx context.Context) string {
	if rid, ok := ctx.Value(ctxKeyRequestID{}).(string); ok {
		return rid
	}
	return ""
}
