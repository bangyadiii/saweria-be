package alertqueue

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"saweria-be/internal/domain"
)

// DonationRepository is the minimal interface the manager needs to fetch donation details.
type DonationRepository interface {
	FindByID(ctx context.Context, id string) (*domain.Donation, error)
}

// OverlayRepository is the minimal interface the manager needs for notification duration and stream key.
type OverlayRepository interface {
	FindByUserID(ctx context.Context, userID string) (*domain.OverlaySettings, error)
}

// WSHub sends a message to all OBS browser source connections for a given stream key hash.
type WSHub interface {
	Broadcast(streamKeyHash string, message []byte)
}

// Manager runs one serialised worker goroutine per streamer.
// Donations are delivered to OBS one at a time, spaced by NotificationDuration.
type Manager struct {
	repo        Repository
	donRepo     DonationRepository
	overlayRepo OverlayRepository
	hub         WSHub
	rootCtx     context.Context

	mu      sync.Mutex
	workers map[string]context.CancelFunc
}

func NewManager(
	rootCtx context.Context,
	repo Repository,
	donRepo DonationRepository,
	overlayRepo OverlayRepository,
	hub WSHub,
) *Manager {
	return &Manager{
		repo:        repo,
		donRepo:     donRepo,
		overlayRepo: overlayRepo,
		hub:         hub,
		rootCtx:     rootCtx,
		workers:     make(map[string]context.CancelFunc),
	}
}

// Enqueue persists the alert in the DB and ensures a worker is running for the streamer.
func (m *Manager) Enqueue(ctx context.Context, streamerID, donationID string) error {
	if err := m.repo.Enqueue(ctx, streamerID, donationID); err != nil {
		return err
	}
	m.ensureWorker(streamerID)
	return nil
}

// RecoverPending should be called once on startup.
// It resets any items left in "processing" state (from a previous crash)
// and restarts workers for all streamers that have pending alerts.
func (m *Manager) RecoverPending(ctx context.Context) error {
	if err := m.repo.ResetStaleProcessing(ctx); err != nil {
		return fmt.Errorf("alertqueue: recover stale processing: %w", err)
	}
	ids, err := m.repo.GetPendingStreamerIDs(ctx)
	if err != nil {
		return fmt.Errorf("alertqueue: get pending streamers: %w", err)
	}
	for _, id := range ids {
		m.ensureWorker(id)
	}
	log.Printf("alertqueue: recovered %d pending streamer queue(s)", len(ids))
	return nil
}

func (m *Manager) ensureWorker(streamerID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, running := m.workers[streamerID]; running {
		return
	}
	ctx, cancel := context.WithCancel(m.rootCtx)
	m.workers[streamerID] = cancel
	go m.runWorker(ctx, streamerID)
}

// runWorker processes one alert at a time for a single streamer.
// It exits when the queue is empty; the next Enqueue call will restart it.
func (m *Manager) runWorker(ctx context.Context, streamerID string) {
	defer func() {
		m.mu.Lock()
		delete(m.workers, streamerID)
		m.mu.Unlock()
	}()

	for {
		item, err := m.repo.ClaimNext(ctx, streamerID)
		if err != nil {
			log.Printf("alertqueue: ClaimNext error (streamer=%s): %v", streamerID, err)
			return
		}
		if item == nil {
			// Queue is empty — exit. Worker will be restarted on next Enqueue.
			return
		}

		donation, err := m.donRepo.FindByID(ctx, item.DonationID)
		if err != nil {
			log.Printf("alertqueue: donation not found %s: %v", item.DonationID, err)
			_ = m.repo.MarkDone(context.Background(), item.ID, "skipped")
			continue
		}

		settings, _ := m.overlayRepo.FindByUserID(ctx, streamerID)

		if settings != nil && settings.StreamKeyHash != nil && *settings.StreamKeyHash != "" {
			m.hub.Broadcast(*settings.StreamKeyHash, buildMessage(donation, settings))
		}

		duration := alertDuration(settings)

		select {
		case <-ctx.Done():
			// Server shutting down — mark skipped so it retries on next start.
			_ = m.repo.MarkDone(context.Background(), item.ID, "skipped")
			return
		case <-time.After(duration):
			_ = m.repo.MarkDone(ctx, item.ID, "sent")
		}
	}
}

func alertDuration(s *domain.OverlaySettings) time.Duration {
	if s != nil && s.NotificationDuration > 0 {
		return time.Duration(s.NotificationDuration) * time.Second
	}
	return 8 * time.Second
}

func buildMessage(d *domain.Donation, s *domain.OverlaySettings) []byte {
	mediaShown := "false"
	if d.MediaShown {
		mediaShown = "true"
	}
	durationMS := int(alertDuration(s).Milliseconds())
	msg := fmt.Sprintf(
		`{"event":"donation","id":%q,"donorName":%q,"amount":%d,"message":%q,"mediaUrl":%s,"mediaShown":%s,"duration":%d}`,
		d.ID, d.DonorName, d.Amount, d.Message, nullableStr(d.MediaURL), mediaShown, durationMS,
	)
	return []byte(msg)
}

func nullableStr(s *string) string {
	if s == nil {
		return "null"
	}
	return fmt.Sprintf("%q", *s)
}
