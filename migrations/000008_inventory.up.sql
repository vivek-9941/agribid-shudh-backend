-- Migration 000008: Inventory Schema
-- Creates: inventory_ledger, inventory_movements

CREATE TABLE inventory_ledger (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partner_id UUID NOT NULL REFERENCES partners(id),
    warehouse_id UUID NOT NULL REFERENCES warehouses(id),
    product_id UUID NOT NULL REFERENCES products(id),
    available_qty INT NOT NULL DEFAULT 0 CHECK (available_qty >= 0),
    reserved_qty INT NOT NULL DEFAULT 0 CHECK (reserved_qty >= 0),
    physical_qty INT NOT NULL DEFAULT 0,
    reorder_level INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(partner_id, warehouse_id, product_id)
);

CREATE UNIQUE INDEX idx_inventory_partner_warehouse_product
    ON inventory_ledger(partner_id, warehouse_id, product_id);

CREATE TABLE inventory_movements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partner_id UUID NOT NULL REFERENCES partners(id),
    warehouse_id UUID NOT NULL REFERENCES warehouses(id),
    product_id UUID NOT NULL REFERENCES products(id),
    movement_type VARCHAR(20) NOT NULL
        CHECK (movement_type IN ('inward', 'outward', 'reservation', 'unreservation', 'adjustment', 'transfer')),
    quantity INT NOT NULL,
    reference_type VARCHAR(50),
    reference_id UUID,
    reason TEXT,
    performed_by UUID REFERENCES users(id),
    moved_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_inventory_movements_partner_id ON inventory_movements(partner_id);
CREATE INDEX idx_inventory_movements_product_id ON inventory_movements(product_id);
CREATE INDEX idx_inventory_movements_moved_at ON inventory_movements(moved_at DESC);
CREATE INDEX idx_inventory_movements_reference ON inventory_movements(reference_type, reference_id);
