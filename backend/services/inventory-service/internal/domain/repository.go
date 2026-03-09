package domain

import (
	"context"
	"errors"
)

var ErrInsufficientStock = errors.New("insufficient stock")

type InventoryRepository interface {
	Reserve(ctx context.Context, productID string, qty int64) error
	GetStock(ctx context.Context, productID string) (*Stock, error)
}
