package kafka

import (
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
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}
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
				session.MarkMessage(msg, "")
				continue
			}
			if err := h.sync.Execute(session.Context(), evt.UserID, evt.PictureURL); err != nil {
				log.Printf("avatar-consumer: sync user=%s: %v", evt.UserID, err)
				// Mark message to avoid infinite retry; adjust if at-least-once semantics are required.
			}
			session.MarkMessage(msg, "")
		case <-session.Context().Done():
			return nil
		}
	}
}

