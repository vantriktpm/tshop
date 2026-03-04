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
)

// XStart redirect user tới X (Twitter) OAuth. Frontend gọi GET /api/auth/x/start?state=<frontend_url>
func (h *UserHandler) XStart(cfg config.XOAuthConfig) gin.HandlerFunc {
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
				AuthURL:   "https://x.com/i/oauth2/authorize",
				TokenURL:  "https://api.twitter.com/2/oauth2/token",
				AuthStyle: oauth2.AuthStyleInHeader,
			},
			Scopes: []string{"tweet.read", "users.read", "offline.access"},
		}
		c.Redirect(http.StatusFound, oauthCfg.AuthCodeURL(state))
	}
}

// XCallback xử lý redirect từ X (Twitter) - đăng nhập / đăng ký.
// X Developer Portal → App → User authentication settings → Callback URI / Redirect URI:
//   thêm: http://localhost:8080/api/auth/x/callback
func (h *UserHandler) XCallback(cfg config.XOAuthConfig, repo domain.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Query("code")
		if code == "" {
			c.Redirect(http.StatusFound, redirectWithError(cfg.FrontendRedirectURL, "missing_code"))
			return
		}
		state := c.Query("state")

		// X (Twitter) OAuth 2.0: token endpoint yêu cầu Basic auth (Client ID + Secret)
		oauthCfg := &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURI(),
			Endpoint: oauth2.Endpoint{
				AuthURL:   "https://x.com/i/oauth2/authorize",
				TokenURL:  "https://api.twitter.com/2/oauth2/token",
				AuthStyle: oauth2.AuthStyleInHeader, // X yêu cầu Basic auth cho token exchange
			},
			Scopes: []string{"tweet.read", "users.read", "offline.access"},
		}
		tok, err := oauthCfg.Exchange(c.Request.Context(), code)
		if err != nil {
			log.Printf("x exchange: %v", err)
			c.Redirect(http.StatusFound, redirectWithError(cfg.FrontendRedirectURL, "exchange_failed"))
			return
		}
		userInfo, err := fetchXUser(c.Request.Context(), tok.AccessToken)
		if err != nil {
			log.Printf("x user: %v", err)
			c.Redirect(http.StatusFound, redirectWithError(cfg.FrontendRedirectURL, "user_failed"))
			return
		}

		ctx := c.Request.Context()
		provider := "x"
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
			// X thường không trả email qua /users/me; dùng username làm identifier
			userName := userInfo.Username
			if userName == nil {
				userName = &userInfo.ID
			}
			user = &domain.User{
				ID:             uuid.New().String(),
				FullName:       userInfo.Name,
				UserName:       userName,
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

// X API v2 GET /2/users/me response
type xUserResponse struct {
	Data xUser `json:"data"`
}

type xUser struct {
	ID       string  `json:"id"`
	Name     *string `json:"name"`
	Username *string `json:"username"`
}

func fetchXUser(ctx context.Context, accessToken string) (*xUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://api.twitter.com/2/users/me?user.fields=username,name", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out xUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if out.Data.Username == nil {
		out.Data.Username = &out.Data.ID
	}
	return &out.Data, nil
}
