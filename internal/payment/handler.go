package payment

import (
	"encoding/json"
	"net/http"

	apperrors "github.com/agribid/agribid-shudh-backend/internal/errors"
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
	r.Route("/payments", func(r chi.Router) {
		r.Post("/", h.recordPayment)
		r.Get("/ledger", h.getLedger)
		r.Get("/aging", h.getAgingReport)
		r.Get("/credit-status", h.getCreditStatus)
	})
}

func (h *Handler) recordPayment(w http.ResponseWriter, r *http.Request) {
	var req RecordPaymentRequest
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

	pay, err := h.svc.RecordPayment(r.Context(), &req, buyerID)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, pay)
}

func (h *Handler) getLedger(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := r.Context().Value("user").(*auth.UserContext)
	partnerID := uuid.Nil
	if userCtx != nil {
		partnerID = userCtx.PartnerID
	}

	entries, _, err := h.svc.GetLedger(r.Context(), partnerID, 1, 50)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, entries)
}

func (h *Handler) getAgingReport(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := r.Context().Value("user").(*auth.UserContext)
	partnerID := uuid.Nil
	if userCtx != nil {
		partnerID = userCtx.PartnerID
	}

	report, err := h.svc.GetAgingReport(r.Context(), partnerID)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, report)
}

func (h *Handler) getCreditStatus(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := r.Context().Value("user").(*auth.UserContext)
	buyerID := uuid.Nil
	if userCtx != nil {
		buyerID = userCtx.PartnerID
	}
	
	sellerIDStr := r.URL.Query().Get("seller_id")
	sellerID, err := uuid.Parse(sellerIDStr)
	if err != nil {
		response.Error(w, apperrors.BadRequest("invalid seller_id"))
		return
	}

	status, err := h.svc.GetCreditStatus(r.Context(), buyerID, sellerID)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, status)
}
