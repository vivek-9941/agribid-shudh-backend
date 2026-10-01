package rbac

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the data access interface for RBAC.
type Repository interface {
	GetPermissionsForRole(ctx context.Context, roleCode string) ([]Permission, error)
}

// pgRepo implements Repository with in-process caching (5-minute TTL).
type pgRepo struct {
	pool  *pgxpool.Pool
	cache map[string]cachedPerms
	mu    sync.RWMutex
}

type cachedPerms struct {
	perms     []Permission
	expiresAt time.Time
}

// NewRepository creates a new pgx-backed RBAC repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepo{
		pool:  pool,
		cache: make(map[string]cachedPerms),
	}
}

func (r *pgRepo) GetPermissionsForRole(ctx context.Context, roleCode string) ([]Permission, error) {
	// Check cache first.
	r.mu.RLock()
	if cached, ok := r.cache[roleCode]; ok && time.Now().Before(cached.expiresAt) {
		r.mu.RUnlock()
		return cached.perms, nil
	}
	r.mu.RUnlock()

	// Query DB.
	query := `SELECT p.id, p.role_id, p.resource, p.action
		FROM permissions p
		JOIN roles r ON r.id = p.role_id
		WHERE r.code = $1`

	rows, err := r.pool.Query(ctx, query, roleCode)
	if err != nil {
		return nil, fmt.Errorf("getPermissionsForRole: %w", err)
	}
	defer rows.Close()

	var perms []Permission
	for rows.Next() {
		var p Permission
		if err := rows.Scan(&p.ID, &p.RoleID, &p.Resource, &p.Action); err != nil {
			return nil, fmt.Errorf("getPermissionsForRole scan: %w", err)
		}
		perms = append(perms, p)
	}

	// Cache for 5 minutes.
	r.mu.Lock()
	r.cache[roleCode] = cachedPerms{
		perms:     perms,
		expiresAt: time.Now().Add(5 * time.Minute),
	}
	r.mu.Unlock()

	return perms, nil
}
