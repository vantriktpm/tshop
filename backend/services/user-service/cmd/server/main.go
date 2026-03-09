package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/tshop/backend/pkg/events"
	"github.com/tshop/backend/services/user-service/internal/core/config"
	"github.com/tshop/backend/services/user-service/internal/delivery/rest"
	"github.com/tshop/backend/services/user-service/internal/domain"
	"github.com/tshop/backend/services/user-service/internal/infrastructure/kafka"
	"github.com/tshop/backend/services/user-service/internal/infrastructure/postgres"
	redisinfra "github.com/tshop/backend/services/user-service/internal/infrastructure/redis"
)

func main() {
	db, err := config.NewPostgresDB()
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	repo := postgres.NewUserRepository(db)

	// Redis: session store (RDB + AOF để khôi phục phiên khi Redis restart)
	var sessionStore *redisinfra.SessionStore
	if addr := getEnv("REDIS_ADDR", "localhost:6379"); addr != "" {
		rdb := redis.NewClient(&redis.Options{Addr: addr})
		if err := rdb.Ping(context.Background()).Err(); err != nil {
			log.Printf("redis: %v (sessions disabled)", err)
		} else {
			sessionStore = redisinfra.NewSessionStore(rdb, 24*time.Hour)
		}
	}

	// Kafka publisher for events (e.g. user.avatar.sync)
	brokers := []string{getEnv("KAFKA_BROKER", "localhost:9092")}
	var eventPub events.Publisher = events.NoopPublisher{}
	if pub, err := kafka.NewPublisher(brokers); err != nil {
		log.Printf("kafka: %v (events disabled)", err)
	} else {
		defer pub.Close()
		eventPub = pub
	}
	baseURL := config.CallbackBaseURL()
	frontURL := config.FrontendRedirectURL()
	jwtSecret := config.JWTSecret()

	fbOAuth := config.FacebookOAuthConfig{
		AppID:               config.FacebookAppID(),
		AppSecret:           config.FacebookAppSecret(),
		CallbackBaseURL:     baseURL,
		FrontendRedirectURL: frontURL,
		JWTSecret:           jwtSecret,
	}
	googleOAuth := config.GoogleOAuthConfig{
		ClientID:            config.GoogleClientID(),
		ClientSecret:        config.GoogleClientSecret(),
		CallbackBaseURL:     baseURL,
		FrontendRedirectURL: frontURL,
		JWTSecret:           jwtSecret,
	}
	xOAuth := config.XOAuthConfig{
		ClientID:            config.XClientID(),
		ClientSecret:        config.XClientSecret(),
		CallbackBaseURL:     baseURL,
		FrontendRedirectURL: frontURL,
		JWTSecret:           jwtSecret,
	}

	var sess domain.SessionStore
	if sessionStore != nil {
		sess = sessionStore
	}
	h := rest.NewUserHandler(eventPub, sess)
	r := gin.Default()
	// CORS chỉ set ở gateway; service phía sau không thêm để tránh header trùng "*, *"
	r.Use(func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	// Google verify: frontend POST credential → tạo/cập nhật user, trả JWT
	googleVerify := h.GoogleVerify(googleOAuth, repo)
	r.POST("/api/auth/google/verify", googleVerify)
	r.POST("/auth/google/verify", googleVerify)

	// Local email/password auth (login & register)
	r.POST("/api/auth/login", h.LocalLogin(jwtSecret, repo))
	r.POST("/api/auth/register", h.LocalRegister(jwtSecret, repo))

	// Lấy thông tin user hiện tại từ JWT
	r.GET("/api/auth", h.GetMe)
	r.GET(config.GoogleStartPath, h.GoogleStart(googleOAuth))
	r.GET(config.GoogleCallbackPath, h.GoogleCallback(googleOAuth, repo))
	r.GET(config.FacebookStartPath, h.FacebookStart(fbOAuth))
	r.GET(config.FacebookCallbackPath, h.FacebookCallback(fbOAuth, repo))
	r.GET(config.XStartPath, h.XStart(xOAuth))
	r.GET(config.XCallbackPath, h.XCallback(xOAuth, repo))
	// Health endpoint specific to user-service
	r.GET("/users/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	port := getEnv("USER_SERVICE_PORT", "5001")
	log.Println("user-service :" + port)
	log.Println("Facebook OAuth Redirect URI:", fbOAuth.RedirectURI())
	log.Println("Google OAuth Redirect URI:", googleOAuth.RedirectURI())
	log.Println("X OAuth Redirect URI:", xOAuth.RedirectURI())
	_ = r.Run(":" + port)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
