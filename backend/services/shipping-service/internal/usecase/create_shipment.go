package usecase

import (
	"context"
	"github.com/tshop/backend/services/shipping-service/internal/domain"
)

type CreateShipment struct{ repo domain.ShippingRepository }

func NewCreateShipment(repo domain.ShippingRepository) *CreateShipment { return &CreateShipment{repo: repo} }

func (u *CreateShipment) Execute(ctx context.Context, orderID, address string) (*domain.Shipment, error) {
	s := &domain.Shipment{ID: "ship-1", OrderID: orderID, Address: address, Status: "pending"}
	return s, u.repo.Create(ctx, s)
}
