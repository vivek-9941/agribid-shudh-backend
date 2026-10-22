package admin

import (
	"net/http"

	"github.com/agribid/agribid-shudh-backend/internal/response"
	"github.com/go-chi/chi/v5"
)

// Handler holds admin HTTP handlers.
type Handler struct {
	service *Service
}

// NewHandler creates a new admin handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers admin routes.
// All admin routes should be protected by RBAC middleware requiring admin:* permission.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/admin", func(r chi.Router) {
		// TODO: Add RBAC middleware here
		
		r.Get("/health", h.healthCheck)
		
		// User management routes (to be implemented)
		// r.Get("/users", h.listUsers)
		// r.Post("/users", h.createUser)
		// r.Patch("/users/{id}", h.updateUser)
		// r.Delete("/users/{id}", h.deactivateUser)
		// r.Post("/users/{id}/reset-password", h.resetPassword)
		
		// Partner management routes (to be implemented)
		// r.Get("/partners/pending", h.listPendingPartners)
		// r.Post("/partners/{id}/approve", h.approvePartner)
		
		// System config routes (to be implemented)
		// r.Get("/config", h.listConfig)
		// r.Put("/config/{key}", h.updateConfig)
		
		// Manual operations routes (to be implemented)
		// r.Post("/operations/reconcile", h.triggerReconciliation)
	})
}

func (h *Handler) healthCheck(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{
		"status": "admin_service_ok",
	})
}
