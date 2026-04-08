package user

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"saweria-be/internal/domain"
)

var ErrNotFound = errors.New("user not found")
var ErrWebhookURLRequired = errors.New("webhook URL required")
var ErrWebhookURLInvalid = errors.New("webhook URL must start with http:// or https://")

type Service interface {
	GetMe(ctx context.Context, userID string) (*domain.User, error)
	GetPublicProfile(ctx context.Context, username string) (*domain.User, error)
	UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) (*domain.User, error)
	UploadProfileImage(ctx context.Context, userID string, imageURL string) (*domain.User, error)
	UpdateWebhookSettings(ctx context.Context, userID string, req UpdateWebhookRequest) (*domain.User, error)
	ResetWebhookToken(ctx context.Context, userID string) (*domain.User, error)
	TestWebhook(ctx context.Context, userID string) error
}

type UpdateProfileRequest struct {
	DisplayName  string  `json:"display_name"`
	Bio          string  `json:"bio"`
	ProfileImage *string `json:"profile_image"`
}

type UpdateWebhookRequest struct {
	Enabled bool    `json:"enabled"`
	URL     *string `json:"url"`
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

func (s *service) UploadProfileImage(ctx context.Context, userID string, imageURL string) (*domain.User, error) {
	current, err := s.repo.FindByID(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("user.UploadProfileImage: %w", err)
	}
	u, err := s.repo.Update(ctx, userID, UpdateFields{
		DisplayName:  current.DisplayName,
		Bio:          current.Bio,
		ProfileImage: &imageURL,
	})
	if err != nil {
		return nil, fmt.Errorf("user.UploadProfileImage: %w", err)
	}
	return u, nil
}

func (s *service) UpdateWebhookSettings(ctx context.Context, userID string, req UpdateWebhookRequest) (*domain.User, error) {
	if req.URL != nil && *req.URL != "" {
		lower := strings.ToLower(*req.URL)
		if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
			return nil, ErrWebhookURLInvalid
		}
	}

	// Preserve existing token (only reset generates a new one)
	existing, err := s.repo.FindByID(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("user.UpdateWebhookSettings: %w", err)
	}

	u, err := s.repo.UpdateWebhook(ctx, userID, WebhookFields{
		Enabled: req.Enabled,
		URL:     req.URL,
		Token:   existing.WebhookToken,
	})
	if err != nil {
		return nil, fmt.Errorf("user.UpdateWebhookSettings: %w", err)
	}
	return u, nil
}

func (s *service) ResetWebhookToken(ctx context.Context, userID string) (*domain.User, error) {
	existing, err := s.repo.FindByID(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("user.ResetWebhookToken: %w", err)
	}

	token, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("user.ResetWebhookToken: generate token: %w", err)
	}

	u, err := s.repo.UpdateWebhook(ctx, userID, WebhookFields{
		Enabled: existing.WebhookEnabled,
		URL:     existing.WebhookURL,
		Token:   &token,
	})
	if err != nil {
		return nil, fmt.Errorf("user.ResetWebhookToken: %w", err)
	}
	return u, nil
}

// TestWebhook sends a fake donation payload to the user's configured webhook URL.
func (s *service) TestWebhook(ctx context.Context, userID string) error {
	u, err := s.repo.FindByID(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("user.TestWebhook: %w", err)
	}
	if u.WebhookURL == nil || *u.WebhookURL == "" {
		return ErrWebhookURLRequired
	}

	payload := map[string]any{
		"id":             "test-donation-id",
		"streamer_id":    u.ID,
		"donor_name":     "Tako Test",
		"amount":         10000,
		"net_amount":     9000,
		"platform_fee":   1000,
		"message":        "Ini adalah notifikasi tes dari Tako!",
		"media_url":      nil,
		"payment_method": "bank_transfer",
		"payment_status": "success",
		"created_at":     time.Now().UTC().Format(time.RFC3339),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("user.TestWebhook: marshal: %w", err)
	}

	signature := ""
	if u.WebhookToken != nil && *u.WebhookToken != "" {
		mac := hmac.New(sha256.New, []byte(*u.WebhookToken))
		mac.Write(body)
		signature = hex.EncodeToString(mac.Sum(nil))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, *u.WebhookURL, strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("user.TestWebhook: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tako-Signature", signature)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("user.TestWebhook: request failed: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

func generateToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
