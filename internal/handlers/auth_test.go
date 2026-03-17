package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"database/sql"

	"github.com/Tonyblaise/to-do/internal/config"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)


func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	
	db, err := sql.Open("postgres", "host=invalid user=invalid dbname=invalid sslmode=disable")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	r := mux.NewRouter()
	cfg := &config.Config{
		JWTSecret:   "test-secret-key",
		JWTExpiry:   time.Hour,
		StoragePath: t.TempDir(),
		MaxUploadSize: 10 << 20,
	}
	RegisterRoutes(r, db, cfg)
	return r
}



func TestSignup_InvalidJSON(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestSignup_ValidationErrors(t *testing.T) {
	srv := newTestServer(t)

	tests := []struct {
		name string
		body map[string]string
	}{
		{"missing email", map[string]string{"password": "password123"}},
		{"invalid email", map[string]string{"email": "notanemail", "password": "password123"}},
		{"short password", map[string]string{"email": "a@b.com", "password": "short"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", w.Code)
			}
		})
	}
}



func TestLogin_InvalidJSON(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
