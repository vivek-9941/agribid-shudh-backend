package config

import (
	"github.com/kelseyhightower/envconfig"
)

// Config holds all application configuration loaded from environment variables.
// Required fields will cause Load() to return an error if they are not set.
// Optional fields have sensible defaults that work for local development.
type Config struct {
	// Server
	ServerPort  int `envconfig:"SERVER_PORT" default:"8080"`
	MetricsPort int `envconfig:"METRICS_PORT" default:"9090"`

	// Database
	DatabaseURL    string `envconfig:"DATABASE_URL" required:"true"`
	DBMaxOpenConns int    `envconfig:"DB_MAX_OPEN_CONNS" default:"100"`
	DBMinOpenConns int    `envconfig:"DB_MIN_OPEN_CONNS" default:"10"`

	// JWT
	JWTPrivateKeyPath string `envconfig:"JWT_PRIVATE_KEY_PATH" required:"true"`
	JWTPublicKeyPath  string `envconfig:"JWT_PUBLIC_KEY_PATH" required:"true"`
	JWTAccessTTLHours int    `envconfig:"JWT_ACCESS_TTL_HOURS" default:"24"`
	JWTRefreshTTLDays int    `envconfig:"JWT_REFRESH_TTL_DAYS" default:"30"`

	// Logging
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`

	// CORS
	CORSAllowedOrigins []string `envconfig:"CORS_ALLOWED_ORIGINS"`

	// Rate limiting
	RateLimitAuthRPM int `envconfig:"RATE_LIMIT_AUTH_RPM" default:"10"`
	RateLimitAPIPRM  int `envconfig:"RATE_LIMIT_API_RPM" default:"100"`

	// Encryption / secret
	EncryptionKey string `envconfig:"ENCRYPTION_KEY" required:"true"`
	SecretKey     string `envconfig:"SECRET_KEY"`

	// Redis (used for OTP cache)
	RedisHost string `envconfig:"REDIS_HOST" default:"localhost"`
	RedisPort int    `envconfig:"REDIS_PORT" default:"6379"`
	RedisDB   int    `envconfig:"REDIS_DB" default:"0"`
	RedisPass string `envconfig:"REDIS_PASSWORD"` // optional

	// SMS Gateway – National Bulk SMS
	SMSAPIURL          string `envconfig:"SMS_API_URL"`
	SMSAPIUsername     string `envconfig:"SMS_API_USERNAME"`
	SMSAPIPassword     string `envconfig:"SMS_API_PASSWORD"`
	SMSAPISenderID     string `envconfig:"SMS_API_SENDER_ID"`
	SMSDLTContentID    string `envconfig:"SMS_API_DLT_CONTENT_ID"`
	SMSDLTPEiD         string `envconfig:"SMS_API_DLT_PEID"`
}

// Load reads configuration from environment variables and returns a populated
// Config. It fails fast: if any required variable is missing or any value is
// malformed, an error is returned immediately so the application can refuse to
// start with an incomplete configuration.
func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

