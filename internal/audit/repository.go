package audit

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditLog struct {
	ID           uuid.UUID `json:"id"`
	UserID       *uuid.UUID `json:"user_id,omitempty"`
	Action       string    `json:"action"`
	Resource     string    `json:"resource"`
	ResourceID   string    `json:"resource_id,omitempty"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	StatusCode   int       `json:"status_code"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	RequestSize  int64     `json:"request_size"`
	ResponseSize int64     `json:"response_size"`
	DurationMs   int64     `json:"duration_ms"`
	RequestData  string    `json:"request_data,omitempty"`
	ResponseData string    `json:"response_data,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type Repository interface {
	Insert(ctx context.Context, log *AuditLog) error
	List(ctx context.Context, filter map[string]interface{}) ([]*AuditLog, error)
}

type pgRepo struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepo{pool: pool}
}

func (r *pgRepo) conn(ctx context.Context) interface {
	Exec(ctx context.Context, sql string, args ...interface{}) (interface{ RowsAffected() int64 }, error)
} {
	return r.pool
}

func (r *pgRepo) Insert(ctx context.Context, log *AuditLog) error {
	query := `INSERT INTO audit_logs (id, user_id, action, resource, resource_id, method, path, status_code, ip_address, user_agent, request_size, response_size, duration_ms, request_data, response_data, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`
	
	log.ID = uuid.New()
	log.CreatedAt = time.Now()
	
	_, err := r.conn(ctx).Exec(ctx, query,
		log.ID, log.UserID, log.Action, log.Resource, log.ResourceID,
		log.Method, log.Path, log.StatusCode, log.IPAddress, log.UserAgent,
		log.RequestSize, log.ResponseSize, log.DurationMs,
		log.RequestData, log.ResponseData, log.CreatedAt)
	return err
}

func (r *pgRepo) List(ctx context.Context, filter map[string]interface{}) ([]*AuditLog, error) {
	// Stub implementation
	return nil, nil
}
