package postgres

import (
	"context"
	"github.com/tshop/backend/services/promotion-service/internal/domain"
	"gorm.io/gorm"
)

type PromotionModel struct {
	gorm.Model
	ID         string  `gorm:"primaryKey"`
	Code       string
	Discount   float64
	ValidUntil string
}

func (PromotionModel) TableName() string { return "promotions" }

type PromotionRepository struct{ db *gorm.DB }

func NewPromotionRepository(db *gorm.DB) *PromotionRepository { return &PromotionRepository{db: db} }

func (r *PromotionRepository) GetByCode(ctx context.Context, code string) (*domain.Promotion, error) {
	var m PromotionModel
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&m).Error; err != nil {
		return nil, err
	}
	return &domain.Promotion{ID: m.ID, Code: m.Code, Discount: m.Discount}, nil
}
