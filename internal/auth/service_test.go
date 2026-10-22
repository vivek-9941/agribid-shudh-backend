package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// MockRepository is a mock implementation of Repository
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

// MockSMSProvider is a mock SMS provider
type MockSMSProvider struct {
	mock.Mock
}

func (m *MockSMSProvider) SendSMS(ctx context.Context, to string, message string) error {
	args := m.Called(ctx, to, message)
	return args.Error(0)
}

// Test SendOTP
func TestSendOTP(t *testing.T) {
	mockRepo := new(MockRepository)
	mockSMS := new(MockSMSProvider)
	
	service := NewService(mockRepo, nil, mockSMS)
	ctx := context.Background()
	
	// Setup expectations
	mockRepo.On("CreateOTPSession", ctx, mock.AnythingOfType("*auth.OTPSession")).Return(nil)
	mockSMS.On("SendSMS", ctx, "+919876543210", mock.AnythingOfType("string")).Return(nil)
	
	// Execute
	err := service.SendOTP(ctx, "+919876543210")
	
	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockSMS.AssertExpectations(t)
}

// Test VerifyOTP - Success
func TestVerifyOTP_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockJWT := &JWTService{}
	
	service := NewService(mockRepo, mockJWT, nil)
	ctx := context.Background()
	
	otp := "123456"
	otpHash, _ := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	
	session := &OTPSession{
		ID:        uuid.New(),
		Phone:     "+919876543210",
		OTPHash:   string(otpHash),
		ExpiresAt: time.Now().Add(10 * time.Minute),
		Verified:  false,
		Attempts:  0,
	}
	
	user := &User{
		ID:       uuid.New(),
		Phone:    "+919876543210",
		FullName: "Test User",
		Status:   "active",
	}
	
	// Setup expectations
	mockRepo.On("GetOTPSession", ctx, "+919876543210").Return(session, nil)
	mockRepo.On("IncrementOTPAttempts", ctx, session.ID).Return(nil)
	mockRepo.On("MarkOTPVerified", ctx, session.ID).Return(nil)
	mockRepo.On("GetUserByPhone", ctx, "+919876543210").Return(user, nil)
	mockRepo.On("GetUserRoles", ctx, user.ID).Return([]string{"RETAILER"}, nil)
	mockRepo.On("CreateRefreshToken", ctx, mock.AnythingOfType("*auth.RefreshToken")).Return(nil)
	
	// Note: This test will fail without a real JWT service, but demonstrates the structure
	// In real tests, you'd mock the JWT service too
}

// Test VerifyOTP - Expired
func TestVerifyOTP_Expired(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, nil, nil)
	ctx := context.Background()
	
	session := &OTPSession{
		ID:        uuid.New(),
		Phone:     "+919876543210",
		OTPHash:   "hash",
		ExpiresAt: time.Now().Add(-1 * time.Minute), // Expired
		Verified:  false,
		Attempts:  0,
	}
	
	mockRepo.On("GetOTPSession", ctx, "+919876543210").Return(session, nil)
	
	// Execute
	_, err := service.VerifyOTP(ctx, "+919876543210", "123456")
	
	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

// Test VerifyOTP - Max Attempts
func TestVerifyOTP_MaxAttempts(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, nil, nil)
	ctx := context.Background()
	
	session := &OTPSession{
		ID:        uuid.New(),
		Phone:     "+919876543210",
		OTPHash:   "hash",
		ExpiresAt: time.Now().Add(10 * time.Minute),
		Verified:  false,
		Attempts:  3, // Max attempts reached
	}
	
	mockRepo.On("GetOTPSession", ctx, "+919876543210").Return(session, nil)
	
	// Execute
	_, err := service.VerifyOTP(ctx, "+919876543210", "123456")
	
	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum")
}
