package inventory

import (
	"context"
	"fmt"

	"github.com/agribid/agribid-shudh-backend/internal/db"
	apperrors "github.com/agribid/agribid-shudh-backend/internal/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Service provides inventory business logic.
// All stock mutations go through db.WithTx + SELECT FOR UPDATE to prevent races.
type Service struct {
	repo Repository
	pool *pgxpool.Pool
}

// NewService creates a new inventory service.
func NewService(repo Repository, pool *pgxpool.Pool) *Service {
	return &Service{repo: repo, pool: pool}
}

// AddStock adds inward inventory (e.g., goods received from manufacturer).
func (s *Service) AddStock(ctx context.Context, partnerID, warehouseID, productID uuid.UUID, qty int, performedBy uuid.UUID) error {
	if qty <= 0 {
		return apperrors.BadRequest(apperrors.CodeValidationFailed, "quantity must be positive")
	}

	return db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		ledger, err := s.repo.GetLedgerForUpdate(ctx, tx, partnerID, warehouseID, productID)
		if err != nil {
			return apperrors.Internal(err)
		}

		if ledger == nil {
			ledger = &InventoryLedger{
				PartnerID:   partnerID,
				WarehouseID: warehouseID,
				ProductID:   productID,
			}
		}

		ledger.AvailableQty += qty
		ledger.PhysicalQty += qty

		if err := s.repo.UpsertLedger(ctx, tx, ledger); err != nil {
			return apperrors.Internal(err)
		}

		movement := &InventoryMovement{
			PartnerID:    partnerID,
			WarehouseID:  warehouseID,
			ProductID:    productID,
			MovementType: "inward",
			Quantity:     qty,
			PerformedBy:  &performedBy,
		}
		return s.repo.RecordMovement(ctx, tx, movement)
	})
}

// ReserveStock reserves inventory for an order (moves from available to reserved).
func (s *Service) ReserveStock(ctx context.Context, partnerID, warehouseID, productID uuid.UUID, qty int, orderID uuid.UUID) error {
	if qty <= 0 {
		return apperrors.BadRequest(apperrors.CodeValidationFailed, "quantity must be positive")
	}

	return db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		ledger, err := s.repo.GetLedgerForUpdate(ctx, tx, partnerID, warehouseID, productID)
		if err != nil {
			return apperrors.Internal(err)
		}
		if ledger == nil {
			return apperrors.UnprocessableEntity(apperrors.CodeInsufficientStock, "no inventory found", nil)
		}

		if ledger.AvailableQty < qty {
			return apperrors.UnprocessableEntity(apperrors.CodeInsufficientStock,
				fmt.Sprintf("insufficient stock: available=%d, requested=%d", ledger.AvailableQty, qty),
				map[string]interface{}{
					"available": ledger.AvailableQty,
					"requested": qty,
				})
		}

		ledger.AvailableQty -= qty
		ledger.ReservedQty += qty

		if err := s.repo.UpsertLedger(ctx, tx, ledger); err != nil {
			return apperrors.Internal(err)
		}

		movement := &InventoryMovement{
			PartnerID:     partnerID,
			WarehouseID:   warehouseID,
			ProductID:     productID,
			MovementType:  "reservation",
			Quantity:      qty,
			ReferenceType: "order",
			ReferenceID:   &orderID,
		}
		return s.repo.RecordMovement(ctx, tx, movement)
	})
}

// ReleaseReservation releases reserved stock back to available (e.g., order cancelled).
func (s *Service) ReleaseReservation(ctx context.Context, partnerID, warehouseID, productID uuid.UUID, qty int, orderID uuid.UUID) error {
	return db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		ledger, err := s.repo.GetLedgerForUpdate(ctx, tx, partnerID, warehouseID, productID)
		if err != nil {
			return apperrors.Internal(err)
		}
		if ledger == nil {
			return apperrors.Internal(fmt.Errorf("no inventory ledger found for release"))
		}

		ledger.AvailableQty += qty
		ledger.ReservedQty -= qty
		if ledger.ReservedQty < 0 {
			ledger.ReservedQty = 0
		}

		if err := s.repo.UpsertLedger(ctx, tx, ledger); err != nil {
			return apperrors.Internal(err)
		}

		movement := &InventoryMovement{
			PartnerID:     partnerID,
			WarehouseID:   warehouseID,
			ProductID:     productID,
			MovementType:  "unreservation",
			Quantity:      qty,
			ReferenceType: "order",
			ReferenceID:   &orderID,
		}
		return s.repo.RecordMovement(ctx, tx, movement)
	})
}

// DeductStock deducts reserved stock on fulfillment (moves from reserved, decrements physical).
func (s *Service) DeductStock(ctx context.Context, partnerID, warehouseID, productID uuid.UUID, qty int, orderID uuid.UUID) error {
	return db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		ledger, err := s.repo.GetLedgerForUpdate(ctx, tx, partnerID, warehouseID, productID)
		if err != nil {
			return apperrors.Internal(err)
		}
		if ledger == nil {
			return apperrors.Internal(fmt.Errorf("no inventory ledger found for deduction"))
		}

		ledger.ReservedQty -= qty
		ledger.PhysicalQty -= qty
		if ledger.ReservedQty < 0 {
			ledger.ReservedQty = 0
		}

		if err := s.repo.UpsertLedger(ctx, tx, ledger); err != nil {
			return apperrors.Internal(err)
		}

		movement := &InventoryMovement{
			PartnerID:     partnerID,
			WarehouseID:   warehouseID,
			ProductID:     productID,
			MovementType:  "outward",
			Quantity:      qty,
			ReferenceType: "order",
			ReferenceID:   &orderID,
		}
		return s.repo.RecordMovement(ctx, tx, movement)
	})
}

// AdjustStock adjusts inventory (e.g., physical count reconciliation).
func (s *Service) AdjustStock(ctx context.Context, partnerID, warehouseID, productID uuid.UUID, qty int, reason string, performedBy uuid.UUID) error {
	return db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		ledger, err := s.repo.GetLedgerForUpdate(ctx, tx, partnerID, warehouseID, productID)
		if err != nil {
			return apperrors.Internal(err)
		}
		if ledger == nil {
			ledger = &InventoryLedger{
				PartnerID:   partnerID,
				WarehouseID: warehouseID,
				ProductID:   productID,
			}
		}

		ledger.AvailableQty += qty
		ledger.PhysicalQty += qty

		// Prevent negative stock.
		if ledger.AvailableQty < 0 {
			return apperrors.UnprocessableEntity(apperrors.CodeNegativeStock,
				"adjustment would result in negative available stock", nil)
		}

		if err := s.repo.UpsertLedger(ctx, tx, ledger); err != nil {
			return apperrors.Internal(err)
		}

		r := reason
		movement := &InventoryMovement{
			PartnerID:    partnerID,
			WarehouseID:  warehouseID,
			ProductID:    productID,
			MovementType: "adjustment",
			Quantity:     qty,
			Reason:       &r,
			PerformedBy:  &performedBy,
		}
		return s.repo.RecordMovement(ctx, tx, movement)
	})
}

// GetStock returns the current inventory level for a product at a warehouse.
func (s *Service) GetStock(ctx context.Context, partnerID, warehouseID, productID uuid.UUID) (*InventoryLedger, error) {
	l, err := s.repo.GetLedger(ctx, partnerID, warehouseID, productID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	return l, nil
}

// ListStock lists all inventory for a partner.
func (s *Service) ListStock(ctx context.Context, partnerID uuid.UUID, warehouseID *uuid.UUID, page, pageSize int) ([]*InventoryLedger, int64, error) {
	return s.repo.ListLedger(ctx, partnerID, warehouseID, page, pageSize)
}

// GetLowStockItems returns items below their reorder level.
func (s *Service) GetLowStockItems(ctx context.Context, partnerID uuid.UUID) ([]*InventoryLedger, error) {
	return s.repo.GetLowStockItems(ctx, partnerID)
}
