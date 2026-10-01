package order

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/agribid/agribid-shudh-backend/internal/db"
	apperrors "github.com/agribid/agribid-shudh-backend/internal/errors"
	"github.com/agribid/agribid-shudh-backend/internal/pricing"
	"github.com/google/uuid"
)

type InventoryService interface {
	Reserve(ctx context.Context, productID uuid.UUID, qty int) error
}

type PaymentService interface {
	CheckCreditLimit(ctx context.Context, buyerID, sellerID uuid.UUID, amountStr string) error
	UpdateCredit(ctx context.Context, buyerID, sellerID uuid.UUID, amountStr string) error
}

type PricingService interface {
	CalculateLinePrice(ctx context.Context, productID uuid.UUID, qty int, buyerPartnerID uuid.UUID, buyerRole string) (*pricing.PriceResult, error)
}

type CatalogService interface {
	GetProduct(ctx context.Context, id uuid.UUID) (interface{}, error)
}

type Service struct {
	repo         Repository
	inventorySvc InventoryService
	paymentSvc   PaymentService
	pricingSvc   PricingService
	catalogSvc   CatalogService
}

func NewService(repo Repository, invSvc InventoryService, paySvc PaymentService, prSvc PricingService, catSvc CatalogService) *Service {
	return &Service{
		repo:         repo,
		inventorySvc: invSvc,
		paymentSvc:   paySvc,
		pricingSvc:   prSvc,
		catalogSvc:   catSvc,
	}
}

func (s *Service) CreateCart(ctx context.Context, buyerID uuid.UUID, req *CreateCartRequest) (*Cart, error) {
	sellerID, err := uuid.Parse(req.SellerID)
	if err != nil {
		return nil, apperrors.BadRequest("invalid seller_id")
	}
	cart := &Cart{
		BuyerID:  buyerID,
		SellerID: sellerID,
	}
	if err := s.repo.CreateCart(ctx, cart); err != nil {
		return nil, apperrors.Internal(err)
	}
	return cart, nil
}

func (s *Service) GetActiveCart(ctx context.Context, buyerID uuid.UUID) (*Cart, error) {
	cart, err := s.repo.GetActiveCart(ctx, buyerID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if cart == nil {
		return nil, apperrors.NotFound("cart", "active")
	}
	return cart, nil
}

func (s *Service) AddCartItem(ctx context.Context, buyerID uuid.UUID, req *AddCartItemRequest) (*Cart, error) {
	cart, err := s.repo.GetActiveCart(ctx, buyerID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if cart == nil {
		return nil, apperrors.BadRequest("no active cart found")
	}

	productID, _ := uuid.Parse(req.ProductID)
	item := &CartItem{
		CartID:    cart.ID,
		ProductID: productID,
		Quantity:  req.Quantity,
		UnitPrice: "0.00",
		LineTotal: "0.00",
	}

	if err := s.repo.AddItem(ctx, item); err != nil {
		return nil, apperrors.Internal(err)
	}
	return s.repo.GetActiveCart(ctx, buyerID)
}

func (s *Service) PlaceOrder(ctx context.Context, buyerPartnerID uuid.UUID, buyerRole string, req *PlaceOrderRequest) (*Order, error) {
	cart, err := s.repo.GetActiveCart(ctx, buyerPartnerID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if cart == nil || len(cart.Items) == 0 {
		return nil, apperrors.BadRequest("no active cart or cart is empty")
	}

	order := &Order{
		BuyerID:         buyerPartnerID,
		SellerID:        cart.SellerID,
		CartID:          &cart.ID,
		DeliveryAddress: req.DeliveryAddress,
	}
	if req.Notes != "" {
		order.Notes = &req.Notes
	}

	grandTotalRat := new(big.Rat).SetInt64(0)
	subtotalRat := new(big.Rat).SetInt64(0)
	discountRat := new(big.Rat).SetInt64(0)
	taxableRat := new(big.Rat).SetInt64(0)
	cgstRat := new(big.Rat).SetInt64(0)
	sgstRat := new(big.Rat).SetInt64(0)
	igstRat := new(big.Rat).SetInt64(0)

	// In a real app we'd lock here or let WithTx handle it
	err = db.WithTx(ctx, func(txCtx context.Context) error {
		for _, item := range cart.Items {
			// 1. Reserve Inventory
			if err := s.inventorySvc.Reserve(txCtx, item.ProductID, item.Quantity); err != nil {
				return fmt.Errorf("inventory reserve failed: %w", err)
			}

			// 2. Compute Price
			priceRes, err := s.pricingSvc.CalculateLinePrice(txCtx, item.ProductID, item.Quantity, buyerPartnerID, buyerRole)
			if err != nil {
				return fmt.Errorf("price calc failed: %w", err)
			}

			// accumulate totals
			bt, _ := new(big.Rat).SetString(priceRes.BasePrice)
			subtotalRat.Add(subtotalRat, new(big.Rat).Mul(bt, new(big.Rat).SetInt64(int64(item.Quantity))))
			
			dt, _ := new(big.Rat).SetString(priceRes.DiscountAmount)
			discountRat.Add(discountRat, dt)
			
			ta, _ := new(big.Rat).SetString(priceRes.TaxableAmount)
			taxableRat.Add(taxableRat, ta)
			
			c, _ := new(big.Rat).SetString(priceRes.CGSTAmount)
			cgstRat.Add(cgstRat, c)
			sAmt, _ := new(big.Rat).SetString(priceRes.SGSTAmount)
			sgstRat.Add(sgstRat, sAmt)
			i, _ := new(big.Rat).SetString(priceRes.IGSTAmount)
			igstRat.Add(igstRat, i)
			
			lt, _ := new(big.Rat).SetString(priceRes.LineTotal)
			grandTotalRat.Add(grandTotalRat, lt)

			line := &OrderLine{
				ProductID:      item.ProductID,
				SKU:            "TODO-SKU",
				ProductName:    "TODO-NAME",
				OrderedQty:     item.Quantity,
				UnitPrice:      priceRes.BasePrice,
				DiscountAmount: priceRes.DiscountAmount,
				TaxableAmount:  priceRes.TaxableAmount,
				CGSTAmount:     priceRes.CGSTAmount,
				SGSTAmount:     priceRes.SGSTAmount,
				IGSTAmount:     priceRes.IGSTAmount,
				LineTotal:      priceRes.LineTotal,
			}
			order.Lines = append(order.Lines, line)
		}

		order.Subtotal = formatDecimal(subtotalRat)
		order.DiscountTotal = formatDecimal(discountRat)
		order.TaxableAmount = formatDecimal(taxableRat)
		order.CGSTTotal = formatDecimal(cgstRat)
		order.SGSTTotal = formatDecimal(sgstRat)
		order.IGSTTotal = formatDecimal(igstRat)
		order.GrandTotal = formatDecimal(grandTotalRat)

		// 3. Check and Update Credit
		if err := s.paymentSvc.CheckCreditLimit(txCtx, order.BuyerID, order.SellerID, order.GrandTotal); err != nil {
			return fmt.Errorf("credit check failed: %w", err)
		}
		if err := s.paymentSvc.UpdateCredit(txCtx, order.BuyerID, order.SellerID, order.GrandTotal); err != nil {
			return fmt.Errorf("credit update failed: %w", err)
		}

		// 4. Create Order
		if err := s.repo.CreateOrder(txCtx, order); err != nil {
			return fmt.Errorf("create order failed: %w", err)
		}

		// 5. Update Cart
		if err := s.repo.UpdateCartStatus(txCtx, cart.ID, "ordered"); err != nil {
			return fmt.Errorf("update cart failed: %w", err)
		}

		// Add history
		hist := &OrderStatusHistory{
			OrderID:    order.ID,
			FromStatus: "",
			ToStatus:   string(StatusPending),
			ChangedBy:  &buyerPartnerID,
			ChangedAt:  time.Now(),
		}
		s.repo.AddOrderStatusHistory(txCtx, hist)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return order, nil
}

func formatDecimal(r *big.Rat) string {
	f, _ := r.Float64()
	return fmt.Sprintf("%.2f", f)
}
