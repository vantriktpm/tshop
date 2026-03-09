package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tshop/backend/services/image-service/internal/domain"
	"github.com/tshop/backend/services/image-service/internal/usecase"
)

type ImageHandler struct {
	createImage    *usecase.CreateImage
	getDownloadURL *usecase.GetDownloadURL
	getImage       *usecase.GetImage
	syncAvatar     *usecase.SyncUserAvatar
	avatarNotifier domain.AvatarSavedNotifier
}

func NewImageHandler(
	createImage *usecase.CreateImage,
	getDownloadURL *usecase.GetDownloadURL,
	getImage *usecase.GetImage,
	syncAvatar *usecase.SyncUserAvatar,
	avatarNotifier domain.AvatarSavedNotifier,
) *ImageHandler {
	return &ImageHandler{
		createImage:    createImage,
		getDownloadURL: getDownloadURL,
		getImage:       getImage,
		syncAvatar:     syncAvatar,
		avatarNotifier: avatarNotifier,
	}
}

// CreateImageRequest body for requesting upload URL (client will PUT to returned upload_url).
type CreateImageRequest struct {
	Bucket      string  `json:"bucket"`     // product-images | user-avatars | order-invoices
	ObjectKey   string  `json:"object_key"` // optional
	ContentType *string `json:"content_type"`
	Size        *int64  `json:"size"`
	CreatedBy   *string `json:"created_by"`
}

func (h *ImageHandler) CreateImage(c *gin.Context) {
	var req CreateImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Bucket == "" {
		req.Bucket = domain.BucketProductImages
	}
	res, err := h.createImage.Execute(c.Request.Context(), usecase.CreateImageInput{
		Bucket:      req.Bucket,
		ObjectKey:   req.ObjectKey,
		ContentType: req.ContentType,
		Size:        req.Size,
		CreatedBy:   req.CreatedBy,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, res)
}

func (h *ImageHandler) GetDownloadURL(c *gin.Context) {
	imageID := c.Param("id")
	if imageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id required"})
		return
	}
	res, err := h.getDownloadURL.Execute(c.Request.Context(), imageID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *ImageHandler) GetImage(c *gin.Context) {
	imageID := c.Param("id")
	if imageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id required"})
		return
	}
	img, err := h.getImage.Execute(c.Request.Context(), imageID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, img)
}

// SyncAvatarRequest body for POST /api/images/sync-avatar (worker-sync-avatar or direct call).
type SyncAvatarRequest struct {
	UserID     string `json:"user_id" binding:"required"`
	PictureURL string `json:"picture_url" binding:"required"`
}

// SyncAvatar downloads avatar from picture_url, saves to MinIO + DB, then notifies via Redis (WebSocket).
func (h *ImageHandler) SyncAvatar(c *gin.Context) {
	var req SyncAvatarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	imageID, err := h.syncAvatar.Execute(c.Request.Context(), req.UserID, req.PictureURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if imageID != "" && h.avatarNotifier != nil {
		_ = h.avatarNotifier.NotifyAvatarSaved(c.Request.Context(), req.UserID, imageID)
	}
	c.JSON(http.StatusOK, gin.H{"image_id": imageID})
}
