-- Migration 000007: Invoice Schema
-- Creates: invoices, invoice_lines

CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_number VARCHAR(50) UNIQUE NOT NULL,
    order_id UUID NOT NULL REFERENCES orders(id),
    seller_id UUID NOT NULL REFERENCES partners(id),
    buyer_id UUID NOT NULL REFERENCES partners(id),
    invoice_type VARCHAR(20) NOT NULL DEFAULT 'tax_invoice'
        CHECK (invoice_type IN ('tax_invoice', 'credit_note', 'debit_note')),
    parent_invoice_id UUID REFERENCES invoices(id),
    seller_gstin VARCHAR(15),
    buyer_gstin VARCHAR(15),
    place_of_supply VARCHAR(2),
    subtotal NUMERIC(12,2) NOT NULL DEFAULT 0,
    discount_total NUMERIC(12,2) NOT NULL DEFAULT 0,
    taxable_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    cgst_total NUMERIC(12,2) NOT NULL DEFAULT 0,
    sgst_total NUMERIC(12,2) NOT NULL DEFAULT 0,
    igst_total NUMERIC(12,2) NOT NULL DEFAULT 0,
    cess_total NUMERIC(12,2) NOT NULL DEFAULT 0,
    grand_total NUMERIC(12,2) NOT NULL DEFAULT 0,
    pdf_url VARCHAR(500),
    status VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'issued', 'cancelled')),
    issued_at TIMESTAMPTZ,
    due_date DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_invoices_seller_id ON invoices(seller_id);
CREATE INDEX idx_invoices_buyer_id ON invoices(buyer_id);
CREATE INDEX idx_invoices_order_id ON invoices(order_id);
CREATE INDEX idx_invoices_issued_at ON invoices(issued_at DESC);
CREATE INDEX idx_invoices_status ON invoices(status);

CREATE TABLE invoice_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    sku VARCHAR(100) NOT NULL,
    description VARCHAR(255),
    hsn_code VARCHAR(20) NOT NULL,
    quantity INT NOT NULL,
    unit_price NUMERIC(12,2) NOT NULL,
    discount_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    taxable_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    cgst_rate NUMERIC(5,2) NOT NULL DEFAULT 0,
    cgst_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    sgst_rate NUMERIC(5,2) NOT NULL DEFAULT 0,
    sgst_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    igst_rate NUMERIC(5,2) NOT NULL DEFAULT 0,
    igst_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    line_total NUMERIC(12,2) NOT NULL DEFAULT 0
);

CREATE INDEX idx_invoice_lines_invoice_id ON invoice_lines(invoice_id);
