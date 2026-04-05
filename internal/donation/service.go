package donation

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"saweria-be/internal/domain"
)

var (
	ErrStreamerNotFound = errors.New("streamer not found")
	ErrInvalidMediaURL  = errors.New("media URL must be from youtube.com or tiktok.com")
	ErrNotFound         = errors.New("donation not found")
)

// UserRepository is the minimal interface donation service needs from the user package
type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
}

// OverlayRepository is the minimal interface donation service needs from the overlay package
type OverlayRepository interface {
	FindByUserID(ctx context.Context, userID string) (*domain.OverlaySettings, error)
}

// MidtransClient defines the external payment interface (Core API)
type MidtransClient interface {
	CreateCharge(orderID string, amount int64, donorName, paymentType, bank string) (*domain.ChargeResult, error)
}

type ChargeResponse struct {
	OrderID     string `json:"orderId"`
	PaymentType string `json:"paymentType"`
	// bank_transfer
	Bank     string `json:"bank,omitempty"`
	VANumber string `json:"vaNumber,omitempty"`
	// mandiri echannel
	BillerCode string `json:"billerCode,omitempty"`
	BillKey    string `json:"billKey,omitempty"`
	// e-wallet
	QRCodeURL   string `json:"qrCodeUrl,omitempty"`
	DeepLinkURL string `json:"deepLinkUrl,omitempty"`
}

type SubmitRequest struct {
	DonorName   string `json:"donorName"   binding:"required"`
	Amount      int64  `json:"amount"      binding:"required,min=1000"`
	Message     string `json:"message"`
	MediaURL    string `json:"mediaUrl"`
	PaymentType string `json:"paymentType" binding:"required"` // bank_transfer | echannel | gopay | shopeepay
	Bank        string `json:"bank"`                           // bca | bni | bri | permata | mandiri
}

type Service interface {
	Submit(ctx context.Context, username string, req SubmitRequest, feePercent float64) (*ChargeResponse, error)
	GetHistory(ctx context.Context, streamerID string, page, pageSize int) ([]*domain.Donation, error)
	GetDetail(ctx context.Context, streamerID, donationID string) (*domain.Donation, error)
}

type service struct {
	repo        Repository
	userRepo    UserRepository
	overlayRepo OverlayRepository
	midtrans    MidtransClient
}

func NewService(repo Repository, userRepo UserRepository, overlayRepo OverlayRepository, mt MidtransClient) Service {
	return &service{repo: repo, userRepo: userRepo, overlayRepo: overlayRepo, midtrans: mt}
}

func (s *service) Submit(ctx context.Context, username string, req SubmitRequest, feePercent float64) (*ChargeResponse, error) {
	streamer, err := s.userRepo.FindByUsername(ctx, username)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrStreamerNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("donation.Submit: find streamer: %w", err)
	}

	// validate media URL domain
	var mediaURL *string
	mediaShown := false
	if req.MediaURL != "" {
		if err := validateMediaURL(req.MediaURL); err != nil {
			return nil, err
		}
		mediaURL = &req.MediaURL

		// check if media qualifies to be shown
		settings, sErr := s.overlayRepo.FindByUserID(ctx, streamer.ID)
		if sErr == nil && req.Amount >= settings.MinimumMediashare {
			mediaShown = true
		}
	}

	platformFee := int64(float64(req.Amount) * feePercent / 100)
	netAmount := req.Amount - platformFee

	orderID := fmt.Sprintf("%s-%d-%s", username, time.Now().UnixMilli(), randomSuffix(6))

	charge, err := s.midtrans.CreateCharge(orderID, req.Amount, req.DonorName, req.PaymentType, req.Bank)
	if err != nil {
		return nil, fmt.Errorf("donation.Submit: create charge: %w", err)
	}

	d := &domain.Donation{
		StreamerID:      streamer.ID,
		DonorName:       req.DonorName,
		Amount:          req.Amount,
		NetAmount:       netAmount,
		PlatformFee:     platformFee,
		Message:         req.Message,
		MediaURL:        mediaURL,
		MediaShown:      mediaShown,
		PaymentMethod:   req.PaymentType,
		PaymentStatus:   domain.PaymentStatusPending,
		MidtransOrderID: orderID,
		PaymentToken:    charge.TransactionID,
	}

	_, err = s.repo.Create(ctx, d)
	if err != nil {
		return nil, fmt.Errorf("donation.Submit: save: %w", err)
	}

	return &ChargeResponse{
		OrderID:     orderID,
		PaymentType: req.PaymentType,
		Bank:        charge.Bank,
		VANumber:    charge.VANumber,
		BillerCode:  charge.BillerCode,
		BillKey:     charge.BillKey,
		QRCodeURL:   charge.QRCodeURL,
		DeepLinkURL: charge.DeepLinkURL,
	}, nil
}

func (s *service) GetHistory(ctx context.Context, streamerID string, page, pageSize int) ([]*domain.Donation, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	list, err := s.repo.FindByStreamerID(ctx, streamerID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("donation.GetHistory: %w", err)
	}
	return list, nil
}

func (s *service) GetDetail(ctx context.Context, streamerID, donationID string) (*domain.Donation, error) {
	d, err := s.repo.FindByID(ctx, donationID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("donation.GetDetail: %w", err)
	}
	if d.StreamerID != streamerID {
		return nil, ErrNotFound
	}
	return d, nil
}

var allowedMediaHosts = []string{"youtube.com", "youtu.be", "tiktok.com", "www.youtube.com", "www.tiktok.com"}

func validateMediaURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ErrInvalidMediaURL
	}
	host := strings.ToLower(parsed.Host)
	for _, allowed := range allowedMediaHosts {
		if host == allowed {
			return nil
		}
	}
	return ErrInvalidMediaURL
}

func randomSuffix(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[time.Now().UnixNano()%int64(len(chars))]
	}
	return string(b)
}
