package overlay

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"

	"saweria-be/internal/domain"
)

var ErrNotFound = errors.New("overlay settings not found")

type Service interface {
	GetSettings(ctx context.Context, userID string) (*domain.OverlaySettings, error)
	UpdateAlertRules(ctx context.Context, userID string, req AlertRulesRequest) error
	UpdateTemplate(ctx context.Context, userID string, req TemplateRequest) error
	UpdateFilterKata(ctx context.Context, userID, filterKata string) error
	UpdateSoundURL(ctx context.Context, userID string, soundURL *string) error
	ResetStreamKey(ctx context.Context, userID string) (string, error)
	UpdateMediashareRules(ctx context.Context, userID string, req MediashareRulesRequest) error
	UpdateMediashareTemplate(ctx context.Context, userID string, req MediashareTemplateRequest) error
	UpdateQRSettings(ctx context.Context, userID string, req QRSettingsRequest) error
	UpdateMilestoneSettings(ctx context.Context, userID string, req MilestoneSettingsRequest) error
}

type AlertRulesRequest struct {
	GifSetting        bool    `json:"gif_setting"`
	TTSVariant        *string `json:"tts_variant"`
	MinimumAlert      int64   `json:"minimum_alert"`
	MinimumMediashare int64   `json:"minimum_mediashare"`
	MinimumTTS        int64   `json:"minimum_tts"`
}

type TemplateRequest struct {
	BackgroundColor      *string `json:"background_color"`
	HighlightColor       *string `json:"highlight_color"`
	TextColor            *string `json:"text_color"`
	TemplateText         string  `json:"template_text"`
	NotificationDuration int     `json:"notification_duration"`
}

type MediashareRulesRequest struct {
	Enabled             bool  `json:"ms_enabled"`
	YtShorts            bool  `json:"ms_yt_shorts"`
	Tiktok              bool  `json:"ms_tiktok"`
	IgReels             bool  `json:"ms_ig_reels"`
	VoiceNote           bool  `json:"ms_voice_note"`
	MaxVideoDuration    int   `json:"ms_max_video_duration"`
	PricePerSecondVideo int64 `json:"ms_price_per_second_video"`
	MaxAudioDuration    int   `json:"ms_max_audio_duration"`
	PricePerSecondAudio int64 `json:"ms_price_per_second_audio"`
	MinAudio            int64 `json:"ms_min_audio"`
	MinimumMediashare   int64 `json:"minimum_mediashare"`
}

type MediashareTemplateRequest struct {
	BackgroundColor *string `json:"ms_background_color"`
	HighlightColor  *string `json:"ms_highlight_color"`
	TextColor       *string `json:"ms_text_color"`
	TemplateText    string  `json:"ms_template_text"`
	NoBorder        bool    `json:"ms_no_border"`
	FontWeight      int     `json:"ms_font_weight"`
	FontFamily      string  `json:"ms_font_family"`
	ShowNsfw        bool    `json:"ms_show_nsfw"`
}

type QRSettingsRequest struct {
	BackgroundColor string `json:"qr_background_color"`
	BarcodeColor    string `json:"qr_barcode_color"`
	LabelTop        string `json:"qr_label_top"`
	LabelBottom     string `json:"qr_label_bottom"`
	FontFamily      string `json:"qr_font_family"`
}

type MilestoneSettingsRequest struct {
	Title       string  `json:"ms_title"`
	Target      int64   `json:"ms_target"`
	StartDate   *string `json:"ms_start_date"`
	BgColor     string  `json:"ms_bg_color"`
	TextColor   string  `json:"ms_text_color_ms"`
	NoBorder    bool    `json:"ms_no_border_ms"`
	FontWeight  int     `json:"ms_font_weight_ms"`
	FontTitle   string  `json:"ms_font_title"`
	FontContent string  `json:"ms_font_content"`
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetSettings(ctx context.Context, userID string) (*domain.OverlaySettings, error) {
	settings, err := s.repo.FindByUserID(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		// auto-create defaults if not yet exist
		settings, err = s.repo.CreateDefault(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("overlay.GetSettings: create default: %w", err)
		}
		return settings, nil
	}
	if err != nil {
		return nil, fmt.Errorf("overlay.GetSettings: %w", err)
	}
	return settings, nil
}

func (s *service) UpdateAlertRules(ctx context.Context, userID string, req AlertRulesRequest) error {
	if _, err := s.GetSettings(ctx, userID); err != nil {
		return err
	}
	return s.repo.UpdateAlertRules(ctx, userID, AlertRulesFields{
		GifSetting:        req.GifSetting,
		TTSVariant:        req.TTSVariant,
		MinimumAlert:      req.MinimumAlert,
		MinimumMediashare: req.MinimumMediashare,
		MinimumTTS:        req.MinimumTTS,
	})
}

func (s *service) UpdateTemplate(ctx context.Context, userID string, req TemplateRequest) error {
	if _, err := s.GetSettings(ctx, userID); err != nil {
		return err
	}
	return s.repo.UpdateTemplate(ctx, userID, TemplateFields{
		BackgroundColor:      req.BackgroundColor,
		HighlightColor:       req.HighlightColor,
		TextColor:            req.TextColor,
		TemplateText:         req.TemplateText,
		NotificationDuration: req.NotificationDuration,
	})
}

func (s *service) UpdateFilterKata(ctx context.Context, userID, filterKata string) error {
	if _, err := s.GetSettings(ctx, userID); err != nil {
		return err
	}
	return s.repo.UpdateFilterKata(ctx, userID, filterKata)
}

func (s *service) UpdateSoundURL(ctx context.Context, userID string, soundURL *string) error {
	if _, err := s.GetSettings(ctx, userID); err != nil {
		return err
	}
	return s.repo.UpdateSoundURL(ctx, userID, soundURL)
}

func (s *service) ResetStreamKey(ctx context.Context, userID string) (string, error) {
	if _, err := s.GetSettings(ctx, userID); err != nil {
		return "", err
	}

	// generate 32 random bytes → hex string (64 chars) as the plain key
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("overlay.ResetStreamKey: generate key: %w", err)
	}
	plainKey := hex.EncodeToString(raw)

	// hash with SHA-256 for storage
	sum := sha256.Sum256([]byte(plainKey))
	keyHash := hex.EncodeToString(sum[:])

	if err := s.repo.UpdateStreamKeyHash(ctx, userID, keyHash, plainKey); err != nil {
		return "", fmt.Errorf("overlay.ResetStreamKey: %w", err)
	}
	return plainKey, nil
}

func (s *service) UpdateMediashareRules(ctx context.Context, userID string, req MediashareRulesRequest) error {
	if _, err := s.GetSettings(ctx, userID); err != nil {
		return err
	}
	return s.repo.UpdateMediashareRules(ctx, userID, MediashareRulesFields{
		Enabled:             req.Enabled,
		YtShorts:            req.YtShorts,
		Tiktok:              req.Tiktok,
		IgReels:             req.IgReels,
		VoiceNote:           req.VoiceNote,
		MaxVideoDuration:    req.MaxVideoDuration,
		PricePerSecondVideo: req.PricePerSecondVideo,
		MaxAudioDuration:    req.MaxAudioDuration,
		PricePerSecondAudio: req.PricePerSecondAudio,
		MinAudio:            req.MinAudio,
		MinimumMediashare:   req.MinimumMediashare,
	})
}

func (s *service) UpdateMediashareTemplate(ctx context.Context, userID string, req MediashareTemplateRequest) error {
	if _, err := s.GetSettings(ctx, userID); err != nil {
		return err
	}
	return s.repo.UpdateMediashareTemplate(ctx, userID, MediashareTemplateFields{
		BackgroundColor: req.BackgroundColor,
		HighlightColor:  req.HighlightColor,
		TextColor:       req.TextColor,
		TemplateText:    req.TemplateText,
		NoBorder:        req.NoBorder,
		FontWeight:      req.FontWeight,
		FontFamily:      req.FontFamily,
		ShowNsfw:        req.ShowNsfw,
	})
}

func (s *service) UpdateQRSettings(ctx context.Context, userID string, req QRSettingsRequest) error {
	if _, err := s.GetSettings(ctx, userID); err != nil {
		return err
	}
	return s.repo.UpdateQRSettings(ctx, userID, QRSettingsFields{
		BackgroundColor: req.BackgroundColor,
		BarcodeColor:    req.BarcodeColor,
		LabelTop:        req.LabelTop,
		LabelBottom:     req.LabelBottom,
		FontFamily:      req.FontFamily,
	})
}

func (s *service) UpdateMilestoneSettings(ctx context.Context, userID string, req MilestoneSettingsRequest) error {
	if _, err := s.GetSettings(ctx, userID); err != nil {
		return err
	}
	return s.repo.UpdateMilestoneSettings(ctx, userID, MilestoneSettingsFields{
		Title:       req.Title,
		Target:      req.Target,
		StartDate:   req.StartDate,
		BgColor:     req.BgColor,
		TextColor:   req.TextColor,
		NoBorder:    req.NoBorder,
		FontWeight:  req.FontWeight,
		FontTitle:   req.FontTitle,
		FontContent: req.FontContent,
	})
}
