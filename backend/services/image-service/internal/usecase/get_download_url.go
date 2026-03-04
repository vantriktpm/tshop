package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/tshop/backend/services/image-service/internal/domain"
)

// GetDownloadURLResult holds presigned GET URL for client to download directly from MinIO.
type GetDownloadURLResult struct {
	DownloadURL string `json:"download_url"`
	ExpiresIn   int    `json:"expires_in_seconds"`
}

type GetDownloadURL struct {
	repo    domain.ImageRepository
	storage DownloadURLProvider
	expiry  time.Duration
}

type DownloadURLProvider interface {
	PresignedGetURL(ctx context.Context, bucket, objectKey string, expiry time.Duration) (string, error)
}

func NewGetDownloadURL(repo domain.ImageRepository, storage DownloadURLProvider, expiry time.Duration) *GetDownloadURL {
	if expiry <= 0 {
		expiry = 15 * time.Minute
	}
	return &GetDownloadURL{repo: repo, storage: storage, expiry: expiry}
}

func (u *GetDownloadURL) Execute(ctx context.Context, imageID string) (*GetDownloadURLResult, error) {
	img, err := u.repo.GetByID(ctx, imageID)
	if err != nil {
		return nil, fmt.Errorf("get image: %w", err)
	}
	if img == nil {
		return nil, fmt.Errorf("image not found: %s", imageID)
	}
	urlStr, err := u.storage.PresignedGetURL(ctx, img.BucketName, img.ObjectKey, u.expiry)
	if err != nil {
		return nil, fmt.Errorf("presigned get: %w", err)
	}
	return &GetDownloadURLResult{
		DownloadURL: urlStr,
		ExpiresIn:   int(u.expiry.Seconds()),
	}, nil
}
