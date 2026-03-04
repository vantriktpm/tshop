package domain

import "context"

type ShippingRepository interface {
	Create(ctx context.Context, s *Shipment) error
}
