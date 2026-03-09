package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tshop/backend/services/product-service/internal/service"
)

type ProductHandler struct {
	productService *service.ProductService
}

func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

func (h *ProductHandler) List(c *gin.Context) {
	limit, _ := c.GetQuery("limit")
	offset, _ := c.GetQuery("offset")
	l, o := 20, 0
	if limit != "" {
		if n, err := strconv.Atoi(limit); err == nil && n > 0 {
			l = n
		}
	}
	if offset != "" {
		if n, err := strconv.Atoi(offset); err == nil && n >= 0 {
			o = n
		}
	}
	list, err := h.productService.List(c.Request.Context(), l, o)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, list)
}
