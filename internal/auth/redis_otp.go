// Package auth – redis_otp.go provides a Redis-backed OTP cache.
// OTP records are stored as JSON with a TTL so no explicit expiry check is needed;
// Redis automatically evicts expired keys.
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	otpTTL       = 10 * time.Minute
	otpKeyPrefix = "otp:"
	maxAttempts  = 3
)

// RedisOTPStore is a thin wrapper around redis.Client that satisfies the
// OTPStore interface used by Service.SendOTP / Service.VerifyOTP.
type RedisOTPStore struct {
	client *redis.Client
}

// redisOTPRecord is what we serialise into Redis.
type redisOTPRecord struct {
	Phone    string `json:"phone"`
	OTPHash  string `json:"otp_hash"` // SHA-256 hex of the OTP (fast comparison)
	Attempts int    `json:"attempts"`
}

// NewRedisOTPStore creates a new store backed by the provided Redis client.
func NewRedisOTPStore(client *redis.Client) *RedisOTPStore {
	return &RedisOTPStore{client: client}
}

// otpKey builds the Redis key for a phone number.
func otpKey(phone string) string {
	return otpKeyPrefix + phone
}

// Save stores a hashed OTP for the given phone with a 10-minute TTL.
// Any previous OTP for the same phone is overwritten.
func (r *RedisOTPStore) Save(ctx context.Context, phone, otpHash string) error {
	rec := redisOTPRecord{
		Phone:    phone,
		OTPHash:  otpHash,
		Attempts: 0,
	}
	data, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("redis otp save marshal: %w", err)
	}
	if err := r.client.Set(ctx, otpKey(phone), data, otpTTL).Err(); err != nil {
		return fmt.Errorf("redis otp save set: %w", err)
	}
	return nil
}

// Get retrieves the OTP record for the given phone.
// Returns (nil, nil) when the key does not exist (expired or never set).
func (r *RedisOTPStore) Get(ctx context.Context, phone string) (*redisOTPRecord, error) {
	data, err := r.client.Get(ctx, otpKey(phone)).Bytes()
	if err == redis.Nil {
		return nil, nil // key expired / not found
	}
	if err != nil {
		return nil, fmt.Errorf("redis otp get: %w", err)
	}
	var rec redisOTPRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, fmt.Errorf("redis otp get unmarshal: %w", err)
	}
	return &rec, nil
}

// IncrementAttempts atomically increments the attempt counter.
// It re-fetches, updates, and writes back with the remaining TTL preserved.
func (r *RedisOTPStore) IncrementAttempts(ctx context.Context, phone string) error {
	key := otpKey(phone)

	// Use a pipeline to read TTL and value atomically.
	pipe := r.client.Pipeline()
	getCmd := pipe.Get(ctx, key)
	ttlCmd := pipe.TTL(ctx, key)
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return fmt.Errorf("redis otp incr pipeline: %w", err)
	}

	data, err := getCmd.Bytes()
	if err != nil {
		return nil // key gone – nothing to do
	}

	ttl, _ := ttlCmd.Result()
	if ttl <= 0 {
		ttl = otpTTL
	}

	var rec redisOTPRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return fmt.Errorf("redis otp incr unmarshal: %w", err)
	}
	rec.Attempts++

	updated, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("redis otp incr marshal: %w", err)
	}
	return r.client.Set(ctx, key, updated, ttl).Err()
}

// Delete removes the OTP record (called after successful verification).
func (r *RedisOTPStore) Delete(ctx context.Context, phone string) error {
	return r.client.Del(ctx, otpKey(phone)).Err()
}

// NewRedisClient creates a redis.Client from individual host/port/db/pass values.
func NewRedisClient(host string, port int, db int, password string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       db,
	})
}

// PingRedis checks that the Redis connection is alive.
func PingRedis(ctx context.Context, client *redis.Client) error {
	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}
	return nil
}
