package postgres

import (
	"context"
	"github.com/tshop/backend/services/shipping-service/internal/domain"
	"gorm.io/gorm"
)

type ShipmentModel struct {
	gorm.Model
	ID      string `gorm:"primaryKey"`
	OrderID string
	Address string
	Status  string
}

func (ShipmentModel) TableName() string { return "shipments" }

type ShippingRepository struct{ db *gorm.DB }

func NewShippingRepository(db *gorm.DB) *ShippingRepository { return &ShippingRepository{db: db} }

func (r *ShippingRepository) Create(ctx context.Context, s *domain.Shipment) error {
	return r.db.WithContext(ctx).Create(&ShipmentModel{ID: s.ID, OrderID: s.OrderID, Address: s.Address, Status: s.Status}).Error
}
