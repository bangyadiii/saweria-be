package domain

import "time"

type User struct {
	ID           string    `db:"id"            json:"id"`
	Email        string    `db:"email"         json:"email"`
	Username     string    `db:"username"      json:"username"`
	PasswordHash *string   `db:"password_hash" json:"-"`
	GoogleID     *string   `db:"google_id"     json:"-"`
	ProfileImage *string   `db:"profile_image" json:"profileImage"`
	DisplayName  string    `db:"display_name"  json:"displayName"`
	Bio          string    `db:"bio"           json:"bio"`
	Balance      int64     `db:"balance"       json:"balance"`
	CreatedAt    time.Time `db:"created_at"    json:"createdAt"`
	UpdatedAt    time.Time `db:"updated_at"    json:"updatedAt"`
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
	MilestoneTitle       string    `db:"ms_title"          json:"ms_title"`
	MilestoneTarget      int64     `db:"ms_target"         json:"ms_target"`
	MilestoneStartDate   *string   `db:"ms_start_date"     json:"ms_start_date"`
	MilestoneBgColor     string    `db:"ms_bg_color"       json:"ms_bg_color"`
	MilestoneTextColor   string    `db:"ms_text_color_ms"  json:"ms_text_color_ms"`
	MilestoneNoBorder    bool      `db:"ms_no_border_ms"   json:"ms_no_border_ms"`
	MilestoneFontWeight  int       `db:"ms_font_weight_ms" json:"ms_font_weight_ms"`
	MilestoneFontTitle   string    `db:"ms_font_title"     json:"ms_font_title"`
	MilestoneFontContent string    `db:"ms_font_content"   json:"ms_font_content"`
	CreatedAt            time.Time `db:"created_at"        json:"created_at"`
	UpdatedAt            time.Time `db:"updated_at"        json:"updated_at"`
}

type Donation struct {
	ID              string    `db:"id"                   json:"id"`
	StreamerID      string    `db:"streamer_id"          json:"streamerId"`
	DonorName       string    `db:"donor_name"           json:"donorName"`
	Amount          int64     `db:"amount"               json:"amount"`
	NetAmount       int64     `db:"net_amount"           json:"netAmount"`
	PlatformFee     int64     `db:"platform_fee"         json:"platformFee"`
	Message         string    `db:"message"              json:"message"`
	MediaURL        *string   `db:"media_url"            json:"mediaUrl"`
	MediaShown      bool      `db:"media_shown"          json:"mediaShown"`
	PaymentMethod   string    `db:"payment_method"       json:"paymentMethod"`
	PaymentStatus   string    `db:"payment_status"       json:"paymentStatus"`
	MidtransOrderID string    `db:"midtrans_order_id" json:"midtransOrderId"`
	PaymentToken    string    `db:"payment_token"     json:"paymentToken"`
	CreatedAt       time.Time `db:"created_at"           json:"createdAt"`
	UpdatedAt       time.Time `db:"updated_at"           json:"updatedAt"`
}

type Cashout struct {
	ID            string    `db:"id"             json:"id"`
	UserID        string    `db:"user_id"        json:"userId"`
	Amount        int64     `db:"amount"         json:"amount"`
	Fee           int64     `db:"fee"            json:"fee"`
	NetAmount     int64     `db:"net_amount"     json:"netAmount"`
	BankName      string    `db:"bank_name"      json:"bankName"`
	AccountNumber string    `db:"account_number" json:"accountNumber"`
	AccountName   string    `db:"account_name"   json:"accountName"`
	Status        string    `db:"status"         json:"status"`
	CreatedAt     time.Time `db:"created_at"     json:"createdAt"`
	UpdatedAt     time.Time `db:"updated_at"     json:"updatedAt"`
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
