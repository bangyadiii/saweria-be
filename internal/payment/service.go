package payment

import (
	"context"
	"crypto/sha512"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"saweria-be/internal/domain"
)

// DonationRepository is the minimal interface the payment service needs
type DonationRepository interface {
	FindByMidtransOrderID(ctx context.Context, orderID string) (*domain.Donation, error)
	UpdateStatus(ctx context.Context, id, status string, netAmount int64) error
}

// WalletRepository is the minimal interface the payment service needs
type WalletRepository interface {
	AddBalance(ctx context.Context, userID string, amount int64) error
}

// AlertQueueEnqueuer enqueues a paid donation into the per-streamer alert queue.
type AlertQueueEnqueuer interface {
	Enqueue(ctx context.Context, streamerID, donationID string) error
}

type WebhookPayload struct {
	OrderID           string `json:"order_id"`
	StatusCode        string `json:"status_code"`
	GrossAmount       string `json:"gross_amount"`
	SignatureKey      string `json:"signature_key"`
	TransactionStatus string `json:"transaction_status"`
	FraudStatus       string `json:"fraud_status"`
	PaymentType       string `json:"payment_type"`
}

type Service interface {
	ProcessWebhook(ctx context.Context, payload WebhookPayload, serverKey string) error
}

type service struct {
	donationRepo DonationRepository
	walletRepo   WalletRepository
	alertQueue   AlertQueueEnqueuer
	feePercent   float64
}

func NewService(dr DonationRepository, wr WalletRepository, alertQueue AlertQueueEnqueuer, feePercent float64) Service {
	return &service{donationRepo: dr, walletRepo: wr, alertQueue: alertQueue, feePercent: feePercent}
}

func (s *service) ProcessWebhook(ctx context.Context, payload WebhookPayload, serverKey string) error {
	if !verifySignature(payload, serverKey) {
		return errors.New("invalid signature")
	}

	donation, err := s.donationRepo.FindByMidtransOrderID(ctx, payload.OrderID)
	if errors.Is(err, sql.ErrNoRows) {
		// idempotency: unknown order — silently ignore
		return nil
	}
	if err != nil {
		return fmt.Errorf("payment.ProcessWebhook: find donation: %w", err)
	}

	// idempotency: already processed
	if donation.PaymentStatus != domain.PaymentStatusPending {
		return nil
	}

	newStatus := resolveStatus(payload.TransactionStatus, payload.FraudStatus)
	if newStatus == "" {
		return nil
	}

	if err := s.donationRepo.UpdateStatus(ctx, donation.ID, newStatus, donation.NetAmount); err != nil {
		return fmt.Errorf("payment.ProcessWebhook: update status: %w", err)
	}

	if newStatus == domain.PaymentStatusSuccess {
		if err := s.walletRepo.AddBalance(ctx, donation.StreamerID, donation.NetAmount); err != nil {
			return fmt.Errorf("payment.ProcessWebhook: add balance: %w", err)
		}
		if err := s.alertQueue.Enqueue(ctx, donation.StreamerID, donation.ID); err != nil {
			log.Printf("payment.ProcessWebhook: enqueue alert: %v", err)
		}
	}

	return nil
}

// verifySignature checks Midtrans signature: SHA-512(order_id + status_code + gross_amount + server_key)
func verifySignature(p WebhookPayload, serverKey string) bool {
	plain := p.OrderID + p.StatusCode + p.GrossAmount + serverKey
	h := sha512.Sum512([]byte(plain))
	expected := fmt.Sprintf("%x", h)
	return expected == p.SignatureKey
}

func resolveStatus(txStatus, fraudStatus string) string {
	switch txStatus {
	case "capture":
		if fraudStatus == "challenge" {
			return domain.PaymentStatusPending
		}
		return domain.PaymentStatusSuccess
	case "settlement":
		return domain.PaymentStatusSuccess
	case "deny", "cancel", "failure":
		return domain.PaymentStatusFailed
	case "expire":
		return domain.PaymentStatusExpired
	}
	return ""
}
