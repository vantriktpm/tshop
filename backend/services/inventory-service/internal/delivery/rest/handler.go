package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/tshop/backend/services/inventory-service/internal/usecase"
)

type InventoryHandler struct{ reserve *usecase.ReserveStock }

func NewInventoryHandler(reserve *usecase.ReserveStock) *InventoryHandler {
	return &InventoryHandler{reserve: reserve}
}

func (h *InventoryHandler) Reserve(c *gin.Context) {
	// consume OrderCreated event in production (Kafka consumer)
	c.JSON(200, gin.H{"status": "ok"})
}
