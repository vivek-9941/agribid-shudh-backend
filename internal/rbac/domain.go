package rbac

import (
	"time"

	"github.com/google/uuid"
)

// Role represents a system role.
type Role struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// Permission represents a role-resource-action permission.
type Permission struct {
	ID       uuid.UUID `json:"id"`
	RoleID   uuid.UUID `json:"role_id"`
	Resource string    `json:"resource"`
	Action   string    `json:"action"`
}
