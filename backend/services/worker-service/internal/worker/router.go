package worker

import (
	"context"
	"log"
	"sync"

	"github.com/IBM/sarama"
	"github.com/tshop/backend/pkg/events"
	"github.com/tshop/backend/services/worker-service/internal/domain"
	redisidem "github.com/tshop/backend/services/worker-service/internal/infrastructure/redis"
)

// Router maps Kafka topics to job handlers and runs them with idempotency.
type Router struct {
	idem    *redisidem.Idempotency
	handlers map[string][]domain.JobHandler // topic -> handlers
	mu      sync.RWMutex
}

func NewRouter(idem *redisidem.Idempotency) *Router {
	return &Router{idem: idem, handlers: make(map[string][]domain.JobHandler)}
}

// Register adds handlers for a topic. Handlers run in order.
func (r *Router) Register(topic string, handlers ...domain.JobHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[topic] = append(r.handlers[topic], handlers...)
}

func (r *Router) Handlers(topic string) []domain.JobHandler {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.handlers[topic]
}

// Process claims the message (idempotency); if claimed, runs all handlers for the topic.
func (r *Router) Process(ctx context.Context, topic string, partition int32, offset int64, key, payload []byte) error {
	claimed, err := r.idem.Claim(ctx, topic, partition, offset)
	if err != nil {
		return err
	}
	if !claimed {
		return nil // already processed
	}
	keyStr := string(key)
	handlers := r.Handlers(topic)
	job := domain.Job{Topic: topic, Key: keyStr, Payload: payload}
	for _, h := range handlers {
		if err := h.Handle(ctx, job); err != nil {
			return err
		}
	}
	return nil
}

// ConsumerGroupHandler implements sarama.ConsumerGroupHandler.
func (r *Router) ConsumerGroupHandler() sarama.ConsumerGroupHandler {
	return &groupHandler{router: r}
}

type groupHandler struct {
	router *Router
}

func (h *groupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *groupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *groupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			ctx := session.Context()
			if err := h.router.Process(ctx, msg.Topic, msg.Partition, msg.Offset, msg.Key, msg.Value); err != nil {
				log.Printf("[worker] process error topic=%s partition=%d offset=%d: %v", msg.Topic, msg.Partition, msg.Offset, err)
				// Don't mark offset so message can be retried
				continue
			}
			session.MarkMessage(msg, "")
		case <-session.Context().Done():
			return nil
		}
	}
}

// DefaultTopics returns topics this worker subscribes to.
func DefaultTopics() []string {
	return []string{
		events.TopicOrderCreated,
		events.TopicOrderPaid,
		events.TopicNotification,
		events.TopicPaymentIntent,
		events.TopicInventoryReserve,
	}
}
