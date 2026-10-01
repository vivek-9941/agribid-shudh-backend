package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, notif *Notification) error
	ListForUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*Notification, int64, error)
	MarkRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	GetPreferences(ctx context.Context, userID uuid.UUID) (*NotificationPreference, error)
	UpsertPreferences(ctx context.Context, pref *NotificationPreference) error
}

type pgRepo struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepo{pool: pool}
}

func (r *pgRepo) conn(ctx context.Context) interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...interface{}) (interface{ RowsAffected() int64 }, error)
} {
	return r.pool
}

func (r *pgRepo) Create(ctx context.Context, notif *Notification) error {
	query := `INSERT INTO notifications (id, recipient_id, type, title, body, data_payload, is_read, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	notif.ID = uuid.New()
	notif.CreatedAt = time.Now()
	_, err := r.conn(ctx).Exec(ctx, query,
		notif.ID, notif.RecipientID, notif.Type, notif.Title, notif.Body, notif.DataPayload, notif.IsRead, notif.CreatedAt)
	return err
}

func (r *pgRepo) ListForUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*Notification, int64, error) {
	// Stub implementation
	return nil, 0, nil
}

func (r *pgRepo) MarkRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	query := `UPDATE notifications SET is_read = TRUE WHERE id = $1 AND recipient_id = $2`
	_, err := r.conn(ctx).Exec(ctx, query, id, userID)
	return err
}

func (r *pgRepo) GetPreferences(ctx context.Context, userID uuid.UUID) (*NotificationPreference, error) {
	query := `SELECT id, user_id, email_enabled, sms_enabled, in_app_enabled, muted_events, updated_at
			  FROM notification_preferences WHERE user_id = $1`
	var p NotificationPreference
	err := r.conn(ctx).QueryRow(ctx, query, userID).Scan(
		&p.ID, &p.UserID, &p.EmailEnabled, &p.SMSEnabled, &p.InAppEnabled, &p.MutedEvents, &p.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Default preferences
			return &NotificationPreference{
				UserID:       userID,
				EmailEnabled: true,
				SMSEnabled:   true,
				InAppEnabled: true,
				MutedEvents:  []string{},
			}, nil
		}
		return nil, fmt.Errorf("GetPreferences: %w", err)
	}
	return &p, nil
}

func (r *pgRepo) UpsertPreferences(ctx context.Context, pref *NotificationPreference) error {
	query := `INSERT INTO notification_preferences (id, user_id, email_enabled, sms_enabled, in_app_enabled, muted_events, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7)
			  ON CONFLICT (user_id) DO UPDATE SET
			  email_enabled = EXCLUDED.email_enabled,
			  sms_enabled = EXCLUDED.sms_enabled,
			  in_app_enabled = EXCLUDED.in_app_enabled,
			  muted_events = EXCLUDED.muted_events,
			  updated_at = EXCLUDED.updated_at`
	if pref.ID == uuid.Nil {
		pref.ID = uuid.New()
	}
	pref.UpdatedAt = time.Now()
	_, err := r.conn(ctx).Exec(ctx, query,
		pref.ID, pref.UserID, pref.EmailEnabled, pref.SMSEnabled, pref.InAppEnabled, pref.MutedEvents, pref.UpdatedAt)
	return err
}
