package repository

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
	"github.com/tshop/backend/services/cart-service/internal/domain"
)

const cartPrefix = "cart:"

type CartRepository struct {
	client *redis.Client
}

func NewCartRepository(client *redis.Client) *CartRepository {
	return &CartRepository{client: client}
}

func (r *CartRepository) Get(ctx context.Context, userID string) (*domain.Cart, error) {
	b, err := r.client.Get(ctx, cartPrefix+userID).Bytes()
	if err == redis.Nil {
		return &domain.Cart{UserID: userID, Items: nil}, nil
	}
	if err != nil {
		return nil, err
	}
	var cart domain.Cart
	if err := json.Unmarshal(b, &cart); err != nil {
		return nil, err
	}
	return &cart, nil
}

func (r *CartRepository) Set(ctx context.Context, cart *domain.Cart) error {
	b, _ := json.Marshal(cart)
	return r.client.Set(ctx, cartPrefix+cart.UserID, b, 0).Err()
}
