package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/tshop/backend/services/cart-service/internal/bootstrap"
)

func main() {
	c := bootstrap.New()

	r := gin.Default()
	r.Use(func(ctx *gin.Context) {
		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(204)
			return
		}
		ctx.Next()
	})

	h := c.CartHandler()
	r.GET("/api/cart", h.Get)
	r.GET("/cart/health", func(ctx *gin.Context) { ctx.JSON(200, gin.H{"status": "ok"}) })

	port := os.Getenv("PORT")
	if port == "" {
		port = "5005"
	}
	log.Println("cart-service :" + port)
	_ = r.Run(":" + port)
}
