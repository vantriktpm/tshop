package usecase

import (
	"context"

	"github.com/tshop/backend/services/inventory-service/internal/domain"
)

type ReserveStock struct {
	repo domain.InventoryRepository
}

func NewReserveStock(repo domain.InventoryRepository) *ReserveStock {
	return &ReserveStock{repo: repo}
}

func (u *ReserveStock) Execute(ctx context.Context, productID string, qty int64) error {
	return u.repo.Reserve(ctx, productID, qty)
}
