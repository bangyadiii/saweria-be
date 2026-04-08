package user

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"saweria-be/internal/domain"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
	baseURL string
}

func NewHandler(service Service, baseURL string) *Handler {
	return &Handler{service: service, baseURL: baseURL}
}

// resolveUser returns a shallow copy of the user with profileImage rewritten to a full URL.
func (h *Handler) resolveUser(u *domain.User) *domain.User {
	if u == nil || u.ProfileImage == nil || *u.ProfileImage == "" {
		return u
	}
	img := *u.ProfileImage
	if !strings.HasPrefix(img, "http") {
		full := h.baseURL + img
		copy := *u
		copy.ProfileImage = &full
		return &copy
	}
	return u
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
	c.JSON(http.StatusOK, gin.H{"message": "ok", "data": h.resolveUser(u)})
}

type publicProfileResponse struct {
	Username     string  `json:"username"`
	DisplayName  string  `json:"display_name"`
	ProfileImage *string `json:"profile_image"`
	Bio          string  `json:"bio"`
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
	u = h.resolveUser(u)
	c.JSON(http.StatusOK, gin.H{"message": "ok", "data": publicProfileResponse{
		Username:     u.Username,
		DisplayName:  u.DisplayName,
		ProfileImage: u.ProfileImage,
		Bio:          u.Bio,
	}})
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
	c.JSON(http.StatusOK, gin.H{"message": "ok", "data": h.resolveUser(u)})
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
	c.JSON(http.StatusOK, gin.H{"message": "ok", "data": h.resolveUser(u)})
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
	c.JSON(http.StatusOK, gin.H{"message": "ok", "data": h.resolveUser(u)})
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

func (h *Handler) UploadProfileImage(c *gin.Context) {
	userID := c.GetString("user_id")

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image file is required"})
		return
	}
	defer file.Close()

	const uploadDir = "uploads/images"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create upload directory"})
		return
	}
	destPath := uploadDir + "/" + header.Filename
	if err := c.SaveUploadedFile(header, destPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot save file"})
		return
	}

	imagePath := "/uploads/images/" + header.Filename
	u, err := h.service.UploadProfileImage(c.Request.Context(), userID, imagePath)
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	resolved := h.resolveUser(u)
	c.JSON(http.StatusOK, gin.H{"message": "ok", "data": gin.H{"profileImage": resolved.ProfileImage}})
}
