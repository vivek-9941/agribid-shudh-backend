package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	RecordPayment(ctx context.Context, p *Payment) error
	GetCreditAccount(ctx context.Context, buyerID, sellerID uuid.UUID) (*CreditAccount, error)
	UpdateCredit(ctx context.Context, buyerID, sellerID uuid.UUID, amountStr string) error
	InsertLedgerEntry(ctx context.Context, entry *LedgerEntry) error
	GetLedger(ctx context.Context, partnerID uuid.UUID, page, pageSize int) ([]*LedgerEntry, int64, error)
	GetAgingReport(ctx context.Context, buyerID uuid.UUID) ([]AgingBucket, error)
}

type pgRepo struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepo{pool: pool}
}


func (r *pgRepo) RecordPayment(ctx context.Context, p *Payment) error {
	query := `INSERT INTO payments (id, payment_number, buyer_id, seller_id, amount, method, reference_number, gateway_txn_id, invoice_id, status, paid_at, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	p.ID = uuid.New()
	p.CreatedAt = time.Now()
	if p.Status == "" {
		p.Status = "confirmed"
	}
	if p.Status == "confirmed" && p.PaidAt == nil {
		p.PaidAt = &p.CreatedAt
	}
	_, err := r.pool.Exec(ctx, query,
		p.ID, p.PaymentNumber, p.BuyerID, p.SellerID, p.Amount, p.Method,
		p.ReferenceNumber, p.GatewayTxnID, p.InvoiceID, p.Status, p.PaidAt, p.CreatedAt)
	return err
}

func (r *pgRepo) GetCreditAccount(ctx context.Context, buyerID, sellerID uuid.UUID) (*CreditAccount, error) {
	query := `SELECT id, buyer_id, seller_id, credit_limit, credit_utilized, payment_terms, overdue_amount, updated_at
			  FROM credit_accounts WHERE buyer_id = $1 AND seller_id = $2`
	var c CreditAccount
	err := r.pool.QueryRow(ctx, query, buyerID, sellerID).Scan(
		&c.ID, &c.BuyerID, &c.SellerID, &c.CreditLimit, &c.CreditUtilized,
		&c.PaymentTerms, &c.OverdueAmount, &c.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("GetCreditAccount: %w", err)
	}
	return &c, nil
}

func (r *pgRepo) UpdateCredit(ctx context.Context, buyerID, sellerID uuid.UUID, amountStr string) error {
	// For stub, just update utilized by amountStr by adding it in DB
	// We'd parse it using NUMERIC in PostgreSQL: credit_utilized = credit_utilized + $1
	query := `UPDATE credit_accounts SET credit_utilized = credit_utilized + $1::numeric, updated_at = $2
			  WHERE buyer_id = $3 AND seller_id = $4`
	_, err := r.pool.Exec(ctx, query, amountStr, time.Now(), buyerID, sellerID)
	return err
}

func (r *pgRepo) InsertLedgerEntry(ctx context.Context, entry *LedgerEntry) error {
	query := `INSERT INTO ledger_entries (id, partner_id, counterparty_id, entry_type, reference_id, debit, credit, balance, entry_date, description, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	entry.ID = uuid.New()
	entry.CreatedAt = time.Now()
	_, err := r.pool.Exec(ctx, query,
		entry.ID, entry.PartnerID, entry.CounterpartyID, entry.EntryType, entry.ReferenceID,
		entry.Debit, entry.Credit, entry.Balance, entry.EntryDate, entry.Description, entry.CreatedAt)
	return err
}

func (r *pgRepo) GetLedger(ctx context.Context, partnerID uuid.UUID, page, pageSize int) ([]*LedgerEntry, int64, error) {
	return nil, 0, nil
}

func (r *pgRepo) GetAgingReport(ctx context.Context, buyerID uuid.UUID) ([]AgingBucket, error) {
	// Stub
	return nil, nil
}
