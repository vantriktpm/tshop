package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tshop/backend/pkg/events"
	"github.com/tshop/backend/services/user-service/internal/domain"
)

type UserHandler struct {
	eventPub     events.Publisher
	sessionStore domain.SessionStore
}

func NewUserHandler(pub events.Publisher, sessionStore domain.SessionStore) *UserHandler {
	if pub == nil {
		pub = events.NoopPublisher{}
	}
	return &UserHandler{eventPub: pub, sessionStore: sessionStore}
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user_id": userID})
}
