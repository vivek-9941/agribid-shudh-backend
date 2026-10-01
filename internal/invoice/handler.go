package invoice

import (
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
	r.Route("/invoices", func(r chi.Router) {
		r.Get("/{id}", h.getInvoice)
		r.Post("/order/{orderId}", h.generateForOrder) // Typically internal or automated, but exposed for testing
	})
}

func (h *Handler) generateForOrder(w http.ResponseWriter, r *http.Request) {
	orderIDStr := chi.URLParam(r, "orderId")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		response.Error(w, err)
		return
	}

	inv, err := h.svc.GenerateForOrder(r.Context(), orderID)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, inv)
}

func (h *Handler) getInvoice(w http.ResponseWriter, r *http.Request) {
	// Stub for getting an invoice
	userCtx, _ := r.Context().Value("user").(*auth.UserContext)
	_ = userCtx
	
	// Not fully implemented, returns 501 Not Implemented
	response.JSON(w, http.StatusNotImplemented, map[string]string{"message": "Not implemented"})
}
