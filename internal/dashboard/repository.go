package dashboard

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	GetSalesSummary(ctx context.Context, partnerID *uuid.UUID) (map[string]interface{}, error)
	GetPendingOrders(ctx context.Context, partnerID *uuid.UUID) (int, error)
	GetLowStockCount(ctx context.Context, partnerID *uuid.UUID) (int, error)
	GetOutstandingPayments(ctx context.Context, partnerID *uuid.UUID) (string, error)
	GetTopProducts(ctx context.Context, partnerID *uuid.UUID, limit int) ([]map[string]interface{}, error)
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
} {
	return r.pool
}

func (r *pgRepo) GetSalesSummary(ctx context.Context, partnerID *uuid.UUID) (map[string]interface{}, error) {
	// Stub
	return map[string]interface{}{"total_sales": "10000.00", "order_count": 50}, nil
}

func (r *pgRepo) GetPendingOrders(ctx context.Context, partnerID *uuid.UUID) (int, error) {
	// Stub
	return 10, nil
}

func (r *pgRepo) GetLowStockCount(ctx context.Context, partnerID *uuid.UUID) (int, error) {
	// Stub
	return 5, nil
}

func (r *pgRepo) GetOutstandingPayments(ctx context.Context, partnerID *uuid.UUID) (string, error) {
	// Stub
	return "5000.00", nil
}

func (r *pgRepo) GetTopProducts(ctx context.Context, partnerID *uuid.UUID, limit int) ([]map[string]interface{}, error) {
	// Stub
	return []map[string]interface{}{
		{"product_id": uuid.New(), "name": "Product A", "sales": 100},
	}, nil
}
