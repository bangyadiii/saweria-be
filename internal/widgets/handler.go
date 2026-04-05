package widgets

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"strconv"

	"saweria-be/internal/domain"

	"github.com/gin-gonic/gin"
)

// OverlayRepository is the minimal interface this handler needs
type OverlayRepository interface {
	FindByStreamKeyHash(ctx context.Context, hash string) (*domain.OverlaySettings, error)
}

// UserRepository is the minimal interface this handler needs
type UserRepository interface {
	FindByID(ctx context.Context, id string) (*domain.User, error)
}

type Handler struct {
	repo        Repository
	overlayRepo OverlayRepository
	userRepo    UserRepository
}

func NewHandler(repo Repository, overlayRepo OverlayRepository, userRepo UserRepository) *Handler {
	return &Handler{repo: repo, overlayRepo: overlayRepo, userRepo: userRepo}
}

type InfoResponse struct {
	Username        string                  `json:"username"`
	DisplayName     string                  `json:"displayName"`
	ProfileImage    *string                 `json:"profileImage"`
	TotalRaised     int64                   `json:"totalRaised"`
	OverlaySettings *domain.OverlaySettings `json:"overlaySettings"`
}

// GET /widgets/info?streamKey=<plainKey>
func (h *Handler) Info(c *gin.Context) {
	settings, user, ok := h.resolveStreamKey(c)
	if !ok {
		return
	}

	total, err := h.repo.GetTotalRaised(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"data": InfoResponse{
			Username:        user.Username,
			DisplayName:     user.DisplayName,
			ProfileImage:    user.ProfileImage,
			TotalRaised:     total,
			OverlaySettings: settings,
		},
	})
}

// GET /widgets/leaderboard?streamKey=<plainKey>&limit=10
func (h *Handler) Leaderboard(c *gin.Context) {
	_, user, ok := h.resolveStreamKey(c)
	if !ok {
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	list, err := h.repo.GetLeaderboard(c.Request.Context(), user.ID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok", "data": list})
}

// resolveStreamKey hashes the plain stream key and looks up overlay_settings + user
func (h *Handler) resolveStreamKey(c *gin.Context) (*domain.OverlaySettings, *domain.User, bool) {
	plainKey := c.Query("streamKey")
	if plainKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "streamKey is required"})
		return nil, nil, false
	}

	hash := hashKey(plainKey)
	settings, err := h.overlayRepo.FindByStreamKeyHash(c.Request.Context(), hash)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid stream key"})
		return nil, nil, false
	}

	user, err := h.userRepo.FindByID(c.Request.Context(), settings.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return nil, nil, false
	}

	return settings, user, true
}

func hashKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%x", sum)
}
