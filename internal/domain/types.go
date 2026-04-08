package domain

import (
	"encoding/json"
	"time"
)

type User struct {
	ID           string  `db:"id"            json:"id"`
	Email        string  `db:"email"         json:"email"`
	Username     string  `db:"username"      json:"username"`
	PasswordHash *string `db:"password_hash" json:"-"`
	GoogleID     *string `db:"google_id"     json:"-"`
	ProfileImage *string `db:"profile_image" json:"profileImage"`
	DisplayName  string  `db:"display_name"  json:"displayName"`
	Bio          string  `db:"bio"           json:"bio"`
	Balance      int64   `db:"balance"       json:"balance"`
	// Webhook settings
	WebhookEnabled bool      `db:"webhook_enabled" json:"webhookEnabled"`
	WebhookURL     *string   `db:"webhook_url"     json:"webhookUrl"`
	WebhookToken   *string   `db:"webhook_token"   json:"webhookToken"`
	CreatedAt      time.Time `db:"created_at"    json:"createdAt"`
	UpdatedAt      time.Time `db:"updated_at"    json:"updatedAt"`
}

type OverlaySettings struct {
	ID                   string  `db:"id"                    json:"id"`
	UserID               string  `db:"user_id"               json:"user_id"`
	GifSetting           bool    `db:"gif_setting"           json:"gif_setting"`
	TTSVariant           *string `db:"tts_variant"           json:"tts_variant"`
	MinimumAlert         int64   `db:"minimum_alert"         json:"minimum_alert"`
	MinimumMediashare    int64   `db:"minimum_mediashare"    json:"minimum_mediashare"`
	MinimumTTS           int64   `db:"minimum_tts"           json:"minimum_tts"`
	BackgroundColor      *string `db:"background_color"      json:"background_color"`
	HighlightColor       *string `db:"highlight_color"       json:"highlight_color"`
	TextColor            *string `db:"text_color"            json:"text_color"`
	TemplateText         string  `db:"template_text"         json:"template_text"`
	NotificationDuration int     `db:"notification_duration" json:"notification_duration"`
	FilterKata           string  `db:"filter_kata"           json:"filter_kata"`
	SoundURL             *string `db:"sound_url"             json:"sound_url"`
	StreamKeyHash        *string `db:"stream_key_hash"       json:"-"`
	StreamKey            *string `db:"stream_key"            json:"stream_key"`
	// Mediashare rules
	MsEnabled             bool  `db:"ms_enabled"                json:"ms_enabled"`
	MsYtShorts            bool  `db:"ms_yt_shorts"              json:"ms_yt_shorts"`
	MsTiktok              bool  `db:"ms_tiktok"                 json:"ms_tiktok"`
	MsIgReels             bool  `db:"ms_ig_reels"               json:"ms_ig_reels"`
	MsVoiceNote           bool  `db:"ms_voice_note"             json:"ms_voice_note"`
	MsMaxVideoDuration    int   `db:"ms_max_video_duration"     json:"ms_max_video_duration"`
	MsPricePerSecondVideo int64 `db:"ms_price_per_second_video" json:"ms_price_per_second_video"`
	MsMaxAudioDuration    int   `db:"ms_max_audio_duration"     json:"ms_max_audio_duration"`
	MsPricePerSecondAudio int64 `db:"ms_price_per_second_audio" json:"ms_price_per_second_audio"`
	MsMinAudio            int64 `db:"ms_min_audio"              json:"ms_min_audio"`
	// Mediashare template
	MsBackgroundColor *string `db:"ms_background_color" json:"ms_background_color"`
	MsHighlightColor  *string `db:"ms_highlight_color"  json:"ms_highlight_color"`
	MsTextColor       *string `db:"ms_text_color"       json:"ms_text_color"`
	MsTemplateText    string  `db:"ms_template_text"    json:"ms_template_text"`
	MsNoBorder        bool    `db:"ms_no_border"        json:"ms_no_border"`
	MsFontWeight      int     `db:"ms_font_weight"      json:"ms_font_weight"`
	MsFontFamily      string  `db:"ms_font_family"      json:"ms_font_family"`
	MsShowNsfw        bool    `db:"ms_show_nsfw"        json:"ms_show_nsfw"`
	// QR code widget settings
	QRBackgroundColor string `db:"qr_background_color" json:"qr_background_color"`
	QRBarcodeColor    string `db:"qr_barcode_color"    json:"qr_barcode_color"`
	QRLabelTop        string `db:"qr_label_top"        json:"qr_label_top"`
	QRLabelBottom     string `db:"qr_label_bottom"     json:"qr_label_bottom"`
	QRFontFamily      string `db:"qr_font_family"      json:"qr_font_family"`
	// Milestone widget settings
	MilestoneTitle       string  `db:"ms_title"          json:"ms_title"`
	MilestoneTarget      int64   `db:"ms_target"         json:"ms_target"`
	MilestoneStartDate   *string `db:"ms_start_date"     json:"ms_start_date"`
	MilestoneBgColor     string  `db:"ms_bg_color"       json:"ms_bg_color"`
	MilestoneTextColor   string  `db:"ms_text_color_ms"  json:"ms_text_color_ms"`
	MilestoneNoBorder    bool    `db:"ms_no_border_ms"   json:"ms_no_border_ms"`
	MilestoneFontWeight  int     `db:"ms_font_weight_ms" json:"ms_font_weight_ms"`
	MilestoneFontTitle   string  `db:"ms_font_title"     json:"ms_font_title"`
	MilestoneFontContent string  `db:"ms_font_content"   json:"ms_font_content"`
	// Subathon widget settings
	SubInitialHours   int             `db:"sub_initial_hours"   json:"sub_initial_hours"`
	SubInitialMinutes int             `db:"sub_initial_minutes" json:"sub_initial_minutes"`
	SubInitialSeconds int             `db:"sub_initial_seconds" json:"sub_initial_seconds"`
	SubNoBorder       bool            `db:"sub_no_border"       json:"sub_no_border"`
	SubBgColor        string          `db:"sub_bg_color"        json:"sub_bg_color"`
	SubAutoPlay       bool            `db:"sub_auto_play"       json:"sub_auto_play"`
	SubTextColor      string          `db:"sub_text_color"      json:"sub_text_color"`
	SubFontWeight     int             `db:"sub_font_weight"     json:"sub_font_weight"`
	SubFontContent    string          `db:"sub_font_content"    json:"sub_font_content"`
	SubTimeRules      json.RawMessage `db:"sub_time_rules"      json:"sub_time_rules"`
	// Leaderboard widget settings
	LbTitle       string `db:"lb_title"        json:"lb_title"`
	LbBgColor     string `db:"lb_bg_color"     json:"lb_bg_color"`
	LbTextColor   string `db:"lb_text_color"   json:"lb_text_color"`
	LbFontWeight  int    `db:"lb_font_weight"  json:"lb_font_weight"`
	LbNoBorder    bool   `db:"lb_no_border"    json:"lb_no_border"`
	LbHideAmount  bool   `db:"lb_hide_amount"  json:"lb_hide_amount"`
	LbFontTitle   string `db:"lb_font_title"   json:"lb_font_title"`
	LbFontContent string `db:"lb_font_content" json:"lb_font_content"`
	LbTimeRange   string `db:"lb_time_range"   json:"lb_time_range"`
	LbLimit       int    `db:"lb_limit"        json:"lb_limit"`
	// Mabar settings
	MabarEnabled         bool      `db:"mabar_enabled"          json:"mabar_enabled"`
	MabarKeyword         string    `db:"mabar_keyword"          json:"mabar_keyword"`
	MabarMinimumAmount   int64     `db:"mabar_minimum_amount"   json:"mabar_minimum_amount"`
	MabarGoldThreshold   int64     `db:"mabar_gold_threshold"   json:"mabar_gold_threshold"`
	MabarSilverThreshold int64     `db:"mabar_silver_threshold" json:"mabar_silver_threshold"`
	CreatedAt            time.Time `db:"created_at"       json:"created_at"`
	UpdatedAt            time.Time `db:"updated_at"       json:"updated_at"`
}

type MabarQueueItem struct {
	ID             string    `db:"id"              json:"id"`
	StreamerID     string    `db:"streamer_id"     json:"streamer_id"`
	DonationID     string    `db:"donation_id"     json:"donation_id"`
	DonorName      string    `db:"donor_name"      json:"donor_name"`
	IngameUsername string    `db:"ingame_username" json:"ingame_username"`
	Amount         int64     `db:"amount"          json:"amount"`
	PriorityTier   string    `db:"priority_tier"   json:"priority_tier"`
	PriorityOrder  int       `db:"priority_order"  json:"priority_order"`
	Status         string    `db:"status"          json:"status"`
	CreatedAt      time.Time `db:"created_at"      json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"      json:"updated_at"`
}

type Donation struct {
	ID              string    `db:"id"                   json:"id"`
	StreamerID      string    `db:"streamer_id"          json:"streamer_id"`
	DonorName       string    `db:"donor_name"           json:"donor_name"`
	Amount          int64     `db:"amount"               json:"amount"`
	NetAmount       int64     `db:"net_amount"           json:"net_amount"`
	PlatformFee     int64     `db:"platform_fee"         json:"platform_fee"`
	Message         string    `db:"message"              json:"message"`
	MediaURL        *string   `db:"media_url"            json:"media_url"`
	MediaShown      bool      `db:"media_shown"          json:"media_shown"`
	PaymentMethod   string    `db:"payment_method"       json:"payment_method"`
	PaymentStatus   string    `db:"payment_status"       json:"payment_status"`
	MidtransOrderID string    `db:"midtrans_order_id" json:"midtrans_order_id"`
	PaymentToken    string    `db:"payment_token"     json:"payment_token"`
	CreatedAt       time.Time `db:"created_at"           json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"           json:"updated_at"`
}

type Cashout struct {
	ID            string    `db:"id"             json:"id"`
	UserID        string    `db:"user_id"        json:"user_id"`
	Amount        int64     `db:"amount"         json:"amount"`
	Fee           int64     `db:"fee"            json:"fee"`
	NetAmount     int64     `db:"net_amount"     json:"net_amount"`
	BankName      string    `db:"bank_name"      json:"bank_name"`
	AccountNumber string    `db:"account_number" json:"account_number"`
	AccountName   string    `db:"account_name"   json:"account_name"`
	Status        string    `db:"status"         json:"status"`
	CreatedAt     time.Time `db:"created_at"     json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"     json:"updated_at"`
}

// Payment status constants
const (
	PaymentStatusPending = "pending"
	PaymentStatusSuccess = "success"
	PaymentStatusFailed  = "failed"
	PaymentStatusExpired = "expired"
)

// Cashout status constants
const (
	CashoutStatusPending   = "pending"
	CashoutStatusProcessed = "processed"
	CashoutStatusFailed    = "failed"
)

// ChargeResult holds the normalised Core API response regardless of payment type.
type ChargeResult struct {
	TransactionID string
	// bank_transfer
	Bank     string
	VANumber string
	// mandiri echannel
	BillerCode string
	BillKey    string
	// e-wallet (gopay, shopeepay)
	QRCodeURL   string
	DeepLinkURL string
}
