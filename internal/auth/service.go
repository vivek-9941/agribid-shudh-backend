package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	apperrors "github.com/agribid/agribid-shudh-backend/internal/errors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Service provides authentication business logic.
type Service struct {
	repo       Repository
	jwt        *JWTService
}

// NewService creates a new auth service.
func NewService(repo Repository, jwt *JWTService) *Service {
	return &Service{repo: repo, jwt: jwt}
}

// SendOTP generates a 6-digit OTP, hashes it, and stores an OTP session.
// In production, this would also dispatch the OTP via SMS.
func (s *Service) SendOTP(ctx context.Context, phone string) error {
	// Generate 6-digit OTP.
	otp, err := generateOTP()
	if err != nil {
		return apperrors.Internal(fmt.Errorf("generate OTP: %w", err))
	}

	// Hash OTP with bcrypt for storage.
	hash, err := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	if err != nil {
		return apperrors.Internal(fmt.Errorf("hash OTP: %w", err))
	}

	session := &OTPSession{
		Phone:     phone,
		OTPHash:   string(hash),
		ExpiresAt: time.Now().Add(10 * time.Minute),
		Verified:  false,
		Attempts:  0,
	}

	if err := s.repo.CreateOTPSession(ctx, session); err != nil {
		return apperrors.Internal(fmt.Errorf("create OTP session: %w", err))
	}

	// TODO: Dispatch OTP via SMS provider (integration layer)
	// For now, OTP is stored and would be visible in logs during development.

	return nil
}

// VerifyOTP checks the OTP against the stored session and returns tokens.
func (s *Service) VerifyOTP(ctx context.Context, phone, otp string) (*TokenPair, error) {
	session, err := s.repo.GetOTPSession(ctx, phone)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if session == nil {
		return nil, apperrors.UnprocessableEntity(apperrors.CodeInvalidOTP, "no OTP session found for this phone", nil)
	}

	// Check max attempts (3).
	if session.Attempts >= 3 {
		return nil, apperrors.UnprocessableEntity(apperrors.CodeOTPMaxAttempts, "maximum OTP attempts exceeded, request a new OTP", nil)
	}

	// Check expiry.
	if time.Now().After(session.ExpiresAt) {
		return nil, apperrors.UnprocessableEntity(apperrors.CodeOTPExpired, "OTP has expired, request a new one", nil)
	}

	// Increment attempts.
	_ = s.repo.IncrementOTPAttempts(ctx, session.ID)

	// Verify OTP.
	if err := bcrypt.CompareHashAndPassword([]byte(session.OTPHash), []byte(otp)); err != nil {
		return nil, apperrors.UnprocessableEntity(apperrors.CodeInvalidOTP, "invalid OTP", nil)
	}

	// Mark verified.
	if err := s.repo.MarkOTPVerified(ctx, session.ID); err != nil {
		return nil, apperrors.Internal(err)
	}

	// Find or create user.
	user, err := s.repo.GetUserByPhone(ctx, phone)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if user == nil {
		// Auto-register user on first OTP verification.
		user = &User{
			Phone:    phone,
			FullName: "New User", // Will be updated via profile
			Status:   "active",
		}
		if err := s.repo.CreateUser(ctx, user); err != nil {
			return nil, apperrors.Internal(fmt.Errorf("create user: %w", err))
		}
	}

	return s.generateTokenPair(ctx, user)
}

// LoginWithPassword authenticates via email + password and returns tokens.
func (s *Service) LoginWithPassword(ctx context.Context, email, password string) (*TokenPair, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if user == nil {
		return nil, apperrors.UnprocessableEntity(apperrors.CodeInvalidCredentials, "invalid email or password", nil)
	}

	if user.PasswordHash == "" {
		return nil, apperrors.UnprocessableEntity(apperrors.CodeInvalidCredentials, "password login not available for this account", nil)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, apperrors.UnprocessableEntity(apperrors.CodeInvalidCredentials, "invalid email or password", nil)
	}

	if user.Status != "active" {
		return nil, apperrors.Forbidden("account is not active")
	}

	return s.generateTokenPair(ctx, user)
}

// RefreshToken validates the refresh token and issues a new token pair.
func (s *Service) RefreshToken(ctx context.Context, refreshTokenStr string) (*TokenPair, error) {
	claims, err := s.jwt.VerifyRefreshToken(refreshTokenStr)
	if err != nil {
		return nil, apperrors.Unauthorized("invalid refresh token")
	}

	tokenHash := hashToken(refreshTokenStr)
	stored, err := s.repo.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if stored == nil || stored.Revoked {
		return nil, apperrors.Unauthorized("refresh token has been revoked")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, apperrors.Unauthorized("invalid token subject")
	}

	// Revoke old refresh token.
	_ = s.repo.RevokeRefreshToken(ctx, stored.ID)

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if user == nil {
		return nil, apperrors.Unauthorized("user not found")
	}

	return s.generateTokenPair(ctx, user)
}

// Logout revokes all refresh tokens for the user.
func (s *Service) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.repo.RevokeAllRefreshTokens(ctx, userID)
}

// GetMe returns the user profile for the given user ID.
func (s *Service) GetMe(ctx context.Context, userID uuid.UUID) (*User, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if user == nil {
		return nil, apperrors.NotFound("user", userID.String())
	}
	return user, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func (s *Service) generateTokenPair(ctx context.Context, user *User) (*TokenPair, error) {
	roles, err := s.repo.GetUserRoles(ctx, user.ID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}

	partnerID := uuid.Nil
	if user.PartnerID != nil {
		partnerID = *user.PartnerID
	}

	accessToken, err := s.jwt.GenerateAccessToken(user.ID, partnerID, roles)
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("generate access token: %w", err))
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("generate refresh token: %w", err))
	}

	// Store refresh token hash.
	rt := &RefreshToken{
		UserID:    user.ID,
		TokenHash: hashToken(refreshToken),
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		Revoked:   false,
	}
	if err := s.repo.CreateRefreshToken(ctx, rt); err != nil {
		return nil, apperrors.Internal(fmt.Errorf("store refresh token: %w", err))
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.jwt.AccessTTLSeconds(),
	}, nil
}

// generateOTP generates a cryptographically random 6-digit OTP.
func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// hashToken creates a SHA-256 hash of a token string for storage.
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
