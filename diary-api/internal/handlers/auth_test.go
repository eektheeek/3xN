package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/eektheeek/dead-lift-project/diary-api/internal/db"
	"github.com/eektheeek/dead-lift-project/diary-api/internal/handlers"
	"github.com/eektheeek/dead-lift-project/diary-api/internal/repository"
)

func newTestMux(t *testing.T) http.Handler {
	t.Helper()
	sqlDB, err := db.Open(filepath.Join(t.TempDir(), "auth.db"), filepath.Join("..", "..", "migrations"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return handlers.CORS(handlers.NewMux(handlers.NewAPI(repository.New(sqlDB))))
}

func postJSON(t *testing.T, mux http.Handler, path string, body any, token string) *httptest.ResponseRecorder {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestRegisterLoginMeAndIsolation(t *testing.T) {
	mux := newTestMux(t)

	rec := postJSON(t, mux, "/v1/auth/register", map[string]string{
		"email": "One@Test.Local", "password": "secret123",
	}, "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("register status=%d body=%s", rec.Code, rec.Body.String())
	}
	var created struct {
		Token string `json:"token"`
		User  struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		} `json:"user"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode register: %v", err)
	}
	if created.Token == "" || created.User.Email != "one@test.local" {
		t.Fatalf("unexpected register: %+v", created)
	}

	rec = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/exercises", nil)
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauth list expected 401, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+created.Token)
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("me status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = postJSON(t, mux, "/v1/exercises", map[string]any{
		"name": "Pull-up", "kind": "reps",
	}, created.Token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create exercise status=%d body=%s", rec.Code, rec.Body.String())
	}
	var ex struct{ ID string `json:"id"` }
	if err := json.Unmarshal(rec.Body.Bytes(), &ex); err != nil {
		t.Fatalf("decode exercise: %v", err)
	}

	rec = postJSON(t, mux, "/v1/auth/register", map[string]string{
		"email": "two@test.local", "password": "secret123",
	}, "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("register two: %d %s", rec.Code, rec.Body.String())
	}
	var other struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &other); err != nil {
		t.Fatalf("decode two: %v", err)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1/exercises/"+ex.ID, nil)
	req.Header.Set("Authorization", "Bearer "+other.Token)
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("other user should get 404, got %d %s", rec.Code, rec.Body.String())
	}

	rec = postJSON(t, mux, "/v1/auth/login", map[string]string{
		"email": "one@test.local", "password": "secret123",
	}, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("login status=%d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	raw, _ := json.Marshal(map[string]string{"currentPassword": "secret123", "newPassword": "newpass99"})
	req = httptest.NewRequest(http.MethodPut, "/v1/auth/password", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+created.Token)
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("change password status=%d body=%s", rec.Code, rec.Body.String())
	}
}
