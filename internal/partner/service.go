package partner

import (
	"context"

	apperrors "github.com/agribid/agribid-shudh-backend/internal/errors"
	"github.com/google/uuid"
)

// Service provides partner business logic.
type Service struct {
	repo Repository
}

// NewService creates a new partner service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Create creates a new partner.
func (s *Service) Create(ctx context.Context, req *CreatePartnerRequest) (*Partner, error) {
	var parentID *uuid.UUID
	if req.ParentID != "" {
		pid, err := uuid.Parse(req.ParentID)
		if err != nil {
			return nil, apperrors.BadRequest(apperrors.CodeValidationFailed, "invalid parent_id")
		}
		parentID = &pid

		// Verify parent exists.
		parent, err := s.repo.GetByID(ctx, pid)
		if err != nil {
			return nil, apperrors.Internal(err)
		}
		if parent == nil {
			return nil, apperrors.NotFound("partner", req.ParentID)
		}
	}

	p := &Partner{
		Type:         req.Type,
		BusinessName: req.BusinessName,
		ParentID:     parentID,
	}
	if req.TradeName != "" {
		p.TradeName = &req.TradeName
	}
	if req.GSTIN != "" {
		p.GSTIN = &req.GSTIN
	}
	if req.PAN != "" {
		p.PAN = &req.PAN
	}
	if req.StateCode != "" {
		p.StateCode = &req.StateCode
	}
	p.Address = req.Address

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, apperrors.Internal(err)
	}
	return p, nil
}

// GetByID retrieves a partner by ID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*Partner, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if p == nil {
		return nil, apperrors.NotFound("partner", id.String())
	}
	return p, nil
}

// Update updates partner information.
func (s *Service) Update(ctx context.Context, id uuid.UUID, req *UpdatePartnerRequest) (*Partner, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if p == nil {
		return nil, apperrors.NotFound("partner", id.String())
	}

	if req.BusinessName != "" {
		p.BusinessName = req.BusinessName
	}
	if req.TradeName != "" {
		p.TradeName = &req.TradeName
	}
	if req.GSTIN != "" {
		p.GSTIN = &req.GSTIN
	}
	if req.PAN != "" {
		p.PAN = &req.PAN
	}
	if req.StateCode != "" {
		p.StateCode = &req.StateCode
	}
	if req.Address != nil {
		p.Address = req.Address
	}

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, apperrors.Internal(err)
	}
	return p, nil
}

// ListChildren returns downstream partners for a given parent.
func (s *Service) ListChildren(ctx context.Context, parentID uuid.UUID) ([]*Partner, error) {
	children, err := s.repo.GetChildren(ctx, parentID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	return children, nil
}

// SubmitKYC submits a KYC document for a partner.
func (s *Service) SubmitKYC(ctx context.Context, partnerID uuid.UUID, req *SubmitKYCRequest) (*KYCDocument, error) {
	p, err := s.repo.GetByID(ctx, partnerID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if p == nil {
		return nil, apperrors.NotFound("partner", partnerID.String())
	}

	doc := &KYCDocument{
		PartnerID: partnerID,
		DocType:   req.DocType,
		FileURL:   req.FileURL,
	}

	if err := s.repo.SubmitKYC(ctx, doc); err != nil {
		return nil, apperrors.Internal(err)
	}

	// Update partner KYC status to submitted if it was pending.
	if p.KYCStatus == "pending" {
		_ = s.repo.UpdatePartnerKYCStatus(ctx, partnerID, "submitted", p.Status)
	}

	return doc, nil
}

// ReviewKYC approves or rejects a KYC document.
func (s *Service) ReviewKYC(ctx context.Context, partnerID uuid.UUID, req *ReviewKYCRequest, reviewerID uuid.UUID) error {
	p, err := s.repo.GetByID(ctx, partnerID)
	if err != nil {
		return apperrors.Internal(err)
	}
	if p == nil {
		return apperrors.NotFound("partner", partnerID.String())
	}

	// Get all KYC docs and update the latest one.
	docs, err := s.repo.GetKYCDocuments(ctx, partnerID)
	if err != nil {
		return apperrors.Internal(err)
	}
	if len(docs) == 0 {
		return apperrors.BadRequest(apperrors.CodeValidationFailed, "no KYC documents found for this partner")
	}

	// Update latest doc.
	latestDoc := docs[0]
	if err := s.repo.UpdateKYCStatus(ctx, latestDoc.ID, req.Status, reviewerID, req.ReviewNote); err != nil {
		return apperrors.Internal(err)
	}

	// On KYC approval: set partner status to active.
	if req.Status == "approved" {
		if err := s.repo.UpdatePartnerKYCStatus(ctx, partnerID, "approved", "active"); err != nil {
			return apperrors.Internal(err)
		}
	} else {
		if err := s.repo.UpdatePartnerKYCStatus(ctx, partnerID, "rejected", p.Status); err != nil {
			return apperrors.Internal(err)
		}
	}

	return nil
}

// GetKYCDocuments returns all KYC documents for a partner.
func (s *Service) GetKYCDocuments(ctx context.Context, partnerID uuid.UUID) ([]*KYCDocument, error) {
	docs, err := s.repo.GetKYCDocuments(ctx, partnerID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	return docs, nil
}
