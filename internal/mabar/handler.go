package mabar

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

var ErrNotFound = errors.New("mabar queue item not found")

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// GET /mabar/queue
func (h *Handler) GetQueue(c *gin.Context) {
	streamerID := c.GetString("user_id")
	items, err := h.service.GetQueue(c.Request.Context(), streamerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok", "data": items})
}

// PUT /mabar/queue/:id/done
func (h *Handler) MarkDone(c *gin.Context) {
	streamerID := c.GetString("user_id")
	id := c.Param("id")
	err := h.service.MarkDone(c.Request.Context(), id, streamerID)
	if errors.Is(err, ErrNotFound) || errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// PUT /mabar/queue/reorder
func (h *Handler) Reorder(c *gin.Context) {
	streamerID := c.GetString("user_id")
	var items []ReorderItem
	if err := c.ShouldBindJSON(&items); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.service.Reorder(c.Request.Context(), streamerID, items)
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "one or more items not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

// DELETE /mabar/queue
func (h *Handler) ClearAll(c *gin.Context) {
	streamerID := c.GetString("user_id")
	if err := h.service.ClearAll(c.Request.Context(), streamerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
