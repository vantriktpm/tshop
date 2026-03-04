package handlers

import (
	"context"
	"log"

	"github.com/tshop/backend/services/worker-service/internal/domain"
)

// SendEmailHandler sends email (e.g. notification). Wire to SMTP or notification-service.
type SendEmailHandler struct {
	// Mailer interface can be injected for real sending
}

func NewSendEmailHandler() *SendEmailHandler {
	return &SendEmailHandler{}
}

func (h *SendEmailHandler) Handle(ctx context.Context, job domain.Job) error {
	log.Printf("[worker] send_email topic=%s key=%s payload_len=%d", job.Topic, job.Key, len(job.Payload))
	// TODO: parse payload (e.g. notification.send event), call SMTP or notification-service API
	return nil
}
