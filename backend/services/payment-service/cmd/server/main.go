package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/tshop/backend/services/payment-service/internal/delivery/rest"
	"github.com/tshop/backend/services/payment-service/internal/infrastructure/postgres"
	"github.com/tshop/backend/services/payment-service/internal/usecase"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=1 dbname=payment_db port=5432 sslmode=disable"
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
	repo := postgres.NewPaymentRepository(db)
	h := rest.NewPaymentHandler(usecase.NewCreatePayment(repo))
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	r.POST("/api/payments/intent", h.CreateIntent)
	// Health endpoint specific to payment-service
	r.GET("/payments/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	log.Println("payment-service :5006")
	_ = r.Run(":5006")
}
