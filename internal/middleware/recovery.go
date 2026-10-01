package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/agribid/agribid-shudh-backend/internal/logger"
	"github.com/agribid/agribid-shudh-backend/internal/response"
	"go.uber.org/zap"
)

// Recovery catches panics in downstream handlers, logs the stack trace,
// and returns a 500 Internal Server Error JSON response.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				stack := string(debug.Stack())
				logger.Get().Error("panic recovered",
					zap.Any("panic", rec),
					zap.String("stack", stack),
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.String("request_id", GetRequestID(r.Context())),
				)

				// Write a generic 500 response.
				response.Error(w, fmt.Errorf("internal server error"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
