package fulfillment

import (
	"context"
	"fmt"
	"time"

	"github.com/agribid/agribid-shudh-backend/internal/db"
	apperrors "github.com/agribid/agribid-shudh-backend/internal/errors"
	"github.com/agribid/agribid-shudh-backend/internal/order"
	"github.com/google/uuid"
)

type InventoryService interface {
	Fulfill(ctx context.Context, productID uuid.UUID, qty int) error
	Release(ctx context.Context, productID uuid.UUID, qty int) error
}

type OrderRepository interface {
	GetByID(ctx context.Context, orderID uuid.UUID) (*order.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status order.OrderStatus) error
}

type Service struct {
	repo         Repository
	orderRepo    OrderRepository
	inventorySvc InventoryService
}

func NewService(repo Repository, orderRepo OrderRepository, invSvc InventoryService) *Service {
	return &Service{
		repo:         repo,
		orderRepo:    orderRepo,
		inventorySvc: invSvc,
	}
}

func (s *Service) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, userID uuid.UUID, newStatus string, reason string) error {
	return db.WithTx(ctx, func(txCtx context.Context) error {
		ord, err := s.orderRepo.GetByID(txCtx, orderID)
		if err != nil {
			return apperrors.Internal(err)
		}
		if ord == nil {
			return apperrors.NotFound("order", orderID.String())
		}

		fromStatus := ord.Status
		toStatus := order.OrderStatus(newStatus)

		if !order.CanTransition(fromStatus, toStatus) {
			return apperrors.BadRequest(fmt.Sprintf("invalid transition from %s to %s", fromStatus, toStatus))
		}

		if err := s.orderRepo.UpdateOrderStatus(txCtx, orderID, toStatus); err != nil {
			return apperrors.Internal(err)
		}

		// Update inventory if cancelled (release) or delivered (fulfill - decrements physical stock)
		if toStatus == order.StatusCancelled {
			for _, line := range ord.Lines {
				if err := s.inventorySvc.Release(txCtx, line.ProductID, line.OrderedQty); err != nil {
					return fmt.Errorf("failed to release inventory: %w", err)
				}
			}
		} else if toStatus == order.StatusDelivered {
			for _, line := range ord.Lines {
				if err := s.inventorySvc.Fulfill(txCtx, line.ProductID, line.OrderedQty); err != nil {
					return fmt.Errorf("failed to fulfill inventory: %w", err)
				}
			}
		}

		// Add history event
		rsn := &reason
		if reason == "" {
			rsn = nil
		}
		evt := &FulfillmentEvent{
			OrderID:    orderID,
			FromStatus: string(fromStatus),
			ToStatus:   string(toStatus),
			ChangedBy:  userID,
			Reason:     rsn,
			ChangedAt:  time.Now(),
		}
		if err := s.repo.AddEvent(txCtx, evt); err != nil {
			return apperrors.Internal(err)
		}

		return nil
	})
}

func (s *Service) RecordPartialFulfillment(ctx context.Context, orderID uuid.UUID, userID uuid.UUID, req *RecordPartialRequest) error {
	return db.WithTx(ctx, func(txCtx context.Context) error {
		ord, err := s.orderRepo.GetByID(txCtx, orderID)
		if err != nil {
			return apperrors.Internal(err)
		}
		if ord == nil {
			return apperrors.NotFound("order", orderID.String())
		}

		if ord.Status != order.StatusPacked && ord.Status != order.StatusPartiallyFulfilled {
			return apperrors.BadRequest("order must be packed or partially fulfilled to record partials")
		}

		var partials []PartialFulfillment
		for _, lineIn := range req.Lines {
			lid, _ := uuid.Parse(lineIn.OrderLineID)
			partials = append(partials, PartialFulfillment{
				OrderLineID:  lid,
				FulfilledQty: lineIn.FulfilledQty,
			})
		}

		if err := s.repo.RecordPartial(txCtx, orderID, partials); err != nil {
			return apperrors.Internal(err)
		}

		// Change status to partially_fulfilled if not already
		if ord.Status != order.StatusPartiallyFulfilled {
			if err := s.orderRepo.UpdateOrderStatus(txCtx, orderID, order.StatusPartiallyFulfilled); err != nil {
				return apperrors.Internal(err)
			}
			evt := &FulfillmentEvent{
				OrderID:    orderID,
				FromStatus: string(ord.Status),
				ToStatus:   string(order.StatusPartiallyFulfilled),
				ChangedBy:  userID,
				ChangedAt:  time.Now(),
			}
			if err := s.repo.AddEvent(txCtx, evt); err != nil {
				return apperrors.Internal(err)
			}
		}

		return nil
	})
}
