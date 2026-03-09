package rest

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/tshop/backend/services/shipping-service/internal/domain"
)

type ShippingRepository interface {
	Create(ctx context.Context, s *domain.Shipment) error
}

type ShippingHandler struct {
	repo ShippingRepository
}

func NewShippingHandler(repo ShippingRepository) *ShippingHandler {
	return &ShippingHandler{repo: repo}
}

func (h *ShippingHandler) Create(c *gin.Context) {
	c.JSON(200, gin.H{"shipment_id": "ship-1", "status": "pending"})
}
