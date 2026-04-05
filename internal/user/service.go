package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"saweria-be/internal/domain"
)

var ErrNotFound = errors.New("user not found")

type Service interface {
	GetMe(ctx context.Context, userID string) (*domain.User, error)
	GetPublicProfile(ctx context.Context, username string) (*domain.User, error)
	UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) (*domain.User, error)
}

type UpdateProfileRequest struct {
	DisplayName  string  `json:"displayName"`
	Bio          string  `json:"bio"`
	ProfileImage *string `json:"profileImage"`
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetMe(ctx context.Context, userID string) (*domain.User, error) {
	u, err := s.repo.FindByID(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("user.GetMe: %w", err)
	}
	return u, nil
}

func (s *service) GetPublicProfile(ctx context.Context, username string) (*domain.User, error) {
	u, err := s.repo.FindByUsername(ctx, username)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("user.GetPublicProfile: %w", err)
	}
	return u, nil
}

func (s *service) UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) (*domain.User, error) {
	u, err := s.repo.Update(ctx, userID, UpdateFields{
		DisplayName:  req.DisplayName,
		Bio:          req.Bio,
		ProfileImage: req.ProfileImage,
	})
	if err != nil {
		return nil, fmt.Errorf("user.UpdateProfile: %w", err)
	}
	return u, nil
}
