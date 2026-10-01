-- Migration 000002: Users, Roles, Permissions Schema
-- Creates: users, roles, permissions, user_roles, otp_sessions, refresh_tokens
-- Seeds: 7 default roles and default permissions

-- ── roles ───────────────────────────────────────────────────────────────
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ── permissions ─────────────────────────────────────────────────────────
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    UNIQUE(role_id, resource, action)
);

-- ── users ───────────────────────────────────────────────────────────────
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone VARCHAR(15) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE,
    password_hash VARCHAR(255),
    full_name VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'inactive', 'suspended')),
    partner_id UUID, -- FK added after partners table exists
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_users_phone ON users(phone);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_partner_id ON users(partner_id);
CREATE INDEX idx_users_status ON users(status) WHERE deleted_at IS NULL;

-- ── user_roles ──────────────────────────────────────────────────────────
CREATE TABLE user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    assigned_by UUID REFERENCES users(id),
    UNIQUE(user_id, role_id)
);

-- ── otp_sessions ────────────────────────────────────────────────────────
CREATE TABLE otp_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone VARCHAR(15) NOT NULL,
    otp_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    attempts INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_otp_sessions_phone ON otp_sessions(phone, created_at DESC);

-- ── refresh_tokens ──────────────────────────────────────────────────────
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);

-- ── Seed: default roles ─────────────────────────────────────────────────
INSERT INTO roles (code, name) VALUES
    ('ADMIN', 'Administrator'),
    ('MANUFACTURER', 'Manufacturer'),
    ('STATE_STOCKIST', 'State Stockist'),
    ('DISTRIBUTOR', 'Distributor'),
    ('SUB_DISTRIBUTOR', 'Sub-Distributor'),
    ('RETAILER', 'Retailer'),
    ('DELIVERY_PARTNER', 'Delivery Partner');

-- ── Seed: default permissions per role ──────────────────────────────────
-- Admin gets all permissions.
INSERT INTO permissions (role_id, resource, action)
SELECT r.id, res.resource, act.action
FROM roles r
CROSS JOIN (VALUES ('products'),('orders'),('invoices'),('shipments'),('inventory'),('payments'),('returns'),('partners'),('users'),('dashboard'),('reports'),('notifications'),('admin'),('config'),('audit')) AS res(resource)
CROSS JOIN (VALUES ('create'),('read'),('update'),('delete'),('approve')) AS act(action)
WHERE r.code = 'ADMIN';

-- Manufacturer: products CRUD, orders read/fulfill, invoices read, inventory manage
INSERT INTO permissions (role_id, resource, action)
SELECT r.id, v.resource, v.action FROM roles r
CROSS JOIN (VALUES
    ('products','create'),('products','read'),('products','update'),('products','delete'),
    ('orders','read'),('orders','update'),
    ('invoices','read'),('invoices','create'),
    ('inventory','create'),('inventory','read'),('inventory','update'),
    ('shipments','read'),
    ('payments','read'),
    ('returns','read'),('returns','update'),
    ('dashboard','read'),('reports','read'),('notifications','read')
) AS v(resource, action) WHERE r.code = 'MANUFACTURER';

-- State Stockist
INSERT INTO permissions (role_id, resource, action)
SELECT r.id, v.resource, v.action FROM roles r
CROSS JOIN (VALUES
    ('products','read'),
    ('orders','create'),('orders','read'),('orders','update'),
    ('invoices','read'),('invoices','create'),
    ('inventory','create'),('inventory','read'),('inventory','update'),
    ('shipments','read'),('shipments','update'),
    ('payments','read'),('payments','create'),
    ('returns','create'),('returns','read'),('returns','update'),
    ('dashboard','read'),('reports','read'),('notifications','read')
) AS v(resource, action) WHERE r.code = 'STATE_STOCKIST';

-- Distributor
INSERT INTO permissions (role_id, resource, action)
SELECT r.id, v.resource, v.action FROM roles r
CROSS JOIN (VALUES
    ('products','read'),
    ('orders','create'),('orders','read'),('orders','update'),
    ('invoices','read'),('invoices','create'),
    ('inventory','create'),('inventory','read'),('inventory','update'),
    ('shipments','read'),('shipments','update'),
    ('payments','read'),('payments','create'),
    ('returns','create'),('returns','read'),('returns','update'),
    ('dashboard','read'),('reports','read'),('notifications','read')
) AS v(resource, action) WHERE r.code = 'DISTRIBUTOR';

-- Sub-Distributor
INSERT INTO permissions (role_id, resource, action)
SELECT r.id, v.resource, v.action FROM roles r
CROSS JOIN (VALUES
    ('products','read'),
    ('orders','create'),('orders','read'),('orders','update'),
    ('invoices','read'),('invoices','create'),
    ('inventory','create'),('inventory','read'),('inventory','update'),
    ('shipments','read'),('shipments','update'),
    ('payments','read'),('payments','create'),
    ('returns','create'),('returns','read'),('returns','update'),
    ('dashboard','read'),('reports','read'),('notifications','read')
) AS v(resource, action) WHERE r.code = 'SUB_DISTRIBUTOR';

-- Retailer
INSERT INTO permissions (role_id, resource, action)
SELECT r.id, v.resource, v.action FROM roles r
CROSS JOIN (VALUES
    ('products','read'),
    ('orders','create'),('orders','read'),
    ('invoices','read'),
    ('payments','read'),('payments','create'),
    ('returns','create'),('returns','read'),
    ('dashboard','read'),('notifications','read')
) AS v(resource, action) WHERE r.code = 'RETAILER';

-- Delivery Partner
INSERT INTO permissions (role_id, resource, action)
SELECT r.id, v.resource, v.action FROM roles r
CROSS JOIN (VALUES
    ('shipments','read'),('shipments','update'),
    ('orders','read'),
    ('dashboard','read'),('notifications','read')
) AS v(resource, action) WHERE r.code = 'DELIVERY_PARTNER';
