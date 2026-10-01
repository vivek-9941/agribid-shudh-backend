package returns

import (
	"time"

	"github.com/google/uuid"
)

// ReturnStatus represents valid return request states.
type ReturnStatus string

const (
	ReturnRequested ReturnStatus = "requested"
	ReturnApproved  ReturnStatus = "approved"
	ReturnRejected  ReturnStatus = "rejected"
	ReturnPickedUp  ReturnStatus = "picked_up"
	ReturnInspected ReturnStatus = "inspected"
	ReturnClosed    ReturnStatus = "closed"
)

// ReturnRequest represents a product return request.
type ReturnRequest struct {
	ID                uuid.UUID     `json:"id"`
	ReturnNumber      string        `json:"return_number"`
	OrderID           uuid.UUID     `json:"order_id"`
	InvoiceID         *uuid.UUID    `json:"invoice_id,omitempty"`
	BuyerID           uuid.UUID     `json:"buyer_id"`
	SellerID          uuid.UUID     `json:"seller_id"`
	Status            ReturnStatus  `json:"status"`
	Reason            string        `json:"reason"` // damaged, expired, wrong_product, quality_issue, other
	Notes             *string       `json:"notes,omitempty"`
	PickupScheduledAt *time.Time    `json:"pickup_scheduled_at,omitempty"`
	CreditNoteID      *uuid.UUID    `json:"credit_note_id,omitempty"`
	Lines             []*ReturnLine `json:"lines,omitempty"`
	RequestedAt       time.Time     `json:"requested_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
}

// ReturnLine represents a line item in a return request.
type ReturnLine struct {
	ID               uuid.UUID `json:"id"`
	ReturnID         uuid.UUID `json:"return_id"`
	OrderLineID      uuid.UUID `json:"order_line_id"`
	ProductID        uuid.UUID `json:"product_id"`
	RequestedQty     int       `json:"requested_qty"`
	ApprovedQty      *int      `json:"approved_qty,omitempty"`
	InspectionResult *string   `json:"inspection_result,omitempty"` // pass, fail, partial
	InspectionNotes  *string   `json:"inspection_notes,omitempty"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Request DTOs
// ─────────────────────────────────────────────────────────────────────────────

// InitiateReturnRequest is the payload for POST /returns.
type InitiateReturnRequest struct {
	OrderID string            `json:"order_id" validate:"required,uuid"`
	Reason  string            `json:"reason" validate:"required,oneof=damaged expired wrong_product quality_issue other"`
	Notes   string            `json:"notes,omitempty"`
	Lines   []ReturnLineInput `json:"lines" validate:"required,min=1,dive"`
}

// ReturnLineInput is a line item in a return request.
type ReturnLineInput struct {
	OrderLineID  string `json:"order_line_id" validate:"required,uuid"`
	RequestedQty int    `json:"requested_qty" validate:"required,min=1"`
}

// InspectionInput is the payload for POST /returns/:id/inspect.
type InspectionInput struct {
	Lines []InspectionLineInput `json:"lines" validate:"required,min=1,dive"`
}

// InspectionLineInput is a line-level inspection result.
type InspectionLineInput struct {
	ReturnLineID    string `json:"return_line_id" validate:"required,uuid"`
	Result          string `json:"result" validate:"required,oneof=pass fail partial"`
	InspectionNotes string `json:"inspection_notes,omitempty"`
}
