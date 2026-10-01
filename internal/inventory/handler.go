package inventory

import (
	"encoding/json"
	"net/http"
	"strconv"

	apperrors "github.com/agribid/agribid-shudh-backend/internal/errors"
	"github.com/agribid/agribid-shudh-backend/internal/middleware"
	"github.com/agribid/agribid-shudh-backend/internal/response"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// Handler holds inventory HTTP handlers.
type Handler struct {
	service  *Service
	validate *validator.Validate
}

// NewHandler creates a new inventory handler.
func NewHandler(service *Service) *Handler {
	return &Handler{
		service:  service,
		validate: validator.New(),
	}
}

// RegisterRoutes registers inventory routes.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/inventory", func(r chi.Router) {
		r.Get("/", h.ListStock)
		r.Post("/inward", h.AddStock)
		r.Post("/adjust", h.AdjustStock)
		r.Get("/low-stock", h.GetLowStock)
	})
}

func (h *Handler) ListStock(w http.ResponseWriter, r *http.Request) {
	userCtx := middleware.GetUserContext(r.Context())
	if userCtx == nil {
		response.Error(w, apperrors.Unauthorized("authentication required"))
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var warehouseID *uuid.UUID
	if whID := r.URL.Query().Get("warehouse_id"); whID != "" {
		if wid, err := uuid.Parse(whID); err == nil {
			warehouseID = &wid
		}
	}

	items, total, err := h.service.ListStock(r.Context(), userCtx.PartnerID, warehouseID, page, pageSize)
	if err != nil {
		response.Error(w, apperrors.Internal(err))
		return
	}
	response.Paginated(w, items, response.PaginationMeta{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	})
}

func (h *Handler) AddStock(w http.ResponseWriter, r *http.Request) {
	userCtx := middleware.GetUserContext(r.Context())
	if userCtx == nil {
		response.Error(w, apperrors.Unauthorized("authentication required"))
		return
	}

	var req AdjustStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apperrors.BadRequest(apperrors.CodeValidationFailed, "invalid request body"))
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.Error(w, apperrors.ValidationFailed(map[string]interface{}{"error": err.Error()}))
		return
	}

	warehouseID, _ := uuid.Parse(req.WarehouseID)
	productID, _ := uuid.Parse(req.ProductID)

	if err := h.service.AddStock(r.Context(), userCtx.PartnerID, warehouseID, productID, req.Quantity, userCtx.UserID); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "stock added successfully"})
}

func (h *Handler) AdjustStock(w http.ResponseWriter, r *http.Request) {
	userCtx := middleware.GetUserContext(r.Context())
	if userCtx == nil {
		response.Error(w, apperrors.Unauthorized("authentication required"))
		return
	}

	var req AdjustStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apperrors.BadRequest(apperrors.CodeValidationFailed, "invalid request body"))
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.Error(w, apperrors.ValidationFailed(map[string]interface{}{"error": err.Error()}))
		return
	}

	warehouseID, _ := uuid.Parse(req.WarehouseID)
	productID, _ := uuid.Parse(req.ProductID)

	if err := h.service.AdjustStock(r.Context(), userCtx.PartnerID, warehouseID, productID, req.Quantity, req.Reason, userCtx.UserID); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "stock adjusted successfully"})
}

func (h *Handler) GetLowStock(w http.ResponseWriter, r *http.Request) {
	userCtx := middleware.GetUserContext(r.Context())
	if userCtx == nil {
		response.Error(w, apperrors.Unauthorized("authentication required"))
		return
	}

	items, err := h.service.GetLowStockItems(r.Context(), userCtx.PartnerID)
	if err != nil {
		response.Error(w, apperrors.Internal(err))
		return
	}
	response.JSON(w, http.StatusOK, items)
}
