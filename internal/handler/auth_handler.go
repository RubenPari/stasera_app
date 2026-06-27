package handler

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"github.com/stasera/stasera-api/internal/middleware"
	"github.com/stasera/stasera-api/internal/model"
	"github.com/stasera/stasera-api/internal/repository"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	userRepo   UserStore
	jwtManager *middleware.JWTManager
}

// NewAuthHandler returns a new AuthHandler with the required dependencies.
func NewAuthHandler(
	userRepo UserStore,
	jwtManager *middleware.JWTManager,
) *AuthHandler {
	return &AuthHandler{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

type registerRequest struct {
	Email       string `json:"email"        validate:"required,email"`
	Password    string `json:"password"     validate:"required,gte=8"`
	DisplayName string `json:"display_name" validate:"required"`
}

type loginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// Register creates a new user account, seeds default staples and returns a token pair.
func (h *AuthHandler) Register(c echo.Context) error {
	var req registerRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}

	user, err := h.userRepo.CreateWithDefaultStaples(c.Request().Context(), req.Email, string(hash), req.DisplayName)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			return c.JSON(http.StatusConflict, map[string]string{"error": "email already registered"})
		}
		return respondError(c, err)
	}

	pair, err := h.jwtManager.GenerateTokenPair(user)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(http.StatusCreated, model.AuthResponse{
		User:         toUserDTO(user),
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	})
}

// Login authenticates an existing user and returns a fresh token pair.
func (h *AuthHandler) Login(c echo.Context) error {
	var req loginRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	user, err := h.userRepo.FindByEmail(c.Request().Context(), req.Email)
	if err != nil {
		return respondError(c, err)
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
	}

	pair, err := h.jwtManager.GenerateTokenPair(*user)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(http.StatusOK, model.AuthResponse{
		User:         toUserDTO(*user),
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	})
}

// Refresh validates a refresh token and returns a rotated token pair.
func (h *AuthHandler) Refresh(c echo.Context) error {
	var req refreshRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	claims, err := h.jwtManager.ValidateToken(req.RefreshToken, "refresh")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid refresh token"})
	}

	uidStr, ok := (*claims)["user_id"].(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid refresh token"})
	}

	uid, err := uuid.Parse(uidStr)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid refresh token"})
	}

	user, err := h.userRepo.GetByID(c.Request().Context(), uid)
	if err != nil {
		return respondError(c, err)
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid refresh token"})
	}

	pair, err := h.jwtManager.GenerateTokenPair(*user)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(http.StatusOK, model.TokenPair{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	})
}

// Me returns the authenticated user's profile.
func (h *AuthHandler) Me(c echo.Context) error {
	uid := mustUserID(c)

	user, err := h.userRepo.GetByID(c.Request().Context(), uid)
	if err != nil {
		return respondError(c, err)
	}
	if user == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "user not found"})
	}

	return c.JSON(http.StatusOK, toUserDTO(*user))
}

type updateProfileRequest struct {
	DisplayName string `json:"display_name" validate:"required,min=1,max=100"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password"     validate:"required,gte=8"`
}

// UpdateMe updates the authenticated user's display name.
func (h *AuthHandler) UpdateMe(c echo.Context) error {
	uid := mustUserID(c)

	var req updateProfileRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	user, err := h.userRepo.UpdateProfile(c.Request().Context(), uid, req.DisplayName)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(http.StatusOK, toUserDTO(user))
}

// ChangePassword rotates the authenticated user's password after verifying the current one.
func (h *AuthHandler) ChangePassword(c echo.Context) error {
	uid := mustUserID(c)

	var req changePasswordRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	user, err := h.userRepo.GetByID(c.Request().Context(), uid)
	if err != nil {
		return respondError(c, err)
	}
	if user == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "user not found"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid current password"})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}

	if err := h.userRepo.UpdatePasswordHash(c.Request().Context(), uid, string(hash)); err != nil {
		return respondError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

func toUserDTO(u model.User) model.UserDTO {
	return model.UserDTO{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		CreatedAt:   u.CreatedAt,
	}
}