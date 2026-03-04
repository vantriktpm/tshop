package usecase

import (
	"context"
	"github.com/tshop/backend/services/promotion-service/internal/domain"
)

type ApplyPromotion struct{ repo domain.PromotionRepository }

func NewApplyPromotion(repo domain.PromotionRepository) *ApplyPromotion { return &ApplyPromotion{repo: repo} }

func (u *ApplyPromotion) Execute(ctx context.Context, code string) (*domain.Promotion, error) {
	return u.repo.GetByCode(ctx, code)
}
