package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/tshop/backend/services/inventory-service/internal/usecase"
)

type InventoryHandler struct {
	reserveStock *usecase.ReserveStock
}

func NewInventoryHandler(reserveStock *usecase.ReserveStock) *InventoryHandler {
	return &InventoryHandler{reserveStock: reserveStock}
}

func (h *InventoryHandler) Reserve(c *gin.Context) {
	var body struct {
		ProductID string `json:"product_id"`
		Quantity  int64  `json:"quantity"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := h.reserveStock.Execute(c.Request.Context(), body.ProductID, body.Quantity); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "reserved"})
}
