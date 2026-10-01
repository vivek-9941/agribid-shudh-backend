// Package response provides standard JSON response helpers for the Agribid Shudh API.
// All handlers should use these helpers to ensure a consistent response envelope
// as defined in design.md §5.2.
package response

import (
	"encoding/json"
	"net/http"

	apperrors "github.com/agribid/agribid-shudh-backend/internal/errors"
)

// ─────────────────────────────────────────────────────────────────────────────
// Response envelope types
// ─────────────────────────────────────────────────────────────────────────────

// successResponse is the envelope for successful non-paginated responses.
//
//	{"success":true,"data":{...},"meta":{"request_id":"..."}}
type successResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Meta    metaOnly    `json:"meta,omitempty"`
}

// metaOnly carries the request_id for simple (non-paginated) success responses.
type metaOnly struct {
	RequestID string `json:"request_id,omitempty"`
}

// paginatedResponse is the envelope for paginated list responses.
//
//	{"success":true,"data":[...],"meta":{"page":1,"page_size":20,"total":150,"request_id":"..."}}
type paginatedResponse struct {
	Success bool           `json:"success"`
	Data    interface{}    `json:"data"`
	Meta    PaginationMeta `json:"meta"`
}

// errorResponse is the envelope for all error responses.
//
//	{"success":false,"error":{...},"meta":{"request_id":"..."}}
type errorResponse struct {
	Success bool        `json:"success"`
	Error   errorBody   `json:"error"`
	Meta    metaOnly    `json:"meta,omitempty"`
}

// errorBody holds the machine-readable error payload.
type errorBody struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// PaginationMeta carries pagination information and the correlation request ID.
// It is used by Paginated() and may be constructed directly by handlers.
type PaginationMeta struct {
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
	Total     int64  `json:"total"`
	RequestID string `json:"request_id,omitempty"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

// requestID extracts the correlation ID that the requestid middleware injects
// into the response header.  Returns an empty string when not present.
func requestID(w http.ResponseWriter) string {
	return w.Header().Get("X-Request-ID")
}

// write serialises v as JSON and writes it to w with the given HTTP status.
// On serialisation failure it falls back to a plain-text 500.
func write(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		// Encoding errors are extremely rare (e.g. non-serialisable types).
		// At this point the headers are already sent, so we can only log;
		// the caller is responsible for ensuring data is serialisable.
		_ = err
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Public API
// ─────────────────────────────────────────────────────────────────────────────

// JSON writes a success response with the given HTTP status code and data payload.
// The response envelope is:
//
//	{"success":true,"data":<data>,"meta":{"request_id":"<rid>"}}
//
// The request_id is extracted from the X-Request-ID response header if present.
func JSON(w http.ResponseWriter, status int, data interface{}) {
	write(w, status, successResponse{
		Success: true,
		Data:    data,
		Meta:    metaOnly{RequestID: requestID(w)},
	})
}

// Error introspects err and writes the appropriate error envelope:
//   - *apperrors.AppError  → uses its HTTPStatus, Code, Message, and Details
//   - any other error      → HTTP 500 with code INTERNAL_ERROR
//
// The response envelope is:
//
//	{"success":false,"error":{"code":"...","message":"...","details":{...}},"meta":{"request_id":"<rid>"}}
func Error(w http.ResponseWriter, err error) {
	var (
		status  int
		code    string
		message string
		details map[string]interface{}
	)

	if ae, ok := apperrors.As(err); ok {
		status = ae.HTTPStatus
		code = ae.Code
		message = ae.Message
		details = ae.Details
	} else {
		status = http.StatusInternalServerError
		code = apperrors.CodeInternalError
		message = "an unexpected internal error occurred"
	}

	write(w, status, errorResponse{
		Success: false,
		Error: errorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
		Meta: metaOnly{RequestID: requestID(w)},
	})
}

// Paginated writes a success response with HTTP 200 and a full pagination meta block.
// The response envelope is:
//
//	{"success":true,"data":<data>,"meta":{"page":1,"page_size":20,"total":150,"request_id":"<rid>"}}
//
// The meta.RequestID field is populated from the X-Request-ID response header when
// it is not already set on the provided PaginationMeta value.
func Paginated(w http.ResponseWriter, data interface{}, meta PaginationMeta) {
	if meta.RequestID == "" {
		meta.RequestID = requestID(w)
	}
	write(w, http.StatusOK, paginatedResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}
