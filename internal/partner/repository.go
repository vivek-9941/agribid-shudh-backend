package partner

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines data access for partner entities.
type Repository interface {
	Create(ctx context.Context, p *Partner) error
	GetByID(ctx context.Context, id uuid.UUID) (*Partner, error)
	Update(ctx context.Context, p *Partner) error
	List(ctx context.Context, partnerType string, page, pageSize int) ([]*Partner, int64, error)
	GetChildren(ctx context.Context, parentID uuid.UUID) ([]*Partner, error)

	SubmitKYC(ctx context.Context, doc *KYCDocument) error
	GetKYCDocuments(ctx context.Context, partnerID uuid.UUID) ([]*KYCDocument, error)
	UpdateKYCStatus(ctx context.Context, docID uuid.UUID, status string, reviewedBy uuid.UUID, note string) error
	UpdatePartnerKYCStatus(ctx context.Context, partnerID uuid.UUID, kycStatus, partnerStatus string) error

	CreateWarehouse(ctx context.Context, w *Warehouse) error
	GetWarehouses(ctx context.Context, partnerID uuid.UUID) ([]*Warehouse, error)
}

type pgRepo struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new pgx-backed partner repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepo{pool: pool}
}

func (r *pgRepo) Create(ctx context.Context, p *Partner) error {
	addressJSON, _ := json.Marshal(p.Address)
	p.ID = uuid.New()
	p.Code = generatePartnerCode(p.Type)
	p.Status = "pending"
	p.KYCStatus = "pending"
	p.CreatedAt = time.Now()
	p.UpdatedAt = p.CreatedAt

	query := `INSERT INTO partners (id, code, type, business_name, trade_name, parent_id, gstin, pan, state_code, address, status, kyc_status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`
	_, err := r.pool.Exec(ctx, query,
		p.ID, p.Code, p.Type, p.BusinessName, p.TradeName, p.ParentID,
		p.GSTIN, p.PAN, p.StateCode, addressJSON, p.Status, p.KYCStatus,
		p.CreatedAt, p.UpdatedAt)
	return err
}

func (r *pgRepo) GetByID(ctx context.Context, id uuid.UUID) (*Partner, error) {
	query := `SELECT id, code, type, business_name, trade_name, parent_id, gstin, pan, state_code, address, status, kyc_status, created_at, updated_at, deleted_at
		FROM partners WHERE id = $1 AND deleted_at IS NULL`
	return r.scanPartner(r.pool.QueryRow(ctx, query, id))
}

func (r *pgRepo) Update(ctx context.Context, p *Partner) error {
	addressJSON, _ := json.Marshal(p.Address)
	p.UpdatedAt = time.Now()
	query := `UPDATE partners SET business_name=$1, trade_name=$2, gstin=$3, pan=$4, state_code=$5, address=$6, updated_at=$7
		WHERE id = $8 AND deleted_at IS NULL`
	_, err := r.pool.Exec(ctx, query,
		p.BusinessName, p.TradeName, p.GSTIN, p.PAN, p.StateCode, addressJSON,
		p.UpdatedAt, p.ID)
	return err
}

func (r *pgRepo) List(ctx context.Context, partnerType string, page, pageSize int) ([]*Partner, int64, error) {
	offset := (page - 1) * pageSize
	where := "WHERE deleted_at IS NULL"
	args := []interface{}{}
	argIdx := 1

	if partnerType != "" {
		where += fmt.Sprintf(" AND type = $%d", argIdx)
		args = append(args, partnerType)
		argIdx++
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM partners %s", where)
	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`SELECT id, code, type, business_name, trade_name, parent_id, gstin, pan, state_code, address, status, kyc_status, created_at, updated_at, deleted_at
		FROM partners %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var partners []*Partner
	for rows.Next() {
		p, err := r.scanPartnerRow(rows)
		if err != nil {
			return nil, 0, err
		}
		partners = append(partners, p)
	}
	return partners, total, nil
}

func (r *pgRepo) GetChildren(ctx context.Context, parentID uuid.UUID) ([]*Partner, error) {
	query := `SELECT id, code, type, business_name, trade_name, parent_id, gstin, pan, state_code, address, status, kyc_status, created_at, updated_at, deleted_at
		FROM partners WHERE parent_id = $1 AND deleted_at IS NULL ORDER BY business_name`
	rows, err := r.pool.Query(ctx, query, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var partners []*Partner
	for rows.Next() {
		p, err := r.scanPartnerRow(rows)
		if err != nil {
			return nil, err
		}
		partners = append(partners, p)
	}
	return partners, nil
}

func (r *pgRepo) SubmitKYC(ctx context.Context, doc *KYCDocument) error {
	doc.ID = uuid.New()
	doc.Status = "pending"
	doc.SubmittedAt = time.Now()
	query := `INSERT INTO kyc_documents (id, partner_id, doc_type, file_url, status, submitted_at)
		VALUES ($1,$2,$3,$4,$5,$6)`
	_, err := r.pool.Exec(ctx, query, doc.ID, doc.PartnerID, doc.DocType, doc.FileURL, doc.Status, doc.SubmittedAt)
	return err
}

func (r *pgRepo) GetKYCDocuments(ctx context.Context, partnerID uuid.UUID) ([]*KYCDocument, error) {
	query := `SELECT id, partner_id, doc_type, file_url, status, reviewed_by, review_note, submitted_at, reviewed_at
		FROM kyc_documents WHERE partner_id = $1 ORDER BY submitted_at DESC`
	rows, err := r.pool.Query(ctx, query, partnerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*KYCDocument
	for rows.Next() {
		var d KYCDocument
		if err := rows.Scan(&d.ID, &d.PartnerID, &d.DocType, &d.FileURL, &d.Status,
			&d.ReviewedBy, &d.ReviewNote, &d.SubmittedAt, &d.ReviewedAt); err != nil {
			return nil, err
		}
		docs = append(docs, &d)
	}
	return docs, nil
}

func (r *pgRepo) UpdateKYCStatus(ctx context.Context, docID uuid.UUID, status string, reviewedBy uuid.UUID, note string) error {
	now := time.Now()
	query := `UPDATE kyc_documents SET status=$1, reviewed_by=$2, review_note=$3, reviewed_at=$4 WHERE id=$5`
	_, err := r.pool.Exec(ctx, query, status, reviewedBy, note, now, docID)
	return err
}

func (r *pgRepo) UpdatePartnerKYCStatus(ctx context.Context, partnerID uuid.UUID, kycStatus, partnerStatus string) error {
	query := `UPDATE partners SET kyc_status=$1, status=$2, updated_at=$3 WHERE id=$4`
	_, err := r.pool.Exec(ctx, query, kycStatus, partnerStatus, time.Now(), partnerID)
	return err
}

func (r *pgRepo) CreateWarehouse(ctx context.Context, w *Warehouse) error {
	addressJSON, _ := json.Marshal(w.Address)
	w.ID = uuid.New()
	w.CreatedAt = time.Now()
	query := `INSERT INTO warehouses (id, partner_id, name, address, is_default, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)`
	_, err := r.pool.Exec(ctx, query, w.ID, w.PartnerID, w.Name, addressJSON, w.IsDefault, w.CreatedAt)
	return err
}

func (r *pgRepo) GetWarehouses(ctx context.Context, partnerID uuid.UUID) ([]*Warehouse, error) {
	query := `SELECT id, partner_id, name, address, is_default, created_at
		FROM warehouses WHERE partner_id = $1`
	rows, err := r.pool.Query(ctx, query, partnerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var warehouses []*Warehouse
	for rows.Next() {
		var w Warehouse
		var addrJSON []byte
		if err := rows.Scan(&w.ID, &w.PartnerID, &w.Name, &addrJSON, &w.IsDefault, &w.CreatedAt); err != nil {
			return nil, err
		}
		if len(addrJSON) > 0 {
			var addr Address
			_ = json.Unmarshal(addrJSON, &addr)
			w.Address = &addr
		}
		warehouses = append(warehouses, &w)
	}
	return warehouses, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func (r *pgRepo) scanPartner(row pgx.Row) (*Partner, error) {
	var p Partner
	var addrJSON []byte
	err := row.Scan(&p.ID, &p.Code, &p.Type, &p.BusinessName, &p.TradeName,
		&p.ParentID, &p.GSTIN, &p.PAN, &p.StateCode, &addrJSON,
		&p.Status, &p.KYCStatus, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if len(addrJSON) > 0 {
		var addr Address
		_ = json.Unmarshal(addrJSON, &addr)
		p.Address = &addr
	}
	return &p, nil
}

func (r *pgRepo) scanPartnerRow(rows pgx.Rows) (*Partner, error) {
	var p Partner
	var addrJSON []byte
	err := rows.Scan(&p.ID, &p.Code, &p.Type, &p.BusinessName, &p.TradeName,
		&p.ParentID, &p.GSTIN, &p.PAN, &p.StateCode, &addrJSON,
		&p.Status, &p.KYCStatus, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt)
	if err != nil {
		return nil, err
	}
	if len(addrJSON) > 0 {
		var addr Address
		_ = json.Unmarshal(addrJSON, &addr)
		p.Address = &addr
	}
	return &p, nil
}

func generatePartnerCode(partnerType string) string {
	prefixes := map[string]string{
		"manufacturer":    "MFR",
		"state_stockist":  "SST",
		"distributor":     "DST",
		"sub_distributor": "SDT",
		"retailer":        "RTL",
		"delivery_partner": "DLV",
	}
	prefix := prefixes[partnerType]
	if prefix == "" {
		prefix = "PTR"
	}
	return fmt.Sprintf("%s-%s", prefix, uuid.New().String()[:8])
}
