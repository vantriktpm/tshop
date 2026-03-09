package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/tshop/backend/services/product-service/internal/bootstrap"
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

	h := c.ProductHandler()
	r.GET("/api/products", h.List)
	r.GET("/products/health", func(ctx *gin.Context) { ctx.JSON(200, gin.H{"status": "ok"}) })

	port := os.Getenv("PORT")
	if port == "" {
		port = "5003"
	}
	log.Println("product-service :" + port)
	_ = r.Run(":" + port)
}
