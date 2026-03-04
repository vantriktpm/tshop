package postgres

import (
	"context"

	"gorm.io/gorm"
	"github.com/tshop/backend/services/order-service/internal/domain"
)

type OrderModel struct {
	gorm.Model
	ID          string  `gorm:"primaryKey"`
	UserID      string
	Status      string
	TotalAmount float64
}

func (OrderModel) TableName() string { return "orders" }

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, order *domain.Order) error {
	m := OrderModel{
		ID: order.ID, UserID: order.UserID, Status: string(order.Status), TotalAmount: order.TotalAmount,
	}
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	var m OrderModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return &domain.Order{ID: m.ID, UserID: m.UserID, Status: domain.OrderStatus(m.Status), TotalAmount: m.TotalAmount}, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	return r.db.WithContext(ctx).Model(&OrderModel{}).Where("id = ?", id).Update("status", string(status)).Error
}
