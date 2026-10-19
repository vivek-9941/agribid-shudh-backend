package payment

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/agribid/agribid-shudh-backend/internal/db"
	apperrors "github.com/agribid/agribid-shudh-backend/internal/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	repo Repository
	pool *pgxpool.Pool
}

func NewService(repo Repository, pool *pgxpool.Pool) *Service {
	return &Service{
		repo: repo,
		pool: pool,
	}
}

func (s *Service) CheckCreditLimit(ctx context.Context, buyerID, sellerID uuid.UUID, amountStr string) error {
	account, err := s.repo.GetCreditAccount(ctx, buyerID, sellerID)
	if err != nil {
		return apperrors.Internal(err)
	}
	if account == nil {
		// Assume zero credit limit if no account exists
		return apperrors.BadRequest(apperrors.CodeCreditLimitExceeded, "no credit account established")
	}

	limitRat, _ := new(big.Rat).SetString(account.CreditLimit)
	utilizedRat, _ := new(big.Rat).SetString(account.CreditUtilized)
	requestedRat, _ := new(big.Rat).SetString(amountStr)

	availableRat := new(big.Rat).Sub(limitRat, utilizedRat)

	if requestedRat.Cmp(availableRat) > 0 {
		return apperrors.BadRequest(apperrors.CodeCreditLimitExceeded, fmt.Sprintf("credit limit exceeded. available: %s, requested: %s", formatDecimal(availableRat), amountStr))
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

	err := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		if err := s.repo.RecordPayment(ctx, pay); err != nil {
			return err
		}

		// Reduce credit utilized
		amtRat, _ := new(big.Rat).SetString(req.Amount)
		negAmt := new(big.Rat).Neg(amtRat)
		if err := s.repo.UpdateCredit(ctx, buyerID, sellerID, formatDecimal(negAmt)); err != nil {
			return err
		}

		// Compute running balance: previous_balance - payment_amount (debit reduces balance owed)
		prevBalance, err := s.repo.GetLastBalance(ctx, buyerID, sellerID)
		if err != nil {
			return fmt.Errorf("get last balance: %w", err)
		}
		prevRat, _ := new(big.Rat).SetString(prevBalance)
		newBalance := new(big.Rat).Sub(prevRat, amtRat)

		// Insert ledger entry for buyer with computed running balance
		entry := &LedgerEntry{
			PartnerID:      buyerID,
			CounterpartyID: sellerID,
			EntryType:      "payment",
			ReferenceID:    &pay.ID,
			Debit:          req.Amount,
			Credit:         "0.00",
			Balance:        formatDecimal(newBalance),
			EntryDate:      time.Now(),
			Description:    fmt.Sprintf("Payment via %s", req.Method),
		}
		if err := s.repo.InsertLedgerEntry(ctx, entry); err != nil {
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

// CreateCreditAccount creates a new credit account between buyer and seller.
func (s *Service) CreateCreditAccount(ctx context.Context, req *CreateCreditAccountRequest) (*CreditAccount, error) {
	buyerID, _ := uuid.Parse(req.BuyerID)
	sellerID, _ := uuid.Parse(req.SellerID)

	// Check if account already exists
	existing, err := s.repo.GetCreditAccount(ctx, buyerID, sellerID)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if existing != nil {
		return nil, apperrors.Conflict(apperrors.CodeConflict, "credit account already exists between these partners")
	}

	account := &CreditAccount{
		BuyerID:        buyerID,
		SellerID:       sellerID,
		CreditLimit:    req.CreditLimit,
		CreditUtilized: "0.00",
		PaymentTerms:   req.PaymentTerms,
		OverdueAmount:  "0.00",
	}

	if err := s.repo.CreateCreditAccount(ctx, account); err != nil {
		return nil, apperrors.Internal(err)
	}

	account.CreditAvailable = req.CreditLimit
	return account, nil
}

// UpdateCreditLimit updates the credit limit for an existing account.
func (s *Service) UpdateCreditLimit(ctx context.Context, accountID uuid.UUID, newLimit string) error {
	return s.repo.UpdateCreditLimit(ctx, accountID, newLimit)
}

// GetPaymentByID retrieves a payment by ID.
func (s *Service) GetPaymentByID(ctx context.Context, id uuid.UUID) (*Payment, error) {
	pay, err := s.repo.GetPaymentByID(ctx, id)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	if pay == nil {
		return nil, apperrors.NotFound("payment", id.String())
	}
	return pay, nil
}

// ListPayments lists payments for a partner.
func (s *Service) ListPayments(ctx context.Context, partnerID uuid.UUID, page, pageSize int) ([]*Payment, int64, error) {
	return s.repo.ListPayments(ctx, partnerID, page, pageSize)
}

func formatDecimal(r *big.Rat) string {
	f, _ := r.Float64()
	return fmt.Sprintf("%.2f", f)
}

