package domain

import "context"

type PromotionRepository interface {
	GetByCode(ctx context.Context, code string) (*Promotion, error)
}
