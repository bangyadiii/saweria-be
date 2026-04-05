package donation

import (
	"context"
	"fmt"

	"saweria-be/internal/domain"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	Create(ctx context.Context, d *domain.Donation) (*domain.Donation, error)
	FindByID(ctx context.Context, id string) (*domain.Donation, error)
	FindByStreamerID(ctx context.Context, streamerID string, limit, offset int) ([]*domain.Donation, error)
	FindByMidtransOrderID(ctx context.Context, orderID string) (*domain.Donation, error)
	UpdateStatus(ctx context.Context, id, status string, netAmount int64) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, d *domain.Donation) (*domain.Donation, error) {
	query := `
		INSERT INTO donations
		  (streamer_id, donor_name, amount, net_amount, platform_fee,
		   message, media_url, media_shown, payment_method,
		   payment_status, midtrans_order_id, payment_token)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING *`
	var created domain.Donation
	err := r.db.QueryRowxContext(ctx, query,
		d.StreamerID, d.DonorName, d.Amount, d.NetAmount, d.PlatformFee,
		d.Message, d.MediaURL, d.MediaShown, d.PaymentMethod,
		d.PaymentStatus, d.MidtransOrderID, d.PaymentToken,
	).StructScan(&created)
	if err != nil {
		return nil, fmt.Errorf("donation.Create: %w", err)
	}
	return &created, nil
}

func (r *repository) FindByID(ctx context.Context, id string) (*domain.Donation, error) {
	var d domain.Donation
	err := r.db.GetContext(ctx, &d, `SELECT * FROM donations WHERE id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("donation.FindByID: %w", err)
	}
	return &d, nil
}

func (r *repository) FindByStreamerID(ctx context.Context, streamerID string, limit, offset int) ([]*domain.Donation, error) {
	var list []*domain.Donation
	err := r.db.SelectContext(ctx, &list,
		`SELECT * FROM donations WHERE streamer_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		streamerID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("donation.FindByStreamerID: %w", err)
	}
	return list, nil
}

func (r *repository) FindByMidtransOrderID(ctx context.Context, orderID string) (*domain.Donation, error) {
	var d domain.Donation
	err := r.db.GetContext(ctx, &d, `SELECT * FROM donations WHERE midtrans_order_id = $1`, orderID)
	if err != nil {
		return nil, fmt.Errorf("donation.FindByMidtransOrderID: %w", err)
	}
	return &d, nil
}

func (r *repository) UpdateStatus(ctx context.Context, id, status string, netAmount int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE donations
		SET payment_status = $1,
		    net_amount     = $2,
		    updated_at     = NOW()
		WHERE id = $3`,
		status, netAmount, id,
	)
	if err != nil {
		return fmt.Errorf("donation.UpdateStatus: %w", err)
	}
	return nil
}
