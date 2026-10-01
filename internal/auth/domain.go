// Package auth implements authentication and session management.
package auth

import (
	"time"

	"github.com/google/uuid"
)

// User represents a registered user in the system.
type User struct {
	ID           uuid.UUID  `json:"id"`
	Phone        string     `json:"phone"`
	Email        *string    `json:"email,omitempty"`
	PasswordHash string     `json:"-"` // never sent to client
	FullName     string     `json:"full_name"`
	Status       string     `json:"status"` // active, inactive, suspended
	PartnerID    *uuid.UUID `json:"partner_id,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

// OTPSession represents a pending OTP verification.
type OTPSession struct {
	ID        uuid.UUID `json:"id"`
	Phone     string    `json:"phone"`
	OTPHash   string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	Verified  bool      `json:"verified"`
	Attempts  int       `json:"attempts"`
	CreatedAt time.Time `json:"created_at"`
}

// RefreshToken represents a stored refresh token.
type RefreshToken struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	TokenHash string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	Revoked   bool      `json:"revoked"`
	CreatedAt time.Time `json:"created_at"`
}

// UserContext is extracted from the JWT and injected into request context.
// It is used by all downstream handlers and services for auth/scoping decisions.
type UserContext struct {
	UserID    uuid.UUID `json:"user_id"`
	PartnerID uuid.UUID `json:"partner_id"`
	Roles     []string  `json:"roles"`
	IsAdmin   bool      `json:"is_admin"`
}

// TokenPair holds the access and refresh tokens returned to the client.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // seconds until access token expiry
}

// ─────────────────────────────────────────────────────────────────────────────
// Request DTOs
// ─────────────────────────────────────────────────────────────────────────────

// SendOTPRequest is the payload for POST /auth/otp/send.
type SendOTPRequest struct {
	Phone string `json:"phone" validate:"required,e164"`
}

// VerifyOTPRequest is the payload for POST /auth/otp/verify.
type VerifyOTPRequest struct {
	Phone string `json:"phone" validate:"required,e164"`
	OTP   string `json:"otp" validate:"required,len=6,numeric"`
}

// LoginRequest is the payload for POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// RefreshRequest is the payload for POST /auth/refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RegisterRequest is the payload for POST /auth/register.
type RegisterRequest struct {
	Phone    string `json:"phone" validate:"required,e164"`
	FullName string `json:"full_name" validate:"required,min=2,max=255"`
	Email    string `json:"email,omitempty" validate:"omitempty,email"`
	Password string `json:"password,omitempty" validate:"omitempty,min=8"`
}
