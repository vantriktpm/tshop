package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tshop/backend/services/payment-service/internal/domain"
)

type CreatePayment struct {
	repo domain.PaymentRepository
}

func NewCreatePayment(repo domain.PaymentRepository) *CreatePayment {
	return &CreatePayment{repo: repo}
}

func (u *CreatePayment) Execute(ctx context.Context, orderID string, amount float64) (*domain.Payment, error) {
	p := &domain.Payment{
		ID:        uuid.New().String(),
		OrderID:   orderID,
		Amount:    amount,
		Status:    "pending",
		CreatedAt: time.Now(),
	}
	if err := u.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}
