package rest

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tshop/backend/pkg/auth"
	"github.com/tshop/backend/services/user-service/internal/domain"
)

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type registerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"fullName" binding:"required"`
}

type authUserResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"fullName"`
}

type authTokensResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken,omitempty"`
	ExpiresIn    int64  `json:"expiresIn,omitempty"`
}

type authResponse struct {
	User   authUserResponse   `json:"user"`
	Tokens authTokensResponse `json:"tokens"`
}

// LocalLogin xử lý đăng nhập bằng email/password (không qua OAuth).
// POST /api/auth/login body: { "email": "...", "password": "..." }
func (h *UserHandler) LocalLogin(jwtSecret string, repo domain.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req loginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		u, _ := repo.GetByUserName(ctx, req.Email)
		if u == nil || u.ID == "" || u.PasswordHash == nil || u.Salt == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials"})
			return
		}

		if !verifyPassword(*u.PasswordHash, *u.Salt, req.Password) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials"})
			return
		}

		// Tạo JWT
		ttl := 24 * time.Hour
		token, err := auth.NewToken(u.ID, req.Email, jwtSecret, ttl)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "token_failed"})
			return
		}

		now := time.Now()
		u.AccessToken = &token
		u.UpdatedAt = &now
		if err := repo.Update(ctx, u); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "update_failed"})
			return
		}

		fullName := ""
		if u.FullName != nil {
			fullName = *u.FullName
		}

		c.JSON(http.StatusOK, authResponse{
			User: authUserResponse{
				ID:       u.ID,
				Email:    req.Email,
				FullName: fullName,
			},
			Tokens: authTokensResponse{
				AccessToken: token,
				ExpiresIn:   int64(ttl.Seconds()),
			},
		})
	}
}

// LocalRegister xử lý đăng ký bằng email/password.
// POST /api/auth/register body: { "email": "...", "password": "...", "fullName": "..." }
func (h *UserHandler) LocalRegister(jwtSecret string, repo domain.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req registerRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Kiểm tra đã tồn tại email chưa
		existing, _ := repo.GetByUserName(ctx, req.Email)
		if existing != nil && existing.ID != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "email_exists"})
			return
		}

		hash, salt, err := hashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "hash_failed"})
			return
		}

		now := time.Now()
		verified := true
		passwordHash := hash
		saltStr := salt
		userName := req.Email
		fullName := req.FullName

		u := &domain.User{
			ID:           uuid.New().String(),
			UserName:     &userName,
			FullName:     &fullName,
			PasswordHash: &passwordHash,
			Salt:         &saltStr,
			IsVerified:   &verified,
			CreatedAt:    &now,
			UpdatedAt:    &now,
		}

		// Tạo JWT cho user mới
		ttl := 24 * time.Hour
		token, err := auth.NewToken(u.ID, req.Email, jwtSecret, ttl)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "token_failed"})
			return
		}
		u.AccessToken = &token

		if err := repo.Create(ctx, u); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "create_failed"})
			return
		}

		c.JSON(http.StatusOK, authResponse{
			User: authUserResponse{
				ID:       u.ID,
				Email:    req.Email,
				FullName: req.FullName,
			},
			Tokens: authTokensResponse{
				AccessToken: token,
				ExpiresIn:   int64(ttl.Seconds()),
			},
		})
	}
}

func hashPassword(password string) (hash string, salt string, err error) {
	buf := make([]byte, 16)
	if _, err = rand.Read(buf); err != nil {
		return "", "", err
	}
	salt = strings.TrimSpace(hex.EncodeToString(buf))
	h := sha256.Sum256([]byte(salt + password))
	hash = hex.EncodeToString(h[:])
	return hash, salt, nil
}

func verifyPassword(storedHash, salt, password string) bool {
	storedHash = strings.TrimSpace(storedHash)
	salt = strings.TrimSpace(salt)
	h := sha256.Sum256([]byte(salt + password))
	return storedHash == hex.EncodeToString(h[:])
}
