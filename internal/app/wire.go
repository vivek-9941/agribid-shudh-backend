package app

import (
	"context"
	"fmt"

	"github.com/agribid/agribid-shudh-backend/internal/admin"
	// "github.com/agribid/agribid-shudh-backend/internal/audit"  // Not needed unless adding audit endpoints
	"github.com/agribid/agribid-shudh-backend/internal/auth"
	"github.com/agribid/agribid-shudh-backend/internal/catalog"
	"github.com/agribid/agribid-shudh-backend/internal/dashboard"
	"github.com/agribid/agribid-shudh-backend/internal/dispatch"
	"github.com/agribid/agribid-shudh-backend/internal/fulfillment"
	"github.com/agribid/agribid-shudh-backend/internal/integration"
	"github.com/agribid/agribid-shudh-backend/internal/integration/sms"
	"github.com/agribid/agribid-shudh-backend/internal/inventory"
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
	// ─────────────────────────────────────────────────────────────────
	// Redis OTP store
	// ─────────────────────────────────────────────────────────────────
	redisClient := auth.NewRedisClient(
		a.Config.RedisHost,
		a.Config.RedisPort,
		a.Config.RedisDB,
		a.Config.RedisPass,
	)
	if err := auth.PingRedis(context.Background(), redisClient); err != nil {
		logger.Warn("Redis not reachable – OTP flow will be unavailable", zap.Error(err))
	} else {
		logger.Info("Redis connected",
			zap.String("host", a.Config.RedisHost),
			zap.Int("port", a.Config.RedisPort),
			zap.Int("db", a.Config.RedisDB),
		)
	}
	otpStore := auth.NewRedisOTPStore(redisClient)

	// ─────────────────────────────────────────────────────────────────
	// SMS provider – use real gateway when credentials are configured,
	// fall back to mock logger in dev/test.
	// ─────────────────────────────────────────────────────────────────
	var smsProv auth.SMSProvider
	if a.Config.SMSAPIURL != "" && a.Config.SMSAPIUsername != "" {
		smsProv = sms.New(sms.Config{
			APIURL:       a.Config.SMSAPIURL,
			Username:     a.Config.SMSAPIUsername,
			Password:     a.Config.SMSAPIPassword,
			SenderID:     a.Config.SMSAPISenderID,
			DLTContentID: a.Config.SMSDLTContentID,
			DLTPEiD:      a.Config.SMSDLTPEiD,
		})
		logger.Info("SMS gateway configured",
			zap.String("url", a.Config.SMSAPIURL),
			zap.String("sender", a.Config.SMSAPISenderID),
		)
	} else {
		smsProv = integration.NewMockSMSProvider(logger)
		logger.Warn("SMS credentials not set – using mock SMS provider (OTPs will only appear in logs)")
	}

	// ─────────────────────────────────────────────────────────────────
	// Repositories
	// ─────────────────────────────────────────────────────────────────
	authRepo := auth.NewRepository(a.Pool)
	partnerRepo := partner.NewRepository(a.Pool)
	catalogRepo := catalog.NewRepository(a.Pool)
	pricingRepo := pricing.NewRepository(a.Pool)
	orderRepo := order.NewRepository(a.Pool)
	fulfillmentRepo := fulfillment.NewRepository(a.Pool)
	invoiceRepo := invoice.NewRepository(a.Pool)
	dispatchRepo := dispatch.NewRepository(a.Pool)
	inventoryRepo := inventory.NewRepository(a.Pool)
	paymentRepo := payment.NewRepository(a.Pool)
	returnsRepo := returns.NewRepository(a.Pool)
	notifRepo := notification.NewRepository(a.Pool)
	dashRepo := dashboard.NewRepository(a.Pool)

	// ─────────────────────────────────────────────────────────────────
	// Gateways (email / PDF / etc.)
	// ─────────────────────────────────────────────────────────────────
	emailProv := integration.NewMockEmailProvider(logger)
	pdfGen := integration.NewMockPDFGenerator()

	// ─────────────────────────────────────────────────────────────────
	// Services
	// ─────────────────────────────────────────────────────────────────
	jwtSvc, err := auth.NewJWTService(
		a.Config.JWTPrivateKeyPath,
		a.Config.JWTPublicKeyPath,
		a.Config.JWTAccessTTLHours,
		a.Config.JWTRefreshTTLDays,
	)
	if err != nil {
		logger.Fatal("failed to initialize JWT service", zap.Error(err))
	}

	// Auth service now receives the Redis OTP store.
	authSvc := auth.NewService(authRepo, jwtSvc, smsProv, otpStore)

	partnerSvc := partner.NewService(partnerRepo)
	catalogSvc := catalog.NewService(catalogRepo)

	inventorySvc := inventory.NewService(inventoryRepo, a.Pool)
	orderInvAdapter := inventory.NewOrderInventoryAdapter(inventorySvc, a.Pool)
	fulfillmentInvAdapter := inventory.NewFulfillmentInventoryAdapter(inventorySvc, a.Pool)
	returnsInvAdapter := inventory.NewReturnsInventoryAdapter(inventorySvc, a.Pool)

	invHSNSvc := invoice.NewHSNStubService()
	hsnSvc := &hsnAdapter{svc: invHSNSvc}
	pricingSvc := pricing.NewService(pricingRepo, catalogSvc, partnerSvc, hsnSvc)

	paySvc := payment.NewService(paymentRepo, a.Pool)
	invcSvc := invoice.NewService(invoiceRepo, nil, partnerSvc, pdfGen, a.Pool)

	orderSvc := order.NewService(orderRepo, orderInvAdapter, paySvc, pricingSvc, catalogSvc, a.Pool)
	invcSvc.SetOrderService(orderSvc)

	fulfillmentSvc := fulfillment.NewService(fulfillmentRepo, orderRepo, fulfillmentInvAdapter, a.Pool)
	dispatchSvc := dispatch.NewService(dispatchRepo, fulfillmentSvc, a.Pool)
	returnsSvc := returns.NewService(returnsRepo, orderSvc, returnsInvAdapter, invcSvc, paySvc, a.Pool)
	notifSvc := notification.NewService(notifRepo, emailProv, smsProv, logger)
	dashSvc := dashboard.NewService(dashRepo)
	adminSvc := admin.NewService()

	// Suppress unused import warning if SMS is mocked and notification uses smsProv.
	_ = fmt.Sprintf("smsProv type: %T", smsProv)

	// ─────────────────────────────────────────────────────────────────
	// Route registration
	// ─────────────────────────────────────────────────────────────────
	authHnd := auth.NewHandler(authSvc, val)
	authHnd.RegisterRoutes(a.Router)

	partnerHnd := partner.NewHandler(partnerSvc)
	partnerHnd.RegisterRoutes(a.Router)

	catalogHnd := catalog.NewHandler(catalogSvc)
	catalogHnd.RegisterRoutes(a.Router)

	pricingHnd := pricing.NewHandler(pricingSvc, val)
	pricingHnd.RegisterRoutes(a.Router)

	inventoryHnd := inventory.NewHandler(inventorySvc)
	inventoryHnd.RegisterRoutes(a.Router)

	orderHnd := order.NewHandler(orderSvc, val)
	orderHnd.RegisterRoutes(a.Router)

	fullHnd := fulfillment.NewHandler(fulfillmentSvc, val)
	fullHnd.RegisterRoutes(a.Router)

	invcHnd := invoice.NewHandler(invcSvc, val)
	invcHnd.RegisterRoutes(a.Router)

	dispHnd := dispatch.NewHandler(dispatchSvc, val)
	dispHnd.RegisterRoutes(a.Router)

	payHnd := payment.NewHandler(paySvc, val)
	payHnd.RegisterRoutes(a.Router)

	retHnd := returns.NewHandler(returnsSvc, val)
	retHnd.RegisterRoutes(a.Router)

	notHnd := notification.NewHandler(notifSvc, val)
	notHnd.RegisterRoutes(a.Router)

	dashHnd := dashboard.NewHandler(dashSvc)
	dashHnd.RegisterRoutes(a.Router)

	adminHnd := admin.NewHandler(adminSvc)
	adminHnd.RegisterRoutes(a.Router)
}
