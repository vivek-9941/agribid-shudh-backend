package dispatch

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, s *Shipment) error
	GetByID(ctx context.Context, id uuid.UUID) (*Shipment, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status ShipmentStatus) error
	AddEvent(ctx context.Context, event *ShipmentEvent) error
	ListByOrder(ctx context.Context, orderID uuid.UUID) ([]*Shipment, error)
}

type pgRepo struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepo{pool: pool}
}


func (r *pgRepo) Create(ctx context.Context, s *Shipment) error {
	query := `INSERT INTO shipments (id, tracking_number, order_id, seller_id, delivery_partner_id, status, vehicle_number, driver_name, driver_phone, pod_url, estimated_delivery, shipped_at, delivered_at, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`
	s.ID = uuid.New()
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()
	if s.Status == "" {
		s.Status = StatusPending
	}
	_, err := r.pool.Exec(ctx, query,
		s.ID, s.TrackingNumber, s.OrderID, s.SellerID, s.DeliveryPartnerID, s.Status,
		s.VehicleNumber, s.DriverName, s.DriverPhone, s.PODURL, s.EstimatedDelivery,
		s.ShippedAt, s.DeliveredAt, s.CreatedAt, s.UpdatedAt)
	return err
}

func (r *pgRepo) GetByID(ctx context.Context, id uuid.UUID) (*Shipment, error) {
	query := `SELECT id, tracking_number, order_id, seller_id, delivery_partner_id, status, vehicle_number, driver_name, driver_phone, pod_url, estimated_delivery, shipped_at, delivered_at, created_at, updated_at
			  FROM shipments WHERE id = $1`
	var s Shipment
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.TrackingNumber, &s.OrderID, &s.SellerID, &s.DeliveryPartnerID, &s.Status,
		&s.VehicleNumber, &s.DriverName, &s.DriverPhone, &s.PODURL, &s.EstimatedDelivery,
		&s.ShippedAt, &s.DeliveredAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("GetByID: %w", err)
	}
	return &s, nil
}

func (r *pgRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status ShipmentStatus) error {
	query := `UPDATE shipments SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, status, time.Now(), id)
	return err
}

func (r *pgRepo) AddEvent(ctx context.Context, event *ShipmentEvent) error {
	query := `INSERT INTO shipment_events (id, shipment_id, status, location, notes, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6)`
	event.ID = uuid.New()
	event.CreatedAt = time.Now()
	_, err := r.pool.Exec(ctx, query, event.ID, event.ShipmentID, event.Status, event.Location, event.Notes, event.CreatedAt)
	return err
}

func (r *pgRepo) ListByOrder(ctx context.Context, orderID uuid.UUID) ([]*Shipment, error) {
	// Stub
	return nil, nil
}
