package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tshop/backend/services/inventory-service/internal/domain"
)

const keyPrefix = "inventory:stock:"
const cacheTTL = 24 * time.Hour

// StockCache lưu/đọc tồn kho trong Redis (RDB+AOF persist, phục hồi khi restart).
type StockCache struct {
	rdb *redis.Client
}

func NewStockCache(rdb *redis.Client) *StockCache {
	return &StockCache{rdb: rdb}
}

func (c *StockCache) Get(ctx context.Context, productID string) (*domain.Stock, error) {
	key := keyPrefix + productID
	val, err := c.rdb.Get(ctx, key).Int64()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &domain.Stock{ProductID: productID, Quantity: val}, nil
}

func (c *StockCache) Set(ctx context.Context, s *domain.Stock) error {
	key := keyPrefix + s.ProductID
	return c.rdb.Set(ctx, key, s.Quantity, cacheTTL).Err()
}

func (c *StockCache) SetQuantity(ctx context.Context, productID string, qty int64) error {
	key := keyPrefix + productID
	return c.rdb.Set(ctx, key, qty, cacheTTL).Err()
}

func (c *StockCache) DecrBy(ctx context.Context, productID string, delta int64) error {
	key := keyPrefix + productID
	return c.rdb.DecrBy(ctx, key, delta).Err()
}

// GetInt64 returns quantity and redis.Nil if key missing.
func (c *StockCache) GetInt64(ctx context.Context, productID string) (int64, error) {
	key := keyPrefix + productID
	return c.rdb.Get(ctx, key).Int64()
}
