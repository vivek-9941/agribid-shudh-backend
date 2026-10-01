package partner

import (
	"time"

	"github.com/google/uuid"
)

// Partner represents a business entity in the supply chain hierarchy.
type Partner struct {
	ID           uuid.UUID  `json:"id"`
	Code         string     `json:"code"`
	Type         string     `json:"type"` // manufacturer, state_stockist, distributor, sub_distributor, retailer, delivery_partner
	BusinessName string     `json:"business_name"`
	TradeName    *string    `json:"trade_name,omitempty"`
	ParentID     *uuid.UUID `json:"parent_id,omitempty"`
	GSTIN        *string    `json:"gstin,omitempty"`
	PAN          *string    `json:"pan,omitempty"`
	StateCode    *string    `json:"state_code,omitempty"`
	Address      *Address   `json:"address,omitempty"`
	Status       string     `json:"status"` // pending, active, inactive, suspended
	KYCStatus    string     `json:"kyc_status"` // pending, submitted, approved, rejected
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

// Address holds structured address fields.
type Address struct {
	Line1   string `json:"line1"`
	Line2   string `json:"line2,omitempty"`
	City    string `json:"city"`
	State   string `json:"state"`
	Pincode string `json:"pincode"`
}

// KYCDocument represents a submitted KYC document.
type KYCDocument struct {
	ID          uuid.UUID  `json:"id"`
	PartnerID   uuid.UUID  `json:"partner_id"`
	DocType     string     `json:"doc_type"` // gst_certificate, pan_card, business_license, bank_details, address_proof
	FileURL     string     `json:"file_url"`
	Status      string     `json:"status"` // pending, approved, rejected
	ReviewedBy  *uuid.UUID `json:"reviewed_by,omitempty"`
	ReviewNote  *string    `json:"review_note,omitempty"`
	SubmittedAt time.Time  `json:"submitted_at"`
	ReviewedAt  *time.Time `json:"reviewed_at,omitempty"`
}

// Warehouse represents a partner's warehouse.
type Warehouse struct {
	ID        uuid.UUID `json:"id"`
	PartnerID uuid.UUID `json:"partner_id"`
	Name      string    `json:"name"`
	Address   *Address  `json:"address,omitempty"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Request DTOs
// ─────────────────────────────────────────────────────────────────────────────

// CreatePartnerRequest is the payload for POST /partners.
type CreatePartnerRequest struct {
	Type         string   `json:"type" validate:"required,oneof=manufacturer state_stockist distributor sub_distributor retailer delivery_partner"`
	BusinessName string   `json:"business_name" validate:"required,min=2,max=255"`
	TradeName    string   `json:"trade_name,omitempty"`
	ParentID     string   `json:"parent_id,omitempty" validate:"omitempty,uuid"`
	GSTIN        string   `json:"gstin,omitempty" validate:"omitempty,len=15"`
	PAN          string   `json:"pan,omitempty" validate:"omitempty,len=10"`
	StateCode    string   `json:"state_code,omitempty" validate:"omitempty,len=2"`
	Address      *Address `json:"address,omitempty"`
}

// UpdatePartnerRequest is the payload for PUT /partners/:id.
type UpdatePartnerRequest struct {
	BusinessName string   `json:"business_name,omitempty" validate:"omitempty,min=2,max=255"`
	TradeName    string   `json:"trade_name,omitempty"`
	GSTIN        string   `json:"gstin,omitempty" validate:"omitempty,len=15"`
	PAN          string   `json:"pan,omitempty" validate:"omitempty,len=10"`
	StateCode    string   `json:"state_code,omitempty" validate:"omitempty,len=2"`
	Address      *Address `json:"address,omitempty"`
}

// SubmitKYCRequest is the payload for POST /partners/:id/kyc.
type SubmitKYCRequest struct {
	DocType string `json:"doc_type" validate:"required,oneof=gst_certificate pan_card business_license bank_details address_proof"`
	FileURL string `json:"file_url" validate:"required,url"`
}

// ReviewKYCRequest is the payload for PATCH /partners/:id/kyc.
type ReviewKYCRequest struct {
	Status     string `json:"status" validate:"required,oneof=approved rejected"`
	ReviewNote string `json:"review_note,omitempty"`
}
