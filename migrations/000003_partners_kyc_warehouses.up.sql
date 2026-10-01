-- Migration 000003: Partners, KYC Documents, Warehouses
-- Creates: partners, kyc_documents, warehouses
-- Adds FK: users.partner_id → partners.id

-- ── partners ────────────────────────────────────────────────────────────
CREATE TABLE partners (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    type VARCHAR(30) NOT NULL
        CHECK (type IN ('manufacturer', 'state_stockist', 'distributor', 'sub_distributor', 'retailer', 'delivery_partner')),
    business_name VARCHAR(255) NOT NULL,
    trade_name VARCHAR(255),
    parent_id UUID REFERENCES partners(id),
    gstin VARCHAR(15),
    pan VARCHAR(10),
    state_code VARCHAR(2),
    address JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'active', 'inactive', 'suspended')),
    kyc_status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (kyc_status IN ('pending', 'submitted', 'approved', 'rejected')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    -- Prevent circular self-reference
    CONSTRAINT chk_no_self_parent CHECK (id != parent_id)
);

CREATE INDEX idx_partners_parent_id ON partners(parent_id);
CREATE INDEX idx_partners_type ON partners(type);
CREATE INDEX idx_partners_status ON partners(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_partners_gstin ON partners(gstin) WHERE gstin IS NOT NULL;

-- Add FK from users.partner_id → partners.id
ALTER TABLE users ADD CONSTRAINT fk_users_partner_id
    FOREIGN KEY (partner_id) REFERENCES partners(id);

-- ── kyc_documents ───────────────────────────────────────────────────────
CREATE TABLE kyc_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partner_id UUID NOT NULL REFERENCES partners(id) ON DELETE CASCADE,
    doc_type VARCHAR(30) NOT NULL
        CHECK (doc_type IN ('gst_certificate', 'pan_card', 'business_license', 'bank_details', 'address_proof')),
    file_url VARCHAR(500) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'approved', 'rejected')),
    reviewed_by UUID REFERENCES users(id),
    review_note TEXT,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at TIMESTAMPTZ
);

CREATE INDEX idx_kyc_documents_partner_id ON kyc_documents(partner_id);

-- ── warehouses ──────────────────────────────────────────────────────────
CREATE TABLE warehouses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partner_id UUID NOT NULL REFERENCES partners(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    address JSONB,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_warehouses_partner_id ON warehouses(partner_id);
