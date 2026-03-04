// Order-service: REST (Gin) + gRPC, PostgreSQL, Kafka (event-driven, Saga).
package main

import (
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/tshop/backend/pkg/events"
	"github.com/tshop/backend/services/order-service/internal/delivery/rest"
	"github.com/tshop/backend/services/order-service/internal/infrastructure/kafka"
	"github.com/tshop/backend/services/order-service/internal/infrastructure/postgres"
	"github.com/tshop/backend/services/order-service/internal/usecase"
	"google.golang.org/grpc"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Infrastructure: PostgreSQL (order DB per service)
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=1 dbname=order_db port=5432 sslmode=disable"
	}
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Printf("postgres: %v (run migrations or use sqlite for dev)", err)
	}
	db, err := gorm.Open(gormpostgres.New(gormpostgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		log.Printf("postgres: %v (run migrations or use sqlite for dev)", err)
		// Optional: use in-memory for local dev
	}

	// Kafka (event-driven: publish OrderCreated for inventory, payment, notification)
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		kafkaBroker = "localhost:9092"
	}
	brokers := []string{kafkaBroker}
	pub, err := kafka.NewPublisher(brokers)
	if err != nil {
		log.Printf("kafka: %v (events disabled)", err)
		pub = nil
	}
	var eventPub events.Publisher = events.NoopPublisher{}
	if pub != nil {
		defer pub.Close()
		eventPub = pub
	}

	orderRepo := postgres.NewOrderRepository(db)
	createOrder := usecase.NewCreateOrder(orderRepo, eventPub)

	// Delivery: REST (Gin) - external API
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
	// Rate limit, JWT auth middleware in production
	handler := rest.NewOrderHandler(createOrder)
	r.POST("/api/orders", handler.Create)
	// Health endpoint specific to order-service
	r.GET("/orders/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	// Delivery: gRPC - internal service-to-service (mTLS in production)
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal("grpc listen:", err)
	}
	grpcServer := grpc.NewServer()
	// grpc.RegisterOrderGRPC(grpcServer, grpcSvc)
	go func() {
		log.Println("order-service gRPC :50051")
		_ = grpcServer.Serve(lis)
	}()

	log.Println("order-service REST :8081")
	if err := r.Run(":8081"); err != nil {
		log.Fatal("gin:", err)
	}
}
