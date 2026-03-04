package minio

import (
	"context"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Storage provides presigned URLs. Client uploads/downloads directly to/from MinIO.
type Storage struct {
	client     *minio.Client
	presignExp time.Duration
}

type Config struct {
	Endpoint      string
	AccessKey     string
	SecretKey     string
	UseSSL        bool
	PresignExpiry time.Duration // e.g. 15 * time.Minute
}

func NewStorage(cfg Config) (*Storage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}
	if cfg.PresignExpiry <= 0 {
		cfg.PresignExpiry = 15 * time.Minute
	}
	return &Storage{client: client, presignExp: cfg.PresignExpiry}, nil
}

// EnsureBucket creates the bucket if it does not exist.
func (s *Storage) EnsureBucket(ctx context.Context, bucket string) error {
	exists, err := s.client.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return s.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
}

// PresignedPutURL returns a URL string for client to upload (PUT) directly to MinIO.
func (s *Storage) PresignedPutURL(ctx context.Context, bucket, objectKey string, expiry time.Duration) (string, error) {
	if expiry <= 0 {
		expiry = s.presignExp
	}
	u, err := s.client.PresignedPutObject(ctx, bucket, objectKey, expiry)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// PresignedGetURL returns a URL string for client to download (GET) directly from MinIO.
func (s *Storage) PresignedGetURL(ctx context.Context, bucket, objectKey string, expiry time.Duration) (string, error) {
	if expiry <= 0 {
		expiry = s.presignExp
	}
	u, err := s.client.PresignedGetObject(ctx, bucket, objectKey, expiry, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// PutObject uploads data directly to MinIO.
func (s *Storage) PutObject(ctx context.Context, bucket, objectKey, contentType string, body io.Reader, size int64) error {
	opts := minio.PutObjectOptions{ContentType: contentType}
	_, err := s.client.PutObject(ctx, bucket, objectKey, body, size, opts)
	return err
}
