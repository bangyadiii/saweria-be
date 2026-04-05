package wallet

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetBalance(c *gin.Context) {
	userID := c.GetString("user_id")
	balance, err := h.service.GetBalance(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok", "data": gin.H{"balance": balance}})
}

func (h *Handler) RequestCashout(c *gin.Context) {
	userID := c.GetString("user_id")
	var req CashoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cashout, err := h.service.RequestCashout(c.Request.Context(), userID, req)
	if errors.Is(err, ErrInsufficientBalance) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient balance"})
		return
	}
	if errors.Is(err, ErrMinimumCashout) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "ok", "data": cashout})
}

func (h *Handler) GetCashoutHistory(c *gin.Context) {
	userID := c.GetString("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	list, err := h.service.GetCashoutHistory(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok", "data": list})
}
