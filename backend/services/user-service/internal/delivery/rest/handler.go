package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tshop/backend/pkg/events"
)

type UserHandler struct {
	eventPub events.Publisher
}

func NewUserHandler(pub events.Publisher) *UserHandler {
	if pub == nil {
		pub = events.NoopPublisher{}
	}
	return &UserHandler{eventPub: pub}
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user_id": userID})
}
