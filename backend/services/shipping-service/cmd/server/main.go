package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/tshop/backend/services/shipping-service/internal/delivery/rest"
	"github.com/tshop/backend/services/shipping-service/internal/infrastructure/postgres"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=1 dbname=shipping_db port=5432 sslmode=disable"
	}
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}
	db, err := gorm.Open(gormpostgres.New(gormpostgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	if err := postgres.EnsureSchema(dsn, db); err != nil {
		log.Fatal(err)
	}
	repo := postgres.NewShippingRepository(db)
	h := rest.NewShippingHandler(repo)
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	r.POST("/api/shipping", h.Create)
	// Health endpoint specific to shipping-service
	r.GET("/shipping/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	log.Println("shipping-service :5007")
	_ = r.Run(":5007")
}
