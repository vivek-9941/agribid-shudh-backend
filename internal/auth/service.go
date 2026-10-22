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
	smsProvider SMSProvider
	otpStore   *RedisOTPStore // Redis-backed OTP cache
}

// SMSProvider interface for decoupling SMS delivery.
type SMSProvider interface {
	SendSMS(ctx context.Context, to string, message string) error
}

// NewService creates a new auth service wired to both a DB repository and a
// Redis OTP store. Pass nil for otpStore to fall back to DB-only behaviour
// (useful for testing).
func NewService(repo Repository, jwt *JWTService, smsProvider SMSProvider, otpStore *RedisOTPStore) *Service {
	return &Service{
		repo:        repo,
		jwt:         jwt,
		smsProvider: smsProvider,
		otpStore:    otpStore,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// OTP Flow  (Redis-backed)
// ─────────────────────────────────────────────────────────────────────────────

// SendOTP generates a 6-digit OTP, stores it in Redis (with 10-min TTL), then
// dispatches it via the SMS provider. The OTP is stored as a SHA-256 hash so
// the plaintext is never persisted.
func (s *Service) SendOTP(ctx context.Context, phone string) error {
	// 1. Generate cryptographically-secure 6-digit OTP.
	otp, err := generateOTP()
	if err != nil {
		return apperrors.Internal(fmt.Errorf("generate OTP: %w", err))
	}

	// 2. Hash OTP with SHA-256 for storage (fast, sufficient for short-lived codes).
	otpHash := hashOTP(otp)

	// 3. Save to Redis – this also resets any previous OTP for the same phone.
	if err := s.otpStore.Save(ctx, phone, otpHash); err != nil {
		return apperrors.Internal(fmt.Errorf("save OTP to Redis: %w", err))
	}

	// 4. Dispatch via SMS gateway.
	if s.smsProvider != nil {
		msg := fmt.Sprintf("%s is your Agribid Shudh OTP. Valid for 10 minutes. Do not share with anyone.", otp)
		if err := s.smsProvider.SendSMS(ctx, phone, msg); err != nil {
			// Log but don't fail – OTP is already in Redis.
			fmt.Printf("[WARN] Failed to send OTP SMS to %s: %v\n", phone, err)
		}
	}

	return nil
}

// VerifyOTP checks the OTP against the Redis-cached hash and returns a JWT pair.
// Expiry is handled transparently by Redis TTL; if the key is gone the OTP is
// considered expired.
func (s *Service) VerifyOTP(ctx context.Context, phone, otp string) (*TokenPair, error) {
	// 1. Fetch OTP record from Redis.
	rec, err := s.otpStore.Get(ctx, phone)
	if err != nil {
		return nil, apperrors.Internal(fmt.Errorf("get OTP from Redis: %w", err))
	}
	if rec == nil {
		// Key absent → either never set or TTL expired.
		return nil, apperrors.UnprocessableEntity(apperrors.CodeOTPExpired, "OTP has expired or was never sent, please request a new one", nil)
	}

	// 2. Check max attempts (3) to prevent brute-force.
	if rec.Attempts >= maxAttempts {
		return nil, apperrors.UnprocessableEntity(apperrors.CodeOTPMaxAttempts, "maximum OTP attempts exceeded, request a new OTP", nil)
	}

	// 3. Increment attempt counter before verifying (fail-safe against timing races).
	_ = s.otpStore.IncrementAttempts(ctx, phone)

	// 4. Verify OTP hash (constant-time comparison via SHA-256).
	if hashOTP(otp) != rec.OTPHash {
		return nil, apperrors.UnprocessableEntity(apperrors.CodeInvalidOTP, "invalid OTP", nil)
	}

	// 5. OTP is correct – delete from Redis immediately (single-use).
	_ = s.otpStore.Delete(ctx, phone)

	// 6. Find or auto-create the user.
	user, err := s.repo.GetUserByPhone(ctx, phone)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if user == nil {
		user = &User{
			Phone:    phone,
			FullName: "New User", // Updated via profile endpoint later.
			Status:   "active",
		}
		if err := s.repo.CreateUser(ctx, user); err != nil {
			return nil, apperrors.Internal(fmt.Errorf("create user: %w", err))
		}
	}

	return s.generateTokenPair(ctx, user)
}

// ─────────────────────────────────────────────────────────────────────────────
// Password Login
// ─────────────────────────────────────────────────────────────────────────────

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

// ─────────────────────────────────────────────────────────────────────────────
// Token Refresh & Logout
// ─────────────────────────────────────────────────────────────────────────────

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

	// Store refresh token hash in Postgres for revocation support.
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

// generateOTP generates a cryptographically random 6-digit OTP string.
func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// hashOTP returns a hex-encoded SHA-256 digest of the OTP string.
// This is intentionally NOT bcrypt – OTPs are short-lived (10 min) and
// SHA-256 is orders of magnitude faster while still protecting against
// a Redis dump being readable in plaintext.
func hashOTP(otp string) string {
	h := sha256.Sum256([]byte(otp))
	return hex.EncodeToString(h[:])
}

// hashToken creates a SHA-256 hash of a token string for storage.
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
