package rest

import (
	"github.com/gin-gonic/gin"
)

type ShippingHandler struct{}

func NewShippingHandler() *ShippingHandler { return &ShippingHandler{} }

func (h *ShippingHandler) Create(c *gin.Context) {
	c.JSON(200, gin.H{"shipment_id": "ship-1", "status": "pending"})
}
