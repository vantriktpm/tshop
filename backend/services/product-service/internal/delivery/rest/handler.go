package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/tshop/backend/services/product-service/internal/usecase"
)

type ProductHandler struct{ listProducts *usecase.ListProducts }

func NewProductHandler(listProducts *usecase.ListProducts) *ProductHandler {
	return &ProductHandler{listProducts: listProducts}
}

func (h *ProductHandler) List(c *gin.Context) {
	list, err := h.listProducts.Execute(c.Request.Context(), 20, 0)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, list)
}
