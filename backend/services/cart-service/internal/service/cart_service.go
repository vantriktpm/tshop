package service

import (
	"context"

	"github.com/tshop/backend/services/cart-service/internal/domain"
)

type CartService struct {
	repo domain.CartRepository
}

func NewCartService(repo domain.CartRepository) *CartService {
	return &CartService{repo: repo}
}

func (s *CartService) GetCart(ctx context.Context, userID string) (*domain.Cart, error) {
	return s.repo.Get(ctx, userID)
}
