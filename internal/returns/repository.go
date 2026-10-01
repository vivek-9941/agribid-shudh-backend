package returns

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, req *ReturnRequest) error
	GetByID(ctx context.Context, id uuid.UUID) (*ReturnRequest, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status ReturnStatus) error
	AddInspectionResult(ctx context.Context, lineID uuid.UUID, passedQty, failedQty int, reason string) error
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

func (r *pgRepo) Create(ctx context.Context, req *ReturnRequest) error {
	query := `INSERT INTO return_requests (id, return_number, order_id, buyer_id, seller_id, status, total_amount, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	
	req.ID = uuid.New()
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()
	if req.Status == "" {
		req.Status = StatusPending
	}
	
	_, err := r.conn(ctx).Exec(ctx, query,
		req.ID, req.ReturnNumber, req.OrderID, req.BuyerID, req.SellerID,
		req.Status, req.TotalAmount, req.CreatedAt, req.UpdatedAt)
	if err != nil {
		return fmt.Errorf("Create return request: %w", err)
	}

	for _, line := range req.Lines {
		lineQuery := `INSERT INTO return_lines (id, return_id, order_line_id, product_id, sku, requested_qty, approved_qty, refund_amount, reason, condition)
					  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
		line.ID = uuid.New()
		line.ReturnID = req.ID
		_, err := r.conn(ctx).Exec(ctx, lineQuery,
			line.ID, line.ReturnID, line.OrderLineID, line.ProductID, line.SKU,
			line.RequestedQty, line.ApprovedQty, line.RefundAmount, line.Reason, line.Condition)
		if err != nil {
			return fmt.Errorf("Create return line: %w", err)
		}
	}
	return nil
}

func (r *pgRepo) GetByID(ctx context.Context, id uuid.UUID) (*ReturnRequest, error) {
	query := `SELECT id, return_number, order_id, buyer_id, seller_id, status, total_amount, created_at, updated_at
			  FROM return_requests WHERE id = $1`
	var req ReturnRequest
	err := r.conn(ctx).QueryRow(ctx, query, id).Scan(
		&req.ID, &req.ReturnNumber, &req.OrderID, &req.BuyerID, &req.SellerID,
		&req.Status, &req.TotalAmount, &req.CreatedAt, &req.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("GetByID: %w", err)
	}

	lineQuery := `SELECT id, return_id, order_line_id, product_id, sku, requested_qty, approved_qty, refund_amount, reason, condition
				  FROM return_lines WHERE return_id = $1`
	rows, err := r.conn(ctx).Query(ctx, lineQuery, req.ID)
	if err != nil {
		return nil, fmt.Errorf("GetByID lines: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var l ReturnLine
		if err := rows.Scan(
			&l.ID, &l.ReturnID, &l.OrderLineID, &l.ProductID, &l.SKU,
			&l.RequestedQty, &l.ApprovedQty, &l.RefundAmount, &l.Reason, &l.Condition,
		); err != nil {
			return nil, err
		}
		req.Lines = append(req.Lines, &l)
	}

	return &req, nil
}

func (r *pgRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status ReturnStatus) error {
	query := `UPDATE return_requests SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.conn(ctx).Exec(ctx, query, status, time.Now(), id)
	return err
}

func (r *pgRepo) AddInspectionResult(ctx context.Context, lineID uuid.UUID, passedQty, failedQty int, reason string) error {
	query := `UPDATE return_lines SET approved_qty = $1, condition = $2 WHERE id = $3`
	// Simplified logic for stub, would normally have a separate inspection table
	_, err := r.conn(ctx).Exec(ctx, query, passedQty, reason, lineID)
	return err
}
