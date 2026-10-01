package pricing

import (
	"encoding/json"
	"fmt"
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
	r.Post("/products/{id}/prices", h.setProductPrice)
	
	r.Route("/schemes", func(r chi.Router) {
		r.Post("/", h.createScheme)
		// r.Put("/{id}", h.updateScheme)
		// r.Patch("/{id}/status", h.setSchemeStatus)
	})
	
	// Testing route for calculate
	r.Get("/products/{id}/calculate-price", h.calculatePrice)
}

func (h *Handler) setProductPrice(w http.ResponseWriter, r *http.Request) {
	productIDStr := chi.URLParam(r, "id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		response.Error(w, err) // map to bad request
		return
	}

	var req SetPriceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, err)
		return
	}
	if err := h.val.Struct(req); err != nil {
		response.Error(w, err)
		return
	}

	price, err := h.svc.SetProductPrice(r.Context(), &req, productID)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, price)
}

func (h *Handler) createScheme(w http.ResponseWriter, r *http.Request) {
	var req CreateSchemeRequest
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

	scheme, err := h.svc.CreateScheme(r.Context(), userID, &req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, scheme)
}

func (h *Handler) calculatePrice(w http.ResponseWriter, r *http.Request) {
	productIDStr := chi.URLParam(r, "id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		response.Error(w, err)
		return
	}
	
	qtyStr := r.URL.Query().Get("qty")
	if qtyStr == "" {
		qtyStr = "1"
	}
	qty := 1
	fmt.Sscanf(qtyStr, "%d", &qty)
	
	buyerPartnerIDStr := r.URL.Query().Get("buyer_partner_id")
	buyerPartnerID, _ := uuid.Parse(buyerPartnerIDStr)
	buyerRole := r.URL.Query().Get("buyer_role")
	
	res, err := h.svc.CalculateLinePrice(r.Context(), productID, qty, buyerPartnerID, buyerRole)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, res)
}
