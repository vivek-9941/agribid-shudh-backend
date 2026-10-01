package dispatch

import (
	"time"

	"github.com/google/uuid"
)

// ShipmentStatus represents valid shipment states.
type ShipmentStatus string

const (
	StatusPending          ShipmentStatus = "pending"
	StatusAssigned         ShipmentStatus = "assigned"
	StatusPickedUp         ShipmentStatus = "picked_up"
	StatusInTransit        ShipmentStatus = "in_transit"
	StatusOutForDelivery   ShipmentStatus = "out_for_delivery"
	StatusDelivered        ShipmentStatus = "delivered"
	StatusFailed           ShipmentStatus = "failed"
)

// ValidTransitions defines valid transitions for shipment status.
var ValidTransitions = map[ShipmentStatus][]ShipmentStatus{
	StatusAssigned:       {StatusPickedUp, StatusFailed},
	StatusPickedUp:       {StatusInTransit, StatusFailed},
	StatusInTransit:      {StatusOutForDelivery, StatusFailed},
	StatusOutForDelivery: {StatusDelivered, StatusFailed},
}

func CanTransition(from, to ShipmentStatus) bool {
	targets, ok := ValidTransitions[from]
	if !ok {
		return false
	}
	for _, t := range targets {
		if t == to {
			return true
		}
	}
	return false
}

// Shipment represents a delivery assignment.
type Shipment struct {
	ID                uuid.UUID      `json:"id"`
	OrderID           uuid.UUID      `json:"order_id"`
	SellerID          uuid.UUID      `json:"seller_id"`
	DeliveryPartnerID *uuid.UUID     `json:"delivery_partner_id,omitempty"`
	AssignedBy        *uuid.UUID     `json:"assigned_by,omitempty"`
	TrackingNumber    string         `json:"tracking_number"`
	Status            ShipmentStatus `json:"status"`
	VehicleNumber     *string        `json:"vehicle_number,omitempty"`
	DriverName        *string        `json:"driver_name,omitempty"`
	DriverPhone       *string        `json:"driver_phone,omitempty"`
	EstimatedDelivery *time.Time     `json:"estimated_delivery,omitempty"`
	ShippedAt         *time.Time     `json:"shipped_at,omitempty"`
	DeliveredAt       *time.Time     `json:"delivered_at,omitempty"`
	PODURL            *string        `json:"pod_url,omitempty"`
	PODType           *string        `json:"pod_type,omitempty"`
	GPSLat            *float64       `json:"gps_lat,omitempty"`
	GPSLng            *float64       `json:"gps_lng,omitempty"`
	AssignedAt        time.Time      `json:"assigned_at"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

// ShipmentEvent represents a tracking event for a shipment.
type ShipmentEvent struct {
	ID         uuid.UUID      `json:"id"`
	ShipmentID uuid.UUID      `json:"shipment_id"`
	Status     ShipmentStatus `json:"status"`
	Location   *string        `json:"location,omitempty"`
	Notes      *string        `json:"notes,omitempty"`
	GPSLat     *float64       `json:"gps_lat,omitempty"`
	GPSLng     *float64       `json:"gps_lng,omitempty"`
	RecordedBy *uuid.UUID     `json:"recorded_by,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Request DTOs
// ─────────────────────────────────────────────────────────────────────────────

type AssignPartnerRequest struct {
	OrderID           string `json:"order_id" validate:"required,uuid"`
	DeliveryPartnerID string `json:"delivery_partner_id" validate:"required,uuid"`
	VehicleNumber     string `json:"vehicle_number,omitempty"`
	DriverName        string `json:"driver_name,omitempty"`
	DriverPhone       string `json:"driver_phone,omitempty"`
}

type UpdateStatusRequest struct {
	Status   string  `json:"status" validate:"required"`
	Location *string `json:"location,omitempty"`
	Notes    *string `json:"notes,omitempty"`
	GPSLat   *float64 `json:"gps_lat,omitempty"`
	GPSLng   *float64 `json:"gps_lng,omitempty"`
}

type RecordPODRequest struct {
	PODURL  string `json:"pod_url" validate:"required,url"`
	PODType string `json:"pod_type" validate:"required,oneof=signature photo"`
}

