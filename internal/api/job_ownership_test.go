package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cronwarden/cronwarden/internal/api"
	"github.com/cronwarden/cronwarden/internal/db"
)

func TestJobOwnership_GetEmpty(t *testing.T) {
	conn := tempDB(t)
	srv := api.New(conn, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/job/ownership", nil)
	w := httptest.NewRecorder()
	srv.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var list []db.JobOwnership
	json.NewDecoder(w.Body).Decode(&list)
	if len(list) != 0 {
		t.Errorf("expected empty list, got %d items", len(list))
	}
}

func TestJobOwnership_PutAndGet(t *testing.T) {
	conn := tempDB(t)
	srv := api.New(conn, nil)

	body, _ := json.Marshal(map[string]string{
		"owner":        "alice",
		"team":         "platform",
		"slack_handle": "@alice",
	})
	req := httptest.NewRequest(http.MethodPut, "/api/job/ownership?job=backup", bytes.NewReader(body))
	w := httptest.NewRecorder()
	srv.Handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("put: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/job/ownership?job=backup", nil)
	w2 := httptest.NewRecorder()
	srv.Handler.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("get: expected 200, got %d", w2.Code)
	}
	var o db.JobOwnership
	json.NewDecoder(w2.Body).Decode(&o)
	if o.Owner != "alice" || o.Team != "platform" {
		t.Errorf("unexpected ownership: %+v", o)
	}
}

func TestJobOwnership_Delete(t *testing.T) {
	conn := tempDB(t)
	srv := api.New(conn, nil)

	body, _ := json.Marshal(map[string]string{"owner": "bob", "team": "ops", "slack_handle": "@bob"})
	req := httptest.NewRequest(http.MethodPut, "/api/job/ownership?job=sync", bytes.NewReader(body))
	w := httptest.NewRecorder()
	srv.Handler.ServeHTTP(w, req)

	req2 := httptest.NewRequest(http.MethodDelete, "/api/job/ownership?job=sync", nil)
	w2 := httptest.NewRecorder()
	srv.Handler.ServeHTTP(w2, req2)
	if w2.Code != http.StatusNoContent {
		t.Fatalf("delete: expected 204, got %d", w2.Code)
	}

	req3 := httptest.NewRequest(http.MethodGet, "/api/job/ownership?job=sync", nil)
	w3 := httptest.NewRecorder()
	srv.Handler.ServeHTTP(w3, req3)
	if w3.Code != http.StatusNotFound {
		t.Fatalf("get after delete: expected 404, got %d", w3.Code)
	}
}

func TestJobOwnership_MethodNotAllowed(t *testing.T) {
	conn := tempDB(t)
	srv := api.New(conn, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/job/ownership?job=x", nil)
	w := httptest.NewRecorder()
	srv.Handler.ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
