package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/tshop/backend/services/product-service/internal/delivery/rest"
	"github.com/tshop/backend/services/product-service/internal/infrastructure/postgres"
	"github.com/tshop/backend/services/product-service/internal/usecase"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=1 dbname=product_db port=5432 sslmode=disable"
	}
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}
	db, err := gorm.Open(gormpostgres.New(gormpostgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	repo := postgres.NewProductRepository(db)
	h := rest.NewProductHandler(usecase.NewListProducts(repo))
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	r.GET("/api/products", h.List)
	// Health endpoint specific to product-service
	r.GET("/products/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	log.Println("product-service :8082")
	_ = r.Run(":8082")
}
