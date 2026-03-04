package domain

import "context"

// Job types the worker can run (triggered by Kafka or internal queue).
const (
	JobSendEmail       = "send_email"
	JobResizeImage     = "resize_image"
	JobProcessPayment  = "process_payment"
	JobSyncInventory   = "sync_inventory"
	JobLogElasticsearch = "log_elasticsearch"
	JobProcessOrder    = "process_order"
)

// Job is a unit of work with optional payload (from Kafka message).
type Job struct {
	Type    string
	Topic   string
	Key     string
	Payload []byte
}

// JobHandler processes a job. Must be idempotent when possible.
type JobHandler interface {
	Handle(ctx context.Context, job Job) error
}
