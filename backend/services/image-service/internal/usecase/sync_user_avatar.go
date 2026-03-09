package usecase

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/tshop/backend/services/image-service/internal/domain"
)

// ObjectUploader abstracts direct upload operations to object storage.
type ObjectUploader interface {
	EnsureBucket(ctx context.Context, bucket string) error
	PutObject(ctx context.Context, bucket, objectKey, contentType string, body io.Reader, size int64) error
}

// SyncUserAvatar downloads a remote avatar (e.g. Google) and stores it in MinIO + Postgres.
type SyncUserAvatar struct {
	repo    domain.ImageRepository
	storage ObjectUploader
}

func NewSyncUserAvatar(repo domain.ImageRepository, storage ObjectUploader) *SyncUserAvatar {
	return &SyncUserAvatar{repo: repo, storage: storage}
}

// Execute downloads pictureURL and saves it as user avatar for userID.
// Returns the created image ID for pushing via WebSocket (avatar.saved).
func (u *SyncUserAvatar) Execute(ctx context.Context, userID, pictureURL string) (imageID string, err error) {
	if userID == "" || pictureURL == "" {
		return "", nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pictureURL, nil)
	if err != nil {
		return "", fmt.Errorf("avatar: new request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("avatar: download: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("avatar: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("avatar: read body: %w", err)
	}
	if len(body) == 0 {
		return "", fmt.Errorf("avatar: empty body for user %s", userID)
	}
	contentType := resp.Header.Get("Content-Type")
	size := int64(len(body))

	bucket := domain.BucketUserAvatars
	objectKey := "user-avatars/" + userID

	if err := u.storage.EnsureBucket(ctx, bucket); err != nil {
		return "", fmt.Errorf("avatar: ensure bucket: %w", err)
	}
	if err := u.storage.PutObject(ctx, bucket, objectKey, contentType, bytes.NewReader(body), size); err != nil {
		return "", fmt.Errorf("avatar: put object: %w", err)
	}

	now := time.Now()
	imgID := uuid.NewString()
	img := &domain.Image{
		ID:          imgID,
		ObjectKey:   objectKey,
		BucketName:  bucket,
		ContentType: &contentType,
		Size:        &size,
		CreatedBy:   &userID,
		CreatedAt:   &now,
	}
	if err := u.repo.Create(ctx, img); err != nil {
		return "", fmt.Errorf("avatar: create image: %w", err)
	}
	return imgID, nil
}
