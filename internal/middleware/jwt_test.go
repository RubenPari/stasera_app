package middleware_test

import (
	"testing"

	"github.com/stasera/stasera-api/internal/middleware"
	"github.com/stasera/stasera-api/internal/model"

	"github.com/google/uuid"
)

func newTestUser() model.User {
	return model.User{
		ID:          uuid.New(),
		Email:       "test@example.com",
		DisplayName: "Test User",
	}
}

func TestValidateToken_AcceptAccess(t *testing.T) {
	mgr := middleware.NewJWTManager("test-secret-32-chars-minimum-ok!", 15, 30)
	user := newTestUser()

	pair, err := mgr.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("GenerateTokenPair: %v", err)
	}

	claims, err := mgr.ValidateToken(pair.AccessToken, "access")
	if err != nil {
		t.Fatalf("ValidateToken(access): %v", err)
	}
	uid, ok := (*claims)["user_id"].(string)
	if !ok || uid != user.ID.String() {
		t.Errorf("user_id claim mismatch: got %q want %q", uid, user.ID.String())
	}
}

func TestValidateToken_RejectRefreshAsAccess(t *testing.T) {
	mgr := middleware.NewJWTManager("test-secret-32-chars-minimum-ok!", 15, 30)
	user := newTestUser()

	pair, err := mgr.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("GenerateTokenPair: %v", err)
	}

	// Refresh token MUST be rejected when expectedType is "access".
	_, err = mgr.ValidateToken(pair.RefreshToken, "access")
	if err == nil {
		t.Fatal("expected error using refresh token as access, got nil")
	}
}

func TestValidateToken_AcceptRefresh(t *testing.T) {
	mgr := middleware.NewJWTManager("test-secret-32-chars-minimum-ok!", 15, 30)
	user := newTestUser()

	pair, err := mgr.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("GenerateTokenPair: %v", err)
	}

	claims, err := mgr.ValidateToken(pair.RefreshToken, "refresh")
	if err != nil {
		t.Fatalf("ValidateToken(refresh): %v", err)
	}
	if tt, _ := (*claims)["token_type"].(string); tt != "refresh" {
		t.Errorf("token_type=%q want refresh", tt)
	}
}

func TestValidateToken_RejectAccessAsRefresh(t *testing.T) {
	mgr := middleware.NewJWTManager("test-secret-32-chars-minimum-ok!", 15, 30)
	user := newTestUser()

	pair, err := mgr.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("GenerateTokenPair: %v", err)
	}

	_, err = mgr.ValidateToken(pair.AccessToken, "refresh")
	if err == nil {
		t.Fatal("expected error using access token as refresh, got nil")
	}
}

func TestAccessTokenHasNoPII(t *testing.T) {
	mgr := middleware.NewJWTManager("test-secret-32-chars-minimum-ok!", 15, 30)
	user := newTestUser()

	pair, err := mgr.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("GenerateTokenPair: %v", err)
	}

	claims, err := mgr.ValidateToken(pair.AccessToken, "access")
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}

	if _, exists := (*claims)["email"]; exists {
		t.Error("access token must not contain 'email' claim")
	}
	if _, exists := (*claims)["display_name"]; exists {
		t.Error("access token must not contain 'display_name' claim")
	}
}