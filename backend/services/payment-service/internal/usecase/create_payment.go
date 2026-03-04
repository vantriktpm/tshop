package usecase

import (
	"context"
	"github.com/tshop/backend/services/payment-service/internal/domain"
)

type CreatePayment struct{ repo domain.PaymentRepository }

func NewCreatePayment(repo domain.PaymentRepository) *CreatePayment { return &CreatePayment{repo: repo} }

func (u *CreatePayment) Execute(ctx context.Context, orderID string, amount float64) (*domain.Payment, error) {
	p := &domain.Payment{ID: "pay-1", OrderID: orderID, Amount: amount, Status: "pending"}
	return p, u.repo.Create(ctx, p)
}
