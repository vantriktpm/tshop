// Package auth provides JWT and auth helpers for API Gateway / services.
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	UserID       string `json:"user_id"`
	Email        string `json:"email"`
	SessionID    string `json:"session_id"`
	JTI          string `json:"jti"`
	TokenVersion int    `json:"token_version"`
	jwt.RegisteredClaims
}

// ValidateJWT validates token and returns claims. Use with API Gateway or REST middleware.
func ValidateJWT(tokenString, secret string) (*Claims, error) {
	tok, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}
	claims, ok := tok.Claims.(*Claims)
	if !ok || !tok.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// NewToken creates a JWT for a user (e.g. from user-service).
// sessionID is the Redis session identifier, tokenVersion is stored in DB and
// used to revoke all tokens when incremented.
func NewToken(userID, email, sessionID string, tokenVersion int, secret string, exp time.Duration) (string, error) {
	if tokenVersion <= 0 {
		tokenVersion = 1
	}
	jti := uuid.New().String()
	claims := &Claims{
		UserID:       userID,
		Email:        email,
		SessionID:    sessionID,
		JTI:          jti,
		TokenVersion: tokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(exp)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        jti,
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(secret))
}
