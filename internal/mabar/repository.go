package mabar

import (
	"context"
	"database/sql"
	"fmt"

	"saweria-be/internal/domain"

	"github.com/jmoiron/sqlx"
)

type ReorderItem struct {
	ID            string `json:"id"`
	PriorityOrder int    `json:"priority_order"`
}

type Repository interface {
	Enqueue(ctx context.Context, streamerID, donationID, donorName, ingameUsername string, amount int64, tier string) error
	GetQueue(ctx context.Context, streamerID string) ([]*domain.MabarQueueItem, error)
	MarkDone(ctx context.Context, id string) error
	Reorder(ctx context.Context, items []ReorderItem) error
	ClearAll(ctx context.Context, streamerID string) error
	MaxPriorityOrder(ctx context.Context, streamerID, tier string) (int, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Enqueue(ctx context.Context, streamerID, donationID, donorName, ingameUsername string, amount int64, tier string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO mabar_queue
		    (streamer_id, donation_id, donor_name, ingame_username, amount, priority_tier, priority_order, status)
		VALUES ($1, $2, $3, $4, $5, $6,
		    COALESCE((SELECT MAX(priority_order) FROM mabar_queue
		              WHERE streamer_id = $1 AND priority_tier = $6 AND status = 'waiting'), 0) + 1,
		    'waiting')`,
		streamerID, donationID, donorName, ingameUsername, amount, tier,
	)
	if err != nil {
		return fmt.Errorf("mabar.Enqueue: %w", err)
	}
	return nil
}

func (r *repository) GetQueue(ctx context.Context, streamerID string) ([]*domain.MabarQueueItem, error) {
	var items []*domain.MabarQueueItem
	err := r.db.SelectContext(ctx, &items, `
		SELECT * FROM mabar_queue
		WHERE streamer_id = $1 AND status = 'waiting'
		ORDER BY
		    CASE priority_tier
		        WHEN 'gold'   THEN 1
		        WHEN 'silver' THEN 2
		        ELSE               3
		    END,
		    priority_order ASC,
		    created_at ASC`,
		streamerID,
	)
	if err != nil {
		return nil, fmt.Errorf("mabar.GetQueue: %w", err)
	}
	return items, nil
}

func (r *repository) MarkDone(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE mabar_queue
		SET status = 'done', updated_at = NOW()
		WHERE id = $1 AND status = 'waiting'`,
		id,
	)
	if err != nil {
		return fmt.Errorf("mabar.MarkDone: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("mabar.MarkDone: %w", sql.ErrNoRows)
	}
	return nil
}

func (r *repository) Reorder(ctx context.Context, items []ReorderItem) error {
	if len(items) == 0 {
		return nil
	}
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("mabar.Reorder: begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	for _, item := range items {
		if _, err := tx.ExecContext(ctx,
			`UPDATE mabar_queue SET priority_order = $1, updated_at = NOW() WHERE id = $2`,
			item.PriorityOrder, item.ID,
		); err != nil {
			return fmt.Errorf("mabar.Reorder: update %s: %w", item.ID, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("mabar.Reorder: commit: %w", err)
	}
	return nil
}

func (r *repository) ClearAll(ctx context.Context, streamerID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM mabar_queue WHERE streamer_id = $1 AND status = 'waiting'`,
		streamerID,
	)
	if err != nil {
		return fmt.Errorf("mabar.ClearAll: %w", err)
	}
	return nil
}

func (r *repository) MaxPriorityOrder(ctx context.Context, streamerID, tier string) (int, error) {
	var max sql.NullInt64
	err := r.db.QueryRowContext(ctx,
		`SELECT MAX(priority_order) FROM mabar_queue WHERE streamer_id = $1 AND priority_tier = $2 AND status = 'waiting'`,
		streamerID, tier,
	).Scan(&max)
	if err != nil {
		return 0, fmt.Errorf("mabar.MaxPriorityOrder: %w", err)
	}
	if !max.Valid {
		return 0, nil
	}
	return int(max.Int64), nil
}
