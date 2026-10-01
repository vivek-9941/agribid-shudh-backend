package catalog

import (
	"time"

	"github.com/google/uuid"
)

// Product represents a product in the catalog.
type Product struct {
	ID             uuid.UUID              `json:"id"`
	SKU            string                 `json:"sku"`
	Name           string                 `json:"name"`
	Description    *string                `json:"description,omitempty"`
	CategoryID     *uuid.UUID             `json:"category_id,omitempty"`
	Brand          *string                `json:"brand,omitempty"`
	ManufacturerID uuid.UUID              `json:"manufacturer_id"`
	HSNCode        string                 `json:"hsn_code"`
	UOM            *string                `json:"uom,omitempty"`
	WeightGrams    *int                   `json:"weight_grams,omitempty"`
	Dimensions     map[string]interface{} `json:"dimensions,omitempty"`
	ParentSKUID    *uuid.UUID             `json:"parent_sku_id,omitempty"`
	VariantAttrs   map[string]interface{} `json:"variant_attrs,omitempty"`
	Status         string                 `json:"status"` // active, inactive, discontinued
	Images         []ProductImage         `json:"images,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	DeletedAt      *time.Time             `json:"deleted_at,omitempty"`
}

// ProductImage represents an image associated with a product.
type ProductImage struct {
	ID        uuid.UUID `json:"id"`
	ProductID uuid.UUID `json:"product_id"`
	URL       string    `json:"url"`
	SortOrder int       `json:"sort_order"`
	IsPrimary bool      `json:"is_primary"`
}

// Category represents a product category with optional hierarchy.
type Category struct {
	ID        uuid.UUID   `json:"id"`
	ParentID  *uuid.UUID  `json:"parent_id,omitempty"`
	Name      string      `json:"name"`
	Slug      string      `json:"slug"`
	Path      string      `json:"path"`
	Level     int         `json:"level"`
	Children  []*Category `json:"children,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Request DTOs
// ─────────────────────────────────────────────────────────────────────────────

// CreateProductRequest is the payload for POST /products.
type CreateProductRequest struct {
	SKU          string                 `json:"sku" validate:"required,min=1,max=100"`
	Name         string                 `json:"name" validate:"required,min=2,max=255"`
	Description  string                 `json:"description,omitempty"`
	CategoryID   string                 `json:"category_id,omitempty" validate:"omitempty,uuid"`
	Brand        string                 `json:"brand,omitempty"`
	HSNCode      string                 `json:"hsn_code" validate:"required,max=20"`
	UOM          string                 `json:"uom,omitempty"`
	WeightGrams  int                    `json:"weight_grams,omitempty"`
	Dimensions   map[string]interface{} `json:"dimensions,omitempty"`
	ParentSKUID  string                 `json:"parent_sku_id,omitempty" validate:"omitempty,uuid"`
	VariantAttrs map[string]interface{} `json:"variant_attrs,omitempty"`
}

// UpdateProductRequest is the payload for PUT /products/:id.
type UpdateProductRequest struct {
	Name         string                 `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	Description  string                 `json:"description,omitempty"`
	CategoryID   string                 `json:"category_id,omitempty" validate:"omitempty,uuid"`
	Brand        string                 `json:"brand,omitempty"`
	HSNCode      string                 `json:"hsn_code,omitempty" validate:"omitempty,max=20"`
	UOM          string                 `json:"uom,omitempty"`
	WeightGrams  int                    `json:"weight_grams,omitempty"`
	Dimensions   map[string]interface{} `json:"dimensions,omitempty"`
	VariantAttrs map[string]interface{} `json:"variant_attrs,omitempty"`
}

// SetStatusRequest is the payload for PATCH /products/:id/status.
type SetStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=active inactive discontinued"`
}

// CreateCategoryRequest is the payload for POST /categories.
type CreateCategoryRequest struct {
	ParentID string `json:"parent_id,omitempty" validate:"omitempty,uuid"`
	Name     string `json:"name" validate:"required,min=2,max=255"`
	Slug     string `json:"slug" validate:"required,min=2,max=255"`
}

// ProductFilter holds query parameters for listing products.
type ProductFilter struct {
	CategoryID     *uuid.UUID
	Brand          string
	Status         string
	Search         string
	ManufacturerID *uuid.UUID
	Page           int
	PageSize       int
}
