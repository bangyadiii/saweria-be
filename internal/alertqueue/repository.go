package alertqueue

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// Item represents one row in the alert_queue table.
type Item struct {
	ID          string     `db:"id"`
	StreamerID  string     `db:"streamer_id"`
	DonationID  string     `db:"donation_id"`
	Status      string     `db:"status"`
	CreatedAt   time.Time  `db:"created_at"`
	ProcessedAt *time.Time `db:"processed_at"`
}

type Repository interface {
	// Enqueue inserts a new pending alert.
	Enqueue(ctx context.Context, streamerID, donationID string) error
	// ClaimNext atomically claims the oldest pending alert for a streamer.
	// Returns nil, nil when the queue is empty.
	ClaimNext(ctx context.Context, streamerID string) (*Item, error)
	// MarkDone sets the status to "sent" or "skipped".
	MarkDone(ctx context.Context, id, status string) error
	// GetPendingStreamerIDs returns distinct streamer IDs that have unfinished alerts.
	GetPendingStreamerIDs(ctx context.Context) ([]string, error)
	// ResetStaleProcessing moves "processing" rows back to "pending" on startup.
	ResetStaleProcessing(ctx context.Context) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Enqueue(ctx context.Context, streamerID, donationID string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO alert_queue (streamer_id, donation_id) VALUES ($1, $2)`,
		streamerID, donationID,
	)
	if err != nil {
		return fmt.Errorf("alertqueue.Enqueue: %w", err)
	}
	return nil
}

func (r *repository) ClaimNext(ctx context.Context, streamerID string) (*Item, error) {
	var item Item
	err := r.db.QueryRowxContext(ctx, `
		UPDATE alert_queue
		SET status       = 'processing',
		    processed_at = NOW()
		WHERE id = (
			SELECT id FROM alert_queue
			WHERE streamer_id = $1 AND status = 'pending'
			ORDER BY created_at ASC
			LIMIT 1
		)
		RETURNING id, streamer_id, donation_id, status, created_at, processed_at`,
		streamerID,
	).StructScan(&item)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("alertqueue.ClaimNext: %w", err)
	}
	return &item, nil
}

func (r *repository) MarkDone(ctx context.Context, id, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE alert_queue SET status = $1, processed_at = NOW() WHERE id = $2`,
		status, id,
	)
	if err != nil {
		return fmt.Errorf("alertqueue.MarkDone: %w", err)
	}
	return nil
}

func (r *repository) GetPendingStreamerIDs(ctx context.Context) ([]string, error) {
	var ids []string
	err := r.db.SelectContext(ctx, &ids,
		`SELECT DISTINCT streamer_id::text FROM alert_queue
		 WHERE status IN ('pending', 'processing')`,
	)
	if err != nil {
		return nil, fmt.Errorf("alertqueue.GetPendingStreamerIDs: %w", err)
	}
	return ids, nil
}

func (r *repository) ResetStaleProcessing(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE alert_queue SET status = 'pending' WHERE status = 'processing'`,
	)
	if err != nil {
		return fmt.Errorf("alertqueue.ResetStaleProcessing: %w", err)
	}
	return nil
}
