package payment

import (
	"time"

	"github.com/google/uuid"
)

// Payment represents a recorded payment.
type Payment struct {
	ID              uuid.UUID  `json:"id"`
	PaymentNumber   string     `json:"payment_number"`
	BuyerID         uuid.UUID  `json:"buyer_id"`
	SellerID        uuid.UUID  `json:"seller_id"`
	Amount          string     `json:"amount"`
	Method          string     `json:"method"` // cash, bank_transfer, cheque, upi, card, gateway
	ReferenceNumber *string    `json:"reference_number,omitempty"`
	GatewayTxnID    *string    `json:"gateway_txn_id,omitempty"`
	InvoiceID       *uuid.UUID `json:"invoice_id,omitempty"`
	Status          string     `json:"status"` // pending, confirmed, failed, refunded
	PaidAt          *time.Time `json:"paid_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// CreditAccount represents credit terms between buyer and seller.
type CreditAccount struct {
	ID             uuid.UUID `json:"id"`
	BuyerID        uuid.UUID `json:"buyer_id"`
	SellerID       uuid.UUID `json:"seller_id"`
	CreditLimit    string    `json:"credit_limit"`
	CreditUtilized string    `json:"credit_utilized"`
	CreditAvailable string   `json:"credit_available"` // computed
	PaymentTerms   int       `json:"payment_terms"`
	OverdueAmount  string    `json:"overdue_amount"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// LedgerEntry represents an entry in the partner ledger.
type LedgerEntry struct {
	ID             uuid.UUID  `json:"id"`
	PartnerID      uuid.UUID  `json:"partner_id"`
	CounterpartyID uuid.UUID  `json:"counterparty_id"`
	EntryType      string     `json:"entry_type"` // invoice, payment, credit_note, adjustment
	ReferenceID    *uuid.UUID `json:"reference_id,omitempty"`
	Debit          string     `json:"debit"`
	Credit         string     `json:"credit"`
	Balance        string     `json:"balance"`
	EntryDate      time.Time  `json:"entry_date"`
	Description    string     `json:"description"`
	CreatedAt      time.Time  `json:"created_at"`
}

// AgingBucket holds an aging report bucket.
type AgingBucket struct {
	Bucket   string `json:"bucket"` // 0-30, 31-60, 61-90, 90+
	Amount   string `json:"amount"`
	Count    int    `json:"count"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Request DTOs
// ─────────────────────────────────────────────────────────────────────────────

// RecordPaymentRequest is the payload for POST /payments.
type RecordPaymentRequest struct {
	SellerID        string `json:"seller_id" validate:"required,uuid"`
	Amount          string `json:"amount" validate:"required"`
	Method          string `json:"method" validate:"required,oneof=cash bank_transfer cheque upi card gateway"`
	ReferenceNumber string `json:"reference_number,omitempty"`
	InvoiceID       string `json:"invoice_id,omitempty" validate:"omitempty,uuid"`
}
