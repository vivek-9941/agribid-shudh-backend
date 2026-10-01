package invoice

import (
	"context"
	"fmt"
	"time"

	"github.com/agribid/agribid-shudh-backend/internal/db"
	apperrors "github.com/agribid/agribid-shudh-backend/internal/errors"
	"github.com/agribid/agribid-shudh-backend/internal/order"
	"github.com/agribid/agribid-shudh-backend/internal/partner"
	"github.com/google/uuid"
)

type OrderService interface {
	GetByID(ctx context.Context, orderID uuid.UUID) (*order.Order, error)
}

type PartnerService interface {
	GetPartner(ctx context.Context, id uuid.UUID) (*partner.Partner, error)
}

// PDFGenerator defines an interface for PDF generation.
type PDFGenerator interface {
	GenerateInvoicePDF(inv *Invoice) (string, error)
}

type Service struct {
	repo      Repository
	orderSvc  OrderService
	partnerSvc PartnerService
	pdfGen    PDFGenerator
}

func NewService(repo Repository, orderSvc OrderService, partnerSvc PartnerService, pdfGen PDFGenerator) *Service {
	return &Service{
		repo:       repo,
		orderSvc:   orderSvc,
		partnerSvc: partnerSvc,
		pdfGen:     pdfGen,
	}
}

func (s *Service) SetOrderService(orderSvc OrderService) {
	s.orderSvc = orderSvc
}

func (s *Service) GenerateForOrder(ctx context.Context, orderID uuid.UUID) (*Invoice, error) {
	// 1. Fetch Order
	ord, err := s.orderSvc.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	if ord == nil {
		return nil, apperrors.NotFound("order", orderID.String())
	}

	// 2. Fetch Seller & Buyer for GSTIN & State Code
	seller, err := s.partnerSvc.GetPartner(ctx, ord.SellerID)
	if err != nil || seller == nil {
		return nil, fmt.Errorf("failed to get seller: %w", err)
	}
	buyer, err := s.partnerSvc.GetPartner(ctx, ord.BuyerID)
	if err != nil || buyer == nil {
		return nil, fmt.Errorf("failed to get buyer: %w", err)
	}

	sellerGSTIN := ""
	if seller.GSTIN != nil {
		sellerGSTIN = *seller.GSTIN
	}
	buyerGSTIN := ""
	if buyer.GSTIN != nil {
		buyerGSTIN = *buyer.GSTIN
	}
	placeOfSupply := ""
	if buyer.StateCode != nil {
		placeOfSupply = *buyer.StateCode
	}

	inv := &Invoice{
		OrderID:       ord.ID,
		SellerID:      ord.SellerID,
		BuyerID:       ord.BuyerID,
		InvoiceType:   "tax_invoice",
		SellerGSTIN:   sellerGSTIN,
		BuyerGSTIN:    buyerGSTIN,
		PlaceOfSupply: placeOfSupply,
		Subtotal:      ord.Subtotal,
		DiscountTotal: ord.DiscountTotal,
		TaxableAmount: ord.TaxableAmount,
		CGSTTotal:     ord.CGSTTotal,
		SGSTTotal:     ord.SGSTTotal,
		IGSTTotal:     ord.IGSTTotal,
		CessTotal:     "0.00",
		GrandTotal:    ord.GrandTotal,
	}

	for _, line := range ord.Lines {
		invLine := &InvoiceLine{
			ProductID:      line.ProductID,
			SKU:            line.SKU,
			Description:    line.ProductName,
			HSNCode:        line.HSNCode,
			Quantity:       line.OrderedQty,
			UnitPrice:      line.UnitPrice,
			DiscountAmount: line.DiscountAmount,
			TaxableAmount:  line.TaxableAmount,
			CGSTRate:       "TODO", // We'd ideally store rates in order line, but this is simplified
			CGSTAmount:     line.CGSTAmount,
			SGSTRate:       "TODO",
			SGSTAmount:     line.SGSTAmount,
			IGSTRate:       "TODO",
			IGSTAmount:     line.IGSTAmount,
			LineTotal:      line.LineTotal,
		}
		inv.Lines = append(inv.Lines, invLine)
	}

	err = db.WithTx(ctx, func(txCtx context.Context) error {
		invoiceNo, err := s.repo.GetNextInvoiceNumber(txCtx, ord.SellerID)
		if err != nil {
			return err
		}
		inv.InvoiceNumber = invoiceNo
		
		now := time.Now()
		inv.IssuedAt = &now
		inv.Status = "issued"
		
		// Due date (e.g. 30 days)
		dueDate := now.AddDate(0, 0, 30)
		inv.DueDate = &dueDate

		if err := s.repo.Create(txCtx, inv); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Generate PDF in goroutine
	go func() {
		bgCtx := context.Background()
		pdfURL, err := s.pdfGen.GenerateInvoicePDF(inv)
		if err != nil {
			// Log error (using fmt for simplicity here)
			fmt.Printf("Failed to generate PDF for invoice %s: %v\n", inv.ID, err)
			return
		}
		if err := s.repo.UpdatePDFURL(bgCtx, inv.ID, pdfURL); err != nil {
			fmt.Printf("Failed to update PDF URL for invoice %s: %v\n", inv.ID, err)
		}
	}()

	return inv, nil
}
