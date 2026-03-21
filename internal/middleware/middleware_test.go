package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Tonyblaise/to-do/internal/auth"
	"github.com/Tonyblaise/to-do/internal/config"
)

const testSecret = "test-secret-key-for-middleware!!"

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestAuth_MissingHeader(t *testing.T) {
	token, err := auth.GenerateToken("uid", "a@b.com", "other-secret", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuth_WrongPrefix(t *testing.T) {
	handler := Auth(testSecret)(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Token sometoken")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	handler := Auth(testSecret)(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}


func TestAuth_WrongSecret(t *testing.T) {
	token, _ := auth.GenerateToken("uid", "a@b.com", "other-secret", time.Hour)
	handler := Auth(testSecret)(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer"+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}
	token, err := auth.GenerateToken("user-123", "test@example.com", testSecret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	var capturedUserID string
	handler := Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer"+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if capturedUserID != "user-123" {
		t.Errorf("userID = %q, want %q", capturedUserID, "user-123")
	}
}

func TestAuth_ExpiredToken(t *testing.T) {
	token, _ := auth.GenerateToken("uid", "a@b.com", testSecret, -time.Second)
	handler := Auth(testSecret)(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer"+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestGetUserID_WithValue(t *testing.T) {
	ctx := context.WithValue(context.Background(), ContextKeyUserID, "abc-123")
	if id := GetUserID(ctx); id != "abc-123" {
		t.Errorf("GetUserID = %q, want %q", id, "abc-123")
	}
}

func TestGetUserID_Empty(t *testing.T) {
	if id := GetUserID(context.Background()); id != "" {
		t.Errorf("expected empty string, got %q", id)
	}
}

func TestRequestID_GeneratesID(t *testing.T) {
	handler := RequestID(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") == "" {
		t.Error("expected X-Request-ID header to be set")
	}
}

func TestRequestID_PreservesIncomingID(t *testing.T) {
	handler := RequestID(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "my-request-id")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if got := w.Header().Get("X-Request-ID"); got != "my-request-id" {
		t.Errorf("X-Request-ID = %q, want %q", got, "my-request-id")
	}
}

func TestRequestID_UniquePerRequest(t *testing.T) {
	handler := RequestID(http.HandlerFunc(okHandler))

	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/", nil))

	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/", nil))

	id1 := w1.Header().Get("X-Request-ID")
	id2 := w2.Header().Get("X-Request-ID")
	if id1 == id2 {
		t.Errorf("expected unique IDs, both got %q", id1)
	}
}

func TestRateLimit_AllowsUnderLimit(t *testing.T) {
	cfg := &config.Config{RateLimitRequests: 5, RateLimitWindow: time.Minute}
	handler := RateLimit(http.HandlerFunc(okHandler), cfg)

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i+1, w.Code)
		}
	}
}

func TestRateLimit_BlocksOverLimit(t *testing.T) {
	cfg := &config.Config{RateLimitRequests: 2, RateLimitWindow: time.Minute}
	handler := RateLimit(http.HandlerFunc(okHandler), cfg)

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
		req.RemoteAddr = "10.0.0.2:9999"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		switch i {
		case 0, 1:
			if w.Code != http.StatusOK {
				t.Errorf("request %d: expected 200, got %d", i+1, w.Code)
			}
		case 2:
			if w.Code != http.StatusTooManyRequests {
				t.Errorf("request %d: expected 429, got %d", i+1, w.Code)
			}
		}
	}
}

func TestRateLimit_DifferentIPsIndependent(t *testing.T) {
	cfg := &config.Config{RateLimitRequests: 1, RateLimitWindow: time.Minute}
	handler := RateLimit(http.HandlerFunc(okHandler), cfg)

	for _, ip := range []string{"192.168.1.1:0", "192.168.1.2:0", "192.168.1.3:0"} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = ip
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("IP %s: expected 200, got %d", ip, w.Code)
		}
	}
}
