package inventory

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines data access for inventory entities.
type Repository interface {
	// GetLedgerForUpdate acquires a row-level lock on the inventory ledger entry.
	// This MUST be called within a db.WithTx transaction.
	GetLedgerForUpdate(ctx context.Context, tx pgx.Tx, partnerID, warehouseID, productID uuid.UUID) (*InventoryLedger, error)

	UpsertLedger(ctx context.Context, tx pgx.Tx, l *InventoryLedger) error
	RecordMovement(ctx context.Context, tx pgx.Tx, m *InventoryMovement) error

	GetLedger(ctx context.Context, partnerID, warehouseID, productID uuid.UUID) (*InventoryLedger, error)
	ListLedger(ctx context.Context, partnerID uuid.UUID, warehouseID *uuid.UUID, page, pageSize int) ([]*InventoryLedger, int64, error)
	GetMovements(ctx context.Context, partnerID, productID uuid.UUID, page, pageSize int) ([]*InventoryMovement, int64, error)
	GetLowStockItems(ctx context.Context, partnerID uuid.UUID) ([]*InventoryLedger, error)
}

type pgRepo struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new pgx-backed inventory repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepo{pool: pool}
}

// GetLedgerForUpdate acquires SELECT FOR UPDATE lock — used inside db.WithTx.
func (r *pgRepo) GetLedgerForUpdate(ctx context.Context, tx pgx.Tx, partnerID, warehouseID, productID uuid.UUID) (*InventoryLedger, error) {
	query := `SELECT id, partner_id, warehouse_id, product_id, available_qty, reserved_qty, physical_qty, reorder_level, updated_at
		FROM inventory_ledger
		WHERE partner_id = $1 AND warehouse_id = $2 AND product_id = $3
		FOR UPDATE`

	var l InventoryLedger
	err := tx.QueryRow(ctx, query, partnerID, warehouseID, productID).Scan(
		&l.ID, &l.PartnerID, &l.WarehouseID, &l.ProductID,
		&l.AvailableQty, &l.ReservedQty, &l.PhysicalQty, &l.ReorderLevel, &l.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getLedgerForUpdate: %w", err)
	}
	return &l, nil
}

func (r *pgRepo) UpsertLedger(ctx context.Context, tx pgx.Tx, l *InventoryLedger) error {
	l.UpdatedAt = time.Now()
	query := `INSERT INTO inventory_ledger (id, partner_id, warehouse_id, product_id, available_qty, reserved_qty, physical_qty, reorder_level, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (partner_id, warehouse_id, product_id)
		DO UPDATE SET available_qty=$5, reserved_qty=$6, physical_qty=$7, reorder_level=$8, updated_at=$9`

	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}

	_, err := tx.Exec(ctx, query,
		l.ID, l.PartnerID, l.WarehouseID, l.ProductID,
		l.AvailableQty, l.ReservedQty, l.PhysicalQty, l.ReorderLevel, l.UpdatedAt)
	return err
}

func (r *pgRepo) RecordMovement(ctx context.Context, tx pgx.Tx, m *InventoryMovement) error {
	m.ID = uuid.New()
	m.MovedAt = time.Now()
	query := `INSERT INTO inventory_movements (id, partner_id, warehouse_id, product_id, movement_type, quantity, reference_type, reference_id, reason, performed_by, moved_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`
	_, err := tx.Exec(ctx, query,
		m.ID, m.PartnerID, m.WarehouseID, m.ProductID, m.MovementType,
		m.Quantity, m.ReferenceType, m.ReferenceID, m.Reason, m.PerformedBy, m.MovedAt)
	return err
}

func (r *pgRepo) GetLedger(ctx context.Context, partnerID, warehouseID, productID uuid.UUID) (*InventoryLedger, error) {
	query := `SELECT id, partner_id, warehouse_id, product_id, available_qty, reserved_qty, physical_qty, reorder_level, updated_at
		FROM inventory_ledger WHERE partner_id = $1 AND warehouse_id = $2 AND product_id = $3`
	var l InventoryLedger
	err := r.pool.QueryRow(ctx, query, partnerID, warehouseID, productID).Scan(
		&l.ID, &l.PartnerID, &l.WarehouseID, &l.ProductID,
		&l.AvailableQty, &l.ReservedQty, &l.PhysicalQty, &l.ReorderLevel, &l.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &l, nil
}

func (r *pgRepo) ListLedger(ctx context.Context, partnerID uuid.UUID, warehouseID *uuid.UUID, page, pageSize int) ([]*InventoryLedger, int64, error) {
	offset := (page - 1) * pageSize
	where := "WHERE partner_id = $1"
	args := []interface{}{partnerID}
	argIdx := 2

	if warehouseID != nil {
		where += fmt.Sprintf(" AND warehouse_id = $%d", argIdx)
		args = append(args, *warehouseID)
		argIdx++
	}

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM inventory_ledger %s", where)
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`SELECT id, partner_id, warehouse_id, product_id, available_qty, reserved_qty, physical_qty, reorder_level, updated_at
		FROM inventory_ledger %s ORDER BY updated_at DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var ledgers []*InventoryLedger
	for rows.Next() {
		var l InventoryLedger
		if err := rows.Scan(&l.ID, &l.PartnerID, &l.WarehouseID, &l.ProductID,
			&l.AvailableQty, &l.ReservedQty, &l.PhysicalQty, &l.ReorderLevel, &l.UpdatedAt); err != nil {
			return nil, 0, err
		}
		ledgers = append(ledgers, &l)
	}
	return ledgers, total, nil
}

func (r *pgRepo) GetMovements(ctx context.Context, partnerID, productID uuid.UUID, page, pageSize int) ([]*InventoryMovement, int64, error) {
	offset := (page - 1) * pageSize

	var total int64
	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM inventory_movements WHERE partner_id = $1 AND product_id = $2",
		partnerID, productID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, partner_id, warehouse_id, product_id, movement_type, quantity, reference_type, reference_id, reason, performed_by, moved_at
		FROM inventory_movements WHERE partner_id = $1 AND product_id = $2
		ORDER BY moved_at DESC LIMIT $3 OFFSET $4`

	rows, err := r.pool.Query(ctx, query, partnerID, productID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var movements []*InventoryMovement
	for rows.Next() {
		var m InventoryMovement
		if err := rows.Scan(&m.ID, &m.PartnerID, &m.WarehouseID, &m.ProductID, &m.MovementType,
			&m.Quantity, &m.ReferenceType, &m.ReferenceID, &m.Reason, &m.PerformedBy, &m.MovedAt); err != nil {
			return nil, 0, err
		}
		movements = append(movements, &m)
	}
	return movements, total, nil
}

func (r *pgRepo) GetLowStockItems(ctx context.Context, partnerID uuid.UUID) ([]*InventoryLedger, error) {
	query := `SELECT id, partner_id, warehouse_id, product_id, available_qty, reserved_qty, physical_qty, reorder_level, updated_at
		FROM inventory_ledger
		WHERE partner_id = $1 AND available_qty <= reorder_level AND reorder_level > 0
		ORDER BY available_qty ASC`

	rows, err := r.pool.Query(ctx, query, partnerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*InventoryLedger
	for rows.Next() {
		var l InventoryLedger
		if err := rows.Scan(&l.ID, &l.PartnerID, &l.WarehouseID, &l.ProductID,
			&l.AvailableQty, &l.ReservedQty, &l.PhysicalQty, &l.ReorderLevel, &l.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, &l)
	}
	return items, nil
}
