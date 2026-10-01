package fulfillment

import (
	"encoding/json"
	"net/http"

	"github.com/agribid/agribid-shudh-backend/internal/auth"
	"github.com/agribid/agribid-shudh-backend/internal/order"
	"github.com/agribid/agribid-shudh-backend/internal/response"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Handler struct {
	svc *Service
	val *validator.Validate
}

func NewHandler(svc *Service, val *validator.Validate) *Handler {
	return &Handler{
		svc: svc,
		val: val,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/fulfillment/orders", func(r chi.Router) {
		r.Patch("/{id}/status", h.updateStatus)
		r.Post("/{id}/partial", h.recordPartial)
	})
}

func (h *Handler) updateStatus(w http.ResponseWriter, r *http.Request) {
	orderIDStr := chi.URLParam(r, "id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		response.Error(w, err)
		return
	}

	var req order.UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, err)
		return
	}
	if err := h.val.Struct(req); err != nil {
		response.Error(w, err)
		return
	}

	userCtx, _ := r.Context().Value("user").(*auth.UserContext)
	userID := uuid.Nil
	if userCtx != nil {
		userID = userCtx.ID
	}

	if err := h.svc.UpdateOrderStatus(r.Context(), orderID, userID, req.Status, req.Reason); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "status updated successfully"})
}

func (h *Handler) recordPartial(w http.ResponseWriter, r *http.Request) {
	orderIDStr := chi.URLParam(r, "id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		response.Error(w, err)
		return
	}

	var req RecordPartialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, err)
		return
	}
	if err := h.val.Struct(req); err != nil {
		response.Error(w, err)
		return
	}

	userCtx, _ := r.Context().Value("user").(*auth.UserContext)
	userID := uuid.Nil
	if userCtx != nil {
		userID = userCtx.ID
	}

	if err := h.svc.RecordPartialFulfillment(r.Context(), orderID, userID, &req); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "partial fulfillment recorded"})
}
