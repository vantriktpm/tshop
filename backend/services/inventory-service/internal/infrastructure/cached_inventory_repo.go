package infrastructure

import (
	"context"

	"github.com/tshop/backend/services/inventory-service/internal/domain"
	postgresinfra "github.com/tshop/backend/services/inventory-service/internal/infrastructure/postgres"
	redisinfra "github.com/tshop/backend/services/inventory-service/internal/infrastructure/redis"
)

// CachedInventoryRepository dùng Postgres làm source of truth, Redis làm cache/backup (RDB+AOF).
// GetStock: đọc Redis trước, miss thì đọc DB rồi ghi Redis.
// Reserve: cập nhật DB rồi cập nhật Redis (số dư mới) để khôi phục đúng sau khi restart.
type CachedInventoryRepository struct {
	pg    *postgresinfra.InventoryRepository
	cache *redisinfra.StockCache
}

func NewCachedInventoryRepository(pg *postgresinfra.InventoryRepository, cache *redisinfra.StockCache) *CachedInventoryRepository {
	return &CachedInventoryRepository{pg: pg, cache: cache}
}

func (r *CachedInventoryRepository) GetStock(ctx context.Context, productID string) (*domain.Stock, error) {
	s, err := r.cache.Get(ctx, productID)
	if err == nil && s != nil {
		return s, nil
	}
	s, err = r.pg.GetStock(ctx, productID)
	if err != nil {
		return nil, err
	}
	_ = r.cache.Set(ctx, s)
	return s, nil
}

func (r *CachedInventoryRepository) Reserve(ctx context.Context, productID string, qty int64) error {
	cur, err := r.pg.GetStock(ctx, productID)
	if err != nil {
		return err
	}
	if cur.Quantity < qty {
		return domain.ErrInsufficientStock
	}
	if err := r.pg.Reserve(ctx, productID, qty); err != nil {
		return err
	}
	newQty := cur.Quantity - qty
	return r.cache.SetQuantity(ctx, productID, newQty)
}
