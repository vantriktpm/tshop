package usecase

import (
	"context"
	"fmt"

	"github.com/tshop/backend/services/image-service/internal/domain"
)

type GetImage struct {
	repo domain.ImageRepository
}

func NewGetImage(repo domain.ImageRepository) *GetImage {
	return &GetImage{repo: repo}
}

func (u *GetImage) Execute(ctx context.Context, imageID string) (*domain.Image, error) {
	img, err := u.repo.GetByID(ctx, imageID)
	if err != nil {
		return nil, fmt.Errorf("get image: %w", err)
	}
	if img == nil {
		return nil, fmt.Errorf("image not found: %s", imageID)
	}
	return img, nil
}
