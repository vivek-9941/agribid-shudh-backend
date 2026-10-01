package notification

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
	r.Route("/notifications", func(r chi.Router) {
		r.Get("/", h.listNotifications)
		r.Patch("/{id}/read", h.markRead)

		r.Route("/preferences", func(r chi.Router) {
			r.Get("/", h.getPreferences)
			r.Put("/", h.upsertPreferences)
		})
	})
}

func (h *Handler) listNotifications(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := r.Context().Value("user").(*auth.UserContext)
	userID := uuid.Nil
	if userCtx != nil {
		userID = userCtx.ID
	}

	notifs, _, err := h.svc.ListForUser(r.Context(), userID, 1, 50)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, notifs)
}

func (h *Handler) markRead(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, err)
		return
	}

	userCtx, _ := r.Context().Value("user").(*auth.UserContext)
	userID := uuid.Nil
	if userCtx != nil {
		userID = userCtx.ID
	}

	if err := h.svc.MarkRead(r.Context(), id, userID); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "marked as read"})
}

func (h *Handler) getPreferences(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := r.Context().Value("user").(*auth.UserContext)
	userID := uuid.Nil
	if userCtx != nil {
		userID = userCtx.ID
	}

	pref, err := h.svc.GetPreferences(r.Context(), userID)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, pref)
}

func (h *Handler) upsertPreferences(w http.ResponseWriter, r *http.Request) {
	var pref NotificationPreference
	if err := json.NewDecoder(r.Body).Decode(&pref); err != nil {
		response.Error(w, err)
		return
	}

	userCtx, _ := r.Context().Value("user").(*auth.UserContext)
	if userCtx != nil {
		pref.UserID = userCtx.ID
	}

	if err := h.svc.UpsertPreferences(r.Context(), &pref); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, pref)
}
