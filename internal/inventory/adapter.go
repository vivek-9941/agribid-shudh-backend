package inventory

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OrderInventoryAdapter bridges the order.InventoryService interface
// (which uses only productID + qty) to the full inventory.Service
// (which requires partnerID + warehouseID + productID + qty + orderID).
//
// It resolves the seller's default warehouse automatically.
type OrderInventoryAdapter struct {
	svc  *Service
	pool *pgxpool.Pool
}

// NewOrderInventoryAdapter creates a new adapter for use by the order module.
func NewOrderInventoryAdapter(svc *Service, pool *pgxpool.Pool) *OrderInventoryAdapter {
	return &OrderInventoryAdapter{svc: svc, pool: pool}
}

// Reserve implements order.InventoryService.Reserve.
// It looks up the seller's default warehouse and delegates to the full service.
// Note: the order module is expected to set the seller's partnerID in context
// via ContextWithSellerID / SellerIDFromContext before calling this.
func (a *OrderInventoryAdapter) Reserve(ctx context.Context, productID uuid.UUID, qty int) error {
	sellerID := SellerIDFromContext(ctx)
	if sellerID == uuid.Nil {
		return fmt.Errorf("inventory adapter: seller_id not found in context")
	}

	warehouseID, err := a.getDefaultWarehouse(ctx, sellerID)
	if err != nil {
		return fmt.Errorf("inventory adapter: resolve warehouse: %w", err)
	}

	// Use uuid.Nil as orderID placeholder — the order hasn't been created yet
	// when Reserve is called in the PlaceOrder transaction.
	return a.svc.ReserveStock(ctx, sellerID, warehouseID, productID, qty, uuid.Nil)
}

// FulfillmentInventoryAdapter bridges the fulfillment.InventoryService interface.
type FulfillmentInventoryAdapter struct {
	svc  *Service
	pool *pgxpool.Pool
}

// NewFulfillmentInventoryAdapter creates a new adapter for use by the fulfillment module.
func NewFulfillmentInventoryAdapter(svc *Service, pool *pgxpool.Pool) *FulfillmentInventoryAdapter {
	return &FulfillmentInventoryAdapter{svc: svc, pool: pool}
}

// Fulfill implements fulfillment.InventoryService.Fulfill.
// Deducts reserved stock on delivery completion.
func (a *FulfillmentInventoryAdapter) Fulfill(ctx context.Context, productID uuid.UUID, qty int) error {
	sellerID := SellerIDFromContext(ctx)
	if sellerID == uuid.Nil {
		return fmt.Errorf("inventory adapter: seller_id not found in context")
	}

	warehouseID, err := a.getDefaultWarehouse(ctx, sellerID)
	if err != nil {
		return fmt.Errorf("inventory adapter: resolve warehouse: %w", err)
	}

	return a.svc.DeductStock(ctx, sellerID, warehouseID, productID, qty, uuid.Nil)
}

// Release implements fulfillment.InventoryService.Release.
// Releases reserved stock back to available on order cancellation.
func (a *FulfillmentInventoryAdapter) Release(ctx context.Context, productID uuid.UUID, qty int) error {
	sellerID := SellerIDFromContext(ctx)
	if sellerID == uuid.Nil {
		return fmt.Errorf("inventory adapter: seller_id not found in context")
	}

	warehouseID, err := a.getDefaultWarehouse(ctx, sellerID)
	if err != nil {
		return fmt.Errorf("inventory adapter: resolve warehouse: %w", err)
	}

	return a.svc.ReleaseReservation(ctx, sellerID, warehouseID, productID, qty, uuid.Nil)
}

func (a *FulfillmentInventoryAdapter) getDefaultWarehouse(ctx context.Context, partnerID uuid.UUID) (uuid.UUID, error) {
	return getDefaultWarehouseFromDB(ctx, a.pool, partnerID)
}

// ReturnsInventoryAdapter bridges the returns.InventoryService interface.
type ReturnsInventoryAdapter struct {
	svc  *Service
	pool *pgxpool.Pool
}

// NewReturnsInventoryAdapter creates a new adapter for use by the returns module.
func NewReturnsInventoryAdapter(svc *Service, pool *pgxpool.Pool) *ReturnsInventoryAdapter {
	return &ReturnsInventoryAdapter{svc: svc, pool: pool}
}

// Adjust implements returns.InventoryService.Adjust.
// Restores inventory on return inspection pass.
func (a *ReturnsInventoryAdapter) Adjust(ctx context.Context, productID uuid.UUID, qty int, reason string) error {
	sellerID := SellerIDFromContext(ctx)
	if sellerID == uuid.Nil {
		return fmt.Errorf("inventory adapter: seller_id not found in context")
	}

	warehouseID, err := a.getDefaultWarehouse(ctx, sellerID)
	if err != nil {
		return fmt.Errorf("inventory adapter: resolve warehouse: %w", err)
	}

	return a.svc.AdjustStock(ctx, sellerID, warehouseID, productID, qty, reason, uuid.Nil)
}

func (a *ReturnsInventoryAdapter) getDefaultWarehouse(ctx context.Context, partnerID uuid.UUID) (uuid.UUID, error) {
	return getDefaultWarehouseFromDB(ctx, a.pool, partnerID)
}

// ─────────────────────────────────────────────────────────────────────────────
// Context helpers for passing seller ID through the call chain
// ─────────────────────────────────────────────────────────────────────────────

type ctxKeySellerID struct{}

// ContextWithSellerID injects a seller partner ID into the context.
func ContextWithSellerID(ctx context.Context, sellerID uuid.UUID) context.Context {
	return context.WithValue(ctx, ctxKeySellerID{}, sellerID)
}

// SellerIDFromContext extracts the seller partner ID from context.
func SellerIDFromContext(ctx context.Context) uuid.UUID {
	id, _ := ctx.Value(ctxKeySellerID{}).(uuid.UUID)
	return id
}

// ─────────────────────────────────────────────────────────────────────────────
// Shared warehouse resolution
// ─────────────────────────────────────────────────────────────────────────────

func (a *OrderInventoryAdapter) getDefaultWarehouse(ctx context.Context, partnerID uuid.UUID) (uuid.UUID, error) {
	return getDefaultWarehouseFromDB(ctx, a.pool, partnerID)
}

// getDefaultWarehouseFromDB finds the default warehouse for a partner.
// Falls back to the first warehouse if none is marked as default.
func getDefaultWarehouseFromDB(ctx context.Context, pool *pgxpool.Pool, partnerID uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID
	// Try default warehouse first
	err := pool.QueryRow(ctx,
		`SELECT id FROM warehouses WHERE partner_id = $1 AND is_default = TRUE LIMIT 1`,
		partnerID).Scan(&id)
	if err == nil {
		return id, nil
	}

	// Fallback: first warehouse
	err = pool.QueryRow(ctx,
		`SELECT id FROM warehouses WHERE partner_id = $1 ORDER BY created_at ASC LIMIT 1`,
		partnerID).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("no warehouse found for partner %s: %w", partnerID, err)
	}
	return id, nil
}
