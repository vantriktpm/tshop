package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
	"github.com/tshop/backend/services/inventory-service/internal/domain"
	"github.com/tshop/backend/services/inventory-service/internal/delivery/rest"
	"github.com/tshop/backend/services/inventory-service/internal/infrastructure"
	"github.com/tshop/backend/services/inventory-service/internal/infrastructure/postgres"
	redisinfra "github.com/tshop/backend/services/inventory-service/internal/infrastructure/redis"
	"github.com/tshop/backend/services/inventory-service/internal/usecase"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=1 dbname=inventory_db port=5432 sslmode=disable"
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
	pgRepo := postgres.NewInventoryRepository(db)
	var repo domain.InventoryRepository = pgRepo
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		rdb := redis.NewClient(&redis.Options{Addr: addr})
		if err := rdb.Ping(context.Background()).Err(); err != nil {
			log.Printf("Redis ping failed, using DB only: %v", err)
		} else {
			cache := redisinfra.NewStockCache(rdb)
			repo = infrastructure.NewCachedInventoryRepository(pgRepo, cache)
			log.Println("inventory-service: Redis cache enabled (RDB+AOF backup)")
		}
	}
	h := rest.NewInventoryHandler(usecase.NewReserveStock(repo))
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	r.POST("/api/inventory/reserve", h.Reserve)
	// Health endpoint specific to inventory-service
	r.GET("/inventory/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	log.Println("inventory-service :5004")
	_ = r.Run(":5004")
}
