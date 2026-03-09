package domain

import "context"

// AvatarSavedNotifier publishes avatar.saved events (e.g. to Redis) so WebSocket can push to frontend.
type AvatarSavedNotifier interface {
	NotifyAvatarSaved(ctx context.Context, userID, imageID string) error
}
