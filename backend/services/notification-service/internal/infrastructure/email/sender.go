package email

import (
	"context"
	"github.com/tshop/backend/services/notification-service/internal/domain"
)

type Sender struct{}

func NewSender() *Sender { return &Sender{} }

func (s *Sender) Send(ctx context.Context, n *domain.Notification) error {
	// TODO: SMTP or SendGrid
	return nil
}
