// worker-sync-avatar: chạy liên tục, consume Kafka topic user.avatar.sync và gọi image-service sync-avatar.
//
// Nếu topic đã có message nhưng worker không lấy được: thường do consumer group đã commit offset (đã chạy trước đó).
// Reset offset để đọc lại từ đầu:
//
//	docker exec <kafka-container> kafka-consumer-groups --bootstrap-server localhost:29092 \
//	  --group worker-sync-avatar --reset-offsets --to-earliest --topic user.avatar.sync --execute
//
// Hoặc dùng group mới: set env KAFKA_GROUP_ID=worker-sync-avatar-v2
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/tshop/backend/pkg/events"
)

func main() {
	brokersStr := os.Getenv("KAFKA_BROKER")
	if brokersStr == "" {
		brokersStr = "localhost:9092"
	}
	brokers := strings.Split(brokersStr, ",")
	imageServiceURL := os.Getenv("IMAGE_SERVICE_URL")
	if imageServiceURL == "" {
		imageServiceURL = "http://image-service:5010"
	}
	imageServiceURL = strings.TrimSuffix(imageServiceURL, "/")

	cfg := sarama.NewConfig()
	cfg.Version = sarama.V2_6_0_0
	cfg.Consumer.Return.Errors = true
	// OffsetOldest: khi group chưa commit offset thì đọc từ đầu topic → lấy được message đã có.
	// Nếu group đã commit offset rồi thì vẫn theo offset đã commit; muốn đọc lại từ đầu cần reset group.
	if os.Getenv("KAFKA_OFFSET_RESET") == "newest" {
		cfg.Consumer.Offsets.Initial = sarama.OffsetNewest
	} else {
		cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	}

	// Tạo topic user.avatar.sync nếu chưa có (Kafka không tự tạo khi chỉ có consumer)
	ensureTopic(brokers, cfg, events.TopicUserAvatarSync)
	// Đợi metadata / __consumer_offsets sẵn sàng (tránh lỗi "Offset's topic has not yet been created")
	time.Sleep(3 * time.Second)

	groupID := os.Getenv("KAFKA_GROUP_ID")
	if groupID == "" {
		groupID = "worker-sync-avatar"
	}
	client, err := sarama.NewConsumerGroup(brokers, groupID, cfg)
	if err != nil {
		log.Fatalf("worker-sync-avatar: NewConsumerGroup: %v", err)
	}
	defer client.Close()

	handler := &avatarSyncHandler{imageServiceURL: imageServiceURL}
	ctx := context.Background()
	topics := []string{events.TopicUserAvatarSync}
	offsetReset := "oldest"
	if cfg.Consumer.Offsets.Initial == sarama.OffsetNewest {
		offsetReset = "newest"
	}
	log.Printf("worker-sync-avatar: subscribing to %v (brokers=%v, group=%s, offset=%s)",
		topics, brokers, groupID, offsetReset)

	var backoff time.Duration
	for {
		if ctx.Err() != nil {
			return
		}
		err := client.Consume(ctx, topics, handler)
		if err == nil {
			backoff = 0
			continue
		}
		log.Printf("worker-sync-avatar: Consume: %v", err)
		// Lỗi "Offset's topic has not yet been created": Kafka __consumer_offsets chưa sẵn sàng → retry với backoff
		if strings.Contains(err.Error(), "not yet been created") {
			if backoff == 0 {
				backoff = 2 * time.Second
			}
			if backoff < 30*time.Second {
				backoff *= 2
			}
			log.Printf("worker-sync-avatar: retry in %v", backoff)
			time.Sleep(backoff)
			continue
		}
		time.Sleep(2 * time.Second)
	}
}

func ensureTopic(brokers []string, cfg *sarama.Config, topic string) {
	admin, err := sarama.NewClusterAdmin(brokers, cfg)
	if err != nil {
		log.Printf("worker-sync-avatar: ClusterAdmin (create topic): %v", err)
		return
	}
	defer admin.Close()
	err = admin.CreateTopic(topic, &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
	}, false)
	if err != nil && !errors.Is(err, sarama.ErrTopicAlreadyExists) {
		log.Printf("worker-sync-avatar: CreateTopic %q: %v", topic, err)
		return
	}
	log.Printf("worker-sync-avatar: topic %q ready", topic)
}

type avatarSyncHandler struct {
	imageServiceURL string
}

func (h *avatarSyncHandler) Setup(session sarama.ConsumerGroupSession) error {
	log.Printf("worker-sync-avatar: Setup claims=%v", session.Claims())
	return nil
}

func (h *avatarSyncHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *avatarSyncHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	log.Printf("worker-sync-avatar: ConsumeClaim topic=%s partition=%d", claim.Topic(), claim.Partition())
	for msg := range claim.Messages() {
		if msg.Topic == events.TopicUserAvatarSync {
			h.handleUserAvatarSync(msg.Value)
		}
		session.MarkMessage(msg, "")
	}
	return nil
}

func (h *avatarSyncHandler) handleUserAvatarSync(payload []byte) {
	var evt events.UserAvatarSyncEvent
	if err := json.Unmarshal(payload, &evt); err != nil {
		log.Printf("worker-sync-avatar: user.avatar.sync unmarshal: %v", err)
		return
	}
	if evt.UserID == "" || evt.PictureURL == "" {
		log.Printf("worker-sync-avatar: user.avatar.sync skip empty user_id=%q picture_url=%q", evt.UserID, evt.PictureURL)
		return
	}
	body, _ := json.Marshal(map[string]string{
		"user_id":     evt.UserID,
		"picture_url": evt.PictureURL,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.imageServiceURL+"/api/images/sync-avatar", bytes.NewReader(body))
	if err != nil {
		log.Printf("worker-sync-avatar: sync-avatar new request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("worker-sync-avatar: sync-avatar call: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("worker-sync-avatar: sync-avatar status %d for user=%s", resp.StatusCode, evt.UserID)
		return
	}
	log.Printf("worker-sync-avatar: user.avatar.sync done user=%s", evt.UserID)
}
