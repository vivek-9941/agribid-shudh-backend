-- Migration 000011: Supporting Schema
-- Creates: notifications, notification_preferences, audit_logs, system_config

CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    event_type VARCHAR(100) NOT NULL,
    channel VARCHAR(20) NOT NULL
        CHECK (channel IN ('in_app', 'sms', 'email', 'push')),
    title VARCHAR(255) NOT NULL,
    body TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'sent', 'failed')),
    read_at TIMESTAMPTZ,
    sent_at TIMESTAMPTZ,
    retry_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notif_user_id_status ON notifications(user_id, status);
CREATE INDEX idx_notif_created_at ON notifications(created_at DESC);

CREATE TABLE notification_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    event_type VARCHAR(100) NOT NULL,
    channel VARCHAR(50) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    UNIQUE(user_id, event_type, channel)
);

CREATE INDEX idx_notif_prefs_user_id ON notification_preferences(user_id);

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID,
    user_id UUID,
    ip_address VARCHAR(45),
    method VARCHAR(10),
    path VARCHAR(500),
    status_code INT,
    duration_ms INT,
    request_body JSONB,
    response_status VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_path ON audit_logs(path);

CREATE TABLE system_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(255) UNIQUE NOT NULL,
    value JSONB NOT NULL DEFAULT '{}',
    description TEXT,
    updated_by UUID REFERENCES users(id),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_system_config_key ON system_config(key);
