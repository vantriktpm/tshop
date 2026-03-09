package domain

import "context"

// SessionStore lưu phiên đăng nhập (session_id -> user_id) để khi Redis restart
// vẫn khôi phục từ RDB/AOF. Mỗi lần đăng nhập thành công tạo session_id mới.
type SessionStore interface {
	SetSession(ctx context.Context, sessionID, userID string) error
	GetSession(ctx context.Context, sessionID string) (userID string, err error)
	DeleteSession(ctx context.Context, sessionID string) error
}
