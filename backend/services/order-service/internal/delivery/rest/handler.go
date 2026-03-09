package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/tshop/backend/services/order-service/internal/domain"
	"github.com/tshop/backend/services/order-service/internal/usecase"
)

type OrderHandler struct {
	createOrder *usecase.CreateOrder
}

func NewOrderHandler(createOrder *usecase.CreateOrder) *OrderHandler {
	return &OrderHandler{createOrder: createOrder}
}

func (h *OrderHandler) Create(c *gin.Context) {
	var body struct {
		UserID      string           `json:"user_id"`
		Items       []domain.OrderItem `json:"items"`
		TotalAmount float64          `json:"total_amount"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	order, err := h.createOrder.Execute(c.Request.Context(), body.UserID, body.Items, body.TotalAmount)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, order)
}
