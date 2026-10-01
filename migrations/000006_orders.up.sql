-- Migration 000006: Order Schema
-- Creates: carts, cart_items, orders, order_lines, order_status_history

-- ── carts ───────────────────────────────────────────────────────────────
CREATE TABLE carts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    buyer_id UUID NOT NULL REFERENCES partners(id),
    seller_id UUID NOT NULL REFERENCES partners(id),
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'ordered', 'abandoned')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_carts_buyer_id ON carts(buyer_id);
CREATE INDEX idx_carts_status ON carts(status);

-- ── cart_items ───────────────────────────────────────────────────────────
CREATE TABLE cart_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id UUID NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    quantity INT NOT NULL CHECK (quantity > 0),
    unit_price NUMERIC(12,2) NOT NULL,
    scheme_ids UUID[],
    line_total NUMERIC(12,2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cart_items_cart_id ON cart_items(cart_id);

-- ── orders ──────────────────────────────────────────────────────────────
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_number VARCHAR(50) UNIQUE NOT NULL,
    buyer_id UUID NOT NULL REFERENCES partners(id),
    seller_id UUID NOT NULL REFERENCES partners(id),
    cart_id UUID REFERENCES carts(id),
    status VARCHAR(30) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'confirmed', 'packed', 'shipped', 'delivered', 'cancelled', 'partially_fulfilled')),
    subtotal NUMERIC(12,2) NOT NULL DEFAULT 0,
    discount_total NUMERIC(12,2) NOT NULL DEFAULT 0,
    taxable_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    cgst_total NUMERIC(12,2) NOT NULL DEFAULT 0,
    sgst_total NUMERIC(12,2) NOT NULL DEFAULT 0,
    igst_total NUMERIC(12,2) NOT NULL DEFAULT 0,
    grand_total NUMERIC(12,2) NOT NULL DEFAULT 0,
    payment_status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (payment_status IN ('pending', 'partial', 'paid', 'overdue')),
    delivery_address JSONB,
    notes TEXT,
    placed_at TIMESTAMPTZ,
    confirmed_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_buyer_id ON orders(buyer_id);
CREATE INDEX idx_orders_seller_id ON orders(seller_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_placed_at ON orders(placed_at DESC);
CREATE INDEX idx_orders_order_number ON orders(order_number);

-- ── order_lines ─────────────────────────────────────────────────────────
CREATE TABLE order_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    sku VARCHAR(100) NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    ordered_qty INT NOT NULL CHECK (ordered_qty > 0),
    fulfilled_qty INT NOT NULL DEFAULT 0,
    unit_price NUMERIC(12,2) NOT NULL,
    discount_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    taxable_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    cgst_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    sgst_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    igst_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    line_total NUMERIC(12,2) NOT NULL DEFAULT 0,
    hsn_code VARCHAR(20) NOT NULL
);

CREATE INDEX idx_order_lines_order_id ON order_lines(order_id);

-- ── order_status_history ────────────────────────────────────────────────
CREATE TABLE order_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    from_status VARCHAR(50),
    to_status VARCHAR(50) NOT NULL,
    changed_by UUID REFERENCES users(id),
    reason TEXT,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_order_status_history_order_id ON order_status_history(order_id);
