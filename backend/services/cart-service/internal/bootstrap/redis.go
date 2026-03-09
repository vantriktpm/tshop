package bootstrap

import (
	"context"
	"os"

	"github.com/redis/go-redis/v9"
)

// NewRedis creates a Redis client using address from config (.env).
// REDIS_ADDR is read after loadEnv() in app.go.
func NewRedis() *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	client := redis.NewClient(&redis.Options{Addr: addr})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil
	}
	return client
}
