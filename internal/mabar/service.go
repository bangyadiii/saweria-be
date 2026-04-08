package mabar

import (
	"context"
	"strings"

	"saweria-be/internal/domain"
)

type Service interface {
	TryEnqueue(ctx context.Context, donation *domain.Donation, settings *domain.OverlaySettings) error
	GetQueue(ctx context.Context, streamerID string) ([]*domain.MabarQueueItem, error)
	MarkDone(ctx context.Context, id, streamerID string) error
	Reorder(ctx context.Context, streamerID string, items []ReorderItem) error
	ClearAll(ctx context.Context, streamerID string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) TryEnqueue(ctx context.Context, donation *domain.Donation, settings *domain.OverlaySettings) error {
	if !settings.MabarEnabled {
		return nil
	}
	if donation.Amount < settings.MabarMinimumAmount {
		return nil
	}

	keyword := strings.TrimSpace(settings.MabarKeyword)
	msg := strings.TrimSpace(donation.Message)
	if !strings.HasPrefix(strings.ToLower(msg), strings.ToLower(keyword)) {
		return nil
	}

	ingameUsername := strings.TrimSpace(msg[len(keyword):])

	tier := "bronze"
	if donation.Amount >= settings.MabarGoldThreshold {
		tier = "gold"
	} else if donation.Amount >= settings.MabarSilverThreshold {
		tier = "silver"
	}

	return s.repo.Enqueue(ctx, donation.StreamerID, donation.ID, donation.DonorName, ingameUsername, donation.Amount, tier)
}

func (s *service) GetQueue(ctx context.Context, streamerID string) ([]*domain.MabarQueueItem, error) {
	return s.repo.GetQueue(ctx, streamerID)
}

func (s *service) MarkDone(ctx context.Context, id, streamerID string) error {
	// Fetch queue first to validate ownership before marking done
	items, err := s.repo.GetQueue(ctx, streamerID)
	if err != nil {
		return err
	}
	for _, item := range items {
		if item.ID == id {
			return s.repo.MarkDone(ctx, id)
		}
	}
	return ErrNotFound
}

func (s *service) Reorder(ctx context.Context, streamerID string, items []ReorderItem) error {
	if len(items) == 0 {
		return nil
	}
	// Validate ownership: all IDs must belong to this streamer's waiting queue
	queue, err := s.repo.GetQueue(ctx, streamerID)
	if err != nil {
		return err
	}
	ownedIDs := make(map[string]struct{}, len(queue))
	for _, item := range queue {
		ownedIDs[item.ID] = struct{}{}
	}
	for _, req := range items {
		if _, ok := ownedIDs[req.ID]; !ok {
			return ErrNotFound
		}
	}
	return s.repo.Reorder(ctx, items)
}

func (s *service) ClearAll(ctx context.Context, streamerID string) error {
	return s.repo.ClearAll(ctx, streamerID)
}
