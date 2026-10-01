package order

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
	r.Route("/cart", func(r chi.Router) {
		r.Post("/", h.createCart)
		r.Get("/", h.getCart)
		// r.Delete("/", h.deleteCart)

		r.Route("/items", func(r chi.Router) {
			r.Post("/", h.addCartItem)
			// r.Put("/{id}", h.updateCartItem)
			// r.Delete("/{id}", h.removeCartItem)
		})
	})

	r.Route("/orders", func(r chi.Router) {
		r.Post("/", h.placeOrder)
		// r.Get("/", h.listOrders)
		// r.Get("/{id}", h.getOrder)
		// r.Delete("/{id}", h.cancelOrder)
	})
}

func (h *Handler) createCart(w http.ResponseWriter, r *http.Request) {
	var req CreateCartRequest
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

	cart, err := h.svc.CreateCart(r.Context(), buyerID, &req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, cart)
}

func (h *Handler) getCart(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := r.Context().Value("user").(*auth.UserContext)
	buyerID := uuid.Nil
	if userCtx != nil {
		buyerID = userCtx.PartnerID
	}

	cart, err := h.svc.GetActiveCart(r.Context(), buyerID)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, cart)
}

func (h *Handler) addCartItem(w http.ResponseWriter, r *http.Request) {
	var req AddCartItemRequest
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

	cart, err := h.svc.AddCartItem(r.Context(), buyerID, &req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, cart)
}

func (h *Handler) placeOrder(w http.ResponseWriter, r *http.Request) {
	var req PlaceOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, err)
		return
	}
	if err := h.val.Struct(req); err != nil {
		response.Error(w, err)
		return
	}

	userCtx, _ := r.Context().Value("user").(*auth.UserContext)
	buyerPartnerID := uuid.Nil
	buyerRole := ""
	if userCtx != nil {
		buyerPartnerID = userCtx.PartnerID
		if len(userCtx.Roles) > 0 {
			buyerRole = userCtx.Roles[0]
		}
	}

	order, err := h.svc.PlaceOrder(r.Context(), buyerPartnerID, buyerRole, &req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, order)
}
