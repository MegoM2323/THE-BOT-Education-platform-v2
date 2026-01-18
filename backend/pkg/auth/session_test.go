package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewSessionManager(t *testing.T) {
	secret := "my-secret-key-32-characters-long!"

	sm := NewSessionManager(secret)

	if sm == nil {
		t.Fatal("SessionManager should not be nil")
	}
}

func TestCreateSessionToken(t *testing.T) {
	secret := "my-secret-key-32-characters-long!"
	sm := NewSessionManager(secret)

	sessionID := uuid.New()
	userID := uuid.New()
	expiresAt := time.Now().Add(1 * time.Hour)

	token, err := sm.CreateSessionToken(sessionID, userID, expiresAt)

	if err != nil {
		t.Fatalf("CreateSessionToken failed: %v", err)
	}

	if token == "" {
		t.Error("Token should not be empty")
	}

	// Token should be reasonably long
	if len(token) < 50 {
		t.Error("Token seems too short")
	}
}

func TestValidateSessionToken(t *testing.T) {
	secret := "my-secret-key-32-characters-long!"
	sm := NewSessionManager(secret)

	sessionID := uuid.New()
	userID := uuid.New()
	expiresAt := time.Now().Add(1 * time.Hour)

	token, _ := sm.CreateSessionToken(sessionID, userID, expiresAt)

	// Validate the token
	data, err := sm.ValidateSessionToken(token)

	if err != nil {
		t.Fatalf("ValidateSessionToken failed: %v", err)
	}

	if data.UserID != userID {
		t.Errorf("Expected userID %s, got %s", userID, data.UserID)
	}

	if data.SessionID != sessionID {
		t.Errorf("Expected sessionID %s, got %s", sessionID, data.SessionID)
	}
}

func TestValidateSessionTokenInvalid(t *testing.T) {
	secret := "my-secret-key-32-characters-long!"
	sm := NewSessionManager(secret)

	// Try to validate invalid token
	_, err := sm.ValidateSessionToken("invalid-token")

	if err == nil {
		t.Error("ValidateSessionToken should fail for invalid token")
	}
}

func TestValidateSessionTokenExpired(t *testing.T) {
	secret := "my-secret-key-32-characters-long!"
	sm := NewSessionManager(secret)

	sessionID := uuid.New()
	userID := uuid.New()
	// Create token with past expiry
	expiresAt := time.Now().Add(-1 * time.Hour)

	token, _ := sm.CreateSessionToken(sessionID, userID, expiresAt)

	// Try to validate expired token
	_, err := sm.ValidateSessionToken(token)

	if err == nil {
		t.Error("ValidateSessionToken should fail for expired token")
	}
}

func TestValidateSessionTokenDifferentSecret(t *testing.T) {
	secret1 := "my-secret-key-32-characters-long!"
	secret2 := "different-secret-key-characters!"

	sm1 := NewSessionManager(secret1)
	sm2 := NewSessionManager(secret2)

	sessionID := uuid.New()
	userID := uuid.New()
	expiresAt := time.Now().Add(1 * time.Hour)

	token, _ := sm1.CreateSessionToken(sessionID, userID, expiresAt)

	// Try to validate with different secret
	_, err := sm2.ValidateSessionToken(token)

	if err == nil {
		t.Error("ValidateSessionToken should fail with different secret")
	}
}

func TestGetDefaultCookieOptions(t *testing.T) {
	token := "test-token"
	maxAge := 3600
	isProduction := false

	opts := GetDefaultCookieOptions(token, maxAge, isProduction)

	if opts.HTTPOnly != true {
		t.Error("HTTPOnly should be true")
	}

	if opts.Path != "/" {
		t.Errorf("Expected Path /, got %s", opts.Path)
	}

	if opts.Name != "session" {
		t.Errorf("Expected Name session, got %s", opts.Name)
	}

	if opts.Value != token {
		t.Errorf("Expected Value %s, got %s", token, opts.Value)
	}
}

func TestFormatSetCookie(t *testing.T) {
	token := "test-token"
	maxAge := 3600
	isProduction := false

	opts := GetDefaultCookieOptions(token, maxAge, isProduction)
	cookieStr := FormatSetCookie(opts)

	if cookieStr == "" {
		t.Error("FormatSetCookie should not return empty string")
	}

	// Should contain token value
	if !strings.Contains(cookieStr, token) {
		t.Error("Cookie string should contain token")
	}

	// Should contain HttpOnly flag
	if !strings.Contains(cookieStr, "HttpOnly") {
		t.Error("Cookie string should contain HttpOnly")
	}
}

func TestFormatDeleteCookie(t *testing.T) {
	cookieName := "session"

	cookieStr := FormatDeleteCookie(cookieName)

	if cookieStr == "" {
		t.Error("FormatDeleteCookie should not return empty string")
	}

	// Should contain Max-Age=0
	if !strings.Contains(cookieStr, "Max-Age=0") {
		t.Error("Delete cookie should contain Max-Age=0")
	}

	// Should contain cookie name
	if !strings.Contains(cookieStr, cookieName) {
		t.Error("Delete cookie should contain cookie name")
	}
}

func TestSessionManagerMultipleTokens(t *testing.T) {
	secret := "my-secret-key-32-characters-long!"
	sm := NewSessionManager(secret)

	sessionID1 := uuid.New()
	user1 := uuid.New()
	token1, _ := sm.CreateSessionToken(sessionID1, user1, time.Now().Add(1*time.Hour))

	sessionID2 := uuid.New()
	user2 := uuid.New()
	token2, _ := sm.CreateSessionToken(sessionID2, user2, time.Now().Add(1*time.Hour))

	if token1 == token2 {
		t.Error("Different sessions should get different tokens")
	}

	data1, _ := sm.ValidateSessionToken(token1)
	data2, _ := sm.ValidateSessionToken(token2)

	if data1.UserID != user1 {
		t.Errorf("Token1 should belong to user1, got %s", data1.UserID)
	}

	if data2.UserID != user2 {
		t.Errorf("Token2 should belong to user2, got %s", data2.UserID)
	}
}

func TestSessionTokenContainsUserID(t *testing.T) {
	secret := "my-secret-key-32-characters-long!"
	sm := NewSessionManager(secret)

	sessionID := uuid.New()
	userID := uuid.New()
	expiresAt := time.Now().Add(1 * time.Hour)

	token, _ := sm.CreateSessionToken(sessionID, userID, expiresAt)

	data, _ := sm.ValidateSessionToken(token)

	if data.UserID != userID {
		t.Errorf("Data should contain correct userID: expected %s, got %s", userID, data.UserID)
	}
}

func TestSecretLength(t *testing.T) {
	// Session manager should work with 32-character secrets (minimum recommended)
	secret := "0123456789abcdef0123456789abcdef" // 32 chars
	sm := NewSessionManager(secret)

	if sm == nil {
		t.Error("SessionManager should be created with 32-char secret")
	}

	// Should be able to create token
	token, err := sm.CreateSessionToken(uuid.New(), uuid.New(), time.Now().Add(1*time.Hour))
	if err != nil {
		t.Errorf("Should work with 32-char secret: %v", err)
	}

	if token == "" {
		t.Error("Token should be created")
	}
}

func TestValidTokenWithinExpiry(t *testing.T) {
	secret := "my-secret-key-32-characters-long!"
	sm := NewSessionManager(secret)

	sessionID := uuid.New()
	userID := uuid.New()
	// Token expires 1 hour from now
	expiresAt := time.Now().Add(1 * time.Hour)

	token, _ := sm.CreateSessionToken(sessionID, userID, expiresAt)

	// Should validate successfully
	data, err := sm.ValidateSessionToken(token)

	if err != nil {
		t.Errorf("Token should be valid: %v", err)
	}

	if data.UserID != userID {
		t.Errorf("UserID should match: expected %s, got %s", userID, data.UserID)
	}
}
