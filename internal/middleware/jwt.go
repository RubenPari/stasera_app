package middleware

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stasera/stasera-api/internal/model"
)

// JWTManager signs and validates JSON Web Tokens.
type JWTManager struct {
	secret          []byte
	accessExpiry    time.Duration
	refreshExpiry   time.Duration
}

// NewJWTManager creates a JWTManager with the given secret and token lifetimes.
func NewJWTManager(secret string, accessExpiryMinutes, refreshExpiryDays int) *JWTManager {
	return &JWTManager{
		secret:        []byte(secret),
		accessExpiry:  time.Duration(accessExpiryMinutes) * time.Minute,
		refreshExpiry: time.Duration(refreshExpiryDays) * 24 * time.Hour,
	}
}

// GenerateTokenPair creates a new access/refresh token pair for the given user.
func (j *JWTManager) GenerateTokenPair(user model.User) (model.TokenPair, error) {
	now := time.Now()

	accessClaims := jwt.MapClaims{
		"user_id":      user.ID.String(),
		"email":        user.Email,
		"display_name": user.DisplayName,
		"token_type":   "access",
		"exp":          now.Add(j.accessExpiry).Unix(),
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(j.secret)
	if err != nil {
		return model.TokenPair{}, fmt.Errorf("sign access token: %w", err)
	}

	refreshClaims := jwt.MapClaims{
		"user_id":    user.ID.String(),
		"token_type": "refresh",
		"exp":        now.Add(j.refreshExpiry).Unix(),
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(j.secret)
	if err != nil {
		return model.TokenPair{}, fmt.Errorf("sign refresh token: %w", err)
	}

	return model.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// ValidateToken parses a token string and returns its claims after verifying the signature.
// It also checks that the token_type claim is present.
func (j *JWTManager) ValidateToken(tokenString string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	if _, ok := claims["token_type"]; !ok {
		return nil, fmt.Errorf("missing token_type claim")
	}

	return &claims, nil
}
