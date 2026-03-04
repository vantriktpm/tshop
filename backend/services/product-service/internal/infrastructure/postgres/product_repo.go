package postgres

import (
	"context"
	"time"

	"github.com/tshop/backend/services/product-service/internal/domain"
	"gorm.io/gorm"
)

type ProductModel struct {
	ProductID   string     `gorm:"column:product_id;primaryKey"`
	ProductCode string     `gorm:"column:product_code;size:50;not null"`
	ProductName string     `gorm:"column:product_name;size:255;not null"`
	Quantity    float64    `gorm:"column:quantity;not null;default:0"`
	Price       float64    `gorm:"column:price;not null;default:0"`
	PriceSale   float64    `gorm:"column:price_sale;default:0"`
	CreatedBy   *string    `gorm:"column:created_by;size:50"`
	UpdatedBy   *string    `gorm:"column:updated_by;size:50"`
	CreatedDate *time.Time `gorm:"column:created_date"`
	UpdatedDate *time.Time `gorm:"column:updated_date"`
	ImageID     *string    `gorm:"column:image_id"`
}

func (ProductModel) TableName() string { return "product" }

type ProductRepository struct{ db *gorm.DB }

func NewProductRepository(db *gorm.DB) *ProductRepository { return &ProductRepository{db: db} }

func (r *ProductRepository) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	var m ProductModel
	if err := r.db.WithContext(ctx).Where("product_id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return modelToDomain(&m), nil
}

func (r *ProductRepository) List(ctx context.Context, limit, offset int) ([]*domain.Product, error) {
	var models []ProductModel
	if err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.Product, len(models))
	for i := range models {
		out[i] = modelToDomain(&models[i])
	}
	return out, nil
}

func modelToDomain(m *ProductModel) *domain.Product {
	return &domain.Product{
		ProductID:   m.ProductID,
		ProductCode: m.ProductCode,
		ProductName: m.ProductName,
		Quantity:    m.Quantity,
		Price:       m.Price,
		PriceSale:   m.PriceSale,
		CreatedBy:   m.CreatedBy,
		UpdatedBy:   m.UpdatedBy,
		CreatedDate: m.CreatedDate,
		UpdatedDate: m.UpdatedDate,
		ImageID:     m.ImageID,
	}
}
