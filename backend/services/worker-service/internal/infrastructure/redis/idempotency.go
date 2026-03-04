package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const idempotencyPrefix = "worker:idempotency:"
const idempotencyTTL = 7 * 24 * time.Hour

// Idempotency stores processed message keys to skip duplicates.
type Idempotency struct {
	client *redis.Client
}

func NewIdempotency(client *redis.Client) *Idempotency {
	return &Idempotency{client: client}
}

// Claim returns true if this message was not yet processed (we claimed it). False = already processed, skip.
func (i *Idempotency) Claim(ctx context.Context, topic string, partition int32, offset int64) (claimed bool, err error) {
	key := idempotencyKey(topic, partition, offset)
	ok, err := i.client.SetNX(ctx, key, "1", idempotencyTTL).Result()
	return ok, err
}

func idempotencyKey(topic string, partition int32, offset int64) string {
	return fmt.Sprintf("%s%s:%d:%d", idempotencyPrefix, topic, partition, offset)
}
