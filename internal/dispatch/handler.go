package dispatch

import (
	"encoding/json"
	"net/http"

	"github.com/agribid/agribid-shudh-backend/internal/auth"
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
	r.Route("/shipments", func(r chi.Router) {
		r.Post("/", h.assignDeliveryPartner)
		r.Patch("/{id}/status", h.updateStatus)
		r.Post("/{id}/pod", h.recordPOD)
	})
}

func (h *Handler) assignDeliveryPartner(w http.ResponseWriter, r *http.Request) {
	var req AssignPartnerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, err)
		return
	}
	if err := h.val.Struct(req); err != nil {
		response.Error(w, err)
		return
	}

	userCtx, _ := r.Context().Value("user").(*auth.UserContext)
	sellerID := uuid.Nil
	if userCtx != nil {
		sellerID = userCtx.PartnerID
	}
	orderID, _ := uuid.Parse(req.OrderID)

	shipment, err := h.svc.AssignDeliveryPartner(r.Context(), &req, orderID, sellerID)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, shipment)
}

func (h *Handler) updateStatus(w http.ResponseWriter, r *http.Request) {
	shipmentIDStr := chi.URLParam(r, "id")
	shipmentID, err := uuid.Parse(shipmentIDStr)
	if err != nil {
		response.Error(w, err)
		return
	}

	var req UpdateStatusRequest
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

	if err := h.svc.UpdateShipmentStatus(r.Context(), &req, shipmentID, userID); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "status updated successfully"})
}

func (h *Handler) recordPOD(w http.ResponseWriter, r *http.Request) {
	shipmentIDStr := chi.URLParam(r, "id")
	shipmentID, err := uuid.Parse(shipmentIDStr)
	if err != nil {
		response.Error(w, err)
		return
	}

	var req RecordPODRequest
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

	if err := h.svc.RecordPOD(r.Context(), &req, shipmentID, userID); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "POD recorded successfully"})
}
