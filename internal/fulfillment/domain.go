package fulfillment

import (
	"time"

	"github.com/google/uuid"
)

// FulfillmentEvent represents a fulfillment status update event.
type FulfillmentEvent struct {
	ID         uuid.UUID `json:"id"`
	OrderID    uuid.UUID `json:"order_id"`
	FromStatus string    `json:"from_status"`
	ToStatus   string    `json:"to_status"`
	ChangedBy  uuid.UUID `json:"changed_by"`
	Reason     *string   `json:"reason,omitempty"`
	ChangedAt  time.Time `json:"changed_at"`
}

// PartialFulfillment records a partial fulfillment for an order.
type PartialFulfillment struct {
	OrderLineID  uuid.UUID `json:"order_line_id"`
	FulfilledQty int       `json:"fulfilled_qty"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Request DTOs
// ─────────────────────────────────────────────────────────────────────────────

// RecordPartialRequest is the payload for POST /fulfillment/orders/:id/partial.
type RecordPartialRequest struct {
	Lines []PartialLineInput `json:"lines" validate:"required,min=1,dive"`
}

// PartialLineInput is a line-level fulfillment quantity.
type PartialLineInput struct {
	OrderLineID  string `json:"order_line_id" validate:"required,uuid"`
	FulfilledQty int    `json:"fulfilled_qty" validate:"required,min=1"`
}
