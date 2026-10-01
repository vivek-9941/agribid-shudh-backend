package auth

import (
	"encoding/json"
	"net/http"

	apperrors "github.com/agribid/agribid-shudh-backend/internal/errors"
	"github.com/agribid/agribid-shudh-backend/internal/middleware"
	"github.com/agribid/agribid-shudh-backend/internal/response"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

// Handler holds auth HTTP handlers.
type Handler struct {
	service  *Service
	validate *validator.Validate
}

// NewHandler creates a new auth handler.
func NewHandler(service *Service) *Handler {
	return &Handler{
		service:  service,
		validate: validator.New(),
	}
}

// RegisterRoutes registers auth routes on the chi router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/otp/send", h.SendOTP)
		r.Post("/otp/verify", h.VerifyOTP)
		r.Post("/login", h.Login)
		r.Post("/refresh", h.Refresh)
		r.Post("/logout", h.Logout)    // requires auth
		r.Get("/me", h.GetMe)          // requires auth
	})
}

// SendOTP handles POST /api/v1/auth/otp/send.
func (h *Handler) SendOTP(w http.ResponseWriter, r *http.Request) {
	var req SendOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apperrors.BadRequest(apperrors.CodeValidationFailed, "invalid request body"))
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.Error(w, apperrors.ValidationFailed(map[string]interface{}{"error": err.Error()}))
		return
	}

	if err := h.service.SendOTP(r.Context(), req.Phone); err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "OTP sent successfully"})
}

// VerifyOTP handles POST /api/v1/auth/otp/verify.
func (h *Handler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req VerifyOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apperrors.BadRequest(apperrors.CodeValidationFailed, "invalid request body"))
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.Error(w, apperrors.ValidationFailed(map[string]interface{}{"error": err.Error()}))
		return
	}

	tokens, err := h.service.VerifyOTP(r.Context(), req.Phone, req.OTP)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, tokens)
}

// Login handles POST /api/v1/auth/login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apperrors.BadRequest(apperrors.CodeValidationFailed, "invalid request body"))
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.Error(w, apperrors.ValidationFailed(map[string]interface{}{"error": err.Error()}))
		return
	}

	tokens, err := h.service.LoginWithPassword(r.Context(), req.Email, req.Password)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, tokens)
}

// Refresh handles POST /api/v1/auth/refresh.
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apperrors.BadRequest(apperrors.CodeValidationFailed, "invalid request body"))
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.Error(w, apperrors.ValidationFailed(map[string]interface{}{"error": err.Error()}))
		return
	}

	tokens, err := h.service.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, tokens)
}

// Logout handles POST /api/v1/auth/logout.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	userCtx := middleware.GetUserContext(r.Context())
	if userCtx == nil {
		response.Error(w, apperrors.Unauthorized("authentication required"))
		return
	}

	if err := h.service.Logout(r.Context(), userCtx.UserID); err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}

// GetMe handles GET /api/v1/auth/me.
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userCtx := middleware.GetUserContext(r.Context())
	if userCtx == nil {
		response.Error(w, apperrors.Unauthorized("authentication required"))
		return
	}

	user, err := h.service.GetMe(r.Context(), userCtx.UserID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, user)
}
