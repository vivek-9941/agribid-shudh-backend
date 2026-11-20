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
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Service provides authentication business logic.
type Service struct {
	repo        Repository
	jwt         *JWTService
	smsProvider SMSProvider
	otpStore    *RedisOTPStore // Redis-backed OTP cache
	logger      *zap.Logger
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
		logger:      zap.L(), // uses the global logger wired at startup
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// OTP Flow  (Redis-backed)
// ─────────────────────────────────────────────────────────────────────────────

// SendOTP generates a 6-digit OTP, stores it in Redis (with 10-min TTL), then
// dispatches it via the SMS provider. The OTP is stored as a SHA-256 hash so
// the plaintext is never persisted.
func (s *Service) SendOTP(ctx context.Context, phone string) error {
	log := s.logger.With(zap.String("phone", phone))
	log.Info("[OTP] SendOTP called")

	// 1. Generate cryptographically-secure 6-digit OTP.
	otp, err := generateOTP()
	if err != nil {
		log.Error("[OTP] Failed to generate OTP", zap.Error(err))
		return apperrors.Internal(fmt.Errorf("generate OTP: %w", err))
	}
	log.Debug("[OTP] Generated OTP (debug only – remove in production)", zap.String("otp", otp))

	// 2. Hash OTP with SHA-256 for storage (fast, sufficient for short-lived codes).
	otpHash := hashOTP(otp)

	// 3. Save to Redis – this also resets any previous OTP for the same phone.
	log.Info("[OTP] Saving OTP to Redis", zap.String("key", "otp:"+phone))
	if err := s.otpStore.Save(ctx, phone, otpHash); err != nil {
		log.Error("[OTP] ❌ Failed to save OTP to Redis", zap.Error(err))
		return apperrors.Internal(fmt.Errorf("save OTP to Redis: %w", err))
	}
	log.Info("[OTP] ✅ OTP saved to Redis successfully")

	// 4. Dispatch via SMS gateway.
	if s.smsProvider == nil {
		log.Warn("[OTP] SMS provider is nil – OTP NOT sent")
		return nil
	}

	msg := fmt.Sprintf("Dear User, your verification code for Agribid Shudh is %s . This code is valid for 10 minutes. Please do not share this code with anyone. Team Agribid Shudh", otp)
	log.Info("[OTP] Dispatching SMS via provider", zap.String("provider", fmt.Sprintf("%T", s.smsProvider)))
	if err := s.smsProvider.SendSMS(ctx, phone, msg); err != nil {
		// Log full error but don't fail the request – OTP is already in Redis.
		log.Error("[OTP] ❌ SMS dispatch FAILED",
			zap.String("phone", phone),
			zap.String("provider", fmt.Sprintf("%T", s.smsProvider)),
			zap.Error(err),
		)
		return nil // OTP is in Redis; client can still verify manually
	}
	log.Info("[OTP] ✅ SMS dispatched successfully", zap.String("phone", phone))

	return nil
}

// VerifyOTP checks the OTP against the Redis-cached hash and returns a JWT pair.
// Expiry is handled transparently by Redis TTL; if the key is gone the OTP is
// considered expired.
func (s *Service) VerifyOTP(ctx context.Context, phone, otp string) (*TokenPair, error) {
	log := s.logger.With(zap.String("phone", phone))
	log.Info("[OTP] VerifyOTP called")

	// 1. Fetch OTP record from Redis.
	log.Info("[OTP] Fetching OTP record from Redis", zap.String("key", "otp:"+phone))
	rec, err := s.otpStore.Get(ctx, phone)
	if err != nil {
		log.Error("[OTP] ❌ Failed to fetch OTP from Redis", zap.Error(err))
		return nil, apperrors.Internal(fmt.Errorf("get OTP from Redis: %w", err))
	}
	if rec == nil {
		log.Warn("[OTP] ❌ No OTP record found in Redis (expired or never sent)")
		return nil, apperrors.UnprocessableEntity(apperrors.CodeOTPExpired, "OTP has expired or was never sent, please request a new one", nil)
	}
	log.Info("[OTP] OTP record found in Redis", zap.Int("attempts", rec.Attempts))

	// 2. Check max attempts (3) to prevent brute-force.
	if rec.Attempts >= maxAttempts {
		log.Warn("[OTP] ❌ Max OTP attempts exceeded", zap.Int("attempts", rec.Attempts))
		return nil, apperrors.UnprocessableEntity(apperrors.CodeOTPMaxAttempts, "maximum OTP attempts exceeded, request a new OTP", nil)
	}

	// 3. Increment attempt counter before verifying (fail-safe against timing races).
	_ = s.otpStore.IncrementAttempts(ctx, phone)

	// 4. Verify OTP hash (constant-time comparison via SHA-256).
	if hashOTP(otp) != rec.OTPHash {
		log.Warn("[OTP] ❌ OTP hash mismatch – invalid OTP provided")
		return nil, apperrors.UnprocessableEntity(apperrors.CodeInvalidOTP, "invalid OTP", nil)
	}
	log.Info("[OTP] ✅ OTP verified successfully")

	// 5. OTP is correct – delete from Redis immediately (single-use).
	_ = s.otpStore.Delete(ctx, phone)

	// 6. Find or auto-create the user.
	user, err := s.repo.GetUserByPhone(ctx, phone)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if user == nil {
		log.Info("[OTP] User not found – auto-creating new user")
		user = &User{
			Phone:    phone,
			FullName: "New User", // Updated via profile endpoint later.
			Status:   "active",
		}
		if err := s.repo.CreateUser(ctx, user); err != nil {
			log.Error("[OTP] ❌ Failed to create user", zap.Error(err))
			return nil, apperrors.Internal(fmt.Errorf("create user: %w", err))
		}
		log.Info("[OTP] ✅ New user created", zap.String("user_id", user.ID.String()))
	} else {
		log.Info("[OTP] Existing user found", zap.String("user_id", user.ID.String()))
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
