package dashboard

import (
	"context"

	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetDashboardMetrics(ctx context.Context, partnerID *uuid.UUID, role string) (map[string]interface{}, error) {
	// If admin, partnerID might be nil
	
	sales, err := s.repo.GetSalesSummary(ctx, partnerID)
	if err != nil {
		return nil, err
	}
	
	pending, err := s.repo.GetPendingOrders(ctx, partnerID)
	if err != nil {
		return nil, err
	}
	
	lowStock, err := s.repo.GetLowStockCount(ctx, partnerID)
	if err != nil {
		return nil, err
	}
	
	outstanding, err := s.repo.GetOutstandingPayments(ctx, partnerID)
	if err != nil {
		return nil, err
	}
	
	topProducts, err := s.repo.GetTopProducts(ctx, partnerID, 5)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"sales_summary":        sales,
		"pending_orders":       pending,
		"low_stock_count":      lowStock,
		"outstanding_payments": outstanding,
		"top_products":         topProducts,
	}, nil
}
