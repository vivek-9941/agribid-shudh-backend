-- Migration 000010: Dispatch and Returns Schema
-- Creates: shipments, shipment_events, return_requests, return_lines

CREATE TABLE shipments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id),
    delivery_partner_id UUID REFERENCES partners(id),
    assigned_by UUID REFERENCES users(id),
    tracking_number VARCHAR(100),
    status VARCHAR(30) NOT NULL DEFAULT 'assigned'
        CHECK (status IN ('assigned', 'picked_up', 'in_transit', 'out_for_delivery', 'delivered', 'failed')),
    estimated_delivery DATE,
    actual_delivery_at TIMESTAMPTZ,
    pod_url VARCHAR(500),
    pod_type VARCHAR(20) CHECK (pod_type IS NULL OR pod_type IN ('signature', 'photo')),
    gps_lat NUMERIC(10,7),
    gps_lng NUMERIC(10,7),
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_shipments_order_id ON shipments(order_id);
CREATE INDEX idx_shipments_delivery_partner_id ON shipments(delivery_partner_id);
CREATE INDEX idx_shipments_status ON shipments(status);

CREATE TABLE shipment_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shipment_id UUID NOT NULL REFERENCES shipments(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL,
    notes TEXT,
    gps_lat NUMERIC(10,7),
    gps_lng NUMERIC(10,7),
    recorded_by UUID REFERENCES users(id),
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_shipment_events_shipment_id ON shipment_events(shipment_id);

CREATE TABLE return_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    return_number VARCHAR(50) UNIQUE NOT NULL,
    order_id UUID NOT NULL REFERENCES orders(id),
    invoice_id UUID REFERENCES invoices(id),
    buyer_id UUID NOT NULL REFERENCES partners(id),
    seller_id UUID NOT NULL REFERENCES partners(id),
    status VARCHAR(20) NOT NULL DEFAULT 'requested'
        CHECK (status IN ('requested', 'approved', 'rejected', 'picked_up', 'inspected', 'closed')),
    reason VARCHAR(30) NOT NULL
        CHECK (reason IN ('damaged', 'expired', 'wrong_product', 'quality_issue', 'other')),
    notes TEXT,
    pickup_scheduled_at TIMESTAMPTZ,
    credit_note_id UUID REFERENCES invoices(id),
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_return_requests_order_id ON return_requests(order_id);
CREATE INDEX idx_return_requests_buyer_id ON return_requests(buyer_id);
CREATE INDEX idx_return_requests_seller_id ON return_requests(seller_id);
CREATE INDEX idx_return_requests_status ON return_requests(status);

CREATE TABLE return_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    return_id UUID NOT NULL REFERENCES return_requests(id) ON DELETE CASCADE,
    order_line_id UUID NOT NULL REFERENCES order_lines(id),
    product_id UUID NOT NULL REFERENCES products(id),
    requested_qty INT NOT NULL CHECK (requested_qty > 0),
    approved_qty INT,
    inspection_result VARCHAR(10)
        CHECK (inspection_result IS NULL OR inspection_result IN ('pass', 'fail', 'partial')),
    inspection_notes TEXT
);

CREATE INDEX idx_return_lines_return_id ON return_lines(return_id);
