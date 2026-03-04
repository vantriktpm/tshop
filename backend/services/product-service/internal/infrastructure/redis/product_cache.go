package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/tshop/backend/services/product-service/internal/domain"
)

const (
	keyFlashList = "product:flash:list"
	keyFlashQty  = "product:flash:qty:"
	keyLock      = "product:lock:"
	defaultLockTTL = 10 * time.Second
)

// Lua: atomic decrement quantity. KEYS[1]=qty key, ARGV[1]=decrement (integer).
// Returns new value after decrement, or -1 if not enough quantity.
var scriptDecrQty = redis.NewScript(`
	local cur = tonumber(redis.call('GET', KEYS[1]) or '0')
	local dec = tonumber(ARGV[1])
	if cur < dec then return -1 end
	local new = cur - dec
	redis.call('SET', KEYS[1], new)
	return new
`)

// Lua: release lock only if token matches. KEYS[1]=lock key, ARGV[1]=token.
var scriptUnlock = redis.NewScript(`
	if redis.call('GET', KEYS[1]) == ARGV[1] then
		return redis.call('DEL', KEYS[1])
	end
	return 0
`)

type ProductCache struct {
	client *redis.Client
}

func NewProductCache(client *redis.Client) *ProductCache {
	return &ProductCache{client: client}
}

// SetAll writes the full product list to Redis and seeds per-product quantity keys for flash sale.
// TTL optional; 0 means no expiry.
func (c *ProductCache) SetAll(ctx context.Context, products []*domain.Product, ttl time.Duration) error {
	data, err := json.Marshal(products)
	if err != nil {
		return err
	}
	pipe := c.client.Pipeline()
	pipe.Set(ctx, keyFlashList, data, ttl)
	for _, p := range products {
		qtyKey := keyFlashQty + p.ProductID
		pipe.Set(ctx, qtyKey, int64(p.Quantity), ttl)
	}
	_, err = pipe.Exec(ctx)
	return err
}

// GetAll returns the full product list from Redis. Nil slice if key missing.
func (c *ProductCache) GetAll(ctx context.Context) ([]*domain.Product, error) {
	b, err := c.client.Get(ctx, keyFlashList).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var list []*domain.Product
	if err := json.Unmarshal(b, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// DecrQuantity atomically decrements flash-sale quantity for productID by delta.
// Returns new quantity after decrement, or ErrInsufficientQuantity if not enough.
func (c *ProductCache) DecrQuantity(ctx context.Context, productID string, delta int64) (newQty int64, err error) {
	key := keyFlashQty + productID
	n, err := scriptDecrQty.Run(ctx, c.client, []string{key}, delta).Int64()
	if err != nil {
		return 0, err
	}
	if n < 0 {
		return 0, ErrInsufficientQuantity
	}
	return n, nil
}

// GetQuantity returns current cached quantity for a product (flash sale key).
func (c *ProductCache) GetQuantity(ctx context.Context, productID string) (int64, error) {
	key := keyFlashQty + productID
	n, err := c.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return n, err
}

var ErrInsufficientQuantity = errors.New("insufficient quantity")

// Lock acquires a distributed lock for the given name with optional ttl.
// Returns a token that must be passed to Unlock. Lock is not reentrant.
func (c *ProductCache) Lock(ctx context.Context, name string, ttl time.Duration) (token string, err error) {
	if ttl <= 0 {
		ttl = defaultLockTTL
	}
	token = uuid.New().String()
	key := keyLock + name
	ok, err := c.client.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return "", err
	}
	if !ok {
		return "", ErrLockNotAcquired
	}
	return token, nil
}

// Unlock releases the lock only if the token matches (Lua script, atomic).
func (c *ProductCache) Unlock(ctx context.Context, name, token string) error {
	key := keyLock + name
	n, err := scriptUnlock.Run(ctx, c.client, []string{key}, token).Int64()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrLockNotHeld
	}
	return nil
}

var (
	ErrLockNotAcquired = errors.New("lock not acquired")
	ErrLockNotHeld     = errors.New("lock not held or wrong token")
)
