package handlers

import (
	"context"
	"encoding/json"
	"log"

	"github.com/tshop/backend/pkg/events"
	"github.com/tshop/backend/services/worker-service/internal/domain"
)

// SyncInventoryHandler reserves/decrements inventory (e.g. on order.created). Wire to inventory-service.
type SyncInventoryHandler struct{}

func NewSyncInventoryHandler() *SyncInventoryHandler {
	return &SyncInventoryHandler{}
}

func (h *SyncInventoryHandler) Handle(ctx context.Context, job domain.Job) error {
	log.Printf("[worker] sync_inventory topic=%s key=%s", job.Topic, job.Key)
	if job.Topic != events.TopicOrderCreated {
		return nil
	}
	var evt events.OrderCreatedEvent
	if err := json.Unmarshal(job.Payload, &evt); err != nil {
		return err
	}
	// TODO: call inventory-service (gRPC or HTTP) to reserve items for evt.OrderID, evt.Items
	_ = evt
	return nil
}
