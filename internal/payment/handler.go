package payment

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service   Service
	serverKey string
}

func NewHandler(service Service, serverKey string) *Handler {
	return &Handler{service: service, serverKey: serverKey}
}

func (h *Handler) Webhook(c *gin.Context) {
	var payload WebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ProcessWebhook(c.Request.Context(), payload, h.serverKey); err != nil {
		// distinguish signature errors
		if err.Error() == "invalid signature" {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
