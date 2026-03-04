package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
	"github.com/tshop/backend/pkg/events"
	"github.com/tshop/backend/services/worker-service/internal/handlers"
	"github.com/tshop/backend/services/worker-service/internal/infrastructure/kafka"
	redisidem "github.com/tshop/backend/services/worker-service/internal/infrastructure/redis"
	"github.com/tshop/backend/services/worker-service/internal/worker"
)

func main() {
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		kafkaBroker = "localhost:9092"
	}
	brokers := []string{kafkaBroker}
	groupID := "tshop-worker"
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	// Redis for idempotency
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer rdb.Close()
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Printf("redis ping: %v (idempotency disabled in practice if Redis down)", err)
	}

	idem := redisidem.NewIdempotency(rdb)
	router := worker.NewRouter(idem)

	// Register handlers per topic (idempotent processing per message)
	emailH := handlers.NewSendEmailHandler()
	resizeH := handlers.NewResizeImageHandler()
	paymentH := handlers.NewProcessPaymentHandler()
	inventoryH := handlers.NewSyncInventoryHandler()
	esH := handlers.NewLogElasticsearchHandler()
	orderH := handlers.NewProcessOrderHandler()

	// order.created -> sync inventory, send notification, log ES, process order
	router.Register(events.TopicOrderCreated, inventoryH, emailH, esH, orderH)
	// order.paid -> email, resize (e.g. invoice image), log ES, process order
	router.Register(events.TopicOrderPaid, emailH, resizeH, esH, orderH)
	// notification.send -> send email
	router.Register(events.TopicNotification, emailH)
	// payment.intent -> process payment
	router.Register(events.TopicPaymentIntent, paymentH)
	// inventory.reserve -> sync inventory (if used as callback topic)
	router.Register(events.TopicInventoryReserve, inventoryH)

	// Kafka consumer group
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V2_8_0_0
	cfg.Consumer.Offsets.Initial = sarama.OffsetNewest
	cfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	consumer, err := kafka.NewConsumerGroup(brokers, groupID, cfg)
	if err != nil {
		log.Fatalf("kafka consumer: %v", err)
	}
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown: signal cancels context, consumer stops, then we exit
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		log.Println("worker: shutdown signal received, draining...")
		cancel()
	}()

	topics := worker.DefaultTopics()
	log.Printf("worker: consuming topics %v group=%s", topics, groupID)

	errCh := make(chan error, 1)
	go func() {
		errCh <- consumer.Consume(ctx, topics, router.ConsumerGroupHandler())
	}()

	select {
	case err := <-errCh:
		if err != nil && err != context.Canceled {
			log.Printf("worker: consume error: %v", err)
		}
	case <-ctx.Done():
		// Allow a short drain for in-flight messages
		time.Sleep(2 * time.Second)
	}

	log.Println("worker: stopped")
}
