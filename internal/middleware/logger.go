package middleware

import (
	"net/http"
	"time"

	"github.com/agribid/agribid-shudh-backend/internal/logger"
	"go.uber.org/zap"
)

// statusRecorder wraps http.ResponseWriter to capture the status code.
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.statusCode = code
	sr.ResponseWriter.WriteHeader(code)
}

// Logger logs the start and completion of each request with duration, status,
// method, path, and request_id.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rid := GetRequestID(r.Context())

		logger.Get().Info("request started",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("request_id", rid),
			zap.String("remote_addr", r.RemoteAddr),
		)

		rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rec, r)

		logger.Get().Info("request completed",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", rec.statusCode),
			zap.Duration("duration", time.Since(start)),
			zap.String("request_id", rid),
		)
	})
}
