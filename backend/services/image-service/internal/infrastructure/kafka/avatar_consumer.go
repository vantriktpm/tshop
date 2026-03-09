package kafka

import (
	"context"
	"encoding/json"

	"github.com/IBM/sarama"
	"github.com/tshop/backend/pkg/events"
	"github.com/tshop/backend/pkg/logger"
	"github.com/tshop/backend/services/image-service/internal/usecase"
)

// AvatarConsumer handles user avatar sync events from Kafka.
type AvatarConsumer struct {
	sync *usecase.SyncUserAvatar
}

func NewAvatarConsumer(sync *usecase.SyncUserAvatar) sarama.ConsumerGroupHandler {
	return &AvatarConsumer{sync: sync}
}

func (h *AvatarConsumer) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *AvatarConsumer) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *AvatarConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		if msg.Topic != events.TopicUserAvatarSync {
			session.MarkMessage(msg, "")
			continue
		}

		var evt events.UserAvatarSyncEvent
		if err := json.Unmarshal(msg.Value, &evt); err != nil {
			logger.Error("avatar_consumer_unmarshal_failed", map[string]interface{}{
				"service": "image-service",
				"topic":   msg.Topic,
				"error":   err.Error(),
			})
			session.MarkMessage(msg, "")
			continue
		}
		if evt.UserID == "" || evt.PictureURL == "" {
			logger.Info("avatar_consumer_skip_empty_payload", map[string]interface{}{
				"service":     "image-service",
				"topic":       msg.Topic,
				"user_id":     evt.UserID,
				"picture_url": evt.PictureURL,
			})
			session.MarkMessage(msg, "")
			continue
		}

		if imageID, err := h.sync.Execute(context.Background(), evt.UserID, evt.PictureURL); err != nil {
			logger.Error("avatar_consumer_sync_failed", map[string]interface{}{
				"service":     "image-service",
				"topic":       msg.Topic,
				"user_id":     evt.UserID,
				"picture_url": evt.PictureURL,
				"error":       err.Error(),
			})
		} else if imageID != "" {
			// Notify via Redis so gateway/WS can push to frontend (handled in handler if AvatarNotifier set)
			logger.Info("avatar_consumer_sync_success", map[string]interface{}{
				"service":     "image-service",
				"topic":       msg.Topic,
				"user_id":     evt.UserID,
				"picture_url": evt.PictureURL,
				"image_id":    imageID,
			})
		}
		session.MarkMessage(msg, "")
	}
	return nil
}

