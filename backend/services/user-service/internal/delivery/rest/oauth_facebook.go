package rest

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tshop/backend/pkg/auth"
	"github.com/tshop/backend/services/user-service/internal/core/config"
	"github.com/tshop/backend/services/user-service/internal/domain"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
)

// FacebookStart redirect user tới Facebook OAuth. Frontend gọi GET /api/auth/facebook/start?state=<frontend_url>
func (h *UserHandler) FacebookStart(cfg config.FacebookOAuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		state := c.Query("state")
		if state == "" {
			state = cfg.FrontendRedirectURL
		}
		oauthCfg := &oauth2.Config{
			ClientID:     cfg.AppID,
			ClientSecret: cfg.AppSecret,
			RedirectURL:  cfg.RedirectURI(),
			Endpoint:     facebook.Endpoint,
			Scopes:       []string{"email", "public_profile"},
		}
		c.Redirect(http.StatusFound, oauthCfg.AuthCodeURL(state))
	}
}

// FacebookCallback xử lý redirect từ Facebook (Valid OAuth Redirect URI).
// Trong Facebook App → Facebook Login → Settings → Valid OAuth Redirect URIs:
//   thêm: http://localhost:8080/api/auth/facebook/callback
func (h *UserHandler) FacebookCallback(cfg config.FacebookOAuthConfig, repo domain.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Query("code")
		if code == "" {
			c.Redirect(http.StatusFound, redirectWithError(cfg.FrontendRedirectURL, "missing_code"))
			return
		}
		state := c.Query("state")

		oauthCfg := &oauth2.Config{
			ClientID:     cfg.AppID,
			ClientSecret: cfg.AppSecret,
			RedirectURL:  cfg.RedirectURI(),
			Endpoint:    facebook.Endpoint,
			Scopes:      []string{"email", "public_profile"},
		}
		tok, err := oauthCfg.Exchange(c.Request.Context(), code)
		if err != nil {
			log.Printf("facebook exchange: %v", err)
			c.Redirect(http.StatusFound, redirectWithError(cfg.FrontendRedirectURL, "exchange_failed"))
			return
		}
		// Lấy thông tin user từ Graph API
		userInfo, err := fetchFacebookUser(c.Request.Context(), tok.AccessToken)
		if err != nil {
			log.Printf("facebook user: %v", err)
			c.Redirect(http.StatusFound, redirectWithError(cfg.FrontendRedirectURL, "user_failed"))
			return
		}

		ctx := c.Request.Context()
		provider := "facebook"
		existing, _ := repo.GetByProviderAndProviderUserID(ctx, provider, userInfo.ID)
		var user *domain.User
		if existing != nil {
			user = existing
			// Có thể cập nhật access_token, full_name
			if userInfo.Name != nil {
				user.FullName = userInfo.Name
			}
			user.AccessToken = &tok.AccessToken
			if tok.Expiry.IsZero() == false {
				user.ExpiresAt = &tok.Expiry
			}
			// TODO: repo.Update nếu có
		} else {
			now := time.Now()
			verified := true
			user = &domain.User{
				ID:             uuid.New().String(),
				FullName:       userInfo.Name,
				UserName:       userInfo.Email, // email từ Facebook (hoặc id nếu không có email)
				Provider:       &provider,
				ProviderUserID: &userInfo.ID,
				AccessToken:    &tok.AccessToken,
				IsVerified:     &verified,
				CreatedAt:      &now,
				UpdatedAt:      &now,
			}
			if tok.Expiry.IsZero() == false {
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
		// Redirect về frontend kèm token (frontend lưu vào storage)
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

func redirectWithError(base, errCode string) string {
	return base + "?error=" + errCode
}

type facebookUser struct {
	ID    string  `json:"id"`
	Name  *string `json:"name"`
	Email *string `json:"email"`
}

func fetchFacebookUser(ctx context.Context, accessToken string) (*facebookUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://graph.facebook.com/me?fields=id,name,email&access_token="+accessToken, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var u facebookUser
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, err
	}
	return &u, nil
}
