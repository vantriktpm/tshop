package domain

import "context"

type NotificationRepository interface {
	Send(ctx context.Context, n *Notification) error
}
