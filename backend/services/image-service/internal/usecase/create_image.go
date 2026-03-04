package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tshop/backend/services/image-service/internal/domain"
)

// CreateImageInput for requesting an upload slot. Backend creates record and returns presigned PUT URL.
type CreateImageInput struct {
	Bucket      string // product-images | user-avatars | order-invoices
	ObjectKey   string // optional; if empty, generated as uuid
	ContentType *string
	Size        *int64
	CreatedBy   *string
}

// CreateImageResult returned after creating metadata; client uses UploadURL to PUT file.
type CreateImageResult struct {
	ImageID    string `json:"image_id"`
	UploadURL  string `json:"upload_url"`
	ObjectKey  string `json:"object_key"`
	BucketName string `json:"bucket_name"`
	ExpiresIn  int    `json:"expires_in_seconds"`
}

type CreateImage struct {
	repo    domain.ImageRepository
	storage PresignedStorage
	expiry  time.Duration
}

type PresignedStorage interface {
	PresignedPutURL(ctx context.Context, bucket, objectKey string, expiry time.Duration) (string, error)
	EnsureBucket(ctx context.Context, bucket string) error
}

func NewCreateImage(repo domain.ImageRepository, storage PresignedStorage, expiry time.Duration) *CreateImage {
	if expiry <= 0 {
		expiry = 15 * time.Minute
	}
	return &CreateImage{repo: repo, storage: storage, expiry: expiry}
}

func (u *CreateImage) Execute(ctx context.Context, in CreateImageInput) (*CreateImageResult, error) {
	if in.Bucket == "" {
		in.Bucket = domain.BucketProductImages
	}
	objectKey := in.ObjectKey
	if objectKey == "" {
		objectKey = uuid.New().String()
	}
	if err := u.storage.EnsureBucket(ctx, in.Bucket); err != nil {
		return nil, fmt.Errorf("ensure bucket: %w", err)
	}
	uploadURL, err := u.storage.PresignedPutURL(ctx, in.Bucket, objectKey, u.expiry)
	if err != nil {
		return nil, fmt.Errorf("presigned put: %w", err)
	}
	imageID := uuid.New().String()
	now := time.Now()
	img := &domain.Image{
		ID:          imageID,
		ObjectKey:   objectKey,
		BucketName:  in.Bucket,
		ContentType: in.ContentType,
		Size:        in.Size,
		CreatedBy:   in.CreatedBy,
		CreatedAt:   &now,
	}
	if err := u.repo.Create(ctx, img); err != nil {
		return nil, fmt.Errorf("create image: %w", err)
	}
	return &CreateImageResult{
		ImageID:    imageID,
		UploadURL:  uploadURL,
		ObjectKey:  objectKey,
		BucketName: in.Bucket,
		ExpiresIn:  int(u.expiry.Seconds()),
	}, nil
}
