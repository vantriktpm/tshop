package rest

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/tshop/backend/services/promotion-service/internal/domain"
)

type PromotionRepository interface {
	GetByCode(ctx context.Context, code string) (*domain.Promotion, error)
}

type PromotionHandler struct {
	repo PromotionRepository
}

func NewPromotionHandler(repo PromotionRepository) *PromotionHandler {
	return &PromotionHandler{repo: repo}
}

func (h *PromotionHandler) Validate(c *gin.Context) {
	c.JSON(200, gin.H{"valid": true, "discount": 0})
}
