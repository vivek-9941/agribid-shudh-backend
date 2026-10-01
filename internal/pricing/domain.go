package pricing

import (
	"time"

	"github.com/google/uuid"
)

// ProductPrice represents a role/partner-specific price for a product.
type ProductPrice struct {
	ID            uuid.UUID  `json:"id"`
	ProductID     uuid.UUID  `json:"product_id"`
	RoleCode      string     `json:"role_code"`
	PartnerID     *uuid.UUID `json:"partner_id,omitempty"`
	MRP           string     `json:"mrp"` // string to avoid float precision
	BasePrice     string     `json:"base_price"`
	EffectiveFrom time.Time  `json:"effective_from"`
	EffectiveTo   *time.Time `json:"effective_to,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// Scheme represents a pricing scheme / promotion.
type Scheme struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Type          string    `json:"type"` // percentage, flat, buy_x_get_y
	DiscountValue string    `json:"discount_value,omitempty"`
	BuyQty        *int      `json:"buy_qty,omitempty"`
	GetQty        *int      `json:"get_qty,omitempty"`
	ValidFrom     time.Time `json:"valid_from"`
	ValidTo       time.Time `json:"valid_to"`
	IsExclusive   bool      `json:"is_exclusive"`
	CreatedBy     uuid.UUID `json:"created_by"`
	Status        string    `json:"status"` // active, inactive
	CreatedAt     time.Time `json:"created_at"`
}

// SchemeApplicability defines which targets a scheme applies to.
type SchemeApplicability struct {
	ID         uuid.UUID `json:"id"`
	SchemeID   uuid.UUID `json:"scheme_id"`
	TargetType string    `json:"target_type"` // role, partner, category, product
	TargetID   string    `json:"target_id"`
}

// PriceResult holds the calculated price for an order line.
type PriceResult struct {
	BasePrice     string     `json:"base_price"`
	DiscountAmount string    `json:"discount_amount"`
	TaxableAmount string     `json:"taxable_amount"`
	CGSTRate      string     `json:"cgst_rate"`
	CGSTAmount    string     `json:"cgst_amount"`
	SGSTRate      string     `json:"sgst_rate"`
	SGSTAmount    string     `json:"sgst_amount"`
	IGSTRate      string     `json:"igst_rate"`
	IGSTAmount    string     `json:"igst_amount"`
	LineTotal     string     `json:"line_total"`
	AppliedSchemes []uuid.UUID `json:"applied_schemes,omitempty"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Request DTOs
// ─────────────────────────────────────────────────────────────────────────────

// SetPriceRequest is the payload for POST /products/:id/prices.
type SetPriceRequest struct {
	RoleCode      string `json:"role_code" validate:"required"`
	PartnerID     string `json:"partner_id,omitempty" validate:"omitempty,uuid"`
	MRP           string `json:"mrp" validate:"required"`
	BasePrice     string `json:"base_price" validate:"required"`
	EffectiveFrom string `json:"effective_from" validate:"required"`
	EffectiveTo   string `json:"effective_to,omitempty"`
}

// CreateSchemeRequest is the payload for POST /schemes.
type CreateSchemeRequest struct {
	Name          string `json:"name" validate:"required,min=2,max=255"`
	Type          string `json:"type" validate:"required,oneof=percentage flat buy_x_get_y"`
	DiscountValue string `json:"discount_value,omitempty"`
	BuyQty        *int   `json:"buy_qty,omitempty"`
	GetQty        *int   `json:"get_qty,omitempty"`
	ValidFrom     string `json:"valid_from" validate:"required"`
	ValidTo       string `json:"valid_to" validate:"required"`
	IsExclusive   bool   `json:"is_exclusive"`
	Targets       []SchemeTarget `json:"targets,omitempty"`
}

// SchemeTarget represents a target for scheme applicability.
type SchemeTarget struct {
	TargetType string `json:"target_type" validate:"required,oneof=role partner category product"`
	TargetID   string `json:"target_id" validate:"required"`
}

// UpdateSchemeRequest is the payload for PUT /schemes/:id.
type UpdateSchemeRequest struct {
	Name          string `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	DiscountValue string `json:"discount_value,omitempty"`
	ValidFrom     string `json:"valid_from,omitempty"`
	ValidTo       string `json:"valid_to,omitempty"`
	IsExclusive   *bool  `json:"is_exclusive,omitempty"`
}

// SetSchemeStatusRequest is the payload for PATCH /schemes/:id/status.
type SetSchemeStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=active inactive"`
}
