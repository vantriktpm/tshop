package usecase

import (
	"context"
	"github.com/tshop/backend/services/notification-service/internal/domain"
)

type SendNotification struct{ repo domain.NotificationRepository }

func NewSendNotification(repo domain.NotificationRepository) *SendNotification {
	return &SendNotification{repo: repo}
}

func (u *SendNotification) Execute(ctx context.Context, userID, channel, payload string) error {
	return u.repo.Send(ctx, &domain.Notification{ID: "n-1", UserID: userID, Channel: channel, Payload: payload})
}
