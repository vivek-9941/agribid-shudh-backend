package dashboard

import (
	"net/http"

	"github.com/agribid/agribid-shudh-backend/internal/auth"
	"github.com/agribid/agribid-shudh-backend/internal/response"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/dashboard", func(r chi.Router) {
		r.Get("/", h.getMetrics)
	})
	
	// Stub for reports
	r.Route("/reports", func(r chi.Router) {
		r.Get("/{type}", h.getReport)
	})
}

func (h *Handler) getMetrics(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := r.Context().Value("user").(*auth.UserContext)
	
	var partnerID *uuid.UUID
	var role string
	
	if userCtx != nil {
		if len(userCtx.Roles) > 0 {
			role = userCtx.Roles[0]
		}
		if role != "admin" {
			pid := userCtx.PartnerID
			if pid != uuid.Nil {
				partnerID = &pid
			}
		}
	}

	metrics, err := h.svc.GetDashboardMetrics(r.Context(), partnerID, role)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, metrics)
}

func (h *Handler) getReport(w http.ResponseWriter, r *http.Request) {
	// reportType := chi.URLParam(r, "type")
	// For stub, just return not implemented
	response.JSON(w, http.StatusNotImplemented, map[string]string{"message": "reports not implemented"})
}
