package rest

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tshop/backend/pkg/auth"
	"github.com/tshop/backend/pkg/events"
	"github.com/tshop/backend/services/user-service/internal/core/config"
	"github.com/tshop/backend/services/user-service/internal/domain"
	"golang.org/x/oauth2"
)

// GoogleStart redirect user tới Google OAuth. Frontend gọi GET /api/auth/google/start?state=<frontend_url>
func (h *UserHandler) GoogleStart(cfg config.GoogleOAuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		state := c.Query("state")
		if state == "" {
			state = cfg.FrontendRedirectURL
		}
		oauthCfg := &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURI(),
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://accounts.google.com/o/oauth2/auth",
				TokenURL: "https://oauth2.googleapis.com/token",
			},
			Scopes: []string{"openid", "email", "profile"},
		}
		url := oauthCfg.AuthCodeURL(state)
		c.Redirect(http.StatusFound, url)
	}
}

// GoogleCallback xử lý redirect từ Google (đăng nhập / đăng ký).
// Google Cloud Console → APIs & Services → Credentials → OAuth 2.0 Client IDs →
// Authorized redirect URIs: http://localhost:8080/api/auth/google/callback
func (h *UserHandler) GoogleCallback(cfg config.GoogleOAuthConfig, repo domain.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Query("code")
		if code == "" {
			c.Redirect(http.StatusFound, redirectWithError(cfg.FrontendRedirectURL, "missing_code"))
			return
		}
		state := c.Query("state")

		oauthCfg := &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURI(),
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://accounts.google.com/o/oauth2/auth",
				TokenURL: "https://oauth2.googleapis.com/token",
			},
			Scopes: []string{"openid", "email", "profile"},
		}
		tok, err := oauthCfg.Exchange(c.Request.Context(), code)
		if err != nil {
			log.Printf("google exchange: %v", err)
			c.Redirect(http.StatusFound, redirectWithError(cfg.FrontendRedirectURL, "exchange_failed"))
			return
		}
		userInfo, err := fetchGoogleUser(c.Request.Context(), tok.AccessToken)
		if err != nil {
			log.Printf("google user: %v", err)
			c.Redirect(http.StatusFound, redirectWithError(cfg.FrontendRedirectURL, "user_failed"))
			return
		}

		ctx := c.Request.Context()
		provider := "google"
		existing, _ := repo.GetByProviderAndProviderUserID(ctx, provider, userInfo.ID)
		var user *domain.User
		if existing != nil {
			user = existing
			if userInfo.Name != nil {
				user.FullName = userInfo.Name
			}
			user.AccessToken = &tok.AccessToken
			if !tok.Expiry.IsZero() {
				user.ExpiresAt = &tok.Expiry
			}
		} else {
			now := time.Now()
			verified := true
			user = &domain.User{
				ID:             uuid.New().String(),
				FullName:       userInfo.Name,
				UserName:       userInfo.Email,
				Provider:       &provider,
				ProviderUserID: &userInfo.ID,
				AccessToken:    &tok.AccessToken,
				IsVerified:     &verified,
				CreatedAt:      &now,
				UpdatedAt:      &now,
			}
			if !tok.Expiry.IsZero() {
				user.ExpiresAt = &tok.Expiry
			}
			if err := repo.Create(ctx, user); err != nil {
				log.Printf("user create: %v", err)
				c.Redirect(http.StatusFound, redirectWithError(cfg.FrontendRedirectURL, "create_failed"))
				return
			}
		}

		email := ""
		if user.UserName != nil {
			email = *user.UserName
		}
		jwtToken, err := auth.NewToken(user.ID, email, cfg.JWTSecret, 24*time.Hour)
		if err != nil {
			c.Redirect(http.StatusFound, redirectWithError(cfg.FrontendRedirectURL, "token_failed"))
			return
		}
		redirectURL := cfg.FrontendRedirectURL
		if state != "" {
			redirectURL = state
		}
		if strings.Contains(redirectURL, "?") {
			redirectURL += "&token=" + jwtToken
		} else {
			redirectURL += "?token=" + jwtToken
		}
		c.Redirect(http.StatusFound, redirectURL)
	}
}

type googleUser struct {
	ID            string  `json:"id"`
	Email         *string `json:"email"`
	Name          *string `json:"name"`
	Picture       string  `json:"picture"`
	VerifiedEmail bool    `json:"verified_email"`
}

// GoogleVerifyRequest body gửi từ frontend sau khi Google Sign-In trả về credential (id_token).
type GoogleVerifyRequest struct {
	Credential string `json:"credential" binding:"required"`
}

// GoogleVerifyResponse trả về access_token cho frontend.
type GoogleVerifyResponse struct {
	AccessToken string `json:"access_token"`
	User        struct {
		ID       string `json:"id"`
		Email    string `json:"email"`
		FullName string `json:"full_name"`
	} `json:"user"`
}

// GoogleVerify nhận Google id_token từ frontend, verify với Google, tạo/cập nhật user, trả về JWT.
// POST /api/auth/google/verify body: { "credential": "<id_token>" }
func (h *UserHandler) GoogleVerify(cfg config.GoogleOAuthConfig, repo domain.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req GoogleVerifyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "credential required"})
			return
		}
		// Verify id_token với Google tokeninfo
		userInfo, err := verifyGoogleIDToken(c.Request.Context(), req.Credential, cfg.ClientID)
		if err != nil {
			log.Printf("google verify: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credential"})
			return
		}
		ctx := c.Request.Context()
		provider := "google"
		existing, _ := repo.GetByProviderAndProviderUserID(ctx, provider, userInfo.ID)
		var user *domain.User
		if existing != nil && existing.ID != "" {
			user = existing
			if userInfo.Name != nil {
				user.FullName = userInfo.Name
			}
			if userInfo.Email != nil {
				user.UserName = userInfo.Email
			}
		} else {
			now := time.Now()
			verified := true
			user = &domain.User{
				ID:             uuid.New().String(),
				FullName:       userInfo.Name,
				UserName:       userInfo.Email,
				Provider:       &provider,
				ProviderUserID: &userInfo.ID,
				IsVerified:     &verified,
				CreatedAt:      &now,
				UpdatedAt:      &now,
			}
		}
		email := ""
		if user.UserName != nil {
			email = *user.UserName
		}
		jwtToken, err := auth.NewToken(user.ID, email, cfg.JWTSecret, 24*time.Hour)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "token_failed"})
			return
		}
		// Lưu JWT vào cột access_token trong bảng users
		user.AccessToken = &jwtToken
		now := time.Now()
		user.UpdatedAt = &now
		if existing != nil && existing.ID != "" {
			if err := repo.Update(ctx, user); err != nil {
				log.Printf("user update: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "update_failed"})
				return
			}
		} else {
			if err := repo.Create(ctx, user); err != nil {
				log.Printf("user create: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "create_failed"})
				return
			}
		}
		// Gửi event Kafka để image-service đọc và tải avatar từ URL (user.avatar.sync)
		if userInfo != nil && userInfo.Picture != "" && h.eventPub != nil {
			evt := events.UserAvatarSyncEvent{
				UserID:     user.ID,
				PictureURL: userInfo.Picture,
			}
			if payload, err := json.Marshal(&evt); err == nil {
				if err := h.eventPub.Publish(c.Request.Context(), events.TopicUserAvatarSync, user.ID, payload); err != nil {
					log.Printf("avatar event publish: %v", err)
				}
			} else {
				log.Printf("avatar event marshal: %v", err)
			}
		}

		fullName := ""
		if user.FullName != nil {
			fullName = *user.FullName
		}
		c.JSON(http.StatusOK, GoogleVerifyResponse{
			AccessToken: jwtToken,
			User: struct {
				ID       string `json:"id"`
				Email    string `json:"email"`
				FullName string `json:"full_name"`
			}{user.ID, email, fullName},
		})
	}
}

// verifyGoogleIDToken gọi Google tokeninfo để xác thực id_token và lấy thông tin user.
func verifyGoogleIDToken(ctx context.Context, idToken, clientID string) (*googleUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://oauth2.googleapis.com/tokeninfo?id_token="+idToken, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("tokeninfo failed")
	}
	var payload struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		EmailVerified string `json:"email_verified"`
		Aud           string `json:"aud"` // client_id
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if payload.Aud != clientID {
		return nil, errors.New("audience mismatch")
	}
	var name, email *string
	if payload.Name != "" {
		name = &payload.Name
	}
	if payload.Email != "" {
		email = &payload.Email
	}
	return &googleUser{
		ID:            payload.Sub,
		Email:         email,
		Name:          name,
		Picture:       payload.Picture,
		VerifiedEmail: payload.EmailVerified == "true",
	}, nil
}

func fetchGoogleUser(ctx context.Context, accessToken string) (*googleUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var u googleUser
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, err
	}
	return &u, nil
}
