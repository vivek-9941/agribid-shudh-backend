package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the data access interface for auth entities.
type Repository interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByPhone(ctx context.Context, phone string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error)

	CreateOTPSession(ctx context.Context, otp *OTPSession) error
	GetOTPSession(ctx context.Context, phone string) (*OTPSession, error)
	IncrementOTPAttempts(ctx context.Context, id uuid.UUID) error
	MarkOTPVerified(ctx context.Context, id uuid.UUID) error

	CreateRefreshToken(ctx context.Context, rt *RefreshToken) error
	GetRefreshToken(ctx context.Context, tokenHash string) (*RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, id uuid.UUID) error
	RevokeAllRefreshTokens(ctx context.Context, userID uuid.UUID) error
}

// pgRepo implements Repository using pgx.
type pgRepo struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new pgx-backed auth repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepo{pool: pool}
}


func (r *pgRepo) CreateUser(ctx context.Context, user *User) error {
	query := `INSERT INTO users (id, phone, email, password_hash, full_name, status, partner_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	now := time.Now()
	user.ID = uuid.New()
	user.CreatedAt = now
	user.UpdatedAt = now
	if user.Status == "" {
		user.Status = "active"
	}
	_, err := r.pool.Exec(ctx, query,
		user.ID, user.Phone, user.Email, user.PasswordHash, user.FullName,
		user.Status, user.PartnerID, user.CreatedAt, user.UpdatedAt)
	return err
}

func (r *pgRepo) GetUserByPhone(ctx context.Context, phone string) (*User, error) {
	return r.getUserBy(ctx, "phone", phone)
}

func (r *pgRepo) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return r.getUserBy(ctx, "email", email)
}

func (r *pgRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return r.getUserBy(ctx, "id", id)
}

func (r *pgRepo) getUserBy(ctx context.Context, field string, value interface{}) (*User, error) {
	query := fmt.Sprintf(`SELECT id, phone, email, password_hash, full_name, status, partner_id, created_at, updated_at, deleted_at
		FROM users WHERE %s = $1 AND deleted_at IS NULL`, field)

	var u User
	err := r.pool.QueryRow(ctx, query, value).Scan(
		&u.ID, &u.Phone, &u.Email, &u.PasswordHash, &u.FullName,
		&u.Status, &u.PartnerID, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getUserBy %s: %w", field, err)
	}
	return &u, nil
}

func (r *pgRepo) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	query := `SELECT r.code FROM roles r
		JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = $1`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("getUserRoles: %w", err)
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, fmt.Errorf("getUserRoles scan: %w", err)
		}
		roles = append(roles, code)
	}
	return roles, nil
}

func (r *pgRepo) CreateOTPSession(ctx context.Context, otp *OTPSession) error {
	query := `INSERT INTO otp_sessions (id, phone, otp_hash, expires_at, verified, attempts, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	otp.ID = uuid.New()
	otp.CreatedAt = time.Now()
	_, err := r.pool.Exec(ctx, query,
		otp.ID, otp.Phone, otp.OTPHash, otp.ExpiresAt, otp.Verified, otp.Attempts, otp.CreatedAt)
	return err
}

func (r *pgRepo) GetOTPSession(ctx context.Context, phone string) (*OTPSession, error) {
	query := `SELECT id, phone, otp_hash, expires_at, verified, attempts, created_at
		FROM otp_sessions
		WHERE phone = $1 AND verified = FALSE
		ORDER BY created_at DESC LIMIT 1`

	var o OTPSession
	err := r.pool.QueryRow(ctx, query, phone).Scan(
		&o.ID, &o.Phone, &o.OTPHash, &o.ExpiresAt, &o.Verified, &o.Attempts, &o.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getOTPSession: %w", err)
	}
	return &o, nil
}

func (r *pgRepo) IncrementOTPAttempts(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE otp_sessions SET attempts = attempts + 1 WHERE id = $1`, id)
	return err
}

func (r *pgRepo) MarkOTPVerified(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE otp_sessions SET verified = TRUE WHERE id = $1`, id)
	return err
}

func (r *pgRepo) CreateRefreshToken(ctx context.Context, rt *RefreshToken) error {
	query := `INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, revoked, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	rt.ID = uuid.New()
	rt.CreatedAt = time.Now()
	_, err := r.pool.Exec(ctx, query,
		rt.ID, rt.UserID, rt.TokenHash, rt.ExpiresAt, rt.Revoked, rt.CreatedAt)
	return err
}

func (r *pgRepo) GetRefreshToken(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	query := `SELECT id, user_id, token_hash, expires_at, revoked, created_at
		FROM refresh_tokens WHERE token_hash = $1`

	var rt RefreshToken
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.Revoked, &rt.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getRefreshToken: %w", err)
	}
	return &rt, nil
}

func (r *pgRepo) RevokeRefreshToken(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE refresh_tokens SET revoked = TRUE WHERE id = $1`, id)
	return err
}

func (r *pgRepo) RevokeAllRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE refresh_tokens SET revoked = TRUE WHERE user_id = $1 AND revoked = FALSE`, userID)
	return err
}
