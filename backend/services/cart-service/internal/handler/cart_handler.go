package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/tshop/backend/services/cart-service/internal/service"
)

type CartHandler struct {
	cartService *service.CartService
}

func NewCartHandler(cartService *service.CartService) *CartHandler {
	return &CartHandler{cartService: cartService}
}

func (h *CartHandler) Get(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		userID = "anonymous"
	}
	cart, err := h.cartService.GetCart(c.Request.Context(), userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, cart)
}
