-- Migration 000005: Pricing Schema
-- Creates: product_prices, schemes, scheme_applicability

-- ── product_prices ──────────────────────────────────────────────────────
CREATE TABLE product_prices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    role_code VARCHAR(50) NOT NULL,
    partner_id UUID REFERENCES partners(id),
    mrp NUMERIC(12,2) NOT NULL,
    base_price NUMERIC(12,2) NOT NULL,
    effective_from DATE NOT NULL DEFAULT CURRENT_DATE,
    effective_to DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_product_prices_product_id ON product_prices(product_id);
CREATE INDEX idx_product_prices_role_code ON product_prices(role_code);
CREATE INDEX idx_product_prices_partner_id ON product_prices(partner_id) WHERE partner_id IS NOT NULL;

-- ── schemes ─────────────────────────────────────────────────────────────
CREATE TABLE schemes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL
        CHECK (type IN ('percentage', 'flat', 'buy_x_get_y')),
    discount_value NUMERIC(10,2),
    buy_qty INT,
    get_qty INT,
    valid_from DATE NOT NULL,
    valid_to DATE NOT NULL,
    is_exclusive BOOLEAN NOT NULL DEFAULT FALSE,
    created_by UUID REFERENCES users(id),
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_schemes_status ON schemes(status);
CREATE INDEX idx_schemes_validity ON schemes(valid_from, valid_to);

-- ── scheme_applicability ────────────────────────────────────────────────
CREATE TABLE scheme_applicability (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scheme_id UUID NOT NULL REFERENCES schemes(id) ON DELETE CASCADE,
    target_type VARCHAR(20) NOT NULL
        CHECK (target_type IN ('role', 'partner', 'category', 'product')),
    target_id VARCHAR(255) NOT NULL
);

CREATE INDEX idx_scheme_applicability_scheme_id ON scheme_applicability(scheme_id);
CREATE INDEX idx_scheme_applicability_target ON scheme_applicability(target_type, target_id);
