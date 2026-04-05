package widgets

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type LeaderboardEntry struct {
	DonorName string `db:"donor_name" json:"donorName"`
	Total     int64  `db:"total"      json:"total"`
}

type Repository interface {
	GetLeaderboard(ctx context.Context, streamerID string, limit int) ([]LeaderboardEntry, error)
	GetTotalRaised(ctx context.Context, streamerID string) (int64, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetLeaderboard(ctx context.Context, streamerID string, limit int) ([]LeaderboardEntry, error) {
	var list []LeaderboardEntry
	err := r.db.SelectContext(ctx, &list, `
		SELECT donor_name, SUM(amount) AS total
		FROM donations
		WHERE streamer_id = $1 AND payment_status = 'success'
		GROUP BY donor_name
		ORDER BY total DESC
		LIMIT $2`,
		streamerID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("widgets.GetLeaderboard: %w", err)
	}
	return list, nil
}

func (r *repository) GetTotalRaised(ctx context.Context, streamerID string) (int64, error) {
	var total int64
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(amount), 0)
		FROM donations
		WHERE streamer_id = $1 AND payment_status = 'success'`,
		streamerID,
	).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("widgets.GetTotalRaised: %w", err)
	}
	return total, nil
}
