package integration

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"

	apperrors "github.com/agribid/agribid-shudh-backend/internal/errors"
	"github.com/agribid/agribid-shudh-backend/internal/logger"
	"github.com/agribid/agribid-shudh-backend/internal/response"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// WebhookHandler handles incoming webhooks from external providers.
type WebhookHandler struct {
	paymentWebhookSecret string
	deliveryWebhookSecret string
}

// NewWebhookHandler creates a new webhook handler.
func NewWebhookHandler(paymentSecret, deliverySecret string) *WebhookHandler {
	return &WebhookHandler{
		paymentWebhookSecret:  paymentSecret,
		deliveryWebhookSecret: deliverySecret,
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

	// TODO: Process payment confirmation → call payment.Service.RecordPayment

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

	// TODO: Process delivery update → call dispatch.Service.UpdateShipmentStatus

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
