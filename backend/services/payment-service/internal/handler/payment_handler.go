package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tshop/backend/services/payment-service/internal/service"
)

type PaymentHandler struct {
	svc *service.PaymentService
}

func NewPaymentHandler(svc *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{svc: svc}
}

func (h *PaymentHandler) CreateIntent(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"intent_id": "pi_xxx", "status": "pending"})
}
