package overlay

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"saweria-be/internal/domain"

	"github.com/gin-gonic/gin"
)

// WSHub broadcasts raw JSON to all WS clients subscribed to a stream key hash.
type WSHub interface {
	Broadcast(streamKeyHash string, message []byte)
	SetSubathonState(hash string, totalSeconds int, running bool)
	GetSubathonSeconds(hash string) (totalSeconds int, running bool, ok bool)
}

// UserRepository is the minimal user-lookup interface needed by overlay handler.
type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
}

type Handler struct {
	service  Service
	hub      WSHub
	userRepo UserRepository
}

func NewHandler(service Service, hub WSHub, userRepo UserRepository) *Handler {
	return &Handler{service: service, hub: hub, userRepo: userRepo}
}

// effectiveHash returns the stream key hash to use for WS broadcasting.
// Always derives from the plain key when available, because the stored hash
// may have been computed from different bytes (legacy CreateDefault bug).
// Returns "" when neither is available (stream key not configured yet).
func effectiveHash(storedHash *string, plainKey *string) string {
	if plainKey != nil && *plainKey != "" {
		sum := sha256.Sum256([]byte(*plainKey))
		return fmt.Sprintf("%x", sum)
	}
	if storedHash != nil && *storedHash != "" {
		return *storedHash
	}
	return ""
}

func (h *Handler) GetSettings(c *gin.Context) {
	userID := c.GetString("user_id")
	settings, err := h.service.GetSettings(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok", "data": settings})
}

func (h *Handler) UpdateAlertRules(c *gin.Context) {
	userID := c.GetString("user_id")
	var req AlertRulesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.UpdateAlertRules(c.Request.Context(), userID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (h *Handler) UpdateTemplate(c *gin.Context) {
	userID := c.GetString("user_id")
	var req TemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.UpdateTemplate(c.Request.Context(), userID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (h *Handler) UpdateFilterKata(c *gin.Context) {
	userID := c.GetString("user_id")
	var body struct {
		FilterKata string `json:"filter_kata"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.UpdateFilterKata(c.Request.Context(), userID, body.FilterKata); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (h *Handler) UpdateSound(c *gin.Context) {
	userID := c.GetString("user_id")

	file, header, err := c.Request.FormFile("sound")
	if err != nil {
		// allow removing sound (no file provided = reset to default)
		if err2 := h.service.UpdateSoundURL(c.Request.Context(), userID, nil); err2 != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
		return
	}
	defer file.Close()
	// Save file to uploads/sounds/ directory
	const uploadDir = "uploads/sounds"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create upload directory"})
		return
	}
	destPath := uploadDir + "/" + header.Filename
	if err := c.SaveUploadedFile(header, destPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot save file"})
		return
	}

	soundURL := "/uploads/sounds/" + header.Filename
	if err := h.service.UpdateSoundURL(c.Request.Context(), userID, &soundURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok", "data": gin.H{"sound_url": soundURL}})
}

func (h *Handler) ResetStreamKey(c *gin.Context) {
	userID := c.GetString("user_id")
	plainKey, err := h.service.ResetStreamKey(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"data":    gin.H{"streamKey": plainKey},
	})
}

func (h *Handler) UpdateMediashareRules(c *gin.Context) {
	userID := c.GetString("user_id")
	var req MediashareRulesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.UpdateMediashareRules(c.Request.Context(), userID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (h *Handler) UpdateMediashareTemplate(c *gin.Context) {
	userID := c.GetString("user_id")
	var req MediashareTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.UpdateMediashareTemplate(c.Request.Context(), userID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (h *Handler) UpdateQRSettings(c *gin.Context) {
	userID := c.GetString("user_id")
	var req QRSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.UpdateQRSettings(c.Request.Context(), userID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (h *Handler) UpdateMilestoneSettings(c *gin.Context) {
	userID := c.GetString("user_id")
	var req MilestoneSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.UpdateMilestoneSettings(c.Request.Context(), userID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (h *Handler) UpdateSubathonSettings(c *gin.Context) {
	userID := c.GetString("user_id")
	var req SubathonSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.UpdateSubathonSettings(c.Request.Context(), userID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// TestAlert broadcasts a fake donation_alert to the user's widgets via WebSocket.
func (h *Handler) TestAlert(c *gin.Context) {
	userID := c.GetString("user_id")
	settings, err := h.service.GetSettings(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	keyHash := effectiveHash(settings.StreamKeyHash, settings.StreamKey)
	if keyHash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "stream key not set"})
		return
	}

	tpl := settings.TemplateText
	if tpl == "" {
		tpl = "[nama] baru saja memberikan [nominal]"
	}
	duration := settings.NotificationDuration
	if duration == 0 {
		duration = 8
	}

	bgColor := ""
	if settings.BackgroundColor != nil {
		bgColor = *settings.BackgroundColor
	}
	hlColor := ""
	if settings.HighlightColor != nil {
		hlColor = *settings.HighlightColor
	}
	tcColor := ""
	if settings.TextColor != nil {
		tcColor = *settings.TextColor
	}
	ttsVariant := ""
	if settings.TTSVariant != nil {
		ttsVariant = *settings.TTSVariant
	}
	soundURL := ""
	if settings.SoundURL != nil {
		soundURL = *settings.SoundURL
	}

	msg := fmt.Sprintf(
		`{"type":"donation_alert","donor_name":"Tes User","amount":10000,"message":"Ini adalah tes alert!","template_text":%q,"background_color":%q,"highlight_color":%q,"text_color":%q,"notification_duration":%d,"tts_variant":%q,"sound_url":%q}`,
		tpl, bgColor, hlColor, tcColor, duration, ttsVariant, soundURL,
	)
	h.hub.Broadcast(keyHash, []byte(msg))
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// TestMediashare broadcasts a fake mediashare donation_alert (Rick Roll) to the user's widgets.
func (h *Handler) TestMediashare(c *gin.Context) {
	userID := c.GetString("user_id")
	settings, err := h.service.GetSettings(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	keyHash := effectiveHash(settings.StreamKeyHash, settings.StreamKey)
	if keyHash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "stream key not set"})
		return
	}

	duration := settings.MsMaxVideoDuration
	if duration == 0 {
		duration = 360
	}

	bgColor := ""
	if settings.MsBackgroundColor != nil {
		bgColor = *settings.MsBackgroundColor
	}
	hlColor := ""
	if settings.MsHighlightColor != nil {
		hlColor = *settings.MsHighlightColor
	}
	tcColor := ""
	if settings.MsTextColor != nil {
		tcColor = *settings.MsTextColor
	}
	tpl := settings.MsTemplateText
	if tpl == "" {
		tpl = "baru saja memberi"
	}
	noBorder := settings.MsNoBorder
	fontWeight := settings.MsFontWeight
	if fontWeight == 0 {
		fontWeight = 500
	}
	fontFamily := settings.MsFontFamily
	if fontFamily == "" {
		fontFamily = "Poppins"
	}

	msg := fmt.Sprintf(
		`{"type":"donation_alert","donor_name":"Tes User","amount":10000,"message":"Tes mediashare!","media_url":"https://www.youtube.com/watch?v=Aq5WXmQQooo","notification_duration":%d,"ms_background_color":%q,"ms_highlight_color":%q,"ms_text_color":%q,"ms_template_text":%q,"ms_no_border":%t,"ms_font_weight":%d,"ms_font_family":%q}`,
		duration, bgColor, hlColor, tcColor, tpl, noBorder, fontWeight, fontFamily,
	)
	h.hub.Broadcast(keyHash, []byte(msg))
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// Control broadcasts a control command (pause/play/skip/refresh) to the user's widgets.
func (h *Handler) Control(c *gin.Context) {
	userID := c.GetString("user_id")

	var body struct {
		Action string `json:"action"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validActions := map[string]bool{"pause": true, "play": true, "skip": true, "refresh": true}
	if !validActions[body.Action] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action"})
		return
	}

	settings, err := h.service.GetSettings(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	keyHash := effectiveHash(settings.StreamKeyHash, settings.StreamKey)
	if keyHash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "stream key not set"})
		return
	}

	msg := fmt.Sprintf(`{"type":"control","action":%q}`, body.Action)
	h.hub.Broadcast(keyHash, []byte(msg))
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// TestSubathon broadcasts a fake donation_alert that triggers the first configured subathon time rule.
func (h *Handler) TestSubathon(c *gin.Context) {
	userID := c.GetString("user_id")
	settings, err := h.service.GetSettings(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	keyHash := effectiveHash(settings.StreamKeyHash, settings.StreamKey)
	if keyHash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "stream key not set"})
		return
	}

	// Use the first configured rule's min_amount so the widget adds time; fallback to 10000
	amount := int64(10000)
	if len(settings.SubTimeRules) > 0 {
		var rules []SubathonTimeRule
		if err := json.Unmarshal(settings.SubTimeRules, &rules); err == nil && len(rules) > 0 {
			amount = rules[0].MinAmount
			if amount == 0 {
				amount = 10000
			}
		}
	}

	msg := fmt.Sprintf(
		`{"type":"donation_alert","donor_name":"Tes User","amount":%d,"message":"Tes Subathon!"}`,
		amount,
	)
	h.hub.Broadcast(keyHash, []byte(msg))
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// SubathonControl mutates the hub's in-memory timer state and broadcasts
// the authoritative subathon_state to all connected widgets.
// Because the hub tracks totalSeconds + lastUpdated, remaining time is correctly
// computed even after the timer has been paused for hours.
func (h *Handler) SubathonControl(c *gin.Context) {
	userID := c.GetString("user_id")

	var body struct {
		Action  string `json:"action"`
		Hours   int    `json:"hours"`
		Minutes int    `json:"minutes"`
		Seconds int    `json:"seconds"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validActions := map[string]bool{"start": true, "pause": true, "add_time": true}
	if !validActions[body.Action] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action"})
		return
	}

	settings, err := h.service.GetSettings(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	keyHash := effectiveHash(settings.StreamKeyHash, settings.StreamKey)
	if keyHash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "stream key not set"})
		return
	}

	// Read current state; if none exists yet, initialize from saved settings.
	curSecs, isRunning, ok := h.hub.GetSubathonSeconds(keyHash)
	if !ok {
		curSecs = settings.SubInitialHours*3600 + settings.SubInitialMinutes*60 + settings.SubInitialSeconds
		isRunning = false
	}

	switch body.Action {
	case "start":
		isRunning = true
	case "pause":
		// curSecs already holds the computed remaining time (hub accounts for elapsed).
		isRunning = false
	case "add_time":
		curSecs += body.Hours*3600 + body.Minutes*60 + body.Seconds
	}

	// Persist new state so future reconnects get the correct value.
	h.hub.SetSubathonState(keyHash, curSecs, isRunning)

	msg := fmt.Sprintf(`{"type":"subathon_state","total_seconds":%d,"running":%t}`, curSecs, isRunning)
	h.hub.Broadcast(keyHash, []byte(msg))
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (h *Handler) UpdateLeaderboardSettings(c *gin.Context) {
	userID := c.GetString("user_id")
	var req LeaderboardSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.UpdateLeaderboardSettings(c.Request.Context(), userID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (h *Handler) UpdateMabarSettings(c *gin.Context) {
	userID := c.GetString("user_id")
	var req MabarSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.UpdateMabarSettings(c.Request.Context(), userID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// GetPublicMabar returns whether mabar is enabled and its trigger info for the donation page.
// This endpoint is intentionally unauthenticated.
func (h *Handler) GetPublicMabar(c *gin.Context) {
	username := c.Param("username")
	settings, err := h.service.GetSettingsByUsername(c.Request.Context(), username)
	if err != nil {
		// Return disabled state gracefully if user not found
		c.JSON(http.StatusOK, gin.H{"message": "ok", "data": gin.H{"mabar_enabled": false}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok", "data": gin.H{
		"mabar_enabled":        settings.MabarEnabled,
		"mabar_keyword":        settings.MabarKeyword,
		"mabar_minimum_amount": settings.MabarMinimumAmount,
	}})
}

// GetPublicMediashare returns the mediashare settings a donor needs to decide
// whether to show the media input section on the public donation page.
// This endpoint is intentionally unauthenticated.
func (h *Handler) GetPublicMediashare(c *gin.Context) {
	username := c.Param("username")
	u, err := h.userRepo.FindByUsername(c.Request.Context(), username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	settings, err := h.service.GetSettings(c.Request.Context(), u.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok", "data": gin.H{
		"ms_enabled":            settings.MsEnabled,
		"minimum_mediashare":    settings.MinimumMediashare,
		"ms_yt_shorts":          settings.MsYtShorts,
		"ms_tiktok":             settings.MsTiktok,
		"ms_ig_reels":           settings.MsIgReels,
		"ms_voice_note":         settings.MsVoiceNote,
		"ms_max_video_duration": settings.MsMaxVideoDuration,
		"ms_max_audio_duration": settings.MsMaxAudioDuration,
	}})
}
