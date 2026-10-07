package returns

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
	r.Route("/returns", func(r chi.Router) {
		r.Post("/", h.initiateReturn)
		r.Patch("/{id}/status", h.updateStatus)
		r.Post("/{id}/inspection", h.recordInspection)
	})
}

func (h *Handler) initiateReturn(w http.ResponseWriter, r *http.Request) {
	var req InitiateReturnRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, err)
		return
	}
	if err := h.val.Struct(req); err != nil {
		response.Error(w, err)
		return
	}

	userCtx, _ := r.Context().Value("user").(*auth.UserContext)
	buyerID := uuid.Nil
	if userCtx != nil {
		buyerID = userCtx.PartnerID
	}

	retReq, err := h.svc.Initiate(r.Context(), buyerID, &req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, retReq)
}

func (h *Handler) updateStatus(w http.ResponseWriter, r *http.Request) {
	returnIDStr := chi.URLParam(r, "id")
	returnID, err := uuid.Parse(returnIDStr)
	if err != nil {
		response.Error(w, err)
		return
	}

	var req struct {
		Status string `json:"status" validate:"required,oneof=approved rejected"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, err)
		return
	}

	if req.Status == "approved" {
		err = h.svc.Approve(r.Context(), returnID)
	} else {
		err = h.svc.Reject(r.Context(), returnID)
	}

	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "status updated"})
}

func (h *Handler) recordInspection(w http.ResponseWriter, r *http.Request) {
	returnIDStr := chi.URLParam(r, "id")
	returnID, err := uuid.Parse(returnIDStr)
	if err != nil {
		response.Error(w, err)
		return
	}

	var req InspectionInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, err)
		return
	}
	if err := h.val.Struct(req); err != nil {
		response.Error(w, err)
		return
	}

	if err := h.svc.RecordInspection(r.Context(), returnID, &req); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "inspection recorded"})
}
