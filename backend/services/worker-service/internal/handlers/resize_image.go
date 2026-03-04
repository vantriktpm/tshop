package handlers

import (
	"context"
	"log"

	"github.com/tshop/backend/services/worker-service/internal/domain"
)

// ResizeImageHandler resizes images (e.g. thumbnails). Wire to image-service or image lib.
type ResizeImageHandler struct{}

func NewResizeImageHandler() *ResizeImageHandler {
	return &ResizeImageHandler{}
}

func (h *ResizeImageHandler) Handle(ctx context.Context, job domain.Job) error {
	log.Printf("[worker] resize_image topic=%s key=%s payload_len=%d", job.Topic, job.Key, len(job.Payload))
	// TODO: parse object_key/bucket, download from MinIO, resize, upload back (or new key)
	return nil
}
