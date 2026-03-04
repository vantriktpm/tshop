package rest

import (
	"github.com/gin-gonic/gin"
)

type PromotionHandler struct{}

func NewPromotionHandler() *PromotionHandler { return &PromotionHandler{} }

func (h *PromotionHandler) Validate(c *gin.Context) {
	c.JSON(200, gin.H{"valid": true, "discount": 0})
}
