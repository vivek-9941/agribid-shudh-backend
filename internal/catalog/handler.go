package catalog

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

// Handler holds catalog HTTP handlers.
type Handler struct {
	service  *Service
	validate *validator.Validate
}

// NewHandler creates a new catalog handler.
func NewHandler(service *Service) *Handler {
	return &Handler{
		service:  service,
		validate: validator.New(),
	}
}

// RegisterRoutes registers catalog routes.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/products", func(r chi.Router) {
		r.Get("/", h.ListProducts)
		r.Post("/", h.CreateProduct)
		r.Get("/{id}", h.GetProduct)
		r.Put("/{id}", h.UpdateProduct)
		r.Patch("/{id}/status", h.SetProductStatus)
	})

	r.Route("/api/v1/categories", func(r chi.Router) {
		r.Get("/", h.ListCategories)
		r.Post("/", h.CreateCategory)
	})
}

func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	userCtx := middleware.GetUserContext(r.Context())
	if userCtx == nil {
		response.Error(w, apperrors.Unauthorized("authentication required"))
		return
	}

	var req CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apperrors.BadRequest(apperrors.CodeValidationFailed, "invalid request body"))
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.Error(w, apperrors.ValidationFailed(map[string]interface{}{"error": err.Error()}))
		return
	}

	product, err := h.service.CreateProduct(r.Context(), userCtx.PartnerID, &req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, product)
}

func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperrors.BadRequest(apperrors.CodeValidationFailed, "invalid product id"))
		return
	}

	product, err := h.service.GetProduct(r.Context(), id)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, product)
}

func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("page_size"))

	filter := ProductFilter{
		Brand:  q.Get("brand"),
		Status: q.Get("status"),
		Search: q.Get("search"),
		Page:   page,
		PageSize: pageSize,
	}
	if catID := q.Get("category_id"); catID != "" {
		if cid, err := uuid.Parse(catID); err == nil {
			filter.CategoryID = &cid
		}
	}
	if mfgID := q.Get("manufacturer_id"); mfgID != "" {
		if mid, err := uuid.Parse(mfgID); err == nil {
			filter.ManufacturerID = &mid
		}
	}

	products, total, err := h.service.ListProducts(r.Context(), filter)
	if err != nil {
		response.Error(w, err)
		return
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	response.Paginated(w, products, response.PaginationMeta{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	})
}

func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperrors.BadRequest(apperrors.CodeValidationFailed, "invalid product id"))
		return
	}

	var req UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apperrors.BadRequest(apperrors.CodeValidationFailed, "invalid request body"))
		return
	}

	product, err := h.service.UpdateProduct(r.Context(), id, &req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, product)
}

func (h *Handler) SetProductStatus(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperrors.BadRequest(apperrors.CodeValidationFailed, "invalid product id"))
		return
	}

	var req SetStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apperrors.BadRequest(apperrors.CodeValidationFailed, "invalid request body"))
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.Error(w, apperrors.ValidationFailed(map[string]interface{}{"error": err.Error()}))
		return
	}

	if err := h.service.SetProductStatus(r.Context(), id, req.Status); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "product status updated"})
}

func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.service.ListCategories(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, cats)
}

func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apperrors.BadRequest(apperrors.CodeValidationFailed, "invalid request body"))
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.Error(w, apperrors.ValidationFailed(map[string]interface{}{"error": err.Error()}))
		return
	}

	cat, err := h.service.CreateCategory(r.Context(), &req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, cat)
}
