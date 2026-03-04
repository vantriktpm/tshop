package avatar

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type createImageRequest struct {
	Bucket      string  `json:"bucket"`
	ObjectKey   string  `json:"object_key"`
	ContentType *string `json:"content_type,omitempty"`
	Size        *int64  `json:"size,omitempty"`
	CreatedBy   *string `json:"created_by,omitempty"`
}

type createImageResponse struct {
	ImageID    string `json:"image_id"`
	UploadURL  string `json:"upload_url"`
	ObjectKey  string `json:"object_key"`
	BucketName string `json:"bucket_name"`
	ExpiresIn  int    `json:"expires_in_seconds"`
}

// SyncGoogleAvatarFromURL downloads the avatar from Google and uploads it to image-service (MinIO + Postgres).
// It runs asynchronously and logs errors; login flow does not depend on its success.
func SyncGoogleAvatarFromURL(ctx context.Context, userID, pictureURL string) {
	if userID == "" || pictureURL == "" {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// 1) Download avatar from Google
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, pictureURL, nil)
		if err != nil {
			log.Printf("avatar: new request: %v", err)
			return
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("avatar: download: %v", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.Printf("avatar: download status %d", resp.StatusCode)
			return
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("avatar: read body: %v", err)
			return
		}
		if len(body) == 0 {
			log.Printf("avatar: empty body for user %s", userID)
			return
		}
		contentType := resp.Header.Get("Content-Type")
		size := int64(len(body))

		// 2) Ask image-service for upload URL
		imageSvcBase := os.Getenv("IMAGE_SERVICE_BASE_URL")
		if imageSvcBase == "" {
			imageSvcBase = "http://localhost:8085"
		}
		objectKey := "user-avatars/" + userID
		createdBy := userID
		reqPayload := createImageRequest{
			Bucket:      "user-avatars",
			ObjectKey:   objectKey,
			ContentType: &contentType,
			Size:        &size,
			CreatedBy:   &createdBy,
		}
		buf, err := json.Marshal(&reqPayload)
		if err != nil {
			log.Printf("avatar: marshal create image: %v", err)
			return
		}
		createReq, err := http.NewRequestWithContext(ctx, http.MethodPost, imageSvcBase+"/api/images", bytes.NewReader(buf))
		if err != nil {
			log.Printf("avatar: new create image request: %v", err)
			return
		}
		createReq.Header.Set("Content-Type", "application/json")
		createResp, err := http.DefaultClient.Do(createReq)
		if err != nil {
			log.Printf("avatar: call image-service: %v", err)
			return
		}
		defer createResp.Body.Close()
		if createResp.StatusCode != http.StatusCreated {
			b, _ := io.ReadAll(createResp.Body)
			log.Printf("avatar: image-service status %d, body=%s", createResp.StatusCode, strconv.Quote(string(b)))
			return
		}
		var createOut createImageResponse
		if err := json.NewDecoder(createResp.Body).Decode(&createOut); err != nil {
			log.Printf("avatar: decode create image response: %v", err)
			return
		}
		if createOut.UploadURL == "" {
			log.Printf("avatar: empty upload_url for user %s", userID)
			return
		}

		// 3) Upload avatar bytes to presigned URL
		putReq, err := http.NewRequestWithContext(ctx, http.MethodPut, createOut.UploadURL, bytes.NewReader(body))
		if err != nil {
			log.Printf("avatar: new PUT request: %v", err)
			return
		}
		if contentType != "" {
			putReq.Header.Set("Content-Type", contentType)
		}
		putResp, err := http.DefaultClient.Do(putReq)
		if err != nil {
			log.Printf("avatar: upload to presigned URL: %v", err)
			return
		}
		defer putResp.Body.Close()
		if putResp.StatusCode != http.StatusOK && putResp.StatusCode != http.StatusNoContent {
			log.Printf("avatar: upload status %d", putResp.StatusCode)
			return
		}

		log.Printf("avatar: synced google avatar for user %s as image %s", userID, createOut.ImageID)
	}()
}

