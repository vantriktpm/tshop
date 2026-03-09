package container

import (
	"github.com/tshop/backend/services/product-service/internal/domain"
	"github.com/tshop/backend/services/product-service/internal/handler"
	"github.com/tshop/backend/services/product-service/internal/repository"
	"github.com/tshop/backend/services/product-service/internal/service"
	"gorm.io/gorm"
)

type Container struct {
	db *gorm.DB

	productRepo    domain.ProductRepository
	productService *service.ProductService
	productHandler *handler.ProductHandler
}

func New(db *gorm.DB) *Container {
	c := &Container{db: db}
	c.productRepo = repository.NewProductRepository(db)
	c.productService = service.NewProductService(c.productRepo)
	c.productHandler = handler.NewProductHandler(c.productService)
	return c
}

func (c *Container) ProductHandler() *handler.ProductHandler { return c.productHandler }
func (c *Container) DB() *gorm.DB                            { return c.db }
