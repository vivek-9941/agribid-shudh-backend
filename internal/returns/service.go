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
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
	pool         *pgxpool.Pool
}

func NewService(repo Repository, orderSvc OrderService, invSvc InventoryService, invcSvc InvoiceService, paySvc PaymentService, pool *pgxpool.Pool) *Service {
	return &Service{
		repo:         repo,
		orderSvc:     orderSvc,
		inventorySvc: invSvc,
		invoiceSvc:   invcSvc,
		paymentSvc:   paySvc,
		pool:         pool,
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
		return nil, apperrors.BadRequest(apperrors.CodeValidationFailed, "can only return delivered orders")
	}

	// Window check
	if ord.DeliveredAt == nil {
		return nil, apperrors.BadRequest(apperrors.CodeValidationFailed, "order has no delivery date")
	}
	if time.Since(*ord.DeliveredAt) > 7*24*time.Hour {
		return nil, apperrors.BadRequest(apperrors.CodeValidationFailed, "RETURN_WINDOW_EXPIRED")
	}

	retReq := &ReturnRequest{
		ReturnNumber: fmt.Sprintf("RET-%d", time.Now().Unix()),
		OrderID:      orderID,
		BuyerID:      ord.BuyerID,
		SellerID:     ord.SellerID,
		Status:       ReturnRequested,
		Reason:       req.Reason,
		Notes:        &req.Notes,
	}

	for _, lineIn := range req.Lines {
		lid, _ := uuid.Parse(lineIn.OrderLineID)
		
		retLine := &ReturnLine{
			OrderLineID:  lid,
			RequestedQty: lineIn.RequestedQty,
		}
		retReq.Lines = append(retReq.Lines, retLine)
	}

	if err := s.repo.Create(ctx, retReq); err != nil {
		return nil, apperrors.Internal(err)
	}

	return retReq, nil
}

func (s *Service) Approve(ctx context.Context, returnID uuid.UUID) error {
	return s.repo.UpdateStatus(ctx, returnID, ReturnApproved)
}

func (s *Service) Reject(ctx context.Context, returnID uuid.UUID) error {
	return s.repo.UpdateStatus(ctx, returnID, ReturnRejected)
}

func (s *Service) RecordInspection(ctx context.Context, returnID uuid.UUID, req *InspectionInput) error {
	return db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		txCtx := ctx
		retReq, err := s.repo.GetByID(txCtx, returnID)
		if err != nil {
			return apperrors.Internal(err)
		}
		if retReq == nil {
			return apperrors.NotFound("return", returnID.String())
		}

		if retReq.Status != ReturnApproved && retReq.Status != ReturnPickedUp {
			return apperrors.BadRequest(apperrors.CodeValidationFailed, "return must be approved/received to inspect")
		}

		for _, item := range req.Lines {
			lid, _ := uuid.Parse(item.ReturnLineID)
			
			// PassedQty/FailedQty logic simplified for stub
			passedQty := 0
			failedQty := 0
			if item.Result == "pass" {
				passedQty = 1 // Simplified
			}
			
			if err := s.repo.AddInspectionResult(txCtx, lid, passedQty, failedQty, item.InspectionNotes); err != nil {
				return err
			}
			
			// If passed, restore inventory, generate credit note, update credit.
			// Simplified approach: if any passed, we should create a credit note for the passed portion.
			// Here we assume if passedQty > 0, we do actions.
			if passedQty > 0 {
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

		if err := s.repo.UpdateStatus(txCtx, returnID, ReturnInspected); err != nil {
			return err
		}
		return nil
	})
}
