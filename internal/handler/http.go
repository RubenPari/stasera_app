package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/stasera/stasera-api/internal/middleware"
	"github.com/stasera/stasera-api/internal/repository"
	"github.com/stasera/stasera-api/internal/service"
)

// echoValidator adapts a go-playground/validator.Validate to Echo's Validator
// interface so handlers can call c.Validate(req) without holding their own
// validator instance.
type echoValidator struct {
	v *validator.Validate
}

func (ev *echoValidator) Validate(i any) error {
	return ev.v.Struct(i)
}

// NewValidator returns the shared validator used as Echo's e.Validator.
func NewValidator() echo.Validator {
	return &echoValidator{v: validator.New()}
}

// bindAndValidate binds the request body to req and runs Echo's configured
// validator. On failure it returns a 400 *echo.HTTPError with a safe message
// (never the raw bind error, which may echo back client bytes).
func bindAndValidate(c echo.Context, req any) error {
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

// mustUserID returns the authenticated user ID from the Echo context.
// AuthMiddleware guarantees its presence on protected routes; in the anomalous
// case it is missing it returns uuid.Nil, which the repository treats as
// not-found rather than leaking data.
func mustUserID(c echo.Context) uuid.UUID {
	uid, err := middleware.GetUserID(c)
	if err != nil {
		return uuid.Nil
	}
	return uid
}

// parsePathUUID extracts and parses a UUID path parameter, returning a 400
// *echo.HTTPError on parse failure.
func parsePathUUID(c echo.Context, key string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param(key))
	if err != nil {
		return uuid.Nil, echo.NewHTTPError(http.StatusBadRequest, "invalid "+key)
	}
	return id, nil
}

// parsePathInt extracts and range-validates an integer path parameter,
// returning a 400 *echo.HTTPError on parse/range failure.
func parsePathInt(c echo.Context, key string, min, max int) (int, error) {
	v, err := strconv.Atoi(c.Param(key))
	if err != nil || v < min || v > max {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid "+key)
	}
	return v, nil
}

// respondError maps a service/repository error to a uniform HTTP response with
// a safe message. Unknown errors are logged server-side and surfaced to the
// client as a generic "internal error" — internal details never leave the API.
func respondError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, service.ErrNotFound):
		return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
	case errors.Is(err, service.ErrConflict):
		return c.JSON(http.StatusConflict, map[string]string{"error": "conflict"})
	case errors.Is(err, service.ErrValidation):
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	case errors.Is(err, service.ErrAIUnavailable):
		return c.JSON(http.StatusBadGateway, map[string]string{"error": "ai unavailable"})
	case errors.Is(err, repository.ErrDuplicateEmail):
		return c.JSON(http.StatusConflict, map[string]string{"error": "email already registered"})
	case errors.Is(err, repository.ErrRecipeInUse):
		return c.JSON(http.StatusConflict, map[string]string{"error": "recipe in use"})
	default:
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}
}