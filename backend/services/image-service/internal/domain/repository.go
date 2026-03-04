package domain

import "context"

type ImageRepository interface {
	Create(ctx context.Context, img *Image) error
	GetByID(ctx context.Context, id string) (*Image, error)
}
