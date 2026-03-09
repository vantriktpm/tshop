package postgres

import (
	"context"

	"github.com/tshop/backend/services/payment-service/internal/domain"
	"gorm.io/gorm"
)

type PaymentModel struct {
	gorm.Model
	ID      string  `gorm:"primaryKey"`
	OrderID string
	Amount  float64
	Status  string
}

func (PaymentModel) TableName() string { return "payments" }

type PaymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(ctx context.Context, p *domain.Payment) error {
	return r.db.WithContext(ctx).Create(&PaymentModel{
		ID: p.ID, OrderID: p.OrderID, Amount: p.Amount, Status: p.Status,
	}).Error
}

func (r *PaymentRepository) GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	var m PaymentModel
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&m).Error; err != nil {
		return nil, err
	}
	return &domain.Payment{ID: m.ID, OrderID: m.OrderID, Amount: m.Amount, Status: m.Status, CreatedAt: m.CreatedAt}, nil
}
