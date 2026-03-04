// Package events provides event types and interfaces for event-driven architecture (Kafka).
package events

import "context"

// Topic names (Kafka)
const (
	TopicOrderCreated     = "order.created"
	TopicOrderPaid        = "order.paid"
	TopicInventoryReserve = "inventory.reserve"
	TopicPaymentIntent    = "payment.intent"
	TopicNotification     = "notification.send"
	TopicUserAvatarSync   = "user.avatar.sync"
)

// OrderCreatedEvent is published by order-service (Saga choreography).
type OrderCreatedEvent struct {
	OrderID     string      `json:"order_id"`
	UserID      string      `json:"user_id"`
	TotalAmount float64     `json:"total_amount"`
	Items       []OrderItem `json:"items"`
}

type OrderItem struct {
	ProductID string `json:"product_id"`
	Quantity  int64  `json:"quantity"`
}

// UserAvatarSyncEvent published by user-service when a third-party avatar (e.g. Google) should be synced to image-service.
type UserAvatarSyncEvent struct {
	UserID     string `json:"user_id"`
	PictureURL string `json:"picture_url"`
}

// Publisher interface for infrastructure (Kafka).
type Publisher interface {
	Publish(ctx context.Context, topic string, key string, payload []byte) error
}

// Consumer interface for subscribing to topics.
type Consumer interface {
	Subscribe(ctx context.Context, topic string, handler func(msg []byte) error) error
}

// NoopPublisher no-op implementation when Kafka is unavailable.
type NoopPublisher struct{}

func (NoopPublisher) Publish(ctx context.Context, topic, key string, payload []byte) error {
	return nil
}
