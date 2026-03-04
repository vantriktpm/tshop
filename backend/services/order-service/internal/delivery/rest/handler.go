package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tshop/backend/services/order-service/internal/usecase"
)

type OrderHandler struct {
	createOrder *usecase.CreateOrder
}

func NewOrderHandler(createOrder *usecase.CreateOrder) *OrderHandler {
	return &OrderHandler{createOrder: createOrder}
}

type CreateOrderReq struct {
	Items []struct {
		ProductID string  `json:"product_id"`
		Quantity  int64   `json:"quantity"`
		Price     float64 `json:"price"`
	} `json:"items"`
}

func (h *OrderHandler) Create(c *gin.Context) {
	userID := c.GetString("user_id") // from JWT middleware
	if userID == "" {
		userID = "anonymous"
	}
	var req CreateOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input := usecase.CreateOrderInput{UserID: userID}
	for _, it := range req.Items {
		input.Items = append(input.Items, struct {
			ProductID string
			Quantity  int64
			Price     float64
		}{ProductID: it.ProductID, Quantity: it.Quantity, Price: it.Price})
	}
	order, err := h.createOrder.Execute(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, order)
}
