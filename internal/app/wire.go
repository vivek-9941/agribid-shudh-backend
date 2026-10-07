package app

import (
	"context"

	"github.com/agribid/agribid-shudh-backend/internal/audit"
	"github.com/agribid/agribid-shudh-backend/internal/auth"
	"github.com/agribid/agribid-shudh-backend/internal/catalog"
	"github.com/agribid/agribid-shudh-backend/internal/dashboard"
	"github.com/agribid/agribid-shudh-backend/internal/dispatch"
	"github.com/agribid/agribid-shudh-backend/internal/fulfillment"
	"github.com/agribid/agribid-shudh-backend/internal/integration"
	"github.com/agribid/agribid-shudh-backend/internal/invoice"
	"github.com/agribid/agribid-shudh-backend/internal/notification"
	"github.com/agribid/agribid-shudh-backend/internal/order"
	"github.com/agribid/agribid-shudh-backend/internal/partner"
	"github.com/agribid/agribid-shudh-backend/internal/payment"
	"github.com/agribid/agribid-shudh-backend/internal/pricing"
	"github.com/agribid/agribid-shudh-backend/internal/returns"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type hsnAdapter struct {
	svc *invoice.HSNStubService
}

func (h *hsnAdapter) GetHSNRate(ctx context.Context, hsnCode string) (*pricing.HSNTaxRate, error) {
	r, err := h.svc.GetHSNRate(ctx, hsnCode)
	if err != nil {
		return nil, err
	}
	return &pricing.HSNTaxRate{
		HSNCode:  r.HSNCode,
		CGSTRate: r.CGSTRate,
		SGSTRate: r.SGSTRate,
		IGSTRate: r.IGSTRate,
		CessRate: r.CessRate,
	}, nil
}

// WireModules wires up all the modules and registers their routes.
func (a *App) WireModules(val *validator.Validate, logger *zap.Logger) {
	// Repositories
	authRepo := auth.NewRepository(a.Pool)
	partnerRepo := partner.NewRepository(a.Pool)
	catalogRepo := catalog.NewRepository(a.Pool)
	pricingRepo := pricing.NewRepository(a.Pool)
	orderRepo := order.NewRepository(a.Pool)
	fulfillmentRepo := fulfillment.NewRepository(a.Pool)
	invoiceRepo := invoice.NewRepository(a.Pool)
	dispatchRepo := dispatch.NewRepository(a.Pool)
	paymentRepo := payment.NewRepository(a.Pool)
	returnsRepo := returns.NewRepository(a.Pool)
	notifRepo := notification.NewRepository(a.Pool)
	dashRepo := dashboard.NewRepository(a.Pool)
	auditRepo := audit.NewRepository(a.Pool)

	// Gateways
	smsProv := integration.NewMockSMSProvider(logger)
	emailProv := integration.NewMockEmailProvider(logger)
	pdfGen := integration.NewMockPDFGenerator()

	// Services
	authSvc := auth.NewService(authRepo, nil, smsProv) // Config injected for tokens etc
	_ = authSvc
	partnerSvc := partner.NewService(partnerRepo)
	catalogSvc := catalog.NewService(catalogRepo)
	
	// Pricing needs catalog, partner, and invoice(hsn)
	invHSNSvc := invoice.NewHSNStubService() // We'd need an implementation
	hsnSvc := &hsnAdapter{svc: invHSNSvc}
	pricingSvc := pricing.NewService(pricingRepo, catalogSvc, partnerSvc, hsnSvc)
	
	paySvc := payment.NewService(paymentRepo, a.Pool)
	invcSvc := invoice.NewService(invoiceRepo, nil, partnerSvc, pdfGen, a.Pool)
	
	// Order needs inventory, payment, pricing, catalog
	orderSvc := order.NewService(orderRepo, nil, paySvc, pricingSvc, catalogSvc, a.Pool)
	
	// Set circular dependencies after init
	invcSvc.SetOrderService(orderSvc)
	
	fulfillmentSvc := fulfillment.NewService(fulfillmentRepo, orderRepo, nil, a.Pool)
	
	dispatchSvc := dispatch.NewService(dispatchRepo, fulfillmentSvc, a.Pool)
	returnsSvc := returns.NewService(returnsRepo, orderSvc, nil, invcSvc, paySvc, a.Pool)
	notifSvc := notification.NewService(notifRepo, emailProv, smsProv, logger)
	dashSvc := dashboard.NewService(dashRepo)

	// Middleware
	a.Router.Use(audit.Middleware(auditRepo))

	// Handlers
	// Auth
	authHnd := auth.NewHandler(authSvc, val)
	authHnd.RegisterRoutes(a.Router)

	// Partner
	partnerHnd := partner.NewHandler(partnerSvc)
	partnerHnd.RegisterRoutes(a.Router)

	// Catalog
	catalogHnd := catalog.NewHandler(catalogSvc)
	catalogHnd.RegisterRoutes(a.Router)
	
	// Pricing
	pricingHnd := pricing.NewHandler(pricingSvc, val)
	pricingHnd.RegisterRoutes(a.Router)

	// Order
	orderHnd := order.NewHandler(orderSvc, val)
	orderHnd.RegisterRoutes(a.Router)

	// Fulfillment
	fullHnd := fulfillment.NewHandler(fulfillmentSvc, val)
	fullHnd.RegisterRoutes(a.Router)

	// Invoice
	invcHnd := invoice.NewHandler(invcSvc, val)
	invcHnd.RegisterRoutes(a.Router)

	// Dispatch
	dispHnd := dispatch.NewHandler(dispatchSvc, val)
	dispHnd.RegisterRoutes(a.Router)

	// Payment
	payHnd := payment.NewHandler(paySvc, val)
	payHnd.RegisterRoutes(a.Router)

	// Returns
	retHnd := returns.NewHandler(returnsSvc, val)
	retHnd.RegisterRoutes(a.Router)

	// Notification
	notHnd := notification.NewHandler(notifSvc, val)
	notHnd.RegisterRoutes(a.Router)

	// Dashboard
	dashHnd := dashboard.NewHandler(dashSvc)
	dashHnd.RegisterRoutes(a.Router)
}
