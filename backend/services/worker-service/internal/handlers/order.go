package handlers

import (
	"context"
	"encoding/json"
	"log"

	"github.com/tshop/backend/pkg/events"
	"github.com/tshop/backend/services/worker-service/internal/domain"
)

// ProcessOrderHandler handles order flow after Kafka event (e.g. order.created -> reserve inventory, trigger payment).
type ProcessOrderHandler struct{}

func NewProcessOrderHandler() *ProcessOrderHandler {
	return &ProcessOrderHandler{}
}

func (h *ProcessOrderHandler) Handle(ctx context.Context, job domain.Job) error {
	log.Printf("[worker] process_order topic=%s key=%s", job.Topic, job.Key)
	switch job.Topic {
	case events.TopicOrderCreated:
		var evt events.OrderCreatedEvent
		if err := json.Unmarshal(job.Payload, &evt); err != nil {
			return err
		}
		// Idempotent: update order state to "inventory_reserved" or "payment_pending" after downstream steps
		log.Printf("[worker] order_created order_id=%s user_id=%s total=%.2f", evt.OrderID, evt.UserID, evt.TotalAmount)
		_ = evt
	case events.TopicOrderPaid:
		// Mark order paid, trigger shipping, etc.
		log.Printf("[worker] order_paid key=%s", job.Key)
	}
	return nil
}
