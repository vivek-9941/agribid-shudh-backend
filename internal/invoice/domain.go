package invoice

import (
	"time"

	"github.com/google/uuid"
)

// Invoice represents a GST-compliant invoice.
type Invoice struct {
	ID              uuid.UUID      `json:"id"`
	InvoiceNumber   string         `json:"invoice_number"`
	OrderID         uuid.UUID      `json:"order_id"`
	SellerID        uuid.UUID      `json:"seller_id"`
	BuyerID         uuid.UUID      `json:"buyer_id"`
	InvoiceType     string         `json:"invoice_type"` // tax_invoice, credit_note, debit_note
	ParentInvoiceID *uuid.UUID     `json:"parent_invoice_id,omitempty"`
	SellerGSTIN     string         `json:"seller_gstin"`
	BuyerGSTIN      string         `json:"buyer_gstin"`
	PlaceOfSupply   string         `json:"place_of_supply"`
	Subtotal        string         `json:"subtotal"`
	DiscountTotal   string         `json:"discount_total"`
	TaxableAmount   string         `json:"taxable_amount"`
	CGSTTotal       string         `json:"cgst_total"`
	SGSTTotal       string         `json:"sgst_total"`
	IGSTTotal       string         `json:"igst_total"`
	CessTotal       string         `json:"cess_total"`
	GrandTotal      string         `json:"grand_total"`
	PDFURL          *string        `json:"pdf_url,omitempty"`
	Status          string         `json:"status"` // draft, issued, cancelled
	IssuedAt        *time.Time     `json:"issued_at,omitempty"`
	DueDate         *time.Time     `json:"due_date,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	Lines           []*InvoiceLine `json:"lines,omitempty"`
}

// InvoiceLine represents a line item in an invoice.
type InvoiceLine struct {
	ID             uuid.UUID `json:"id"`
	InvoiceID      uuid.UUID `json:"invoice_id"`
	ProductID      uuid.UUID `json:"product_id"`
	SKU            string    `json:"sku"`
	Description    string    `json:"description"`
	HSNCode        string    `json:"hsn_code"`
	Quantity       int       `json:"quantity"`
	UnitPrice      string    `json:"unit_price"`
	DiscountAmount string    `json:"discount_amount"`
	TaxableAmount  string    `json:"taxable_amount"`
	CGSTRate       string    `json:"cgst_rate"`
	CGSTAmount     string    `json:"cgst_amount"`
	SGSTRate       string    `json:"sgst_rate"`
	SGSTAmount     string    `json:"sgst_amount"`
	IGSTRate       string    `json:"igst_rate"`
	IGSTAmount     string    `json:"igst_amount"`
	LineTotal      string    `json:"line_total"`
}

// CreditNote is an alias for Invoice with type=credit_note.
type CreditNote = Invoice

// GSTType represents intra-state vs inter-state supply.
type GSTType string

const (
	GSTTypeIntra GSTType = "intra" // CGST + SGST
	GSTTypeInter GSTType = "inter" // IGST
)

// TaxBreakdown holds computed tax amounts for a line.
type TaxBreakdown struct {
	TaxableAmount string  `json:"taxable_amount"`
	CGSTRate      string  `json:"cgst_rate"`
	CGSTAmount    string  `json:"cgst_amount"`
	SGSTRate      string  `json:"sgst_rate"`
	SGSTAmount    string  `json:"sgst_amount"`
	IGSTRate      string  `json:"igst_rate"`
	IGSTAmount    string  `json:"igst_amount"`
	CessRate      string  `json:"cess_rate"`
	CessAmount    string  `json:"cess_amount"`
	TotalTax      string  `json:"total_tax"`
}

// HSNTaxRate holds the tax rates for an HSN code.
type HSNTaxRate struct {
	HSNCode       string `json:"hsn_code"`
	Description   string `json:"description"`
	CGSTRate      string `json:"cgst_rate"`
	SGSTRate      string `json:"sgst_rate"`
	IGSTRate      string `json:"igst_rate"`
	CessRate      string `json:"cess_rate"`
	EffectiveFrom string `json:"effective_from"`
}
