package inventory

import (
	"time"

	"github.com/google/uuid"
)

// InventoryLedger represents stock levels for a product at a warehouse.
type InventoryLedger struct {
	ID           uuid.UUID `json:"id"`
	PartnerID    uuid.UUID `json:"partner_id"`
	WarehouseID  uuid.UUID `json:"warehouse_id"`
	ProductID    uuid.UUID `json:"product_id"`
	AvailableQty int       `json:"available_qty"`
	ReservedQty  int       `json:"reserved_qty"`
	PhysicalQty  int       `json:"physical_qty"`
	ReorderLevel int       `json:"reorder_level"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// InventoryMovement represents a stock movement event.
type InventoryMovement struct {
	ID            uuid.UUID  `json:"id"`
	PartnerID     uuid.UUID  `json:"partner_id"`
	WarehouseID   uuid.UUID  `json:"warehouse_id"`
	ProductID     uuid.UUID  `json:"product_id"`
	MovementType  string     `json:"movement_type"` // inward, outward, reservation, unreservation, adjustment, transfer
	Quantity      int        `json:"quantity"`
	ReferenceType string     `json:"reference_type,omitempty"`
	ReferenceID   *uuid.UUID `json:"reference_id,omitempty"`
	Reason        *string    `json:"reason,omitempty"`
	PerformedBy   *uuid.UUID `json:"performed_by,omitempty"`
	MovedAt       time.Time  `json:"moved_at"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Request DTOs
// ─────────────────────────────────────────────────────────────────────────────

// AdjustStockRequest is the payload for POST /inventory/adjust.
type AdjustStockRequest struct {
	WarehouseID string `json:"warehouse_id" validate:"required,uuid"`
	ProductID   string `json:"product_id" validate:"required,uuid"`
	Quantity    int    `json:"quantity" validate:"required"`
	Reason      string `json:"reason" validate:"required,min=2"`
}

// TransferStockRequest is the payload for POST /inventory/transfer.
type TransferStockRequest struct {
	FromWarehouseID string `json:"from_warehouse_id" validate:"required,uuid"`
	ToWarehouseID   string `json:"to_warehouse_id" validate:"required,uuid"`
	ProductID       string `json:"product_id" validate:"required,uuid"`
	Quantity        int    `json:"quantity" validate:"required,min=1"`
}

// CreateWarehouseRequest is the payload for POST /warehouses.
type CreateWarehouseRequest struct {
	Name      string                 `json:"name" validate:"required,min=2,max=255"`
	Address   map[string]interface{} `json:"address,omitempty"`
	IsDefault bool                   `json:"is_default"`
}
