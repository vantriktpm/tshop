package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/tshop/backend/services/cart-service/internal/delivery/rest"
	redisinfra "github.com/tshop/backend/services/cart-service/internal/infrastructure/redis"
	"github.com/tshop/backend/services/cart-service/internal/usecase"
)

func main() {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Printf("redis: %v", err)
	}
	repo := redisinfra.NewCartRepository(rdb)
	h := rest.NewCartHandler(usecase.NewGetCart(repo))
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
	r.GET("/api/cart", h.Get)
	// Health endpoint specific to cart-service
	r.GET("/cart/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	log.Println("cart-service :8084")
	_ = r.Run(":8084")
}
