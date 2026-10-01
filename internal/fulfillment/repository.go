package fulfillment

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	ListIncomingOrders(ctx context.Context, sellerID uuid.UUID) ([]interface{}, error) // Interface simplified
	AddEvent(ctx context.Context, event *FulfillmentEvent) error
	RecordPartial(ctx context.Context, orderID uuid.UUID, partials []PartialFulfillment) error
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

func (r *pgRepo) ListIncomingOrders(ctx context.Context, sellerID uuid.UUID) ([]interface{}, error) {
	// Query to return orders for the seller... Simplified for stub.
	return nil, nil
}

func (r *pgRepo) AddEvent(ctx context.Context, event *FulfillmentEvent) error {
	query := `INSERT INTO order_status_history (id, order_id, from_status, to_status, changed_by, reason, changed_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`
	event.ID = uuid.New()
	event.ChangedAt = time.Now()
	_, err := r.conn(ctx).Exec(ctx, query, event.ID, event.OrderID, event.FromStatus, event.ToStatus, event.ChangedBy, event.Reason, event.ChangedAt)
	return err
}

func (r *pgRepo) RecordPartial(ctx context.Context, orderID uuid.UUID, partials []PartialFulfillment) error {
	for _, p := range partials {
		query := `UPDATE order_lines SET fulfilled_qty = fulfilled_qty + $1 WHERE id = $2 AND order_id = $3`
		_, err := r.conn(ctx).Exec(ctx, query, p.FulfilledQty, p.OrderLineID, orderID)
		if err != nil {
			return fmt.Errorf("record partial failed for line %s: %w", p.OrderLineID, err)
		}
	}
	return nil
}
