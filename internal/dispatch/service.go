package dispatch

import (
	"context"
	"fmt"
	"time"

	"github.com/agribid/agribid-shudh-backend/internal/db"
	apperrors "github.com/agribid/agribid-shudh-backend/internal/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FulfillmentService interface {
	UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, userID uuid.UUID, newStatus string, reason string) error
}

type Service struct {
	repo           Repository
	fulfillmentSvc FulfillmentService
	pool           *pgxpool.Pool
}

func NewService(repo Repository, fulfillmentSvc FulfillmentService, pool *pgxpool.Pool) *Service {
	return &Service{
		repo:           repo,
		fulfillmentSvc: fulfillmentSvc,
		pool:           pool,
	}
}

func (s *Service) AssignDeliveryPartner(ctx context.Context, req *AssignPartnerRequest, orderID, sellerID uuid.UUID) (*Shipment, error) {
	dpID, _ := uuid.Parse(req.DeliveryPartnerID)
	
	shipment := &Shipment{
		TrackingNumber:    fmt.Sprintf("TRK-%d", time.Now().Unix()), // Simplified tracking logic
		OrderID:           orderID,
		SellerID:          sellerID,
		DeliveryPartnerID: &dpID,
		Status:            StatusAssigned,
		VehicleNumber:     &req.VehicleNumber,
		DriverName:        &req.DriverName,
		DriverPhone:       &req.DriverPhone,
	}

	err := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		if err := s.repo.Create(ctx, shipment); err != nil {
			return err
		}

		evt := &ShipmentEvent{
			ShipmentID: shipment.ID,
			Status:     StatusAssigned,
			Notes:      func(s string) *string { return &s }("Assigned to delivery partner"),
		}
		if err := s.repo.AddEvent(ctx, evt); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, apperrors.Internal(err)
	}

	return shipment, nil
}

func (s *Service) UpdateShipmentStatus(ctx context.Context, req *UpdateStatusRequest, shipmentID, userID uuid.UUID) error {
	return db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		shipment, err := s.repo.GetByID(ctx, shipmentID)
		if err != nil {
			return apperrors.Internal(err)
		}
		if shipment == nil {
			return apperrors.NotFound("shipment", shipmentID.String())
		}

		newStatus := ShipmentStatus(req.Status)
		if !CanTransition(shipment.Status, newStatus) {
			return apperrors.BadRequest(apperrors.CodeValidationFailed, fmt.Sprintf("invalid transition from %s to %s", shipment.Status, newStatus))
		}

		if err := s.repo.UpdateStatus(ctx, shipmentID, newStatus); err != nil {
			return apperrors.Internal(err)
		}

		evt := &ShipmentEvent{
			ShipmentID: shipmentID,
			Status:     newStatus,
			Location:   req.Location,
			Notes:      req.Notes,
		}
		if err := s.repo.AddEvent(ctx, evt); err != nil {
			return apperrors.Internal(err)
		}

		if newStatus == StatusDelivered {
			if err := s.fulfillmentSvc.UpdateOrderStatus(ctx, shipment.OrderID, userID, "delivered", "Shipment Delivered"); err != nil {
				return fmt.Errorf("failed to update order status: %w", err)
			}
		}

		return nil
	})
}

func (s *Service) RecordPOD(ctx context.Context, req *RecordPODRequest, shipmentID, userID uuid.UUID) error {
	// First update the shipment with POD
	// In a real app we'd have a specific repo method for UpdatePOD
	// For this stub, we just update status to delivered which we'll handle through UpdateShipmentStatus
	statusReq := &UpdateStatusRequest{
		Status: string(StatusDelivered),
		Notes:  func(s string) *string { return &s }("POD uploaded: " + req.PODURL),
	}
	return s.UpdateShipmentStatus(ctx, statusReq, shipmentID, userID)
}
