package dispatch

import (
	"time"

	"github.com/google/uuid"
)

// ShipmentStatus represents valid shipment states.
type ShipmentStatus string

const (
	ShipmentAssigned       ShipmentStatus = "assigned"
	ShipmentPickedUp       ShipmentStatus = "picked_up"
	ShipmentInTransit      ShipmentStatus = "in_transit"
	ShipmentOutForDelivery ShipmentStatus = "out_for_delivery"
	ShipmentDelivered      ShipmentStatus = "delivered"
	ShipmentFailed         ShipmentStatus = "failed"
)

// Shipment represents a delivery assignment.
type Shipment struct {
	ID                uuid.UUID      `json:"id"`
	OrderID           uuid.UUID      `json:"order_id"`
	DeliveryPartnerID *uuid.UUID     `json:"delivery_partner_id,omitempty"`
	AssignedBy        *uuid.UUID     `json:"assigned_by,omitempty"`
	TrackingNumber    *string        `json:"tracking_number,omitempty"`
	Status            ShipmentStatus `json:"status"`
	EstimatedDelivery *time.Time     `json:"estimated_delivery,omitempty"`
	ActualDeliveryAt  *time.Time     `json:"actual_delivery_at,omitempty"`
	PODURL            *string        `json:"pod_url,omitempty"`
	PODType           *string        `json:"pod_type,omitempty"`
	GPSLat            *float64       `json:"gps_lat,omitempty"`
	GPSLng            *float64       `json:"gps_lng,omitempty"`
	AssignedAt        time.Time      `json:"assigned_at"`
	CreatedAt         time.Time      `json:"created_at"`
}

// ShipmentEvent represents a tracking event for a shipment.
type ShipmentEvent struct {
	ID         uuid.UUID  `json:"id"`
	ShipmentID uuid.UUID  `json:"shipment_id"`
	Status     string     `json:"status"`
	Notes      *string    `json:"notes,omitempty"`
	GPSLat     *float64   `json:"gps_lat,omitempty"`
	GPSLng     *float64   `json:"gps_lng,omitempty"`
	RecordedBy *uuid.UUID `json:"recorded_by,omitempty"`
	RecordedAt time.Time  `json:"recorded_at"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Request DTOs
// ─────────────────────────────────────────────────────────────────────────────

// CreateShipmentRequest is the payload for POST /shipments.
type CreateShipmentRequest struct {
	OrderID           string `json:"order_id" validate:"required,uuid"`
	DeliveryPartnerID string `json:"delivery_partner_id" validate:"required,uuid"`
	EstimatedDelivery string `json:"estimated_delivery,omitempty"`
}

// UpdateShipmentStatusRequest is the payload for PATCH /shipments/:id/status.
type UpdateShipmentStatusRequest struct {
	Status string   `json:"status" validate:"required,oneof=picked_up in_transit out_for_delivery delivered failed"`
	Notes  string   `json:"notes,omitempty"`
	GPSLat *float64 `json:"gps_lat,omitempty"`
	GPSLng *float64 `json:"gps_lng,omitempty"`
}

// UploadPODRequest is the payload for POST /shipments/:id/pod.
type UploadPODRequest struct {
	PODURL  string `json:"pod_url" validate:"required,url"`
	PODType string `json:"pod_type" validate:"required,oneof=signature photo"`
}
