package pricing

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/agribid/agribid-shudh-backend/internal/catalog"
	apperrors "github.com/agribid/agribid-shudh-backend/internal/errors"
	"github.com/agribid/agribid-shudh-backend/internal/partner"
	"github.com/google/uuid"
)

type CatalogService interface {
	GetProduct(ctx context.Context, id uuid.UUID) (*catalog.Product, error)
}

type PartnerService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*partner.Partner, error)
}

type HSNTaxRate struct {
	HSNCode  string
	CGSTRate string
	SGSTRate string
	IGSTRate string
	CessRate string
}

type HSNSvc interface {
	GetHSNRate(ctx context.Context, hsnCode string) (*HSNTaxRate, error)
}

type Service struct {
	repo       Repository
	catalogSvc CatalogService
	partnerSvc PartnerService
	hsnSvc     HSNSvc
}

func NewService(repo Repository, catalogSvc CatalogService, partnerSvc PartnerService, hsnSvc HSNSvc) *Service {
	return &Service{
		repo:       repo,
		catalogSvc: catalogSvc,
		partnerSvc: partnerSvc,
		hsnSvc:     hsnSvc,
	}
}

// CalculateLinePrice computes the final price for a line item applying schemes and taxes.
func (s *Service) CalculateLinePrice(ctx context.Context, productID uuid.UUID, qty int, buyerPartnerID uuid.UUID, buyerRole string) (*PriceResult, error) {
	date := time.Now()
	
	// 1. Get base price
	price, err := s.repo.GetPriceForRole(ctx, productID, &buyerPartnerID, buyerRole, date)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if price == nil {
		return nil, apperrors.NotFound("product_price", productID.String())
	}

	// 2. Get applicable schemes
	schemes, err := s.repo.GetApplicableSchemes(ctx, productID, buyerPartnerID, buyerRole, date)
	if err != nil {
		return nil, apperrors.Internal(err)
	}

	// 3. Evaluate schemes (simplified: just taking first applicable for now, or exclusive)
	basePriceRat, _ := new(big.Rat).SetString(price.BasePrice)
	qtyRat := new(big.Rat).SetInt64(int64(qty))
	
	totalAmountRat := new(big.Rat).Mul(basePriceRat, qtyRat)
	discountAmountRat := new(big.Rat).SetInt64(0)
	var appliedSchemes []uuid.UUID

	for _, scheme := range schemes {
		// apply scheme
		if scheme.Type == "percentage" && scheme.DiscountValue != "" {
			pct, _ := new(big.Rat).SetString(scheme.DiscountValue)
			hundred := new(big.Rat).SetInt64(100)
			
			discount := new(big.Rat).Mul(totalAmountRat, pct)
			discount.Quo(discount, hundred)
			discountAmountRat.Add(discountAmountRat, discount)
			appliedSchemes = append(appliedSchemes, scheme.ID)
		} else if scheme.Type == "flat" && scheme.DiscountValue != "" {
			flat, _ := new(big.Rat).SetString(scheme.DiscountValue)
			discountAmountRat.Add(discountAmountRat, flat)
			appliedSchemes = append(appliedSchemes, scheme.ID)
		}
		
		if scheme.IsExclusive {
			break
		}
	}

	taxableAmountRat := new(big.Rat).Sub(totalAmountRat, discountAmountRat)
	if taxableAmountRat.Sign() < 0 {
		taxableAmountRat.SetInt64(0)
	}

	// 4. Calculate Taxes
	// Get product to get HSN Code
	prod, err := s.catalogSvc.GetProduct(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	
	// Get buyer to get state code
	buyer, err := s.partnerSvc.GetByID(ctx, buyerPartnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get buyer: %w", err)
	}
	
	// Get seller (parent partner) to get state code.
	// For simplicity in this implementation, we assume parent is seller.
	var sellerStateCode string
	if buyer.ParentID != nil {
		seller, err := s.partnerSvc.GetByID(ctx, *buyer.ParentID)
		if err == nil && seller != nil && seller.StateCode != nil {
			sellerStateCode = *seller.StateCode
		}
	}
	var buyerStateCode string
	if buyer.StateCode != nil {
		buyerStateCode = *buyer.StateCode
	}
	
	hsnRate, err := s.hsnSvc.GetHSNRate(ctx, prod.HSNCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get HSN rate: %w", err)
	}
	
	taxableStr := formatDecimal(taxableAmountRat)
	cgstRate, sgstRate, igstRate := "0.00", "0.00", "0.00"
	cgstAmtRat, sgstAmtRat, igstAmtRat := new(big.Rat).SetInt64(0), new(big.Rat).SetInt64(0), new(big.Rat).SetInt64(0)
	hundred := new(big.Rat).SetInt64(100)

	if hsnRate != nil {
		if sellerStateCode != "" && sellerStateCode == buyerStateCode {
			cgstRate = hsnRate.CGSTRate
			sgstRate = hsnRate.SGSTRate
			rC, _ := new(big.Rat).SetString(cgstRate)
			cgstAmtRat = new(big.Rat).Quo(new(big.Rat).Mul(taxableAmountRat, rC), hundred)
			rS, _ := new(big.Rat).SetString(sgstRate)
			sgstAmtRat = new(big.Rat).Quo(new(big.Rat).Mul(taxableAmountRat, rS), hundred)
		} else {
			igstRate = hsnRate.IGSTRate
			rI, _ := new(big.Rat).SetString(igstRate)
			igstAmtRat = new(big.Rat).Quo(new(big.Rat).Mul(taxableAmountRat, rI), hundred)
		}
	}

	res := &PriceResult{
		BasePrice:      price.BasePrice,
		DiscountAmount: formatDecimal(discountAmountRat),
		TaxableAmount:  taxableStr,
		CGSTRate:       cgstRate,
		CGSTAmount:     formatDecimal(cgstAmtRat),
		SGSTRate:       sgstRate,
		SGSTAmount:     formatDecimal(sgstAmtRat),
		IGSTRate:       igstRate,
		IGSTAmount:     formatDecimal(igstAmtRat),
		AppliedSchemes: appliedSchemes,
	}

	totalTaxRat := new(big.Rat).Add(cgstAmtRat, sgstAmtRat)
	totalTaxRat.Add(totalTaxRat, igstAmtRat)
	lineTotalRat := new(big.Rat).Add(taxableAmountRat, totalTaxRat)
	res.LineTotal = formatDecimal(lineTotalRat)

	return res, nil
}

func formatDecimal(r *big.Rat) string {
	f, _ := r.Float64()
	return fmt.Sprintf("%.2f", f)
}

func (s *Service) SetProductPrice(ctx context.Context, req *SetPriceRequest, productID uuid.UUID) (*ProductPrice, error) {
	effectiveFrom, err := time.Parse(time.RFC3339, req.EffectiveFrom)
	if err != nil {
		return nil, apperrors.BadRequest(apperrors.CodeValidationFailed, "invalid effective_from format")
	}
	var effectiveTo *time.Time
	if req.EffectiveTo != "" {
		et, err := time.Parse(time.RFC3339, req.EffectiveTo)
		if err != nil {
			return nil, apperrors.BadRequest(apperrors.CodeValidationFailed, "invalid effective_to format")
		}
		effectiveTo = &et
	}
	var partnerID *uuid.UUID
	if req.PartnerID != "" {
		pid, _ := uuid.Parse(req.PartnerID)
		partnerID = &pid
	}

	price := &ProductPrice{
		ProductID:     productID,
		RoleCode:      req.RoleCode,
		PartnerID:     partnerID,
		MRP:           req.MRP,
		BasePrice:     req.BasePrice,
		EffectiveFrom: effectiveFrom,
		EffectiveTo:   effectiveTo,
	}

	if err := s.repo.SetProductPrice(ctx, price); err != nil {
		return nil, apperrors.Internal(err)
	}
	return price, nil
}

func (s *Service) CreateScheme(ctx context.Context, userID uuid.UUID, req *CreateSchemeRequest) (*Scheme, error) {
	vf, err := time.Parse(time.RFC3339, req.ValidFrom)
	if err != nil {
		return nil, apperrors.BadRequest(apperrors.CodeValidationFailed, "invalid valid_from")
	}
	vt, err := time.Parse(time.RFC3339, req.ValidTo)
	if err != nil {
		return nil, apperrors.BadRequest(apperrors.CodeValidationFailed, "invalid valid_to")
	}

	scheme := &Scheme{
		Name:          req.Name,
		Type:          req.Type,
		DiscountValue: req.DiscountValue,
		BuyQty:        req.BuyQty,
		GetQty:        req.GetQty,
		ValidFrom:     vf,
		ValidTo:       vt,
		IsExclusive:   req.IsExclusive,
		CreatedBy:     userID,
		Status:        "active",
	}

	if err := s.repo.CreateScheme(ctx, scheme); err != nil {
		return nil, apperrors.Internal(err)
	}

	for _, target := range req.Targets {
		app := &SchemeApplicability{
			SchemeID:   scheme.ID,
			TargetType: target.TargetType,
			TargetID:   target.TargetID,
		}
		if err := s.repo.AddSchemeApplicability(ctx, app); err != nil {
			return nil, apperrors.Internal(err)
		}
	}

	return scheme, nil
}
