package usecase

import (
	"context"
	"github.com/tshop/backend/services/cart-service/internal/domain"
)

type GetCart struct{ repo domain.CartRepository }

func NewGetCart(repo domain.CartRepository) *GetCart { return &GetCart{repo: repo} }

func (u *GetCart) Execute(ctx context.Context, userID string) (*domain.Cart, error) {
	return u.repo.Get(ctx, userID)
}
