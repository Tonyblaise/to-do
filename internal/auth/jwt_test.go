package auth

import (
	"testing"
	"time"
)

const testSecret = "test-secret-key-for-unit-tests!!"

func TestGenerateToken_ReturnsNonEmptyToken(t *testing.T) {
	token, err := GenerateToken("user-1", "a@b.com", testSecret, time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestParseToken_Valid(t *testing.T) {
	userID := "user-123"
	email := "test@example.com"

	token, err := GenerateToken(userID, email, testSecret, time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	claims, err := ParseToken(token, testSecret)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("UserID = %q, want %q", claims.UserID, userID)
	}
	if claims.Email != email {
		t.Errorf("Email = %q, want %q", claims.Email, email)
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	token, _ := GenerateToken("uid", "e@e.com", "secret-a", time.Hour)
	_, err := ParseToken(token, "secret-b")
	if err == nil {
		t.Error("expected error for wrong secret, got nil")
	}
}

func TestParseToken_Expired(t *testing.T) {
	token, _ := GenerateToken("uid", "e@e.com", testSecret, -time.Second)
	_, err := ParseToken(token, testSecret)
	if err == nil {
		t.Error("expected error for expired token, got nil")
	}
}

func TestParseToken_Malformed(t *testing.T) {
	cases := []string{"", "not.a.token", "Bearer abc", "abc"}
	for _, tc := range cases {
		_, err := ParseToken(tc, testSecret)
		if err == nil {
			t.Errorf("ParseToken(%q): expected error, got nil", tc)
		}
	}
}
