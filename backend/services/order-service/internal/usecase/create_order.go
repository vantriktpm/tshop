package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/tshop/backend/pkg/events"
	"github.com/tshop/backend/services/order-service/internal/domain"
)

type CreateOrder struct {
	repo domain.OrderRepository
	pub  events.Publisher
}

func NewCreateOrder(repo domain.OrderRepository, pub events.Publisher) *CreateOrder {
	return &CreateOrder{repo: repo, pub: pub}
}

func (u *CreateOrder) Execute(ctx context.Context, userID string, items []domain.OrderItem, totalAmount float64) (*domain.Order, error) {
	now := time.Now()
	order := &domain.Order{
		ID:          uuid.New().String(),
		UserID:      userID,
		Status:      domain.OrderStatusPending,
		TotalAmount: totalAmount,
		Items:       items,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := u.repo.Create(ctx, order); err != nil {
		return nil, err
	}
	if u.pub != nil {
		payload, _ := json.Marshal(map[string]interface{}{
			"order_id": order.ID, "user_id": userID, "total": totalAmount,
		})
		_ = u.pub.Publish(ctx, "order.created", order.ID, payload)
	}
	return order, nil
}
