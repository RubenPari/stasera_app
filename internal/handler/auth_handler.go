package handler

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"github.com/stasera/stasera-api/internal/middleware"
	"github.com/stasera/stasera-api/internal/model"
	"github.com/stasera/stasera-api/internal/repository"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	userRepo   *repository.UserRepository
	stapleRepo *repository.StapleRepository
	jwtManager *middleware.JWTManager
	validator  *validator.Validate
}

// NewAuthHandler returns a new AuthHandler with the required dependencies.
func NewAuthHandler(
	userRepo *repository.UserRepository,
	stapleRepo *repository.StapleRepository,
	jwtManager *middleware.JWTManager,
) *AuthHandler {
	return &AuthHandler{
		userRepo:   userRepo,
		stapleRepo: stapleRepo,
		jwtManager: jwtManager,
		validator:  validator.New(),
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
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if err := h.validator.Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to hash password"})
	}

	user, err := h.userRepo.Create(c.Request().Context(), req.Email, string(hash), req.DisplayName)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			return c.JSON(http.StatusConflict, map[string]string{"error": "email already registered"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
	}

	if err := h.stapleRepo.SeedDefaults(c.Request().Context(), user.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to seed staples"})
	}

	pair, err := h.jwtManager.GenerateTokenPair(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate tokens"})
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
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if err := h.validator.Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	user, err := h.userRepo.FindByEmail(c.Request().Context(), req.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to lookup user"})
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
	}

	pair, err := h.jwtManager.GenerateTokenPair(*user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate tokens"})
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
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if err := h.validator.Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	claims, err := h.jwtManager.ValidateToken(req.RefreshToken)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid refresh token"})
	}

	if tokenType, _ := (*claims)["token_type"].(string); tokenType != "refresh" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid refresh token"})
	}

	uidStr, ok := (*claims)["user_id"].(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid refresh token"})
	}

	user, err := h.userRepo.FindByID(c.Request().Context(), uuidFromString(uidStr))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to lookup user"})
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid refresh token"})
	}

	pair, err := h.jwtManager.GenerateTokenPair(*user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate tokens"})
	}

	return c.JSON(http.StatusOK, model.TokenPair{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	})
}

// Me returns the authenticated user's profile.
func (h *AuthHandler) Me(c echo.Context) error {
	uid, err := middleware.GetUserID(c)
	if err != nil {
		return middleware.Unauthorized(c, "missing or invalid token")
	}

	user, err := h.userRepo.FindByID(c.Request().Context(), uid)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to lookup user"})
	}
	if user == nil {
		return middleware.Unauthorized(c, "user not found")
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
	uid, err := middleware.GetUserID(c)
	if err != nil {
		return middleware.Unauthorized(c, "missing or invalid token")
	}

	var req updateProfileRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if err := h.validator.Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	user, err := h.userRepo.UpdateProfile(c.Request().Context(), uid, req.DisplayName)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update profile"})
	}

	return c.JSON(http.StatusOK, toUserDTO(user))
}

// ChangePassword rotates the authenticated user's password after verifying the current one.
func (h *AuthHandler) ChangePassword(c echo.Context) error {
	uid, err := middleware.GetUserID(c)
	if err != nil {
		return middleware.Unauthorized(c, "missing or invalid token")
	}

	var req changePasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if err := h.validator.Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	user, err := h.userRepo.FindByID(c.Request().Context(), uid)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to lookup user"})
	}
	if user == nil {
		return middleware.Unauthorized(c, "user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid current password"})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to hash password"})
	}

	if err := h.userRepo.UpdatePasswordHash(c.Request().Context(), uid, string(hash)); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update password"})
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

func uuidFromString(s string) uuid.UUID {
	id, _ := uuid.Parse(s)
	return id
}
