package wallet

import (
	"context"
	"fmt"

	"saweria-be/internal/domain"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	GetBalance(ctx context.Context, userID string) (int64, error)
	AddBalance(ctx context.Context, userID string, amount int64) error
	DeductBalance(ctx context.Context, userID string, amount int64) error
	CreateCashout(ctx context.Context, c *domain.Cashout) (*domain.Cashout, error)
	GetCashoutHistory(ctx context.Context, userID string, limit, offset int) ([]*domain.Cashout, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetBalance(ctx context.Context, userID string) (int64, error) {
	var balance int64
	err := r.db.QueryRowContext(ctx, `SELECT balance FROM users WHERE id = $1`, userID).Scan(&balance)
	if err != nil {
		return 0, fmt.Errorf("wallet.GetBalance: %w", err)
	}
	return balance, nil
}

func (r *repository) AddBalance(ctx context.Context, userID string, amount int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET balance = balance + $1, updated_at = NOW() WHERE id = $2`,
		amount, userID,
	)
	if err != nil {
		return fmt.Errorf("wallet.AddBalance: %w", err)
	}
	return nil
}

func (r *repository) DeductBalance(ctx context.Context, userID string, amount int64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE users SET balance = balance - $1, updated_at = NOW()
		WHERE id = $2 AND balance >= $1`,
		amount, userID,
	)
	if err != nil {
		return fmt.Errorf("wallet.DeductBalance: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrInsufficientBalance
	}
	return nil
}

func (r *repository) CreateCashout(ctx context.Context, c *domain.Cashout) (*domain.Cashout, error) {
	query := `
		INSERT INTO cashouts (user_id, amount, fee, net_amount, bank_name, account_number, account_name, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING *`
	var created domain.Cashout
	err := r.db.QueryRowxContext(ctx, query,
		c.UserID, c.Amount, c.Fee, c.NetAmount, c.BankName, c.AccountNumber, c.AccountName, c.Status,
	).StructScan(&created)
	if err != nil {
		return nil, fmt.Errorf("wallet.CreateCashout: %w", err)
	}
	return &created, nil
}

func (r *repository) GetCashoutHistory(ctx context.Context, userID string, limit, offset int) ([]*domain.Cashout, error) {
	var list []*domain.Cashout
	err := r.db.SelectContext(ctx, &list,
		`SELECT * FROM cashouts WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("wallet.GetCashoutHistory: %w", err)
	}
	return list, nil
}
