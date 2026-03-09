package container

import (
	"github.com/redis/go-redis/v9"
	"github.com/tshop/backend/services/cart-service/internal/domain"
	"github.com/tshop/backend/services/cart-service/internal/handler"
	"github.com/tshop/backend/services/cart-service/internal/repository"
	"github.com/tshop/backend/services/cart-service/internal/service"
)

type Container struct {
	redis *redis.Client

	cartRepo    domain.CartRepository
	cartService *service.CartService
	cartHandler *handler.CartHandler
}

func New(redisClient *redis.Client) *Container {
	c := &Container{redis: redisClient}
	c.cartRepo = repository.NewCartRepository(redisClient)
	c.cartService = service.NewCartService(c.cartRepo)
	c.cartHandler = handler.NewCartHandler(c.cartService)
	return c
}

func (c *Container) CartHandler() *handler.CartHandler { return c.cartHandler }
func (c *Container) Redis() *redis.Client               { return c.redis }
