package postgres

import (
	"context"
	"github.com/tshop/backend/services/inventory-service/internal/domain"
	"gorm.io/gorm"
)

type StockModel struct {
	gorm.Model
	ProductID string `gorm:"primaryKey"`
	Quantity  int64
}

func (StockModel) TableName() string { return "inventory" }

type InventoryRepository struct{ db *gorm.DB }

func NewInventoryRepository(db *gorm.DB) *InventoryRepository { return &InventoryRepository{db: db} }

func (r *InventoryRepository) Reserve(ctx context.Context, productID string, qty int64) error {
	return r.db.WithContext(ctx).Model(&StockModel{}).Where("product_id = ?", productID).
		Update("quantity", gorm.Expr("quantity - ?", qty)).Error
}

func (r *InventoryRepository) GetStock(ctx context.Context, productID string) (*domain.Stock, error) {
	var m StockModel
	if err := r.db.WithContext(ctx).Where("product_id = ?", productID).First(&m).Error; err != nil {
		return nil, err
	}
	return &domain.Stock{ProductID: m.ProductID, Quantity: m.Quantity}, nil
}
