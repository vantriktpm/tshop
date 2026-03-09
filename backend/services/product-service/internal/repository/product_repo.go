package repository

import (
	"context"
	"time"

	"github.com/tshop/backend/services/product-service/internal/domain"
	"gorm.io/gorm"
)

// ProductModel GORM model for product table.
type ProductModel struct {
	ProductID   string     `gorm:"column:product_id;primaryKey"`
	ProductCode string     `gorm:"column:product_code;size:50"`
	ProductName string     `gorm:"column:product_name;size:255"`
	Quantity    float64    `gorm:"column:quantity"`
	Price       float64    `gorm:"column:price"`
	PriceSale   float64    `gorm:"column:price_sale"`
	CreatedBy   *string    `gorm:"column:created_by;size:50"`
	UpdatedBy   *string    `gorm:"column:updated_by;size:50"`
	CreatedDate *time.Time `gorm:"column:created_date"`
	UpdatedDate *time.Time `gorm:"column:updated_date"`
	ImageID     *string    `gorm:"column:image_id"`
}

func (ProductModel) TableName() string { return "product" }

// Migrate runs AutoMigrate for product-service tables.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&ProductModel{})
}

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	var m ProductModel
	if err := r.db.WithContext(ctx).Where("product_id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return modelToProduct(&m), nil
}

func (r *ProductRepository) List(ctx context.Context, limit, offset int) ([]*domain.Product, error) {
	var list []ProductModel
	if err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.Product, len(list))
	for i := range list {
		out[i] = modelToProduct(&list[i])
	}
	return out, nil
}

func modelToProduct(m *ProductModel) *domain.Product {
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
