-- Migration 000004: Catalog Schema
-- Creates: categories, products, product_images, hsn_tax_rates

-- ── categories ──────────────────────────────────────────────────────────
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id UUID REFERENCES categories(id),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    path VARCHAR(1000),
    level INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_categories_parent_id ON categories(parent_id);
CREATE INDEX idx_categories_slug ON categories(slug);

-- ── products ────────────────────────────────────────────────────────────
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sku VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category_id UUID REFERENCES categories(id),
    brand VARCHAR(100),
    manufacturer_id UUID NOT NULL REFERENCES partners(id),
    hsn_code VARCHAR(20) NOT NULL,
    uom VARCHAR(20),
    weight_grams INT,
    dimensions JSONB,
    parent_sku_id UUID REFERENCES products(id),
    variant_attrs JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'inactive', 'discontinued')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_products_sku ON products(sku);
CREATE INDEX idx_products_category_id ON products(category_id);
CREATE INDEX idx_products_manufacturer_id ON products(manufacturer_id);
CREATE INDEX idx_products_status ON products(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_hsn_code ON products(hsn_code);
CREATE INDEX idx_products_brand ON products(brand);

-- ── product_images ──────────────────────────────────────────────────────
CREATE TABLE product_images (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    url VARCHAR(500) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_product_images_product_id ON product_images(product_id);

-- ── hsn_tax_rates ───────────────────────────────────────────────────────
CREATE TABLE hsn_tax_rates (
    hsn_code VARCHAR(20) PRIMARY KEY,
    description VARCHAR(500) NOT NULL,
    cgst_rate NUMERIC(5,2) NOT NULL DEFAULT 0,
    sgst_rate NUMERIC(5,2) NOT NULL DEFAULT 0,
    igst_rate NUMERIC(5,2) NOT NULL DEFAULT 0,
    cess_rate NUMERIC(5,2) NOT NULL DEFAULT 0,
    effective_from DATE NOT NULL DEFAULT CURRENT_DATE
);

-- Seed common FMCG HSN codes
INSERT INTO hsn_tax_rates (hsn_code, description, cgst_rate, sgst_rate, igst_rate) VALUES
    ('0401', 'Milk and cream, not concentrated', 0.00, 0.00, 0.00),
    ('0402', 'Milk and cream, concentrated', 2.50, 2.50, 5.00),
    ('0901', 'Coffee', 2.50, 2.50, 5.00),
    ('0902', 'Tea', 2.50, 2.50, 5.00),
    ('1006', 'Rice', 2.50, 2.50, 5.00),
    ('1101', 'Wheat flour', 2.50, 2.50, 5.00),
    ('1507', 'Soybean oil', 2.50, 2.50, 5.00),
    ('1509', 'Olive oil', 6.00, 6.00, 12.00),
    ('1517', 'Margarine and fats', 6.00, 6.00, 12.00),
    ('1701', 'Sugar', 2.50, 2.50, 5.00),
    ('1806', 'Chocolate and cocoa preparations', 9.00, 9.00, 18.00),
    ('1905', 'Bread, pastry, cakes', 9.00, 9.00, 18.00),
    ('2009', 'Fruit juices', 6.00, 6.00, 12.00),
    ('2106', 'Food preparations (instant mixes)', 9.00, 9.00, 18.00),
    ('2201', 'Water, mineral and aerated', 9.00, 9.00, 18.00),
    ('2202', 'Beverages containing sugar', 14.00, 14.00, 28.00),
    ('3301', 'Essential oils', 9.00, 9.00, 18.00),
    ('3304', 'Beauty and skin care preparations', 9.00, 9.00, 18.00),
    ('3305', 'Hair care preparations', 9.00, 9.00, 18.00),
    ('3306', 'Oral hygiene preparations', 9.00, 9.00, 18.00),
    ('3401', 'Soap and detergents', 9.00, 9.00, 18.00),
    ('3402', 'Surface-active agents', 9.00, 9.00, 18.00);
