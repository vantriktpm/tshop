package domain

import "context"

type ProductRepository interface {
	GetByID(ctx context.Context, id string) (*Product, error)
	List(ctx context.Context, limit, offset int) ([]*Product, error)
}
