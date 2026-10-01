package notification

import (
	"time"

	"github.com/google/uuid"
)

// Event type constants for business events.
const (
	EventOrderPlaced    = "order.placed"
	EventOrderConfirmed = "order.confirmed"
	EventOrderShipped   = "order.shipped"
	EventOrderDelivered = "order.delivered"
	EventPaymentReceived = "payment.received"
	EventLowStock        = "inventory.low_stock"
	EventReturnRequested = "return.requested"
	EventKYCApproved     = "kyc.approved"
	EventKYCRejected     = "kyc.rejected"
)

// Channel constants.
const (
	ChannelInApp = "in_app"
	ChannelSMS   = "sms"
	ChannelEmail = "email"
	ChannelPush  = "push"
)

// Notification represents a notification sent to a user.
type Notification struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	EventType  string     `json:"event_type"`
	Channel    string     `json:"channel"`
	Title      string     `json:"title"`
	Body       string     `json:"body"`
	Status     string     `json:"status"` // pending, sent, failed
	ReadAt     *time.Time `json:"read_at,omitempty"`
	SentAt     *time.Time `json:"sent_at,omitempty"`
	RetryCount int        `json:"retry_count"`
	CreatedAt  time.Time  `json:"created_at"`
}

// NotificationPreference represents a user's channel preference for an event.
type NotificationPreference struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	EventType string    `json:"event_type"`
	Channel   string    `json:"channel"`
	Enabled   bool      `json:"enabled"`
}

// NotificationTemplate holds a message template for an event.
type NotificationTemplate struct {
	EventType string `json:"event_type"`
	Channel   string `json:"channel"`
	Title     string `json:"title"`
	Body      string `json:"body"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Request DTOs
// ─────────────────────────────────────────────────────────────────────────────

// UpdatePreferencesRequest is the payload for PUT /notifications/preferences.
type UpdatePreferencesRequest struct {
	Preferences []PreferenceInput `json:"preferences" validate:"required,min=1,dive"`
}

// PreferenceInput is a single preference update.
type PreferenceInput struct {
	EventType string `json:"event_type" validate:"required"`
	Channel   string `json:"channel" validate:"required,oneof=in_app sms email push"`
	Enabled   bool   `json:"enabled"`
}
