package admin

// Service provides admin operations by delegating to underlying module services.
// This is a coordinator service that doesn't have its own repository.
type Service struct {
	// Add dependencies to other services as needed for admin operations
}

// NewService creates a new admin service.
func NewService() *Service {
	return &Service{}
}

// Additional admin operations will be added here as needed:
// - User management (create, deactivate, reset password)
// - Partner approval
// - System configuration
// - Manual reconciliation operations
// - Bulk operations
