package payment

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/agribid/agribid-shudh-backend/internal/db"
	apperrors "github.com/agribid/agribid-shudh-backend/internal/errors"
	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) CheckCreditLimit(ctx context.Context, buyerID, sellerID uuid.UUID, amountStr string) error {
	account, err := s.repo.GetCreditAccount(ctx, buyerID, sellerID)
	if err != nil {
		return apperrors.Internal(err)
	}
	if account == nil {
		// Assume zero credit limit if no account exists
		return apperrors.BadRequest("no credit account established")
	}

	limitRat, _ := new(big.Rat).SetString(account.CreditLimit)
	utilizedRat, _ := new(big.Rat).SetString(account.CreditUtilized)
	requestedRat, _ := new(big.Rat).SetString(amountStr)

	availableRat := new(big.Rat).Sub(limitRat, utilizedRat)

	if requestedRat.Cmp(availableRat) > 0 {
		return apperrors.BadRequest(fmt.Sprintf("credit limit exceeded. available: %s, requested: %s", formatDecimal(availableRat), amountStr))
	}

	return nil
}

func (s *Service) UpdateCredit(ctx context.Context, buyerID, sellerID uuid.UUID, amountStr string) error {
	// amountStr could be positive (ordering) or negative (payment)
	return s.repo.UpdateCredit(ctx, buyerID, sellerID, amountStr)
}

func (s *Service) RecordPayment(ctx context.Context, req *RecordPaymentRequest, buyerID uuid.UUID) (*Payment, error) {
	sellerID, _ := uuid.Parse(req.SellerID)
	
	pay := &Payment{
		PaymentNumber:   fmt.Sprintf("PAY-%d", time.Now().Unix()),
		BuyerID:         buyerID,
		SellerID:        sellerID,
		Amount:          req.Amount,
		Method:          req.Method,
		ReferenceNumber: &req.ReferenceNumber,
	}
	if req.InvoiceID != "" {
		iid, _ := uuid.Parse(req.InvoiceID)
		pay.InvoiceID = &iid
	}

	err := db.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.RecordPayment(txCtx, pay); err != nil {
			return err
		}

		// Reduce credit utilized
		amtRat, _ := new(big.Rat).SetString(req.Amount)
		negAmt := new(big.Rat).Neg(amtRat)
		if err := s.repo.UpdateCredit(txCtx, buyerID, sellerID, formatDecimal(negAmt)); err != nil {
			return err
		}

		// Insert ledger entry for buyer
		entry := &LedgerEntry{
			PartnerID:      buyerID,
			CounterpartyID: sellerID,
			EntryType:      "payment",
			ReferenceID:    &pay.ID,
			Debit:          req.Amount,
			Credit:         "0.00",
			Balance:        "0.00", // Needs actual balance calculation
			EntryDate:      time.Now(),
			Description:    fmt.Sprintf("Payment via %s", req.Method),
		}
		if err := s.repo.InsertLedgerEntry(txCtx, entry); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, apperrors.Internal(err)
	}

	return pay, nil
}

func (s *Service) GetCreditStatus(ctx context.Context, buyerID, sellerID uuid.UUID) (*CreditAccount, error) {
	account, err := s.repo.GetCreditAccount(ctx, buyerID, sellerID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if account == nil {
		return nil, apperrors.NotFound("credit_account", buyerID.String())
	}
	
	limitRat, _ := new(big.Rat).SetString(account.CreditLimit)
	utilizedRat, _ := new(big.Rat).SetString(account.CreditUtilized)
	account.CreditAvailable = formatDecimal(new(big.Rat).Sub(limitRat, utilizedRat))
	
	return account, nil
}

func (s *Service) GetLedger(ctx context.Context, partnerID uuid.UUID, page, pageSize int) ([]*LedgerEntry, int64, error) {
	return s.repo.GetLedger(ctx, partnerID, page, pageSize)
}

func (s *Service) GetAgingReport(ctx context.Context, buyerID uuid.UUID) ([]AgingBucket, error) {
	return s.repo.GetAgingReport(ctx, buyerID)
}

func formatDecimal(r *big.Rat) string {
	f, _ := r.Float64()
	return fmt.Sprintf("%.2f", f)
}
