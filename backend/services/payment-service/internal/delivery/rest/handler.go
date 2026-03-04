package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/tshop/backend/services/payment-service/internal/usecase"
)

type PaymentHandler struct{ createPayment *usecase.CreatePayment }

func NewPaymentHandler(createPayment *usecase.CreatePayment) *PaymentHandler {
	return &PaymentHandler{createPayment: createPayment}
}

func (h *PaymentHandler) CreateIntent(c *gin.Context) {
	c.JSON(200, gin.H{"intent_id": "pi_xxx", "status": "pending"})
}
