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
	"github.com/tshop/backend/pkg/logger"
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
	SessionID    string `json:"sessionId,omitempty"` // Phiên làm việc, lưu Redis (RDB+AOF)
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
			logger.Error("auth_local_login_bad_request", map[string]interface{}{
				"service": "user-service",
				"action":  "login",
				"email":   req.Email,
				"error":   err.Error(),
			})
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		u, _ := repo.GetByUserName(ctx, req.Email)
		if u == nil || u.ID == "" || u.PasswordHash == nil || u.Salt == nil {
			logger.Info("auth_local_login_invalid_user", map[string]interface{}{
				"service": "user-service",
				"action":  "login",
				"email":   req.Email,
			})
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials"})
			return
		}

		if !verifyPassword(*u.PasswordHash, *u.Salt, req.Password) {
			logger.Info("auth_local_login_invalid_password", map[string]interface{}{
				"service": "user-service",
				"action":  "login",
				"user_id": u.ID,
				"email":   req.Email,
			})
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials"})
			return
		}

		// Chuẩn bị session & token_version
		sessionID := uuid.New().String()
		tokenVersion := 1
		if u.TokenVersion != nil && *u.TokenVersion > 0 {
			tokenVersion = *u.TokenVersion
		}

		// Tạo access token ngắn hạn
		accessTTL := 15 * time.Minute
		token, err := auth.NewToken(u.ID, req.Email, sessionID, tokenVersion, jwtSecret, accessTTL)
		if err != nil {
			logger.Error("auth_local_login_token_failed", map[string]interface{}{
				"service": "user-service",
				"action":  "login",
				"user_id": u.ID,
				"email":   req.Email,
				"error":   err.Error(),
			})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "token_failed"})
			return
		}

		// Tạo refresh token (7 ngày)
		refreshTTL := 7 * 24 * time.Hour
		refreshToken, err := generateRefreshToken()
		if err != nil {
			logger.Error("auth_local_login_refresh_failed", map[string]interface{}{
				"service": "user-service",
				"action":  "login",
				"user_id": u.ID,
				"email":   req.Email,
				"error":   err.Error(),
			})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "refresh_failed"})
			return
		}

		now := time.Now()
		u.AccessToken = &token
		u.RefreshToken = &refreshToken
		u.ExpiresAt = ptrTime(now.Add(refreshTTL))
		u.UpdatedAt = &now
		if err := repo.Update(ctx, u); err != nil {
			logger.Error("auth_local_login_update_failed", map[string]interface{}{
				"service": "user-service",
				"action":  "login",
				"user_id": u.ID,
				"email":   req.Email,
				"error":   err.Error(),
			})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "update_failed"})
			return
		}

		fullName := ""
		if u.FullName != nil {
			fullName = *u.FullName
		}

		tokens := authTokensResponse{
			AccessToken:  token,
			RefreshToken: refreshToken,
			SessionID:    sessionID,
			ExpiresIn:    int64(accessTTL.Seconds()),
		}
		if h.sessionStore != nil {
			if err := h.sessionStore.SetSession(ctx, sessionID, u.ID); err != nil {
				logger.Error("auth_local_login_session_set_failed", map[string]interface{}{
					"service":    "user-service",
					"action":     "login",
					"user_id":    u.ID,
					"session_id": sessionID,
					"error":      err.Error(),
				})
			}
		}

		logger.Info("auth_local_login_success", map[string]interface{}{
			"service":    "user-service",
			"action":     "login",
			"user_id":    u.ID,
			"email":      req.Email,
			"session_id": sessionID,
		})

		c.JSON(http.StatusOK, authResponse{
			User:   authUserResponse{ID: u.ID, Email: req.Email, FullName: fullName},
			Tokens: tokens,
		})
	}
}

// LocalRegister xử lý đăng ký bằng email/password.
// POST /api/auth/register body: { "email": "...", "password": "...", "fullName": "..." }
func (h *UserHandler) LocalRegister(jwtSecret string, repo domain.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req registerRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Error("auth_local_register_bad_request", map[string]interface{}{
				"service": "user-service",
				"action":  "register",
				"email":   req.Email,
				"error":   err.Error(),
			})
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Kiểm tra đã tồn tại email chưa
		existing, _ := repo.GetByUserName(ctx, req.Email)
		if existing != nil && existing.ID != "" {
			logger.Info("auth_local_register_email_exists", map[string]interface{}{
				"service": "user-service",
				"action":  "register",
				"email":   req.Email,
			})
			c.JSON(http.StatusBadRequest, gin.H{"error": "email_exists"})
			return
		}

		hash, salt, err := hashPassword(req.Password)
		if err != nil {
			logger.Error("auth_local_register_hash_failed", map[string]interface{}{
				"service": "user-service",
				"action":  "register",
				"email":   req.Email,
				"error":   err.Error(),
			})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "hash_failed"})
			return
		}

		now := time.Now()
		verified := true
		passwordHash := hash
		saltStr := salt
		userName := req.Email
		fullName := req.FullName
		tokenVersion := 1

		// Chuẩn bị session & refresh token cho user mới
		sessionID := uuid.New().String()
		accessTTL := 15 * time.Minute
		refreshTTL := 7 * 24 * time.Hour

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

		token, err := auth.NewToken(u.ID, req.Email, sessionID, tokenVersion, jwtSecret, accessTTL)
		if err != nil {
			logger.Error("auth_local_register_token_failed", map[string]interface{}{
				"service": "user-service",
				"action":  "register",
				"user_id": u.ID,
				"email":   req.Email,
				"error":   err.Error(),
			})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "token_failed"})
			return
		}
		u.AccessToken = &token

		refreshToken, err := generateRefreshToken()
		if err != nil {
			logger.Error("auth_local_register_refresh_failed", map[string]interface{}{
				"service": "user-service",
				"action":  "register",
				"user_id": u.ID,
				"email":   req.Email,
				"error":   err.Error(),
			})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "refresh_failed"})
			return
		}
		u.RefreshToken = &refreshToken
		u.ExpiresAt = ptrTime(now.Add(refreshTTL))

		if err := repo.Create(ctx, u); err != nil {
			logger.Error("auth_local_register_create_failed", map[string]interface{}{
				"service": "user-service",
				"action":  "register",
				"user_id": u.ID,
				"email":   req.Email,
				"error":   err.Error(),
			})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "create_failed"})
			return
		}

		tokens := authTokensResponse{
			AccessToken:  token,
			RefreshToken: refreshToken,
			SessionID:    sessionID,
			ExpiresIn:    int64(accessTTL.Seconds()),
		}
		if h.sessionStore != nil {
			if err := h.sessionStore.SetSession(ctx, sessionID, u.ID); err != nil {
				logger.Error("auth_local_register_session_set_failed", map[string]interface{}{
					"service":    "user-service",
					"action":     "register",
					"user_id":    u.ID,
					"session_id": sessionID,
					"error":      err.Error(),
				})
			}
		}

		logger.Info("auth_local_register_success", map[string]interface{}{
			"service":    "user-service",
			"action":     "register",
			"user_id":    u.ID,
			"email":      req.Email,
			"session_id": sessionID,
		})

		c.JSON(http.StatusOK, authResponse{
			User:   authUserResponse{ID: u.ID, Email: req.Email, FullName: req.FullName},
			Tokens: tokens,
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

func generateRefreshToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
