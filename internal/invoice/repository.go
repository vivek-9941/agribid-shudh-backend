package invoice

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, inv *Invoice) error
	GetByID(ctx context.Context, id uuid.UUID) (*Invoice, error)
	ListBySeller(ctx context.Context, sellerID uuid.UUID) ([]*Invoice, error)
	ListByBuyer(ctx context.Context, buyerID uuid.UUID) ([]*Invoice, error)
	GetGSTData(ctx context.Context, from, to time.Time, partnerID uuid.UUID) ([]interface{}, error) // simplified for now
	CreateCreditNote(ctx context.Context, cn *CreditNote) error
	UpdatePDFURL(ctx context.Context, id uuid.UUID, url string) error
	GetNextInvoiceNumber(ctx context.Context, sellerID uuid.UUID) (string, error)
}

type pgRepo struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepo{pool: pool}
}

func (r *pgRepo) conn(ctx context.Context) interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
} {
	return r.pool
}

func (r *pgRepo) Create(ctx context.Context, inv *Invoice) error {
	query := `INSERT INTO invoices (id, invoice_number, order_id, seller_id, buyer_id, invoice_type, parent_invoice_id, seller_gstin, buyer_gstin, place_of_supply, subtotal, discount_total, taxable_amount, cgst_total, sgst_total, igst_total, cess_total, grand_total, pdf_url, status, issued_at, due_date, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)`
	
	inv.ID = uuid.New()
	inv.CreatedAt = time.Now()
	if inv.Status == "" {
		inv.Status = "draft"
	}

	_, err := r.conn(ctx).Exec(ctx, query,
		inv.ID, inv.InvoiceNumber, inv.OrderID, inv.SellerID, inv.BuyerID, inv.InvoiceType, inv.ParentInvoiceID,
		inv.SellerGSTIN, inv.BuyerGSTIN, inv.PlaceOfSupply, inv.Subtotal, inv.DiscountTotal, inv.TaxableAmount,
		inv.CGSTTotal, inv.SGSTTotal, inv.IGSTTotal, inv.CessTotal, inv.GrandTotal, inv.PDFURL, inv.Status,
		inv.IssuedAt, inv.DueDate, inv.CreatedAt)
	if err != nil {
		return fmt.Errorf("Create invoice: %w", err)
	}

	for _, line := range inv.Lines {
		lineQuery := `INSERT INTO invoice_lines (id, invoice_id, product_id, sku, description, hsn_code, quantity, unit_price, discount_amount, taxable_amount, cgst_rate, cgst_amount, sgst_rate, sgst_amount, igst_rate, igst_amount, line_total)
					  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`
		line.ID = uuid.New()
		line.InvoiceID = inv.ID
		_, err := r.conn(ctx).Exec(ctx, lineQuery,
			line.ID, line.InvoiceID, line.ProductID, line.SKU, line.Description, line.HSNCode, line.Quantity,
			line.UnitPrice, line.DiscountAmount, line.TaxableAmount, line.CGSTRate, line.CGSTAmount,
			line.SGSTRate, line.SGSTAmount, line.IGSTRate, line.IGSTAmount, line.LineTotal)
		if err != nil {
			return fmt.Errorf("Create invoice line: %w", err)
		}
	}
	return nil
}

func (r *pgRepo) GetByID(ctx context.Context, id uuid.UUID) (*Invoice, error) {
	// Simple stub for GetByID to unblock compilation
	return nil, nil
}

func (r *pgRepo) ListBySeller(ctx context.Context, sellerID uuid.UUID) ([]*Invoice, error) {
	return nil, nil
}

func (r *pgRepo) ListByBuyer(ctx context.Context, buyerID uuid.UUID) ([]*Invoice, error) {
	return nil, nil
}

func (r *pgRepo) GetGSTData(ctx context.Context, from, to time.Time, partnerID uuid.UUID) ([]interface{}, error) {
	return nil, nil
}

func (r *pgRepo) CreateCreditNote(ctx context.Context, cn *CreditNote) error {
	return r.Create(ctx, cn)
}

func (r *pgRepo) UpdatePDFURL(ctx context.Context, id uuid.UUID, url string) error {
	query := `UPDATE invoices SET pdf_url = $1 WHERE id = $2`
	_, err := r.conn(ctx).Exec(ctx, query, url, id)
	return err
}

func (r *pgRepo) GetNextInvoiceNumber(ctx context.Context, sellerID uuid.UUID) (string, error) {
	// Simplified sequence implementation
	query := `INSERT INTO system_config (key, value) VALUES ($1, '1')
			  ON CONFLICT (key) DO UPDATE SET value = (system_config.value::int + 1)::text
			  RETURNING value`
	key := fmt.Sprintf("invoice_seq_%s", sellerID.String())
	var seq string
	err := r.conn(ctx).QueryRow(ctx, query, key).Scan(&seq)
	if err != nil {
		return "", fmt.Errorf("GetNextInvoiceNumber: %w", err)
	}
	return fmt.Sprintf("INV-%s-%s", sellerID.String()[:8], seq), nil
}
