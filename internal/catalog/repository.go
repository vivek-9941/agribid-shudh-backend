package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines data access for catalog entities.
type Repository interface {
	CreateProduct(ctx context.Context, p *Product) error
	GetProductByID(ctx context.Context, id uuid.UUID) (*Product, error)
	GetProductBySKU(ctx context.Context, sku string) (*Product, error)
	UpdateProduct(ctx context.Context, p *Product) error
	ListProducts(ctx context.Context, filter ProductFilter) ([]*Product, int64, error)
	SetProductStatus(ctx context.Context, id uuid.UUID, status string) error

	CreateCategory(ctx context.Context, c *Category) error
	GetCategoryByID(ctx context.Context, id uuid.UUID) (*Category, error)
	ListCategories(ctx context.Context) ([]*Category, error)
	GetCategoryBySlug(ctx context.Context, slug string) (*Category, error)

	AddProductImage(ctx context.Context, img *ProductImage) error
	GetProductImages(ctx context.Context, productID uuid.UUID) ([]ProductImage, error)
	DeleteProductImage(ctx context.Context, imageID uuid.UUID) error
}

type pgRepo struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new pgx-backed catalog repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepo{pool: pool}
}

func (r *pgRepo) CreateProduct(ctx context.Context, p *Product) error {
	dims, _ := json.Marshal(p.Dimensions)
	variants, _ := json.Marshal(p.VariantAttrs)
	p.ID = uuid.New()
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now
	if p.Status == "" {
		p.Status = "active"
	}

	query := `INSERT INTO products (id, sku, name, description, category_id, brand, manufacturer_id, hsn_code, uom, weight_grams, dimensions, parent_sku_id, variant_attrs, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`
	_, err := r.pool.Exec(ctx, query,
		p.ID, p.SKU, p.Name, p.Description, p.CategoryID, p.Brand, p.ManufacturerID,
		p.HSNCode, p.UOM, p.WeightGrams, dims, p.ParentSKUID, variants,
		p.Status, p.CreatedAt, p.UpdatedAt)
	return err
}

func (r *pgRepo) GetProductByID(ctx context.Context, id uuid.UUID) (*Product, error) {
	query := `SELECT id, sku, name, description, category_id, brand, manufacturer_id, hsn_code, uom, weight_grams, dimensions, parent_sku_id, variant_attrs, status, created_at, updated_at, deleted_at
		FROM products WHERE id = $1 AND deleted_at IS NULL`
	return r.scanProduct(r.pool.QueryRow(ctx, query, id))
}

func (r *pgRepo) GetProductBySKU(ctx context.Context, sku string) (*Product, error) {
	query := `SELECT id, sku, name, description, category_id, brand, manufacturer_id, hsn_code, uom, weight_grams, dimensions, parent_sku_id, variant_attrs, status, created_at, updated_at, deleted_at
		FROM products WHERE sku = $1 AND deleted_at IS NULL`
	return r.scanProduct(r.pool.QueryRow(ctx, query, sku))
}

func (r *pgRepo) UpdateProduct(ctx context.Context, p *Product) error {
	dims, _ := json.Marshal(p.Dimensions)
	variants, _ := json.Marshal(p.VariantAttrs)
	p.UpdatedAt = time.Now()

	query := `UPDATE products SET name=$1, description=$2, category_id=$3, brand=$4, hsn_code=$5, uom=$6, weight_grams=$7, dimensions=$8, variant_attrs=$9, updated_at=$10
		WHERE id = $11 AND deleted_at IS NULL`
	_, err := r.pool.Exec(ctx, query,
		p.Name, p.Description, p.CategoryID, p.Brand, p.HSNCode, p.UOM,
		p.WeightGrams, dims, variants, p.UpdatedAt, p.ID)
	return err
}

func (r *pgRepo) ListProducts(ctx context.Context, filter ProductFilter) ([]*Product, int64, error) {
	where := "WHERE deleted_at IS NULL"
	args := []interface{}{}
	argIdx := 1

	if filter.CategoryID != nil {
		where += fmt.Sprintf(" AND category_id = $%d", argIdx)
		args = append(args, *filter.CategoryID)
		argIdx++
	}
	if filter.Brand != "" {
		where += fmt.Sprintf(" AND brand = $%d", argIdx)
		args = append(args, filter.Brand)
		argIdx++
	}
	if filter.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.ManufacturerID != nil {
		where += fmt.Sprintf(" AND manufacturer_id = $%d", argIdx)
		args = append(args, *filter.ManufacturerID)
		argIdx++
	}
	if filter.Search != "" {
		where += fmt.Sprintf(" AND (name ILIKE $%d OR sku ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}

	// Count.
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products %s", where)
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Paginate.
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	query := fmt.Sprintf(`SELECT id, sku, name, description, category_id, brand, manufacturer_id, hsn_code, uom, weight_grams, dimensions, parent_sku_id, variant_attrs, status, created_at, updated_at, deleted_at
		FROM products %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		p, err := r.scanProductRow(rows)
		if err != nil {
			return nil, 0, err
		}
		products = append(products, p)
	}
	return products, total, nil
}

func (r *pgRepo) SetProductStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE products SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, status, time.Now(), id)
	return err
}

func (r *pgRepo) CreateCategory(ctx context.Context, c *Category) error {
	c.ID = uuid.New()
	c.CreatedAt = time.Now()
	query := `INSERT INTO categories (id, parent_id, name, slug, path, level, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`
	_, err := r.pool.Exec(ctx, query, c.ID, c.ParentID, c.Name, c.Slug, c.Path, c.Level, c.CreatedAt)
	return err
}

func (r *pgRepo) GetCategoryByID(ctx context.Context, id uuid.UUID) (*Category, error) {
	query := `SELECT id, parent_id, name, slug, path, level, created_at
		FROM categories WHERE id = $1`
	var c Category
	err := r.pool.QueryRow(ctx, query, id).Scan(&c.ID, &c.ParentID, &c.Name, &c.Slug, &c.Path, &c.Level, &c.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *pgRepo) GetCategoryBySlug(ctx context.Context, slug string) (*Category, error) {
	query := `SELECT id, parent_id, name, slug, path, level, created_at
		FROM categories WHERE slug = $1`
	var c Category
	err := r.pool.QueryRow(ctx, query, slug).Scan(&c.ID, &c.ParentID, &c.Name, &c.Slug, &c.Path, &c.Level, &c.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *pgRepo) ListCategories(ctx context.Context) ([]*Category, error) {
	query := `SELECT id, parent_id, name, slug, path, level, created_at
		FROM categories ORDER BY path, level`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []*Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.ParentID, &c.Name, &c.Slug, &c.Path, &c.Level, &c.CreatedAt); err != nil {
			return nil, err
		}
		cats = append(cats, &c)
	}
	return cats, nil
}

func (r *pgRepo) AddProductImage(ctx context.Context, img *ProductImage) error {
	img.ID = uuid.New()
	query := `INSERT INTO product_images (id, product_id, url, sort_order, is_primary)
		VALUES ($1,$2,$3,$4,$5)`
	_, err := r.pool.Exec(ctx, query, img.ID, img.ProductID, img.URL, img.SortOrder, img.IsPrimary)
	return err
}

func (r *pgRepo) GetProductImages(ctx context.Context, productID uuid.UUID) ([]ProductImage, error) {
	query := `SELECT id, product_id, url, sort_order, is_primary
		FROM product_images WHERE product_id = $1 ORDER BY sort_order`
	rows, err := r.pool.Query(ctx, query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []ProductImage
	for rows.Next() {
		var img ProductImage
		if err := rows.Scan(&img.ID, &img.ProductID, &img.URL, &img.SortOrder, &img.IsPrimary); err != nil {
			return nil, err
		}
		images = append(images, img)
	}
	return images, nil
}

func (r *pgRepo) DeleteProductImage(ctx context.Context, imageID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM product_images WHERE id = $1`, imageID)
	return err
}

// ─────────────────────────────────────────────────────────────────────────────
// Scan helpers
// ─────────────────────────────────────────────────────────────────────────────

func (r *pgRepo) scanProduct(row pgx.Row) (*Product, error) {
	var p Product
	var dims, variants []byte
	err := row.Scan(&p.ID, &p.SKU, &p.Name, &p.Description, &p.CategoryID, &p.Brand,
		&p.ManufacturerID, &p.HSNCode, &p.UOM, &p.WeightGrams, &dims, &p.ParentSKUID,
		&variants, &p.Status, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if len(dims) > 0 {
		_ = json.Unmarshal(dims, &p.Dimensions)
	}
	if len(variants) > 0 {
		_ = json.Unmarshal(variants, &p.VariantAttrs)
	}
	return &p, nil
}

func (r *pgRepo) scanProductRow(rows pgx.Rows) (*Product, error) {
	var p Product
	var dims, variants []byte
	err := rows.Scan(&p.ID, &p.SKU, &p.Name, &p.Description, &p.CategoryID, &p.Brand,
		&p.ManufacturerID, &p.HSNCode, &p.UOM, &p.WeightGrams, &dims, &p.ParentSKUID,
		&variants, &p.Status, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt)
	if err != nil {
		return nil, err
	}
	if len(dims) > 0 {
		_ = json.Unmarshal(dims, &p.Dimensions)
	}
	if len(variants) > 0 {
		_ = json.Unmarshal(variants, &p.VariantAttrs)
	}
	return &p, nil
}
