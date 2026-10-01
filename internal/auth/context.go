package auth

import (
	"context"
)

type ctxKeyUserContext struct{}

// WithUserContext returns a new context with UserContext injected.
func WithUserContext(ctx context.Context, uc *UserContext) context.Context {
	return context.WithValue(ctx, ctxKeyUserContext{}, uc)
}

// GetUserContext extracts the UserContext from context.
// Returns nil if not present (e.g., unauthenticated request).
func GetUserContext(ctx context.Context) *UserContext {
	if uc, ok := ctx.Value(ctxKeyUserContext{}).(*UserContext); ok {
		return uc
	}
	return nil
}
