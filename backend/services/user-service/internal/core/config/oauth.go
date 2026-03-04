package config

import "os"

// GoogleClientID returns Google OAuth client ID from config or env.
// Priority: user.config (GOOGLE_CLIENT_ID=...) > env GOOGLE_CLIENT_ID.
func GoogleClientID() string {
	if v, ok := loadFromUserConfig("GOOGLE_CLIENT_ID"); ok && v != "" {
		return v
	}
	return os.Getenv("GOOGLE_CLIENT_ID")
}

// FacebookAppID returns Facebook App ID from config or env.
// Priority: user.config (FACEBOOK_APP_ID=...) > env FACEBOOK_APP_ID.
func FacebookAppID() string {
	if v, ok := loadFromUserConfig("FACEBOOK_APP_ID"); ok && v != "" {
		return v
	}
	return os.Getenv("FACEBOOK_APP_ID")
}

// XClientID returns X (Twitter) OAuth client ID from config or env.
// Priority: user.config (X_CLIENT_ID=...) > env X_CLIENT_ID.
func XClientID() string {
	if v, ok := loadFromUserConfig("X_CLIENT_ID"); ok && v != "" {
		return v
	}
	return os.Getenv("X_CLIENT_ID")
}

// FacebookAppSecret returns Facebook App Secret from config or env.
// Priority: user.config (FACEBOOK_APP_SECRET=...) > env FACEBOOK_APP_SECRET.
func FacebookAppSecret() string {
	if v, ok := loadFromUserConfig("FACEBOOK_APP_SECRET"); ok && v != "" {
		return v
	}
	return os.Getenv("FACEBOOK_APP_SECRET")
}

// GoogleClientSecret returns Google OAuth client secret from config or env.
// Priority: user.config (GOOGLE_CLIENT_SECRET=...) > env GOOGLE_CLIENT_SECRET.
func GoogleClientSecret() string {
	if v, ok := loadFromUserConfig("GOOGLE_CLIENT_SECRET"); ok && v != "" {
		return v
	}
	return os.Getenv("GOOGLE_CLIENT_SECRET")
}

// XClientSecret returns X (Twitter) client secret from config or env.
// Priority: user.config (X_CLIENT_SECRET=...) > env X_CLIENT_SECRET.
func XClientSecret() string {
	if v, ok := loadFromUserConfig("X_CLIENT_SECRET"); ok && v != "" {
		return v
	}
	return os.Getenv("X_CLIENT_SECRET")
}

// CallbackBaseURL returns base URL for OAuth callbacks.
// Priority: user.config (CALLBACK_BASE_URL=...) > env CALLBACK_BASE_URL > default http://localhost:8080.
func CallbackBaseURL() string {
	if v, ok := loadFromUserConfig("CALLBACK_BASE_URL"); ok && v != "" {
		return v
	}
	if v := os.Getenv("CALLBACK_BASE_URL"); v != "" {
		return v
	}
	return "http://localhost:8080"
}

// FrontendRedirectURL returns frontend base URL for redirects.
// Priority: user.config (FRONTEND_REDIRECT_URL=...) > env FRONTEND_REDIRECT_URL > default http://localhost:3000.
func FrontendRedirectURL() string {
	if v, ok := loadFromUserConfig("FRONTEND_REDIRECT_URL"); ok && v != "" {
		return v
	}
	if v := os.Getenv("FRONTEND_REDIRECT_URL"); v != "" {
		return v
	}
	return "http://localhost:3000"
}

// JWTSecret returns JWT signing secret.
// Priority: user.config (JWT_SECRET=...) > env JWT_SECRET > default "your-jwt-secret".
func JWTSecret() string {
	if v, ok := loadFromUserConfig("JWT_SECRET"); ok && v != "" {
		return v
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		return v
	}
	return "your-jwt-secret"
}

// OAuth callback paths (phải khớp với redirect URI đăng ký tại từng provider).
// Facebook: http://localhost:8080/api/auth/facebook/callback
// Google:   http://localhost:8080/api/auth/google/callback (Google Cloud Console → Credentials → Authorized redirect URIs)
// X (Twitter): http://localhost:8080/api/auth/x/callback (X Developer Portal → App → User authentication settings → Callback URI)
const (
	FacebookCallbackPath = "/api/auth/facebook/callback"
	GoogleCallbackPath   = "/api/auth/google/callback"
	XCallbackPath        = "/api/auth/x/callback"
	// Start paths: frontend redirect user tới đây, backend redirect tiếp tới provider
	GoogleStartPath   = "/api/auth/google/start"
	FacebookStartPath = "/api/auth/facebook/start"
	XStartPath        = "/api/auth/x/start"
)

// FacebookOAuthConfig cấu hình Facebook Login (lấy từ env).
type FacebookOAuthConfig struct {
	AppID               string
	AppSecret           string
	CallbackBaseURL     string
	FrontendRedirectURL string
	JWTSecret           string
}

func (c *FacebookOAuthConfig) RedirectURI() string {
	return c.CallbackBaseURL + FacebookCallbackPath
}

// GoogleOAuthConfig cấu hình Google Sign-In (đăng nhập / đăng ký).
type GoogleOAuthConfig struct {
	ClientID            string
	ClientSecret        string
	CallbackBaseURL     string
	FrontendRedirectURL string
	JWTSecret           string
}

func (c *GoogleOAuthConfig) RedirectURI() string {
	return c.CallbackBaseURL + GoogleCallbackPath
}

// XOAuthConfig cấu hình X (Twitter) Sign-In (đăng nhập / đăng ký). OAuth 2.0 tại developer.x.com
type XOAuthConfig struct {
	ClientID            string
	ClientSecret        string
	CallbackBaseURL     string
	FrontendRedirectURL string
	JWTSecret           string
}

func (c *XOAuthConfig) RedirectURI() string {
	return c.CallbackBaseURL + XCallbackPath
}
