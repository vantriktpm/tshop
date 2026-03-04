package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/tshop/backend/services/cart-service/internal/usecase"
)

type CartHandler struct{ getCart *usecase.GetCart }

func NewCartHandler(getCart *usecase.GetCart) *CartHandler { return &CartHandler{getCart: getCart} }

func (h *CartHandler) Get(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		userID = "anonymous"
	}
	cart, err := h.getCart.Execute(c.Request.Context(), userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, cart)
}
