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
	"github.com/tshop/backend/pkg/logger"
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
		logger.Error("worker_sync_avatar_new_consumer_group_failed", map[string]interface{}{
			"service": "worker-sync-avatar",
			"error":   err.Error(),
		})
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
	logger.Info("worker_sync_avatar_subscribe", map[string]interface{}{
		"service": "worker-sync-avatar",
		"topics":  topics,
		"brokers": brokers,
		"group":   groupID,
		"offset":  offsetReset,
	})

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
		logger.Error("worker_sync_avatar_consume_failed", map[string]interface{}{
			"service": "worker-sync-avatar",
			"error":   err.Error(),
		})
		// Lỗi "Offset's topic has not yet been created": Kafka __consumer_offsets chưa sẵn sàng → retry với backoff
		if strings.Contains(err.Error(), "not yet been created") {
			if backoff == 0 {
				backoff = 2 * time.Second
			}
			if backoff < 30*time.Second {
				backoff *= 2
			}
			logger.Info("worker_sync_avatar_retry", map[string]interface{}{
				"service": "worker-sync-avatar",
				"backoff": backoff.String(),
			})
			time.Sleep(backoff)
			continue
		}
		time.Sleep(2 * time.Second)
	}
}

func ensureTopic(brokers []string, cfg *sarama.Config, topic string) {
	admin, err := sarama.NewClusterAdmin(brokers, cfg)
	if err != nil {
		logger.Error("worker_sync_avatar_cluster_admin_failed", map[string]interface{}{
			"service": "worker-sync-avatar",
			"error":   err.Error(),
		})
		return
	}
	defer admin.Close()
	err = admin.CreateTopic(topic, &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
	}, false)
	if err != nil && !errors.Is(err, sarama.ErrTopicAlreadyExists) {
		logger.Error("worker_sync_avatar_create_topic_failed", map[string]interface{}{
			"service": "worker-sync-avatar",
			"topic":   topic,
			"error":   err.Error(),
		})
		return
	}
	logger.Info("worker_sync_avatar_topic_ready", map[string]interface{}{
		"service": "worker-sync-avatar",
		"topic":   topic,
	})
}

type avatarSyncHandler struct {
	imageServiceURL string
}

func (h *avatarSyncHandler) Setup(session sarama.ConsumerGroupSession) error {
	logger.Info("worker_sync_avatar_setup", map[string]interface{}{
		"service": "worker-sync-avatar",
		"claims":  session.Claims(),
	})
	return nil
}

func (h *avatarSyncHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *avatarSyncHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	logger.Info("worker_sync_avatar_consume_claim", map[string]interface{}{
		"service":   "worker-sync-avatar",
		"topic":     claim.Topic(),
		"partition": claim.Partition(),
	})
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
		logger.Error("worker_sync_avatar_unmarshal_failed", map[string]interface{}{
			"service": "worker-sync-avatar",
			"error":   err.Error(),
		})
		return
	}
	if evt.UserID == "" || evt.PictureURL == "" {
		logger.Info("worker_sync_avatar_skip_empty_payload", map[string]interface{}{
			"service":     "worker-sync-avatar",
			"user_id":     evt.UserID,
			"picture_url": evt.PictureURL,
		})
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
		logger.Error("worker_sync_avatar_new_request_failed", map[string]interface{}{
			"service": "worker-sync-avatar",
			"url":     h.imageServiceURL + "/api/images/sync-avatar",
			"error":   err.Error(),
		})
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error("worker_sync_avatar_call_failed", map[string]interface{}{
			"service": "worker-sync-avatar",
			"url":     h.imageServiceURL + "/api/images/sync-avatar",
			"error":   err.Error(),
		})
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logger.Error("worker_sync_avatar_call_bad_status", map[string]interface{}{
			"service":    "worker-sync-avatar",
			"status":     resp.StatusCode,
			"user_id":    evt.UserID,
			"image_url":  h.imageServiceURL + "/api/images/sync-avatar",
		})
		return
	}
	logger.Info("worker_sync_avatar_done", map[string]interface{}{
		"service": "worker-sync-avatar",
		"user_id": evt.UserID,
	})
}
