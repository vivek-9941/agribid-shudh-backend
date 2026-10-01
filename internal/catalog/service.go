package catalog

import (
	"context"

	apperrors "github.com/agribid/agribid-shudh-backend/internal/errors"
	"github.com/google/uuid"
)

// Service provides catalog business logic.
type Service struct {
	repo Repository
}

// NewService creates a new catalog service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// CreateProduct creates a new product in the catalog.
func (s *Service) CreateProduct(ctx context.Context, manufacturerID uuid.UUID, req *CreateProductRequest) (*Product, error) {
	// Check for duplicate SKU.
	existing, err := s.repo.GetProductBySKU(ctx, req.SKU)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if existing != nil {
		return nil, apperrors.Conflict(apperrors.CodeDuplicateSKU, "product with this SKU already exists")
	}

	p := &Product{
		SKU:            req.SKU,
		Name:           req.Name,
		ManufacturerID: manufacturerID,
		HSNCode:        req.HSNCode,
		Status:         "active",
	}
	if req.Description != "" {
		p.Description = &req.Description
	}
	if req.CategoryID != "" {
		cid, _ := uuid.Parse(req.CategoryID)
		p.CategoryID = &cid
	}
	if req.Brand != "" {
		p.Brand = &req.Brand
	}
	if req.UOM != "" {
		p.UOM = &req.UOM
	}
	if req.WeightGrams > 0 {
		p.WeightGrams = &req.WeightGrams
	}
	p.Dimensions = req.Dimensions
	p.VariantAttrs = req.VariantAttrs
	if req.ParentSKUID != "" {
		psid, _ := uuid.Parse(req.ParentSKUID)
		p.ParentSKUID = &psid
	}

	if err := s.repo.CreateProduct(ctx, p); err != nil {
		return nil, apperrors.Internal(err)
	}
	return p, nil
}

// GetProduct returns a product by ID.
func (s *Service) GetProduct(ctx context.Context, id uuid.UUID) (*Product, error) {
	p, err := s.repo.GetProductByID(ctx, id)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if p == nil {
		return nil, apperrors.NotFound("product", id.String())
	}

	// Load images.
	images, err := s.repo.GetProductImages(ctx, id)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	p.Images = images

	return p, nil
}

// ListProducts returns products matching the given filter.
func (s *Service) ListProducts(ctx context.Context, filter ProductFilter) ([]*Product, int64, error) {
	products, total, err := s.repo.ListProducts(ctx, filter)
	if err != nil {
		return nil, 0, apperrors.Internal(err)
	}
	return products, total, nil
}

// UpdateProduct updates a product's information.
func (s *Service) UpdateProduct(ctx context.Context, id uuid.UUID, req *UpdateProductRequest) (*Product, error) {
	p, err := s.repo.GetProductByID(ctx, id)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if p == nil {
		return nil, apperrors.NotFound("product", id.String())
	}

	if req.Name != "" {
		p.Name = req.Name
	}
	if req.Description != "" {
		p.Description = &req.Description
	}
	if req.CategoryID != "" {
		cid, _ := uuid.Parse(req.CategoryID)
		p.CategoryID = &cid
	}
	if req.Brand != "" {
		p.Brand = &req.Brand
	}
	if req.HSNCode != "" {
		p.HSNCode = req.HSNCode
	}
	if req.UOM != "" {
		p.UOM = &req.UOM
	}
	if req.WeightGrams > 0 {
		p.WeightGrams = &req.WeightGrams
	}
	if req.Dimensions != nil {
		p.Dimensions = req.Dimensions
	}
	if req.VariantAttrs != nil {
		p.VariantAttrs = req.VariantAttrs
	}

	if err := s.repo.UpdateProduct(ctx, p); err != nil {
		return nil, apperrors.Internal(err)
	}
	return p, nil
}

// SetProductStatus sets the status of a product.
func (s *Service) SetProductStatus(ctx context.Context, id uuid.UUID, status string) error {
	p, err := s.repo.GetProductByID(ctx, id)
	if err != nil {
		return apperrors.Internal(err)
	}
	if p == nil {
		return apperrors.NotFound("product", id.String())
	}

	if err := s.repo.SetProductStatus(ctx, id, status); err != nil {
		return apperrors.Internal(err)
	}
	return nil
}

// CreateCategory creates a new category.
func (s *Service) CreateCategory(ctx context.Context, req *CreateCategoryRequest) (*Category, error) {
	// Check slug uniqueness.
	existing, err := s.repo.GetCategoryBySlug(ctx, req.Slug)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if existing != nil {
		return nil, apperrors.Conflict(apperrors.CodeConflict, "category with this slug already exists")
	}

	c := &Category{
		Name: req.Name,
		Slug: req.Slug,
	}

	if req.ParentID != "" {
		pid, _ := uuid.Parse(req.ParentID)
		c.ParentID = &pid

		parent, err := s.repo.GetCategoryByID(ctx, pid)
		if err != nil {
			return nil, apperrors.Internal(err)
		}
		if parent == nil {
			return nil, apperrors.NotFound("category", req.ParentID)
		}
		c.Level = parent.Level + 1
		c.Path = parent.Path + "/" + req.Slug
	} else {
		c.Level = 0
		c.Path = req.Slug
	}

	if err := s.repo.CreateCategory(ctx, c); err != nil {
		return nil, apperrors.Internal(err)
	}
	return c, nil
}

// ListCategories returns all categories.
func (s *Service) ListCategories(ctx context.Context) ([]*Category, error) {
	cats, err := s.repo.ListCategories(ctx)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	return cats, nil
}
