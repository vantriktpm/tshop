package usecase

import (
	"context"
	"encoding/json"

	"github.com/tshop/backend/pkg/events"
	"github.com/tshop/backend/services/order-service/internal/domain"
)

type CreateOrderInput struct {
	UserID string
	Items  []struct {
		ProductID string
		Quantity  int64
		Price     float64
	}
}

type CreateOrder struct {
	repo      domain.OrderRepository
	publisher events.Publisher
}

func NewCreateOrder(repo domain.OrderRepository, pub events.Publisher) *CreateOrder {
	return &CreateOrder{repo: repo, publisher: pub}
}

func (u *CreateOrder) Execute(ctx context.Context, input CreateOrderInput) (*domain.Order, error) {
	var total float64
	items := make([]domain.OrderItem, len(input.Items))
	for i, it := range input.Items {
		items[i] = domain.OrderItem{ProductID: it.ProductID, Quantity: it.Quantity, Price: it.Price}
		total += it.Price * float64(it.Quantity)
	}
	order := &domain.Order{
		ID:          generateOrderID(),
		UserID:       input.UserID,
		Status:       domain.OrderStatusPending,
		TotalAmount:  total,
		Items:        items,
	}
	if err := u.repo.Create(ctx, order); err != nil {
		return nil, err
	}
	// Saga: publish OrderCreated for inventory, payment, notification
	if u.publisher != nil {
		evt := events.OrderCreatedEvent{OrderID: order.ID, UserID: order.UserID, TotalAmount: order.TotalAmount}
		for _, it := range order.Items {
			evt.Items = append(evt.Items, events.OrderItem{ProductID: it.ProductID, Quantity: it.Quantity})
		}
		payload, _ := json.Marshal(evt)
		_ = u.publisher.Publish(ctx, events.TopicOrderCreated, order.ID, payload)
	}
	return order, nil
}

func generateOrderID() string {
	// TODO: use UUID or ULID
	return "ord-1"
}
