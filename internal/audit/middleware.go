package audit

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/agribid/agribid-shudh-backend/internal/auth"
	"github.com/google/uuid"
)

type responseWriterInterceptor struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (w *responseWriterInterceptor) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriterInterceptor) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func Middleware(repo Repository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()

			var reqBody []byte
			if r.Body != nil {
				reqBody, _ = io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(reqBody))
			}

			wi := &responseWriterInterceptor{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				body:           bytes.NewBuffer(nil),
			}

			next.ServeHTTP(wi, r)

			duration := time.Since(start).Milliseconds()

			go func() {
				bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				var userID *uuid.UUID
				userCtx, _ := r.Context().Value("user").(*auth.UserContext)
				if userCtx != nil {
					userID = &userCtx.ID
				}

				logEntry := &AuditLog{
					UserID:       userID,
					Action:       r.Method,
					Resource:     strings.Split(r.URL.Path, "/")[1],
					Method:       r.Method,
					Path:         r.URL.Path,
					StatusCode:   wi.statusCode,
					IPAddress:    r.RemoteAddr,
					UserAgent:    r.UserAgent(),
					RequestSize:  r.ContentLength,
					ResponseSize: int64(wi.body.Len()),
					DurationMs:   duration,
					RequestData:  maskSensitive(string(reqBody)),
					ResponseData: maskSensitive(wi.body.String()),
				}

				_ = repo.Insert(bgCtx, logEntry)
			}()
		})
	}
}

var sensitivePatterns = []string{
	`"password"\s*:\s*"[^"]+"`,
	`"token"\s*:\s*"[^"]+"`,
	`"otp"\s*:\s*"[^"]+"`,
	`"pan"\s*:\s*"[^"]+"`,
	`"gstin"\s*:\s*"[^"]+"`,
}

func maskSensitive(data string) string {
	for _, pat := range sensitivePatterns {
		re := regexp.MustCompile(pat)
		data = re.ReplaceAllStringFunc(data, func(match string) string {
			parts := strings.SplitN(match, ":", 2)
			if len(parts) == 2 {
				return parts[0] + `: "***"`
			}
			return match
		})
	}
	return data
}
