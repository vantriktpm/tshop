package domain

import "context"

type InventoryRepository interface {
	Reserve(ctx context.Context, productID string, qty int64) error
	GetStock(ctx context.Context, productID string) (*Stock, error)
}
