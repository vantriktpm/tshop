package service

import (
	"context"

	"github.com/tshop/backend/services/product-service/internal/domain"
)

type ProductService struct {
	repo domain.ProductRepository
}

func NewProductService(repo domain.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) List(ctx context.Context, limit, offset int) ([]*domain.Product, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.repo.List(ctx, limit, offset)
}
