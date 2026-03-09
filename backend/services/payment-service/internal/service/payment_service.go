package service

// PaymentService is used by internal/handler (optional alternative to usecase).
type PaymentService struct{}

func NewPaymentService() *PaymentService {
	return &PaymentService{}
}
