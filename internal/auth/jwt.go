package auth

import (
	"crypto/rsa"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTService handles JWT token creation and verification using RS256.
type JWTService struct {
	privateKey    *rsa.PrivateKey
	publicKey     *rsa.PublicKey
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

// CustomClaims extends jwt.RegisteredClaims with application-specific fields.
type CustomClaims struct {
	PartnerID string   `json:"partner_id,omitempty"`
	Roles     []string `json:"roles"`
	jwt.RegisteredClaims
}

// NewJWTService loads RSA keys from file paths and creates a new JWTService.
func NewJWTService(privateKeyPath, publicKeyPath string, accessTTLHours, refreshTTLDays int) (*JWTService, error) {
	privBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}
	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(privBytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	pubBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read public key: %w", err)
	}
	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubBytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	return &JWTService{
		privateKey: privKey,
		publicKey:  pubKey,
		accessTTL:  time.Duration(accessTTLHours) * time.Hour,
		refreshTTL: time.Duration(refreshTTLDays) * 24 * time.Hour,
	}, nil
}

// GenerateAccessToken creates a signed JWT access token for the given user.
func (j *JWTService) GenerateAccessToken(userID, partnerID uuid.UUID, roles []string) (string, error) {
	now := time.Now()
	claims := CustomClaims{
		PartnerID: partnerID.String(),
		Roles:     roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessTTL)),
			ID:        uuid.New().String(),
			Issuer:    "agribid-shudh",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(j.privateKey)
}

// GenerateRefreshToken creates a signed JWT refresh token.
func (j *JWTService) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshTTL)),
		ID:        uuid.New().String(),
		Issuer:    "agribid-shudh",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(j.privateKey)
}

// VerifyAccessToken parses and validates a JWT access token, returning the claims.
func (j *JWTService) VerifyAccessToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.publicKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("verify token: %w", err)
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// VerifyRefreshToken parses and validates a JWT refresh token.
func (j *JWTService) VerifyRefreshToken(tokenString string) (*jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.publicKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("verify refresh token: %w", err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid refresh token claims")
	}

	return claims, nil
}

// AccessTTLSeconds returns the access token TTL in seconds (for response).
func (j *JWTService) AccessTTLSeconds() int64 {
	return int64(j.accessTTL.Seconds())
}
