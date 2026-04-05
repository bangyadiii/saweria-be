package overlay

import (
	"context"
	"encoding/json"
	"fmt"

	"saweria-be/internal/domain"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	FindByUserID(ctx context.Context, userID string) (*domain.OverlaySettings, error)
	FindByStreamKeyHash(ctx context.Context, hash string) (*domain.OverlaySettings, error)
	CreateDefault(ctx context.Context, userID string) (*domain.OverlaySettings, error)
	UpdateAlertRules(ctx context.Context, userID string, f AlertRulesFields) error
	UpdateTemplate(ctx context.Context, userID string, f TemplateFields) error
	UpdateFilterKata(ctx context.Context, userID, filterKata string) error
	UpdateSoundURL(ctx context.Context, userID string, soundURL *string) error
	UpdateStreamKeyHash(ctx context.Context, userID, hash, plainKey string) error
	UpdateMediashareRules(ctx context.Context, userID string, f MediashareRulesFields) error
	UpdateMediashareTemplate(ctx context.Context, userID string, f MediashareTemplateFields) error
	UpdateQRSettings(ctx context.Context, userID string, f QRSettingsFields) error
	UpdateMilestoneSettings(ctx context.Context, userID string, f MilestoneSettingsFields) error
	UpdateSubathonSettings(ctx context.Context, userID string, f SubathonSettingsFields) error
}

type AlertRulesFields struct {
	GifSetting        bool
	TTSVariant        *string
	MinimumAlert      int64
	MinimumMediashare int64
	MinimumTTS        int64
}

type TemplateFields struct {
	BackgroundColor      *string
	HighlightColor       *string
	TextColor            *string
	TemplateText         string
	NotificationDuration int
}

type MediashareRulesFields struct {
	Enabled             bool
	YtShorts            bool
	Tiktok              bool
	IgReels             bool
	VoiceNote           bool
	MaxVideoDuration    int
	PricePerSecondVideo int64
	MaxAudioDuration    int
	PricePerSecondAudio int64
	MinAudio            int64
	MinimumMediashare   int64
}

type MediashareTemplateFields struct {
	BackgroundColor *string
	HighlightColor  *string
	TextColor       *string
	TemplateText    string
	NoBorder        bool
	FontWeight      int
	FontFamily      string
	ShowNsfw        bool
}

type QRSettingsFields struct {
	BackgroundColor string
	BarcodeColor    string
	LabelTop        string
	LabelBottom     string
	FontFamily      string
}

type MilestoneSettingsFields struct {
	Title       string
	Target      int64
	StartDate   *string
	BgColor     string
	TextColor   string
	NoBorder    bool
	FontWeight  int
	FontTitle   string
	FontContent string
}

type SubathonTimeRule struct {
	MinAmount int64 `json:"min_amount"`
	Hours     int   `json:"hours"`
	Minutes   int   `json:"minutes"`
	Seconds   int   `json:"seconds"`
}

type SubathonSettingsFields struct {
	InitialHours   int
	InitialMinutes int
	InitialSeconds int
	NoBorder       bool
	BgColor        string
	AutoPlay       bool
	TextColor      string
	FontWeight     int
	FontContent    string
	TimeRules      []SubathonTimeRule
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindByUserID(ctx context.Context, userID string) (*domain.OverlaySettings, error) {
	var s domain.OverlaySettings
	err := r.db.GetContext(ctx, &s, `SELECT * FROM overlay_settings WHERE user_id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("overlay.FindByUserID: %w", err)
	}
	return &s, nil
}

func (r *repository) FindByStreamKeyHash(ctx context.Context, hash string) (*domain.OverlaySettings, error) {
	var s domain.OverlaySettings
	// Match either the pre-computed hash column OR compute SHA-256 on the fly for older rows
	err := r.db.GetContext(ctx, &s, `
		SELECT * FROM overlay_settings
		WHERE stream_key_hash = $1
		   OR encode(sha256(stream_key::bytea), 'hex') = $1
		LIMIT 1`, hash)
	if err != nil {
		return nil, fmt.Errorf("overlay.FindByStreamKeyHash: %w", err)
	}
	return &s, nil
}

func (r *repository) CreateDefault(ctx context.Context, userID string) (*domain.OverlaySettings, error) {
	var s domain.OverlaySettings
	err := r.db.QueryRowxContext(ctx,
		`WITH new_key AS (
		    SELECT encode(gen_random_bytes(16), 'hex') AS k
		)
		INSERT INTO overlay_settings (user_id, stream_key, stream_key_hash)
		SELECT $1, k, encode(sha256(k::bytea), 'hex') FROM new_key
		ON CONFLICT (user_id) DO UPDATE
		    SET updated_at      = NOW(),
		        stream_key      = COALESCE(overlay_settings.stream_key, EXCLUDED.stream_key),
		        stream_key_hash = encode(sha256(COALESCE(overlay_settings.stream_key, EXCLUDED.stream_key)::bytea), 'hex')
		RETURNING *`, userID,
	).StructScan(&s)
	if err != nil {
		return nil, fmt.Errorf("overlay.CreateDefault: %w", err)
	}
	return &s, nil
}

func (r *repository) UpdateAlertRules(ctx context.Context, userID string, f AlertRulesFields) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE overlay_settings
		SET gif_setting         = $1,
		    tts_variant         = $2,
		    minimum_alert       = $3,
		    minimum_mediashare  = $4,
		    minimum_tts         = $5,
		    updated_at          = NOW()
		WHERE user_id = $6`,
		f.GifSetting, f.TTSVariant, f.MinimumAlert, f.MinimumMediashare, f.MinimumTTS, userID,
	)
	if err != nil {
		return fmt.Errorf("overlay.UpdateAlertRules: %w", err)
	}
	return nil
}

func (r *repository) UpdateTemplate(ctx context.Context, userID string, f TemplateFields) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE overlay_settings
		SET background_color      = $1,
		    highlight_color       = $2,
		    text_color            = $3,
		    template_text         = $4,
		    notification_duration = $5,
		    updated_at            = NOW()
		WHERE user_id = $6`,
		f.BackgroundColor, f.HighlightColor, f.TextColor, f.TemplateText, f.NotificationDuration, userID,
	)
	if err != nil {
		return fmt.Errorf("overlay.UpdateTemplate: %w", err)
	}
	return nil
}

func (r *repository) UpdateFilterKata(ctx context.Context, userID, filterKata string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE overlay_settings SET filter_kata = $1, updated_at = NOW() WHERE user_id = $2`,
		filterKata, userID,
	)
	if err != nil {
		return fmt.Errorf("overlay.UpdateFilterKata: %w", err)
	}
	return nil
}

func (r *repository) UpdateSoundURL(ctx context.Context, userID string, soundURL *string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE overlay_settings SET sound_url = $1, updated_at = NOW() WHERE user_id = $2`,
		soundURL, userID,
	)
	if err != nil {
		return fmt.Errorf("overlay.UpdateSoundURL: %w", err)
	}
	return nil
}

// UpdateStreamKeyHash stores both the hash (for WS auth lookup) and the plain key.
func (r *repository) UpdateStreamKeyHash(ctx context.Context, userID, hash, plainKey string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE overlay_settings SET stream_key_hash = $1, stream_key = $2, updated_at = NOW() WHERE user_id = $3`,
		hash, plainKey, userID,
	)
	if err != nil {
		return fmt.Errorf("overlay.UpdateStreamKeyHash: %w", err)
	}
	return nil
}

func (r *repository) UpdateMediashareRules(ctx context.Context, userID string, f MediashareRulesFields) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE overlay_settings
		SET ms_enabled                = $1,
		    ms_yt_shorts              = $2,
		    ms_tiktok                 = $3,
		    ms_ig_reels               = $4,
		    ms_voice_note             = $5,
		    ms_max_video_duration     = $6,
		    ms_price_per_second_video = $7,
		    ms_max_audio_duration     = $8,
		    ms_price_per_second_audio = $9,
		    ms_min_audio              = $10,
		    minimum_mediashare        = $11,
		    updated_at                = NOW()
		WHERE user_id = $12`,
		f.Enabled, f.YtShorts, f.Tiktok, f.IgReels, f.VoiceNote,
		f.MaxVideoDuration, f.PricePerSecondVideo,
		f.MaxAudioDuration, f.PricePerSecondAudio, f.MinAudio,
		f.MinimumMediashare, userID,
	)
	if err != nil {
		return fmt.Errorf("overlay.UpdateMediashareRules: %w", err)
	}
	return nil
}

func (r *repository) UpdateMediashareTemplate(ctx context.Context, userID string, f MediashareTemplateFields) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE overlay_settings
		SET ms_background_color = $1,
		    ms_highlight_color  = $2,
		    ms_text_color       = $3,
		    ms_template_text    = $4,
		    ms_no_border        = $5,
		    ms_font_weight      = $6,
		    ms_font_family      = $7,
		    ms_show_nsfw        = $8,
		    updated_at          = NOW()
		WHERE user_id = $9`,
		f.BackgroundColor, f.HighlightColor, f.TextColor, f.TemplateText,
		f.NoBorder, f.FontWeight, f.FontFamily, f.ShowNsfw, userID,
	)
	if err != nil {
		return fmt.Errorf("overlay.UpdateMediashareTemplate: %w", err)
	}
	return nil
}

func (r *repository) UpdateQRSettings(ctx context.Context, userID string, f QRSettingsFields) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE overlay_settings
		SET qr_background_color = $1,
		    qr_barcode_color    = $2,
		    qr_label_top        = $3,
		    qr_label_bottom     = $4,
		    qr_font_family      = $5,
		    updated_at          = NOW()
		WHERE user_id = $6`,
		f.BackgroundColor, f.BarcodeColor, f.LabelTop, f.LabelBottom, f.FontFamily, userID,
	)
	if err != nil {
		return fmt.Errorf("overlay.UpdateQRSettings: %w", err)
	}
	return nil
}

func (r *repository) UpdateMilestoneSettings(ctx context.Context, userID string, f MilestoneSettingsFields) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE overlay_settings
		SET ms_title         = $1,
		    ms_target        = $2,
		    ms_start_date    = $3,
		    ms_bg_color      = $4,
		    ms_text_color_ms = $5,
		    ms_no_border_ms  = $6,
		    ms_font_weight_ms= $7,
		    ms_font_title    = $8,
		    ms_font_content  = $9,
		    updated_at       = NOW()
		WHERE user_id = $10`,
		f.Title, f.Target, f.StartDate, f.BgColor, f.TextColor,
		f.NoBorder, f.FontWeight, f.FontTitle, f.FontContent, userID,
	)
	if err != nil {
		return fmt.Errorf("overlay.UpdateMilestoneSettings: %w", err)
	}
	return nil
}

func (r *repository) UpdateSubathonSettings(ctx context.Context, userID string, f SubathonSettingsFields) error {
	rulesJSON, err := json.Marshal(f.TimeRules)
	if err != nil {
		return fmt.Errorf("overlay.UpdateSubathonSettings: marshal rules: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `
		UPDATE overlay_settings
		SET sub_initial_hours   = $1,
		    sub_initial_minutes = $2,
		    sub_initial_seconds = $3,
		    sub_no_border       = $4,
		    sub_bg_color        = $5,
		    sub_auto_play       = $6,
		    sub_text_color      = $7,
		    sub_font_weight     = $8,
		    sub_font_content    = $9,
		    sub_time_rules      = $10,
		    updated_at          = NOW()
		WHERE user_id = $11`,
		f.InitialHours, f.InitialMinutes, f.InitialSeconds,
		f.NoBorder, f.BgColor, f.AutoPlay, f.TextColor,
		f.FontWeight, f.FontContent, rulesJSON, userID,
	)
	if err != nil {
		return fmt.Errorf("overlay.UpdateSubathonSettings: %w", err)
	}
	return nil
}
