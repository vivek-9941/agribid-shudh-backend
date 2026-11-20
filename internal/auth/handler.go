package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	apperrors "github.com/agribid/agribid-shudh-backend/internal/errors"
	"github.com/agribid/agribid-shudh-backend/internal/response"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// Handler holds auth HTTP handlers.
type Handler struct {
	service    *Service
	validate   *validator.Validate
	jwtService *JWTService
}

// NewHandler creates a new auth handler.
func NewHandler(service *Service, val *validator.Validate) *Handler {
	return &Handler{
		service:    service,
		validate:   val,
		jwtService: service.jwt,
	}
}

// authMiddleware validates Bearer token and injects UserContext.
func (h *Handler) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.Error(w, apperrors.Unauthorized("missing authorization header"))
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			response.Error(w, apperrors.Unauthorized("invalid authorization header format"))
			return
		}
		claims, err := h.jwtService.VerifyAccessToken(parts[1])
		if err != nil {
			response.Error(w, apperrors.Unauthorized("invalid or expired token"))
			return
		}
		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			response.Error(w, apperrors.Unauthorized("invalid token subject"))
			return
		}
		partnerID := uuid.Nil
		if claims.PartnerID != "" {
			if pid, err := uuid.Parse(claims.PartnerID); err == nil {
				partnerID = pid
			}
		}
		isAdmin := false
		for _, role := range claims.Roles {
			if role == "ADMIN" {
				isAdmin = true
				break
			}
		}
		uc := &UserContext{
			ID:        userID,
			UserID:    userID,
			PartnerID: partnerID,
			Roles:     claims.Roles,
			IsAdmin:   isAdmin,
		}
		ctx := WithUserContext(r.Context(), uc)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RegisterRoutes registers auth routes on the chi router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/auth", func(r chi.Router) {
		// Public routes (no auth required)
		r.Post("/otp/send", h.SendOTP)
		r.Post("/otp/verify", h.VerifyOTP)
		r.Post("/login", h.Login)
		r.Post("/refresh", h.Refresh)

		// Protected routes (auth required)
		r.Group(func(r chi.Router) {
			r.Use(h.authMiddleware)
			r.Post("/logout", h.Logout)
			r.Get("/me", h.GetMe)
		})
	})
}

// SendOTP handles POST /api/v1/auth/otp/send.
// @Summary      Send OTP
// @Description  Send a one-time password to the specified phone number
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body  SendOTPRequest  true  "Phone number payload"
// @Success      200  {object}  response.successResponse{data=map[string]string}
// @Failure      400  {object}  response.errorResponse
// @Failure      500  {object}  response.errorResponse
// @Router       /api/v1/auth/otp/send [post]
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
// @Summary      Verify OTP
// @Description  Verify the OTP and receive access/refresh tokens
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body  VerifyOTPRequest  true  "Phone + OTP payload"
// @Success      200  {object}  response.successResponse{data=TokenPair}
// @Failure      400  {object}  response.errorResponse
// @Failure      401  {object}  response.errorResponse
// @Router       /api/v1/auth/otp/verify [post]
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
// @Summary      Login with email/password
// @Description  Authenticate using email and password credentials
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body  LoginRequest  true  "Email + Password payload"
// @Success      200  {object}  response.successResponse{data=TokenPair}
// @Failure      400  {object}  response.errorResponse
// @Failure      401  {object}  response.errorResponse
// @Router       /api/v1/auth/login [post]
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
// @Summary      Refresh access token
// @Description  Get a new access token using a valid refresh token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body  RefreshRequest  true  "Refresh token payload"
// @Success      200  {object}  response.successResponse{data=TokenPair}
// @Failure      400  {object}  response.errorResponse
// @Failure      401  {object}  response.errorResponse
// @Router       /api/v1/auth/refresh [post]
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
// @Summary      Logout
// @Description  Invalidate the current user session
// @Tags         Auth
// @Produce      json
// @Success      200  {object}  response.successResponse{data=map[string]string}
// @Failure      401  {object}  response.errorResponse
// @Security     BearerAuth
// @Router       /api/v1/auth/logout [post]
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	userCtx := GetUserContext(r.Context())
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
// @Summary      Get current user
// @Description  Get the authenticated user's profile information
// @Tags         Auth
// @Produce      json
// @Success      200  {object}  response.successResponse{data=User}
// @Failure      401  {object}  response.errorResponse
// @Security     BearerAuth
// @Router       /api/v1/auth/me [get]
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userCtx := GetUserContext(r.Context())
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
