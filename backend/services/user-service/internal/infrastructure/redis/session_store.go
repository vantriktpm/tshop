package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	sessionKeyPrefix = "session:"
	defaultSessionTTL = 24 * time.Hour
)

// SessionStore lưu phiên đăng nhập (session_id -> user_id) trong Redis.
// Dùng RDB + AOF để khi Redis restart vẫn khôi phục được session từ dump.rdb / appendonly.aof.
type SessionStore struct {
	client *redis.Client
	ttl    time.Duration
}

func NewSessionStore(client *redis.Client, ttl time.Duration) *SessionStore {
	if ttl <= 0 {
		ttl = defaultSessionTTL
	}
	return &SessionStore{client: client, ttl: ttl}
}

// SetSession tạo phiên sau khi đăng nhập thành công. sessionID nên là UUID.
func (s *SessionStore) SetSession(ctx context.Context, sessionID, userID string) error {
	key := sessionKeyPrefix + sessionID
	return s.client.Set(ctx, key, userID, s.ttl).Err()
}

// GetSession trả về userID nếu session còn hiệu lực.
func (s *SessionStore) GetSession(ctx context.Context, sessionID string) (userID string, err error) {
	key := sessionKeyPrefix + sessionID
	userID, err = s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return userID, err
}

// DeleteSession xóa phiên (đăng xuất).
func (s *SessionStore) DeleteSession(ctx context.Context, sessionID string) error {
	return s.client.Del(ctx, sessionKeyPrefix+sessionID).Err()
}

// KeyPrefix trả về prefix dùng cho session (để debug hoặc scan).
func (s *SessionStore) KeyPrefix() string { return sessionKeyPrefix }
