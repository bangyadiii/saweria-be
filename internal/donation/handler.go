package donation

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service    Service
	feePercent float64
}

func NewHandler(service Service, feePercent float64) *Handler {
	return &Handler{service: service, feePercent: feePercent}
}

func (h *Handler) Submit(c *gin.Context) {
	username := c.Param("username")
	var req SubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.Submit(c.Request.Context(), username, req, h.feePercent)
	if errors.Is(err, ErrStreamerNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "streamer not found"})
		return
	}
	if errors.Is(err, ErrInvalidMediaURL) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "ok", "data": resp})
}

func (h *Handler) GetHistory(c *gin.Context) {
	streamerID := c.GetString("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	list, err := h.service.GetHistory(c.Request.Context(), streamerID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok", "data": list})
}

func (h *Handler) GetDetail(c *gin.Context) {
	streamerID := c.GetString("user_id")
	donationID := c.Param("id")

	d, err := h.service.GetDetail(c.Request.Context(), streamerID, donationID)
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "donation not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok", "data": d})
}
