package handlers

import (
	"context"
	"log"

	"github.com/tshop/backend/services/worker-service/internal/domain"
)

// ProcessPaymentHandler processes payment (e.g. payment.intent). Wire to payment-service.
type ProcessPaymentHandler struct{}

func NewProcessPaymentHandler() *ProcessPaymentHandler {
	return &ProcessPaymentHandler{}
}

func (h *ProcessPaymentHandler) Handle(ctx context.Context, job domain.Job) error {
	log.Printf("[worker] process_payment topic=%s key=%s payload_len=%d", job.Topic, job.Key, len(job.Payload))
	// TODO: call payment provider / payment-service, update order status on success
	return nil
}
