package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tshop/backend/services/image-service/internal/domain"
	"github.com/tshop/backend/services/image-service/internal/usecase"
)

type ImageHandler struct {
	createImage     *usecase.CreateImage
	getDownloadURL  *usecase.GetDownloadURL
	getImage        *usecase.GetImage
}

func NewImageHandler(
	createImage *usecase.CreateImage,
	getDownloadURL *usecase.GetDownloadURL,
	getImage *usecase.GetImage,
) *ImageHandler {
	return &ImageHandler{
		createImage:    createImage,
		getDownloadURL: getDownloadURL,
		getImage:       getImage,
	}
}

// CreateImageRequest body for requesting upload URL (client will PUT to returned upload_url).
type CreateImageRequest struct {
	Bucket      string  `json:"bucket"`       // product-images | user-avatars | order-invoices
	ObjectKey   string  `json:"object_key"`   // optional
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
