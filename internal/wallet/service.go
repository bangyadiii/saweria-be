package wallet

import (
	"context"
	"errors"
	"fmt"

	"saweria-be/internal/domain"
)

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrMinimumCashout      = errors.New("minimum cashout amount is 50000")
)

const minimumCashout int64 = 50_000

type CashoutRequest struct {
	Amount        int64  `json:"amount"          binding:"required,min=50000"`
	BankName      string `json:"bank_name"       binding:"required"`
	AccountNumber string `json:"account_number"  binding:"required"`
	AccountName   string `json:"account_name"    binding:"required"`
}

type Service interface {
	GetBalance(ctx context.Context, userID string) (int64, error)
	RequestCashout(ctx context.Context, userID string, req CashoutRequest) (*domain.Cashout, error)
	GetCashoutHistory(ctx context.Context, userID string, page, pageSize int) ([]*domain.Cashout, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetBalance(ctx context.Context, userID string) (int64, error) {
	balance, err := s.repo.GetBalance(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("wallet.GetBalance: %w", err)
	}
	return balance, nil
}

func (s *service) RequestCashout(ctx context.Context, userID string, req CashoutRequest) (*domain.Cashout, error) {
	if req.Amount < minimumCashout {
		return nil, ErrMinimumCashout
	}

	if err := s.repo.DeductBalance(ctx, userID, req.Amount); err != nil {
		return nil, err
	}

	cashout := &domain.Cashout{
		UserID:        userID,
		Amount:        req.Amount,
		Fee:           0,
		NetAmount:     req.Amount,
		BankName:      req.BankName,
		AccountNumber: req.AccountNumber,
		AccountName:   req.AccountName,
		Status:        domain.CashoutStatusPending,
	}
	created, err := s.repo.CreateCashout(ctx, cashout)
	if err != nil {
		// attempt to roll back balance deduction on failure
		_ = s.repo.AddBalance(ctx, userID, req.Amount)
		return nil, fmt.Errorf("wallet.RequestCashout: %w", err)
	}
	return created, nil
}

func (s *service) GetCashoutHistory(ctx context.Context, userID string, page, pageSize int) ([]*domain.Cashout, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	list, err := s.repo.GetCashoutHistory(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("wallet.GetCashoutHistory: %w", err)
	}
	return list, nil
}
