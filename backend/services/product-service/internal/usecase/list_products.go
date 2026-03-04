package usecase

import (
	"context"
	"github.com/tshop/backend/services/product-service/internal/domain"
)

type ListProducts struct{ repo domain.ProductRepository }

func NewListProducts(repo domain.ProductRepository) *ListProducts { return &ListProducts{repo: repo} }

func (u *ListProducts) Execute(ctx context.Context, limit, offset int) ([]*domain.Product, error) {
	return u.repo.List(ctx, limit, offset)
}
