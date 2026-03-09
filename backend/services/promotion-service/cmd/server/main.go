package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/tshop/backend/services/promotion-service/internal/delivery/rest"
	"github.com/tshop/backend/services/promotion-service/internal/infrastructure/postgres"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=1 dbname=promotion_db port=5432 sslmode=disable"
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
	repo := postgres.NewPromotionRepository(db)
	h := rest.NewPromotionHandler(repo)
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	r.GET("/api/promotions/validate", h.Validate)
	// Health endpoint specific to promotion-service
	r.GET("/promotions/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	log.Println("promotion-service :5008")
	_ = r.Run(":5008")
}
