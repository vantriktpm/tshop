package domain

import "time"

// Order aggregate root (DDD)
type Order struct {
	ID         string
	UserID     string
	Status     OrderStatus
	TotalAmount float64
	Items      []OrderItem
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type OrderItem struct {
	ProductID string
	Quantity  int64
	Price     float64
}

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusCancelled OrderStatus = "cancelled"
)
