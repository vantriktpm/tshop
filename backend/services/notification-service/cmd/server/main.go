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
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	r.POST("/api/notifications/send", h.Send)
	// Health endpoint specific to notification-service
	r.GET("/notifications/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	log.Println("notification-service :8088")
	_ = r.Run(":8088")
}
