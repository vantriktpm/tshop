package redis

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

const channelAvatarSaved = "avatar.saved"

// AvatarSavedNotifier publishes { user_id, image_id } to Redis channel for WebSocket gateway.
type AvatarSavedNotifier struct {
	rdb *redis.Client
}

func NewAvatarSavedNotifier(rdb *redis.Client) *AvatarSavedNotifier {
	return &AvatarSavedNotifier{rdb: rdb}
}

func (n *AvatarSavedNotifier) NotifyAvatarSaved(ctx context.Context, userID, imageID string) error {
	payload, _ := json.Marshal(map[string]string{"user_id": userID, "image_id": imageID})
	return n.rdb.Publish(ctx, channelAvatarSaved, payload).Err()
}
