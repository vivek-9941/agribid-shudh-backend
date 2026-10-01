package pricing

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines data access for pricing and schemes.
type Repository interface {
	GetPriceForRole(ctx context.Context, productID uuid.UUID, partnerID *uuid.UUID, roleCode string, date time.Time) (*ProductPrice, error)
	GetApplicableSchemes(ctx context.Context, productID uuid.UUID, partnerID uuid.UUID, roleCode string, date time.Time) ([]Scheme, error)
	
	SetProductPrice(ctx context.Context, price *ProductPrice) error
	CreateScheme(ctx context.Context, scheme *Scheme) error
	AddSchemeApplicability(ctx context.Context, app *SchemeApplicability) error
	UpdateScheme(ctx context.Context, scheme *Scheme) error
	SetSchemeStatus(ctx context.Context, id uuid.UUID, status string) error
	GetSchemeByID(ctx context.Context, id uuid.UUID) (*Scheme, error)
}

type pgRepo struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepo{pool: pool}
}


func (r *pgRepo) GetPriceForRole(ctx context.Context, productID uuid.UUID, partnerID *uuid.UUID, roleCode string, date time.Time) (*ProductPrice, error) {
	// Query to find the most specific applicable price (partner-specific first, then role-specific).
	query := `
		SELECT id, product_id, role_code, partner_id, mrp, base_price, effective_from, effective_to, created_at
		FROM product_prices
		WHERE product_id = $1
		  AND (role_code = $2 OR role_code = 'ALL')
		  AND effective_from <= $3
		  AND (effective_to IS NULL OR effective_to >= $3)
		  AND (partner_id IS NULL OR partner_id = $4)
		ORDER BY partner_id NULLS LAST, effective_from DESC
		LIMIT 1
	`
	
	var p ProductPrice
	err := r.pool.QueryRow(ctx, query, productID, roleCode, date, partnerID).Scan(
		&p.ID, &p.ProductID, &p.RoleCode, &p.PartnerID, &p.MRP, &p.BasePrice,
		&p.EffectiveFrom, &p.EffectiveTo, &p.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // No price found
		}
		return nil, fmt.Errorf("GetPriceForRole: %w", err)
	}
	return &p, nil
}

func (r *pgRepo) GetApplicableSchemes(ctx context.Context, productID uuid.UUID, partnerID uuid.UUID, roleCode string, date time.Time) ([]Scheme, error) {
	// This is a simplified query; in reality, we'd need to match against category hierarchy too.
	query := `
		SELECT s.id, s.name, s.type, s.discount_value, s.buy_qty, s.get_qty, s.valid_from, s.valid_to, s.is_exclusive, s.created_by, s.status, s.created_at
		FROM schemes s
		JOIN scheme_applicability sa ON sa.scheme_id = s.id
		WHERE s.status = 'active'
		  AND s.valid_from <= $1
		  AND s.valid_to >= $1
		  AND (
			  (sa.target_type = 'role' AND sa.target_id = $2)
			  OR (sa.target_type = 'partner' AND sa.target_id = $3)
			  OR (sa.target_type = 'product' AND sa.target_id = $4)
			  -- Category logic can be added here or resolved in service
		  )
	`
	rows, err := r.pool.Query(ctx, query, date, roleCode, partnerID.String(), productID.String())
	if err != nil {
		return nil, fmt.Errorf("GetApplicableSchemes: %w", err)
	}
	defer rows.Close()

	var schemes []Scheme
	for rows.Next() {
		var s Scheme
		if err := rows.Scan(
			&s.ID, &s.Name, &s.Type, &s.DiscountValue, &s.BuyQty, &s.GetQty,
			&s.ValidFrom, &s.ValidTo, &s.IsExclusive, &s.CreatedBy, &s.Status, &s.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("GetApplicableSchemes row scan: %w", err)
		}
		schemes = append(schemes, s)
	}
	return schemes, nil
}

func (r *pgRepo) SetProductPrice(ctx context.Context, price *ProductPrice) error {
	query := `INSERT INTO product_prices (id, product_id, role_code, partner_id, mrp, base_price, effective_from, effective_to, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	price.ID = uuid.New()
	price.CreatedAt = time.Now()
	_, err := r.pool.Exec(ctx, query, price.ID, price.ProductID, price.RoleCode, price.PartnerID, price.MRP, price.BasePrice, price.EffectiveFrom, price.EffectiveTo, price.CreatedAt)
	return err
}

func (r *pgRepo) CreateScheme(ctx context.Context, scheme *Scheme) error {
	query := `INSERT INTO schemes (id, name, type, discount_value, buy_qty, get_qty, valid_from, valid_to, is_exclusive, created_by, status, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	scheme.ID = uuid.New()
	scheme.CreatedAt = time.Now()
	if scheme.Status == "" {
		scheme.Status = "active"
	}
	_, err := r.pool.Exec(ctx, query, scheme.ID, scheme.Name, scheme.Type, scheme.DiscountValue, scheme.BuyQty, scheme.GetQty, scheme.ValidFrom, scheme.ValidTo, scheme.IsExclusive, scheme.CreatedBy, scheme.Status, scheme.CreatedAt)
	return err
}

func (r *pgRepo) AddSchemeApplicability(ctx context.Context, app *SchemeApplicability) error {
	query := `INSERT INTO scheme_applicability (id, scheme_id, target_type, target_id) VALUES ($1, $2, $3, $4)`
	app.ID = uuid.New()
	_, err := r.pool.Exec(ctx, query, app.ID, app.SchemeID, app.TargetType, app.TargetID)
	return err
}

func (r *pgRepo) UpdateScheme(ctx context.Context, scheme *Scheme) error {
	query := `UPDATE schemes SET name = $1, discount_value = $2, valid_from = $3, valid_to = $4, is_exclusive = $5 WHERE id = $6`
	_, err := r.pool.Exec(ctx, query, scheme.Name, scheme.DiscountValue, scheme.ValidFrom, scheme.ValidTo, scheme.IsExclusive, scheme.ID)
	return err
}

func (r *pgRepo) SetSchemeStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE schemes SET status = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, status, id)
	return err
}

func (r *pgRepo) GetSchemeByID(ctx context.Context, id uuid.UUID) (*Scheme, error) {
	query := `SELECT id, name, type, discount_value, buy_qty, get_qty, valid_from, valid_to, is_exclusive, created_by, status, created_at FROM schemes WHERE id = $1`
	var s Scheme
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.Name, &s.Type, &s.DiscountValue, &s.BuyQty, &s.GetQty,
		&s.ValidFrom, &s.ValidTo, &s.IsExclusive, &s.CreatedBy, &s.Status, &s.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("GetSchemeByID: %w", err)
	}
	return &s, nil
}
