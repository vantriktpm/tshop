package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tshop/backend/services/payment-service/internal/usecase"
)

type PaymentHandler struct {
	createPayment *usecase.CreatePayment
}

func NewPaymentHandler(createPayment *usecase.CreatePayment) *PaymentHandler {
	return &PaymentHandler{createPayment: createPayment}
}

func (h *PaymentHandler) CreateIntent(c *gin.Context) {
	var body struct {
		OrderID string  `json:"order_id"`
		Amount  float64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := h.createPayment.Execute(c.Request.Context(), body.OrderID, body.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"intent_id": p.ID, "status": p.Status})
}
