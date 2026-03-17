package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Tonyblaise/to-do/internal/auth"
)


func validAuthHeader(t *testing.T) string {
	t.Helper()
	token, err := auth.GenerateToken("user-test-id", "test@example.com", "test-secret-key", time.Hour)
	if err != nil {
		t.Fatalf("auth.GenerateToken: %v", err)
	}
	return "Bearer" + token
}


func TestProtectedRoutes_NoAuth(t *testing.T) {
	srv := newTestServer(t)

	routes := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, "/api/v1/tasks", ""},
		{http.MethodPost, "/api/v1/tasks", `{"title":"t"}`},
		{http.MethodGet, "/api/v1/tasks/some-id", ""},
		{http.MethodPatch, "/api/v1/tasks/some-id", `{"title":"t"}`},
		{http.MethodDelete, "/api/v1/tasks/some-id", ""},
		{http.MethodPatch, "/api/v1/tasks/some-id/status", `{"status":"pending"}`},
		{http.MethodGet, "/api/v1/tasks/export", ""},
		{http.MethodGet, "/api/v1/tasks/sync", ""},
		{http.MethodPatch, "/api/v1/tasks/bulk", `{"task_ids":["id1"]}`},
		{http.MethodDelete, "/api/v1/tasks/bulk", `{"task_ids":["id1"]}`},
		{http.MethodGet, "/api/v1/tags", ""},
		{http.MethodPost, "/api/v1/tags", `{"name":"t"}`},
		{http.MethodDelete, "/api/v1/tags/some-id", ""},
		{http.MethodGet, "/api/v1/attachments/some-id", ""},
		{http.MethodDelete, "/api/v1/attachments/some-id", ""},
	}

	for _, rt := range routes {
		t.Run(fmt.Sprintf("%s %s", rt.method, rt.path), func(t *testing.T) {
			var body *bytes.Buffer
			if rt.body != "" {
				body = bytes.NewBufferString(rt.body)
			} else {
				body = bytes.NewBuffer(nil)
			}
			req := httptest.NewRequest(rt.method, rt.path, body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("expected 401, got %d", w.Code)
			}
		})
	}
}


func TestProtectedRoutes_InvalidToken(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
	req.Header.Set("Authorization", "Bearer invalid.jwt.token")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}



func TestCreateTask_InvalidJSON(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString("not json"))
	req.Header.Set("Authorization", validAuthHeader(t))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateTask_MissingTitle(t *testing.T) {
	srv := newTestServer(t)

	body, _ := json.Marshal(map[string]string{"priority": "high"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBuffer(body))
	req.Header.Set("Authorization", validAuthHeader(t))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateTask_InvalidPriority(t *testing.T) {
	srv := newTestServer(t)

	body, _ := json.Marshal(map[string]string{"title": "My task", "priority": "urgent"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBuffer(body))
	req.Header.Set("Authorization", validAuthHeader(t))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}



func TestUpdateStatus_InvalidJSON(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/tasks/some-id/status", bytes.NewBufferString("{bad"))
	req.Header.Set("Authorization", validAuthHeader(t))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateStatus_InvalidStatus(t *testing.T) {
	srv := newTestServer(t)

	body, _ := json.Marshal(map[string]string{"status": "done"})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/tasks/some-id/status", bytes.NewBuffer(body))
	req.Header.Set("Authorization", validAuthHeader(t))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}


func TestBulkUpdate_InvalidJSON(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/tasks/bulk", bytes.NewBufferString("{bad"))
	req.Header.Set("Authorization", validAuthHeader(t))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestBulkUpdate_EmptyTaskIDs(t *testing.T) {
	srv := newTestServer(t)

	body, _ := json.Marshal(map[string]interface{}{"task_ids": []string{}, "update": map[string]string{}})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/tasks/bulk", bytes.NewBuffer(body))
	req.Header.Set("Authorization", validAuthHeader(t))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}



func TestBulkDelete_InvalidJSON(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/bulk", bytes.NewBufferString("not json"))
	req.Header.Set("Authorization", validAuthHeader(t))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestBulkDelete_EmptyTaskIDs(t *testing.T) {
	srv := newTestServer(t)

	body, _ := json.Marshal(map[string]interface{}{"task_ids": []string{}})
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/bulk", bytes.NewBuffer(body))
	req.Header.Set("Authorization", validAuthHeader(t))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}



func TestSync_MissingLastSyncedAt(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/sync", nil)
	req.Header.Set("Authorization", validAuthHeader(t))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestSync_InvalidTimestamp(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/sync?last_synced_at=not-a-date", nil)
	req.Header.Set("Authorization", validAuthHeader(t))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}



func TestCreateTag_InvalidJSON(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tags", bytes.NewBufferString("not json"))
	req.Header.Set("Authorization", validAuthHeader(t))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateTag_MissingName(t *testing.T) {
	srv := newTestServer(t)

	body, _ := json.Marshal(map[string]string{"color": "#ff0000"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tags", bytes.NewBuffer(body))
	req.Header.Set("Authorization", validAuthHeader(t))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
