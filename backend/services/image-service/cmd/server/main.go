package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/tshop/backend/pkg/events"
	"github.com/tshop/backend/services/image-service/internal/delivery/rest"
	"github.com/tshop/backend/services/image-service/internal/domain"
	kafkainfra "github.com/tshop/backend/services/image-service/internal/infrastructure/kafka"
	"github.com/tshop/backend/services/image-service/internal/infrastructure/minio"
	"github.com/tshop/backend/services/image-service/internal/infrastructure/postgres"
	"github.com/tshop/backend/services/image-service/internal/usecase"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=1 dbname=postgres port=5432 sslmode=disable"
	}
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}
	db, err := gorm.Open(gormpostgres.New(gormpostgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	if err := db.AutoMigrate(&postgres.ImageModel{}); err != nil {
		log.Fatal(err)
	}

	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:9000"
	}
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	if accessKey == "" {
		accessKey = "minioadmin"
	}
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	if secretKey == "" {
		secretKey = "minioadmin"
	}
	useSSL := false
	if v := os.Getenv("MINIO_USE_SSL"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			useSSL = b
		}
	}
	presignMinutes := 15
	if v := os.Getenv("MINIO_PRESIGN_MINUTES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			presignMinutes = n
		}
	}

	storage, err := minio.NewStorage(minio.Config{
		Endpoint:      endpoint,
		AccessKey:     accessKey,
		SecretKey:     secretKey,
		UseSSL:        useSSL,
		PresignExpiry: time.Duration(presignMinutes) * time.Minute,
	})
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	for _, b := range []string{domain.BucketProductImages, domain.BucketUserAvatars, domain.BucketOrderInvoices} {
		if err := storage.EnsureBucket(ctx, b); err != nil {
			log.Printf("ensure bucket %s: %v", b, err)
		}
	}

	repo := postgres.NewImageRepository(db)
	presignExp := 15 * time.Minute
	h := rest.NewImageHandler(
		usecase.NewCreateImage(repo, storage, presignExp),
		usecase.NewGetDownloadURL(repo, storage, presignExp),
		usecase.NewGetImage(repo),
	)

	// Kafka consumer for user avatar sync events
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		kafkaBroker = "localhost:9092"
	}
	groupID := "image-service-avatar"
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V2_8_0_0
	cfg.Consumer.Offsets.Initial = sarama.OffsetNewest
	cfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}

	consumer, err := sarama.NewConsumerGroup([]string{kafkaBroker}, groupID, cfg)
	if err != nil {
		log.Printf("kafka consumer: %v (avatar sync disabled)", err)
	} else {
		syncAvatar := usecase.NewSyncUserAvatar(repo, storage)
		handler := kafkainfra.NewAvatarConsumer(syncAvatar)
		go func() {
			defer consumer.Close()
			ctx := context.Background()
			topics := []string{events.TopicUserAvatarSync}
			for {
				if err := consumer.Consume(ctx, topics, handler); err != nil && err != context.Canceled {
					log.Printf("avatar-consumer: %v", err)
					time.Sleep(5 * time.Second)
					continue
				}
				if ctx.Err() != nil {
					return
				}
			}
		}()
	}

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
	r.POST("/api/images", h.CreateImage)
	r.GET("/api/images/:id/download-url", h.GetDownloadURL)
	r.GET("/api/images/:id", h.GetImage)
	// Health check specific to image-service
	r.GET("/images/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	log.Println("image-service :8085")
	_ = r.Run(":8085")
}
