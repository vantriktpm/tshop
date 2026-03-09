package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
	"github.com/tshop/backend/pkg/events"
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
			log.Printf("avatar-consumer: unmarshal: %v", err)
			session.MarkMessage(msg, "")
			continue
		}
		if evt.UserID == "" || evt.PictureURL == "" {
			log.Printf("avatar-consumer: skip empty payload user_id=%q picture_url=%q", evt.UserID, evt.PictureURL)
			session.MarkMessage(msg, "")
			continue
		}

		if imageID, err := h.sync.Execute(context.Background(), evt.UserID, evt.PictureURL); err != nil {
			log.Printf("avatar-consumer: sync user=%s: %v", evt.UserID, err)
		} else if imageID != "" {
			// Notify via Redis so gateway/WS can push to frontend (handled in handler if AvatarNotifier set)
			_ = imageID
		}
		session.MarkMessage(msg, "")
	}
	return nil
}

