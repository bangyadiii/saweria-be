package user

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetMe(c *gin.Context) {
	userID := c.GetString("user_id")
	u, err := h.service.GetMe(c.Request.Context(), userID)
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok", "data": u})
}

func (h *Handler) GetPublicProfile(c *gin.Context) {
	username := c.Param("username")
	u, err := h.service.GetPublicProfile(c.Request.Context(), username)
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok", "data": u})
}

func (h *Handler) UpdateMe(c *gin.Context) {
	userID := c.GetString("user_id")

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.service.UpdateProfile(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok", "data": u})
}

func (h *Handler) UpdateWebhookSettings(c *gin.Context) {
	userID := c.GetString("user_id")

	var req UpdateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.service.UpdateWebhookSettings(c.Request.Context(), userID, req)
	if errors.Is(err, ErrWebhookURLInvalid) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok", "data": u})
}

func (h *Handler) ResetWebhookToken(c *gin.Context) {
	userID := c.GetString("user_id")

	u, err := h.service.ResetWebhookToken(c.Request.Context(), userID)
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok", "data": u})
}

func (h *Handler) TestWebhook(c *gin.Context) {
	userID := c.GetString("user_id")

	err := h.service.TestWebhook(c.Request.Context(), userID)
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if errors.Is(err, ErrWebhookURLRequired) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "webhook URL belum dikonfigurasi"})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "gagal mengirim notifikasi: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "notifikasi tes berhasil dikirim"})
}
