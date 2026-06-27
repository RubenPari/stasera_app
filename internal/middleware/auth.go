package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// AuthMiddleware returns an Echo middleware that validates JWT ACCESS tokens
// and stores the authenticated user ID (and claims) in the request context.
// Refresh tokens are explicitly rejected on protected routes: ValidateToken is
// called with expectedType "access", so a refresh token (long-lived) cannot be
// used to bypass the short access-token lifetime.
func AuthMiddleware(jwtManager *JWTManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header.Get(echo.HeaderAuthorization)
			if auth == "" {
				return Unauthorized(c, "missing or invalid token")
			}
			tokenString, ok := strings.CutPrefix(auth, "Bearer ")
			if !ok {
				return Unauthorized(c, "missing or invalid token")
			}

			claims, err := jwtManager.ValidateToken(strings.TrimSpace(tokenString), "access")
			if err != nil {
				return Unauthorized(c, "invalid or expired token")
			}

			uidStr, ok := (*claims)["user_id"].(string)
			if !ok {
				return Unauthorized(c, "invalid or expired token")
			}
			uid, err := uuid.Parse(uidStr)
			if err != nil {
				return Unauthorized(c, "invalid or expired token")
			}

			c.Set("userID", uid)
			c.Set("claims", claims)
			return next(c)
		}
	}
}

// GetUserID returns the authenticated user UUID from the Echo context.
func GetUserID(c echo.Context) (uuid.UUID, error) {
	uid, ok := c.Get("userID").(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("missing user id in context")
	}
	return uid, nil
}

// Unauthorized is a small helper to return 401 responses.
func Unauthorized(c echo.Context, msg string) error {
	return c.JSON(http.StatusUnauthorized, map[string]string{"error": msg})
}