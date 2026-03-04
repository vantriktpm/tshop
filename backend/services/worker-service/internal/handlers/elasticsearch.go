package handlers

import (
	"context"
	"log"
	"time"

	"github.com/tshop/backend/services/worker-service/internal/domain"
)

// LogElasticsearchHandler writes event log to Elasticsearch. Wire to ES client.
type LogElasticsearchHandler struct{}

func NewLogElasticsearchHandler() *LogElasticsearchHandler {
	return &LogElasticsearchHandler{}
}

func (h *LogElasticsearchHandler) Handle(ctx context.Context, job domain.Job) error {
	log.Printf("[worker] log_elasticsearch topic=%s key=%s payload_len=%d", job.Topic, job.Key, len(job.Payload))
	// TODO: index document { "@timestamp": now(), "topic": job.Topic, "key": job.Key, "payload": base64 or raw }
	_ = time.Now()
	return nil
}
