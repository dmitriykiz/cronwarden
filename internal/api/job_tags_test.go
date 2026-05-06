package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/cronwarden/internal/api"
	"github.com/example/cronwarden/internal/db"
)

func TestJobTags_GetEmpty(t *testing.T) {
	sqlDB := tempDB(t)
	srv := api.New(sqlDB, nil)

	req := httptest.NewRequest(http.MethodGet, "/jobs/myjob/tags", nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp struct {
		Tags []string `json:"tags"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Tags) != 0 {
		t.Errorf("expected empty tags, got %v", resp.Tags)
	}
}

func TestJobTags_PutAndGet(t *testing.T) {
	sqlDB := tempDB(t)
	srv := api.New(sqlDB, nil)

	body, _ := json.Marshal(map[string][]string{"tags": {"prod", "critical"}})
	req := httptest.NewRequest(http.MethodPut, "/jobs/deploy/tags", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("PUT expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodGet, "/jobs/deploy/tags", nil)
	rec2 := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec2, req2)

	var resp struct {
		Tags []string `json:"tags"`
	}
	json.NewDecoder(rec2.Body).Decode(&resp) //nolint:errcheck
	if len(resp.Tags) != 2 {
		t.Errorf("expected 2 tags, got %v", resp.Tags)
	}
}

func TestJobTags_Delete(t *testing.T) {
	sqlDB := tempDB(t)
	srv := api.New(sqlDB, nil)

	_ = db.SetJobTags(sqlDB, "cleanup", []string{"old"})

	req := httptest.NewRequest(http.MethodDelete, "/jobs/cleanup/tags", nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestJobTags_MethodNotAllowed(t *testing.T) {
	sqlDB := tempDB(t)
	srv := api.New(sqlDB, nil)

	req := httptest.NewRequest(http.MethodPost, "/jobs/myjob/tags", nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
