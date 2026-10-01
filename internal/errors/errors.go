// Package errors provides the application error model for Agribid Shudh.
// It defines AppError, sentinel constructors, and the error code catalogue.
package errors

import (
	"fmt"
	"strings"
)

// ─────────────────────────────────────────────────────────────────────────────
// Error code catalogue — §8.3
// ─────────────────────────────────────────────────────────────────────────────

// Common codes
const (
	CodeInternalError    = "INTERNAL_ERROR"
	CodeValidationFailed = "VALIDATION_FAILED"
	CodeUnauthorized     = "UNAUTHORIZED"
	CodeForbidden        = "FORBIDDEN"
	CodeNotFound         = "NOT_FOUND"
	CodeConflict         = "CONFLICT"
)

// Auth codes
const (
	CodeInvalidOTP         = "INVALID_OTP"
	CodeOTPExpired         = "OTP_EXPIRED"
	CodeOTPMaxAttempts     = "OTP_MAX_ATTEMPTS"
	CodeInvalidToken       = "INVALID_TOKEN"
	CodeTokenExpired       = "TOKEN_EXPIRED"
	CodeInvalidCredentials = "INVALID_CREDENTIALS"
)

// RBAC codes
const (
	CodePermissionDenied = "PERMISSION_DENIED"
)

// Partner codes
const (
	CodePartnerNotFound      = "PARTNER_NOT_FOUND"
	CodeKYCAlreadySubmitted  = "KYC_ALREADY_SUBMITTED"
	CodeKYCNotApproved       = "KYC_NOT_APPROVED"
	CodePartnerInactive      = "PARTNER_INACTIVE"
)

// Catalog codes
const (
	CodeProductNotFound  = "PRODUCT_NOT_FOUND"
	CodeDuplicateSKU     = "DUPLICATE_SKU"
	CodeCategoryNotFound = "CATEGORY_NOT_FOUND"
)

// Pricing codes
const (
	CodePriceNotFound  = "PRICE_NOT_FOUND"
	CodeSchemeNotFound = "SCHEME_NOT_FOUND"
)

// Order codes
const (
	CodeOrderNotFound            = "ORDER_NOT_FOUND"
	CodeCartEmpty                = "CART_EMPTY"
	CodeCartNotFound             = "CART_NOT_FOUND"
	CodeOrderCreditLimitExceeded = "ORDER_CREDIT_LIMIT_EXCEEDED"
	CodeOrderInsufficientStock   = "ORDER_INSUFFICIENT_STOCK"
	CodeInvalidOrderTransition   = "INVALID_ORDER_TRANSITION"
)

// Invoice codes
const (
	CodeInvoiceNotFound = "INVOICE_NOT_FOUND"
	CodeInvoiceImmutable = "INVOICE_IMMUTABLE"
)

// Inventory codes
const (
	CodeInsufficientStock = "INSUFFICIENT_STOCK"
	CodeNegativeStock     = "NEGATIVE_STOCK"
)

// Payment codes
const (
	CodePaymentNotFound    = "PAYMENT_NOT_FOUND"
	CodeCreditLimitExceeded = "CREDIT_LIMIT_EXCEEDED"
)

// Returns codes
const (
	CodeReturnNotFound        = "RETURN_NOT_FOUND"
	CodeReturnWindowExpired   = "RETURN_WINDOW_EXPIRED"
	CodeReturnAlreadyProcessed = "RETURN_ALREADY_PROCESSED"
)

// Dispatch codes
const (
	CodeShipmentNotFound = "SHIPMENT_NOT_FOUND"
)

// Notification codes
const (
	CodeNotificationNotFound = "NOTIFICATION_NOT_FOUND"
)

// ─────────────────────────────────────────────────────────────────────────────
// AppError — §8.1
// ─────────────────────────────────────────────────────────────────────────────

// AppError is the canonical application error type. It carries a machine-readable
// Code, a human-readable Message, the suggested HTTP status code, optional
// structured Details, and an optional wrapped underlying error (not exposed to
// the client).
type AppError struct {
	Code       string                 // machine-readable error code e.g. "ORDER_NOT_FOUND"
	Message    string                 // human-readable message
	HTTPStatus int                    // suggested HTTP status code
	Details    map[string]interface{} // optional extra context
	Err        error                  // wrapped underlying error (optional, never sent to client)
}

// Error implements the error interface. It returns a descriptive string combining
// the code, message, and any wrapped error.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the wrapped underlying error so that errors.Is / errors.As
// can traverse the chain.
func (e *AppError) Unwrap() error {
	return e.Err
}

// Is returns true when target is an *AppError with the same Code. This allows
// errors.Is(err, &AppError{Code: "ORDER_NOT_FOUND"}) to work correctly.
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// ─────────────────────────────────────────────────────────────────────────────
// Sentinel constructors — §8.1
// ─────────────────────────────────────────────────────────────────────────────

// NotFound returns a 404 AppError. The code is derived from the resource name,
// e.g. resource="order" → code="ORDER_NOT_FOUND".
func NotFound(resource, id string) *AppError {
	code := strings.ToUpper(resource) + "_NOT_FOUND"
	return &AppError{
		Code:       code,
		Message:    fmt.Sprintf("%s with id %q not found", resource, id),
		HTTPStatus: 404,
	}
}

// Unauthorized returns a 401 AppError with code UNAUTHORIZED.
func Unauthorized(msg string) *AppError {
	return &AppError{
		Code:       CodeUnauthorized,
		Message:    msg,
		HTTPStatus: 401,
	}
}

// Forbidden returns a 403 AppError with code FORBIDDEN.
func Forbidden(msg string) *AppError {
	return &AppError{
		Code:       CodeForbidden,
		Message:    msg,
		HTTPStatus: 403,
	}
}

// BadRequest returns a 400 AppError. The caller supplies the code and message.
func BadRequest(code, msg string) *AppError {
	return &AppError{
		Code:       code,
		Message:    msg,
		HTTPStatus: 400,
	}
}

// Conflict returns a 409 AppError. The caller supplies the code and message.
func Conflict(code, msg string) *AppError {
	return &AppError{
		Code:       code,
		Message:    msg,
		HTTPStatus: 409,
	}
}

// Internal returns a 500 AppError with code INTERNAL_ERROR. The underlying
// error is stored for logging but is never serialised to the client.
func Internal(err error) *AppError {
	return &AppError{
		Code:       CodeInternalError,
		Message:    "an unexpected internal error occurred",
		HTTPStatus: 500,
		Err:        err,
	}
}

// ValidationFailed returns a 400 AppError with code VALIDATION_FAILED. The
// details map should contain field-level validation messages.
func ValidationFailed(details map[string]interface{}) *AppError {
	return &AppError{
		Code:       CodeValidationFailed,
		Message:    "request validation failed",
		HTTPStatus: 400,
		Details:    details,
	}
}

// UnprocessableEntity returns a 422 AppError. The caller supplies the code,
// message, and optional structured details.
func UnprocessableEntity(code, msg string, details map[string]interface{}) *AppError {
	return &AppError{
		Code:       code,
		Message:    msg,
		HTTPStatus: 422,
		Details:    details,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Helper — wrap an existing error as an AppError if it isn't one already
// ─────────────────────────────────────────────────────────────────────────────

// As attempts to cast err to *AppError. Returns (appErr, true) on success.
func As(err error) (*AppError, bool) {
	var ae *AppError
	if ok := errorsAs(err, &ae); ok {
		return ae, true
	}
	return nil, false
}

// errorsAs is a thin wrapper around the stdlib errors.As to avoid importing
// the stdlib "errors" package under the same name as this package in callers.
func errorsAs(err error, target interface{}) bool {
	type asIface interface {
		As(interface{}) bool
	}
	// Walk the chain manually using Unwrap to avoid a stdlib import collision.
	for err != nil {
		if t, ok := target.(**AppError); ok {
			if ae, ok := err.(*AppError); ok {
				*t = ae
				return true
			}
		}
		type unwrapper interface{ Unwrap() error }
		u, ok := err.(unwrapper)
		if !ok {
			break
		}
		err = u.Unwrap()
	}
	return false
}
