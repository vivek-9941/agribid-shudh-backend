package integration

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/agribid/agribid-shudh-backend/internal/dispatch"
	apperrors "github.com/agribid/agribid-shudh-backend/internal/errors"
	"github.com/agribid/agribid-shudh-backend/internal/logger"
	"github.com/agribid/agribid-shudh-backend/internal/payment"
	"github.com/agribid/agribid-shudh-backend/internal/response"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type PaymentService interface {
	RecordPayment(ctx context.Context, req *payment.RecordPaymentRequest, buyerID uuid.UUID) (*payment.Payment, error)
}

type DispatchService interface {
	UpdateShipmentStatus(ctx context.Context, req *dispatch.UpdateStatusRequest, shipmentID, userID uuid.UUID) error
}

type WebhookHandler struct {
	paymentWebhookSecret  string
	deliveryWebhookSecret string
	paymentSvc            PaymentService
	dispatchSvc           DispatchService
}

// NewWebhookHandler creates a new webhook handler.
func NewWebhookHandler(paymentSecret, deliverySecret string, paymentSvc PaymentService, dispatchSvc DispatchService) *WebhookHandler {
	return &WebhookHandler{
		paymentWebhookSecret:  paymentSecret,
		deliveryWebhookSecret: deliverySecret,
		paymentSvc:            paymentSvc,
		dispatchSvc:           dispatchSvc,
	}
}

// RegisterRoutes registers webhook routes.
func (h *WebhookHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/webhooks", func(r chi.Router) {
		r.Post("/payment", h.HandlePaymentWebhook)
		r.Post("/delivery", h.HandleDeliveryWebhook)
	})
}

// HandlePaymentWebhook handles POST /api/v1/webhooks/payment.
func (h *WebhookHandler) HandlePaymentWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, apperrors.BadRequest("WEBHOOK_ERROR", "failed to read request body"))
		return
	}

	// Validate HMAC signature.
	signature := r.Header.Get("X-Webhook-Signature")
	if !h.verifySignature(body, signature, h.paymentWebhookSecret) {
		response.Error(w, apperrors.Unauthorized("invalid webhook signature"))
		return
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		response.Error(w, apperrors.BadRequest("WEBHOOK_ERROR", "invalid JSON payload"))
		return
	}

	logger.Get().Info("payment webhook received",
		zap.Any("payload", payload),
	)

	// Process payment confirmation → call payment.Service.RecordPayment
	if pld, ok := payload["payload"].(map[string]interface{}); ok {
		if payInfo, ok := pld["payment"].(map[string]interface{}); ok {
			if entity, ok := payInfo["entity"].(map[string]interface{}); ok {
				notes, _ := entity["notes"].(map[string]interface{})
				if notes != nil {
					buyerIDStr, _ := notes["buyer_id"].(string)
					sellerIDStr, _ := notes["seller_id"].(string)
					amountF, _ := entity["amount"].(float64)
					amountStr := fmt.Sprintf("%.2f", amountF/100)
					refNum, _ := entity["id"].(string)
					method, _ := entity["method"].(string)
					
					buyerID, err := uuid.Parse(buyerIDStr)
					if err == nil && h.paymentSvc != nil {
						req := &payment.RecordPaymentRequest{
							SellerID:        sellerIDStr,
							Amount:          amountStr,
							Method:          method,
							ReferenceNumber: refNum,
						}
						h.paymentSvc.RecordPayment(r.Context(), req, buyerID)
					}
				}
			}
		}
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "received"})
}

// HandleDeliveryWebhook handles POST /api/v1/webhooks/delivery.
func (h *WebhookHandler) HandleDeliveryWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, apperrors.BadRequest("WEBHOOK_ERROR", "failed to read request body"))
		return
	}

	signature := r.Header.Get("X-Webhook-Signature")
	if !h.verifySignature(body, signature, h.deliveryWebhookSecret) {
		response.Error(w, apperrors.Unauthorized("invalid webhook signature"))
		return
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		response.Error(w, apperrors.BadRequest("WEBHOOK_ERROR", "invalid JSON payload"))
		return
	}

	logger.Get().Info("delivery webhook received",
		zap.Any("payload", payload),
	)

	// Process delivery update → call dispatch.Service.UpdateShipmentStatus
	if shipmentIDStr, ok := payload["shipment_id"].(string); ok {
		shipmentID, err := uuid.Parse(shipmentIDStr)
		if err == nil && h.dispatchSvc != nil {
			status, _ := payload["status"].(string)
			req := &dispatch.UpdateStatusRequest{
				Status: status,
			}
			h.dispatchSvc.UpdateShipmentStatus(r.Context(), req, shipmentID, uuid.Nil)
		}
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "received"})
}

// verifySignature validates an HMAC-SHA256 webhook signature.
func (h *WebhookHandler) verifySignature(body []byte, signature, secret string) bool {
	if secret == "" || signature == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}
