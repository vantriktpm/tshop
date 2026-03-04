package domain

import "context"

// OrderRepository port (driven) - implemented in infrastructure
type OrderRepository interface {
	Create(ctx context.Context, order *Order) error
	GetByID(ctx context.Context, id string) (*Order, error)
	UpdateStatus(ctx context.Context, id string, status OrderStatus) error
}
