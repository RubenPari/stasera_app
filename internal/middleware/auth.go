package middleware

import (
	"errors"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

// AuthMiddleware returns an Echo middleware that validates JWT access tokens
// and stores the authenticated user ID in the request context.
func AuthMiddleware(jwtManager *JWTManager) echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		SigningKey: jwtManager.secret,
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(jwt.MapClaims)
		},
		SuccessHandler: func(c echo.Context) {
			token, ok := c.Get("user").(*jwt.Token)
			if !ok {
				return
			}
			claims, ok := token.Claims.(*jwt.MapClaims)
			if !ok {
				return
			}
			uid, err := uuid.Parse((*claims)["user_id"].(string))
			if err != nil {
				return
			}
			c.Set("userID", uid)
		},
	})
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
