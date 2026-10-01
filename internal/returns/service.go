package returns

import (
	"context"
	"fmt"
	"time"

	"github.com/agribid/agribid-shudh-backend/internal/db"
	apperrors "github.com/agribid/agribid-shudh-backend/internal/errors"
	"github.com/agribid/agribid-shudh-backend/internal/invoice"
	"github.com/agribid/agribid-shudh-backend/internal/order"
	"github.com/google/uuid"
)

type OrderService interface {
	GetByID(ctx context.Context, orderID uuid.UUID) (*order.Order, error)
}

type InventoryService interface {
	Adjust(ctx context.Context, productID uuid.UUID, qty int, reason string) error
}

type InvoiceService interface {
	CreateCreditNote(ctx context.Context, cn *invoice.CreditNote) error
}

type PaymentService interface {
	UpdateCredit(ctx context.Context, buyerID, sellerID uuid.UUID, amountStr string) error
}

type Service struct {
	repo         Repository
	orderSvc     OrderService
	inventorySvc InventoryService
	invoiceSvc   InvoiceService
	paymentSvc   PaymentService
}

func NewService(repo Repository, orderSvc OrderService, invSvc InventoryService, invcSvc InvoiceService, paySvc PaymentService) *Service {
	return &Service{
		repo:         repo,
		orderSvc:     orderSvc,
		inventorySvc: invSvc,
		invoiceSvc:   invcSvc,
		paymentSvc:   paySvc,
	}
}

func (s *Service) Initiate(ctx context.Context, buyerID uuid.UUID, req *InitiateReturnRequest) (*ReturnRequest, error) {
	orderID, _ := uuid.Parse(req.OrderID)
	
	ord, err := s.orderSvc.GetByID(ctx, orderID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if ord == nil {
		return nil, apperrors.NotFound("order", req.OrderID)
	}

	if ord.Status != order.StatusDelivered {
		return nil, apperrors.BadRequest("can only return delivered orders")
	}

	// Window check
	if ord.DeliveredAt == nil {
		return nil, apperrors.BadRequest("order has no delivery date")
	}
	if time.Since(*ord.DeliveredAt) > 7*24*time.Hour {
		return nil, apperrors.BadRequest("RETURN_WINDOW_EXPIRED")
	}

	retReq := &ReturnRequest{
		ReturnNumber: fmt.Sprintf("RET-%d", time.Now().Unix()),
		OrderID:      orderID,
		BuyerID:      ord.BuyerID,
		SellerID:     ord.SellerID,
		Status:       StatusPending,
		TotalAmount:  "0.00", // Needs actual calculation based on lines
	}

	for _, lineIn := range req.Lines {
		lid, _ := uuid.Parse(lineIn.OrderLineID)
		
		retLine := &ReturnLine{
			OrderLineID:  lid,
			RequestedQty: lineIn.Quantity,
			Reason:       lineIn.Reason,
			// Simplified: other fields like ProductID, SKU should be looked up from ord.Lines
		}
		retReq.Lines = append(retReq.Lines, retLine)
	}

	if err := s.repo.Create(ctx, retReq); err != nil {
		return nil, apperrors.Internal(err)
	}

	return retReq, nil
}

func (s *Service) Approve(ctx context.Context, returnID uuid.UUID) error {
	return s.repo.UpdateStatus(ctx, returnID, StatusApproved)
}

func (s *Service) Reject(ctx context.Context, returnID uuid.UUID) error {
	return s.repo.UpdateStatus(ctx, returnID, StatusRejected)
}

func (s *Service) RecordInspection(ctx context.Context, returnID uuid.UUID, req *RecordInspectionRequest) error {
	return db.WithTx(ctx, func(txCtx context.Context) error {
		retReq, err := s.repo.GetByID(txCtx, returnID)
		if err != nil {
			return apperrors.Internal(err)
		}
		if retReq == nil {
			return apperrors.NotFound("return", returnID.String())
		}

		if retReq.Status != StatusApproved && retReq.Status != StatusReceived {
			return apperrors.BadRequest("return must be approved/received to inspect")
		}

		for _, item := range req.Items {
			lid, _ := uuid.Parse(item.LineID)
			
			if err := s.repo.AddInspectionResult(txCtx, lid, item.PassedQty, item.FailedQty, item.Condition); err != nil {
				return err
			}
			
			// If passed, restore inventory, generate credit note, update credit.
			// Simplified approach: if any passed, we should create a credit note for the passed portion.
			// Here we assume if passedQty > 0, we do actions.
			if item.PassedQty > 0 {
				// 1. Restore Inventory
				// We need ProductID. For this stub, we just skip it or assume we fetched it.
				// s.inventorySvc.Adjust(...)

				// 2. Generate Credit Note
				cn := &invoice.CreditNote{
					InvoiceType: "credit_note",
					// populate details
				}
				if err := s.invoiceSvc.CreateCreditNote(txCtx, cn); err != nil {
					// return err // skipped in stub
				}

				// 3. Update Credit (negative to restore utilized credit)
				// s.paymentSvc.UpdateCredit(txCtx, retReq.BuyerID, retReq.SellerID, "-amount")
			}
		}

		if err := s.repo.UpdateStatus(txCtx, returnID, StatusInspected); err != nil {
			return err
		}
		return nil
	})
}
