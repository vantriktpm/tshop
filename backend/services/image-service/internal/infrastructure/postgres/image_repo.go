package postgres

import (
	"context"
	"time"

	"github.com/tshop/backend/services/image-service/internal/domain"
	"gorm.io/gorm"
)

type ImageModel struct {
	ID         string     `gorm:"column:id;primaryKey;type:uuid"`
	OwnerID    *string    `gorm:"column:owner_id;type:uuid"` // product_id hoặc user_id
	OwnerType  *string    `gorm:"column:owner_type;size:50"` // product, user, banner...
	FileName   *string    `gorm:"column:file_name;size:255"`
	ObjectKey  string     `gorm:"column:object_key;size:500"` // path trong minio
	BucketName string     `gorm:"column:bucket_name;size:100"`
	MimeType   *string    `gorm:"column:mime_type;size:100"`
	Size       *int64     `gorm:"column:size"`
	IsPrimary  *bool      `gorm:"column:is_primary"`
	SortOrder  *int       `gorm:"column:sort_order"`
	Status     *string    `gorm:"column:status;size:20"`
	CreatedAt  *time.Time `gorm:"column:created_at"`
	CreatedBy  *string    `gorm:"column:created_by;type:uuid"`
	UpdatedAt  *time.Time `gorm:"column:updated_at"`
	UpdatedBy  *string    `gorm:"column:updated_by;type:uuid"`
}

func (ImageModel) TableName() string { return "images" }

type ImageRepository struct{ db *gorm.DB }

func NewImageRepository(db *gorm.DB) *ImageRepository {
	return &ImageRepository{db: db}
}

func (r *ImageRepository) Create(ctx context.Context, img *domain.Image) error {
	m := imageToModel(img)
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *ImageRepository) GetByID(ctx context.Context, id string) (*domain.Image, error) {
	var m ImageModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return modelToImage(&m), nil
}

func imageToModel(img *domain.Image) *ImageModel {
	return &ImageModel{
		ID:         img.ID,
		ObjectKey:  img.ObjectKey,
		BucketName: img.BucketName,
		MimeType:   img.ContentType,
		Size:       img.Size,
		CreatedBy:  img.CreatedBy,
		UpdatedBy:  img.UpdatedBy,
		CreatedAt:  img.CreatedAt,
		UpdatedAt:  img.UpdatedAt,
	}
}

func modelToImage(m *ImageModel) *domain.Image {
	return &domain.Image{
		ID:          m.ID,
		ObjectKey:   m.ObjectKey,
		BucketName:  m.BucketName,
		ContentType: m.MimeType,
		Size:        m.Size,
		CreatedBy:   m.CreatedBy,
		UpdatedBy:   m.UpdatedBy,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
