package auth

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─────────────────────────────────────────────────────────────────────────────
// Mock: Repository
// ─────────────────────────────────────────────────────────────────────────────

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateUser(ctx context.Context, user *User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) GetUserByPhone(ctx context.Context, phone string) (*User, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

// OTP session methods kept for interface compliance – no longer called in the
// new Redis-backed flow, but the Repository interface still declares them.
func (m *MockRepository) CreateOTPSession(ctx context.Context, session *OTPSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockRepository) GetOTPSession(ctx context.Context, phone string) (*OTPSession, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*OTPSession), args.Error(1)
}

func (m *MockRepository) MarkOTPVerified(ctx context.Context, sessionID uuid.UUID) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockRepository) IncrementOTPAttempts(ctx context.Context, sessionID uuid.UUID) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockRepository) CreateRefreshToken(ctx context.Context, token *RefreshToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	args := m.Called(ctx, tokenHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*RefreshToken), args.Error(1)
}

func (m *MockRepository) RevokeRefreshToken(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) RevokeAllRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// ─────────────────────────────────────────────────────────────────────────────
// Mock: SMSProvider
// ─────────────────────────────────────────────────────────────────────────────

type MockSMSProvider struct {
	mock.Mock
}

func (m *MockSMSProvider) SendSMS(ctx context.Context, to string, message string) error {
	args := m.Called(ctx, to, message)
	return args.Error(0)
}

// ─────────────────────────────────────────────────────────────────────────────
// Mock: RedisOTPStore (in-memory for unit tests – no real Redis needed)
// ─────────────────────────────────────────────────────────────────────────────

// testOTPStore is a simple in-memory OTP store used only in tests.
type testOTPStore struct {
	data map[string]*redisOTPRecord
}

func newTestOTPStore() *RedisOTPStore {
	// We return nil so Service uses in-memory fallback paths in tests;
	// tests that need real Redis behaviour should run with a real client.
	// For unit tests we embed a fake via a thin adapter below.
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

// hashOTPTest exposes hashOTP for test assertions.
func hashOTPTest(otp string) string { return hashOTP(otp) }

// ─────────────────────────────────────────────────────────────────────────────
// Tests
// ─────────────────────────────────────────────────────────────────────────────

// TestHashOTP verifies that the same OTP always hashes to the same value.
func TestHashOTP(t *testing.T) {
	otp := "123456"
	h1 := hashOTPTest(otp)
	h2 := hashOTPTest(otp)
	assert.Equal(t, h1, h2, "same OTP must always produce the same hash")
	assert.NotEmpty(t, h1)
}

// TestGenerateOTP verifies that generated OTPs are 6-digit strings.
func TestGenerateOTP(t *testing.T) {
	for i := 0; i < 20; i++ {
		otp, err := generateOTP()
		assert.NoError(t, err)
		assert.Len(t, otp, 6)
		for _, ch := range otp {
			assert.True(t, ch >= '0' && ch <= '9', "OTP must be numeric")
		}
	}
}

// TestLoginWithPassword_InvalidUser tests that login with unknown email returns error.
func TestLoginWithPassword_InvalidUser(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, nil, nil, nil)
	ctx := context.Background()

	mockRepo.On("GetUserByEmail", ctx, "unknown@test.com").Return(nil, nil)

	_, err := service.LoginWithPassword(ctx, "unknown@test.com", "password123")
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

// TestGetMe_NotFound verifies a 404-like error when user doesn't exist.
func TestGetMe_NotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, nil, nil, nil)
	ctx := context.Background()
	uid := uuid.New()

	mockRepo.On("GetUserByID", ctx, uid).Return(nil, nil)

	_, err := service.GetMe(ctx, uid)
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}
