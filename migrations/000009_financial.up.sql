-- Migration 000009: Financial Schema
-- Creates: credit_accounts, payments, ledger_entries

CREATE TABLE credit_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    buyer_id UUID NOT NULL REFERENCES partners(id),
    seller_id UUID NOT NULL REFERENCES partners(id),
    credit_limit NUMERIC(12,2) NOT NULL DEFAULT 0,
    credit_utilized NUMERIC(12,2) NOT NULL DEFAULT 0,
    payment_terms INT NOT NULL DEFAULT 30,
    overdue_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(buyer_id, seller_id)
);

CREATE INDEX idx_credit_accounts_buyer_id ON credit_accounts(buyer_id);
CREATE INDEX idx_credit_accounts_seller_id ON credit_accounts(seller_id);

CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_number VARCHAR(50) UNIQUE NOT NULL,
    buyer_id UUID NOT NULL REFERENCES partners(id),
    seller_id UUID NOT NULL REFERENCES partners(id),
    amount NUMERIC(12,2) NOT NULL,
    method VARCHAR(20) NOT NULL
        CHECK (method IN ('cash', 'bank_transfer', 'cheque', 'upi', 'card', 'gateway')),
    reference_number VARCHAR(100),
    gateway_txn_id VARCHAR(255),
    invoice_id UUID REFERENCES invoices(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'confirmed', 'failed', 'refunded')),
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payments_buyer_id ON payments(buyer_id);
CREATE INDEX idx_payments_seller_id ON payments(seller_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_invoice_id ON payments(invoice_id) WHERE invoice_id IS NOT NULL;

CREATE TABLE ledger_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partner_id UUID NOT NULL REFERENCES partners(id),
    counterparty_id UUID NOT NULL REFERENCES partners(id),
    entry_type VARCHAR(20) NOT NULL
        CHECK (entry_type IN ('invoice', 'payment', 'credit_note', 'adjustment')),
    reference_id UUID,
    debit NUMERIC(12,2) NOT NULL DEFAULT 0,
    credit NUMERIC(12,2) NOT NULL DEFAULT 0,
    balance NUMERIC(12,2) NOT NULL DEFAULT 0,
    entry_date DATE NOT NULL DEFAULT CURRENT_DATE,
    description VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ledger_partner_id ON ledger_entries(partner_id, entry_date DESC);
CREATE INDEX idx_ledger_counterparty_id ON ledger_entries(counterparty_id);
