package domain

import "time"

// Image metadata. DB stores image_id, object_key, bucket_name (no hardcoded URL).
// Presigned URLs are generated on demand.
type Image struct {
	ID          string     // uuid NOT NULL
	ObjectKey   string     // character(255)
	BucketName  string     // character(255)
	ContentType *string    // character(50), nullable
	Size        *int64     // bigint, nullable
	CreatedBy   *string    // character(50), nullable
	UpdatedBy   *string    // character(50), nullable
	CreatedAt   *time.Time // timestamp with time zone, nullable
	UpdatedAt   *time.Time // timestamp with time zone, nullable
}

// Bucket names (separate buckets)
const (
	BucketProductImages = "product-images"
	BucketUserAvatars   = "user-avatars"
	BucketOrderInvoices = "order-invoices"
)
