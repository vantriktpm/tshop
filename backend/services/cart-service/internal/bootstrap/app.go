package bootstrap

import (
	"log"

	"github.com/tshop/backend/services/cart-service/internal/container"
)

// New loads .env, initializes all connections (postgres, redis, kafka, minio as needed),
// and returns a container with repositories, services, and handlers.
func New() *container.Container {
	if err := loadEnv(); err != nil {
		log.Printf("bootstrap: load .env: %v (using OS env)", err)
	}

	redisClient := NewRedis()
	if redisClient == nil {
		log.Fatal("bootstrap: redis is required")
	}

	return container.New(redisClient)
}
