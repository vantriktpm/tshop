package domain

import "context"

type CartRepository interface {
	Get(ctx context.Context, userID string) (*Cart, error)
	Set(ctx context.Context, cart *Cart) error
}
