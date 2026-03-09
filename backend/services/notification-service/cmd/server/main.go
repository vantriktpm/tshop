package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/tshop/backend/services/notification-service/internal/delivery/rest"
)

func main() {
	h := rest.NewNotificationHandler()
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	r.POST("/api/notifications/send", h.Send)
	// Health endpoint specific to notification-service
	r.GET("/notifications/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	log.Println("notification-service :5009")
	_ = r.Run(":5009")
}
