package postgres

import (
	"context"
	"time"

	"github.com/tshop/backend/services/order-service/internal/domain"
	"gorm.io/gorm"
)

type OrderModel struct {
	gorm.Model
	ID          string  `gorm:"primaryKey;column:id"`
	UserID      string  `gorm:"column:user_id"`
	Status      string  `gorm:"column:status"`
	TotalAmount float64 `gorm:"column:total_amount"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type OrderItemModel struct {
	gorm.Model
	OrderID   string  `gorm:"column:order_id"`
	ProductID string  `gorm:"column:product_id"`
	Quantity  int64   `gorm:"column:quantity"`
	Price     float64 `gorm:"column:price"`
}

func (OrderModel) TableName() string { return "orders" }
func (OrderItemModel) TableName() string { return "order_items" }

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, order *domain.Order) error {
	if err := r.db.WithContext(ctx).Create(&OrderModel{
		ID:          order.ID,
		UserID:      order.UserID,
		Status:      string(order.Status),
		TotalAmount: order.TotalAmount,
		CreatedAt:   order.CreatedAt,
		UpdatedAt:   order.UpdatedAt,
	}).Error; err != nil {
		return err
	}
	for i := range order.Items {
		if err := r.db.WithContext(ctx).Create(&OrderItemModel{
			OrderID:   order.ID,
			ProductID: order.Items[i].ProductID,
			Quantity:  order.Items[i].Quantity,
			Price:     order.Items[i].Price,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	var m OrderModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return &domain.Order{
		ID:          m.ID,
		UserID:      m.UserID,
		Status:      domain.OrderStatus(m.Status),
		TotalAmount: m.TotalAmount,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	return r.db.WithContext(ctx).Model(&OrderModel{}).
		Where("id = ?", id).
		Update("status", string(status)).Error
}
