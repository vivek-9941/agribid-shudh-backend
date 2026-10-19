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
	GetLastBalance(ctx context.Context, partnerID, counterpartyID uuid.UUID) (string, error)
	GetLedger(ctx context.Context, partnerID uuid.UUID, page, pageSize int) ([]*LedgerEntry, int64, error)
	GetAgingReport(ctx context.Context, buyerID uuid.UUID) ([]AgingBucket, error)
	CreateCreditAccount(ctx context.Context, account *CreditAccount) error
	UpdateCreditLimit(ctx context.Context, accountID uuid.UUID, newLimit string) error
	GetPaymentByID(ctx context.Context, id uuid.UUID) (*Payment, error)
	ListPayments(ctx context.Context, partnerID uuid.UUID, page, pageSize int) ([]*Payment, int64, error)
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

func (r *pgRepo) GetLastBalance(ctx context.Context, partnerID, counterpartyID uuid.UUID) (string, error) {
	var balance string
	err := r.pool.QueryRow(ctx,
		`SELECT balance FROM ledger_entries
		 WHERE partner_id = $1 AND counterparty_id = $2
		 ORDER BY created_at DESC LIMIT 1`,
		partnerID, counterpartyID).Scan(&balance)
	if err != nil {
		// If no entries exist, balance starts at zero
		if err == pgx.ErrNoRows {
			return "0.00", nil
		}
		return "0.00", fmt.Errorf("GetLastBalance: %w", err)
	}
	return balance, nil
}

func (r *pgRepo) GetLedger(ctx context.Context, partnerID uuid.UUID, page, pageSize int) ([]*LedgerEntry, int64, error) {
	offset := (page - 1) * pageSize

	var total int64
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM ledger_entries WHERE partner_id = $1`,
		partnerID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("GetLedger count: %w", err)
	}

	query := `SELECT id, partner_id, counterparty_id, entry_type, reference_id, debit, credit, balance, entry_date, description, created_at
			  FROM ledger_entries WHERE partner_id = $1
			  ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, partnerID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("GetLedger query: %w", err)
	}
	defer rows.Close()

	var entries []*LedgerEntry
	for rows.Next() {
		var e LedgerEntry
		if err := rows.Scan(
			&e.ID, &e.PartnerID, &e.CounterpartyID, &e.EntryType, &e.ReferenceID,
			&e.Debit, &e.Credit, &e.Balance, &e.EntryDate, &e.Description, &e.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		entries = append(entries, &e)
	}
	return entries, total, nil
}

func (r *pgRepo) GetAgingReport(ctx context.Context, buyerID uuid.UUID) ([]AgingBucket, error) {
	query := `
		SELECT
			CASE
				WHEN age(NOW(), i.issued_at) <= INTERVAL '30 days' THEN '0-30'
				WHEN age(NOW(), i.issued_at) <= INTERVAL '60 days' THEN '31-60'
				WHEN age(NOW(), i.issued_at) <= INTERVAL '90 days' THEN '61-90'
				ELSE '90+'
			END AS bucket,
			COALESCE(SUM(i.grand_total::NUMERIC), 0)::TEXT AS amount,
			COUNT(*)::INT AS count
		FROM invoices i
		WHERE i.buyer_id = $1
			AND i.status = 'issued'
		GROUP BY bucket
		ORDER BY bucket`

	rows, err := r.pool.Query(ctx, query, buyerID)
	if err != nil {
		return nil, fmt.Errorf("GetAgingReport: %w", err)
	}
	defer rows.Close()

	var buckets []AgingBucket
	for rows.Next() {
		var b AgingBucket
		if err := rows.Scan(&b.Bucket, &b.Amount, &b.Count); err != nil {
			return nil, err
		}
		buckets = append(buckets, b)
	}
	return buckets, nil
}

func (r *pgRepo) CreateCreditAccount(ctx context.Context, account *CreditAccount) error {
	account.ID = uuid.New()
	account.UpdatedAt = time.Now()
	query := `INSERT INTO credit_accounts (id, buyer_id, seller_id, credit_limit, credit_utilized, payment_terms, overdue_amount, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.pool.Exec(ctx, query,
		account.ID, account.BuyerID, account.SellerID, account.CreditLimit,
		account.CreditUtilized, account.PaymentTerms, account.OverdueAmount, account.UpdatedAt)
	return err
}

func (r *pgRepo) UpdateCreditLimit(ctx context.Context, accountID uuid.UUID, newLimit string) error {
	query := `UPDATE credit_accounts SET credit_limit = $1, updated_at = $2 WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, newLimit, time.Now(), accountID)
	return err
}

func (r *pgRepo) GetPaymentByID(ctx context.Context, id uuid.UUID) (*Payment, error) {
	query := `SELECT id, payment_number, buyer_id, seller_id, amount, method, reference_number, gateway_txn_id, invoice_id, status, paid_at, created_at
			  FROM payments WHERE id = $1`
	var p Payment
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.PaymentNumber, &p.BuyerID, &p.SellerID, &p.Amount, &p.Method,
		&p.ReferenceNumber, &p.GatewayTxnID, &p.InvoiceID, &p.Status, &p.PaidAt, &p.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("GetPaymentByID: %w", err)
	}
	return &p, nil
}

func (r *pgRepo) ListPayments(ctx context.Context, partnerID uuid.UUID, page, pageSize int) ([]*Payment, int64, error) {
	offset := (page - 1) * pageSize

	var total int64
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM payments WHERE buyer_id = $1 OR seller_id = $1`,
		partnerID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, payment_number, buyer_id, seller_id, amount, method, reference_number, gateway_txn_id, invoice_id, status, paid_at, created_at
			  FROM payments WHERE buyer_id = $1 OR seller_id = $1
			  ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, partnerID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var payments []*Payment
	for rows.Next() {
		var p Payment
		if err := rows.Scan(
			&p.ID, &p.PaymentNumber, &p.BuyerID, &p.SellerID, &p.Amount, &p.Method,
			&p.ReferenceNumber, &p.GatewayTxnID, &p.InvoiceID, &p.Status, &p.PaidAt, &p.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		payments = append(payments, &p)
	}
	return payments, total, nil
}
