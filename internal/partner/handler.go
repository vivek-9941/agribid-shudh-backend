package partner

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

// Handler holds partner HTTP handlers.
type Handler struct {
	service  *Service
	validate *validator.Validate
}

// NewHandler creates a new partner handler.
func NewHandler(service *Service) *Handler {
	return &Handler{
		service:  service,
		validate: validator.New(),
	}
}

// RegisterRoutes registers partner routes on the chi router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/partners", func(r chi.Router) {
		r.Post("/", h.Create)
		r.Get("/", h.List)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}", h.Update)
		r.Post("/{id}/kyc", h.SubmitKYC)
		r.Get("/{id}/kyc", h.GetKYC)
		r.Patch("/{id}/kyc", h.ReviewKYC)
		r.Get("/{id}/children", h.GetChildren)
	})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreatePartnerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apperrors.BadRequest(apperrors.CodeValidationFailed, "invalid request body"))
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.Error(w, apperrors.ValidationFailed(map[string]interface{}{"error": err.Error()}))
		return
	}

	p, err := h.service.Create(r.Context(), &req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, p)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperrors.BadRequest(apperrors.CodeValidationFailed, "invalid partner id"))
		return
	}

	p, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, p)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	partnerType := r.URL.Query().Get("type")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	partners, total, err := h.service.repo.List(r.Context(), partnerType, page, pageSize)
	if err != nil {
		response.Error(w, apperrors.Internal(err))
		return
	}
	response.Paginated(w, partners, response.PaginationMeta{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	})
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperrors.BadRequest(apperrors.CodeValidationFailed, "invalid partner id"))
		return
	}

	var req UpdatePartnerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apperrors.BadRequest(apperrors.CodeValidationFailed, "invalid request body"))
		return
	}

	p, err := h.service.Update(r.Context(), id, &req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, p)
}

func (h *Handler) SubmitKYC(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperrors.BadRequest(apperrors.CodeValidationFailed, "invalid partner id"))
		return
	}

	var req SubmitKYCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apperrors.BadRequest(apperrors.CodeValidationFailed, "invalid request body"))
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.Error(w, apperrors.ValidationFailed(map[string]interface{}{"error": err.Error()}))
		return
	}

	doc, err := h.service.SubmitKYC(r.Context(), id, &req)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, doc)
}

func (h *Handler) GetKYC(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperrors.BadRequest(apperrors.CodeValidationFailed, "invalid partner id"))
		return
	}

	docs, err := h.service.GetKYCDocuments(r.Context(), id)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, docs)
}

func (h *Handler) ReviewKYC(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperrors.BadRequest(apperrors.CodeValidationFailed, "invalid partner id"))
		return
	}

	userCtx := middleware.GetUserContext(r.Context())
	if userCtx == nil {
		response.Error(w, apperrors.Unauthorized("authentication required"))
		return
	}

	var req ReviewKYCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apperrors.BadRequest(apperrors.CodeValidationFailed, "invalid request body"))
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.Error(w, apperrors.ValidationFailed(map[string]interface{}{"error": err.Error()}))
		return
	}

	if err := h.service.ReviewKYC(r.Context(), id, &req, userCtx.UserID); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "KYC review completed"})
}

func (h *Handler) GetChildren(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperrors.BadRequest(apperrors.CodeValidationFailed, "invalid partner id"))
		return
	}

	children, err := h.service.ListChildren(r.Context(), id)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, children)
}
