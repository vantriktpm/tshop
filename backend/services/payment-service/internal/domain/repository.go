package domain

import "context"

type PaymentRepository interface {
	Create(ctx context.Context, p *Payment) error
	GetByOrderID(ctx context.Context, orderID string) (*Payment, error)
}
