package rest

import (
	"github.com/gin-gonic/gin"
)

type NotificationHandler struct{}

func NewNotificationHandler() *NotificationHandler { return &NotificationHandler{} }

func (h *NotificationHandler) Send(c *gin.Context) {
	c.JSON(200, gin.H{"sent": true})
}
