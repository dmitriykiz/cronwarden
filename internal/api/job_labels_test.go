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

func TestJobLabels_GetEmpty(t *testing.T) {
	sqlDB := tempDB(t)
	srv := api.New(sqlDB)

	req := httptest.NewRequest(http.MethodGet, "/jobs/backup/labels", nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var got map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty map, got %v", got)
	}
}

func TestJobLabels_PutAndGet(t *testing.T) {
	sqlDB := tempDB(t)
	srv := api.New(sqlDB)

	body, _ := json.Marshal(map[string]string{"env": "prod", "team": "ops"})
	req := httptest.NewRequest(http.MethodPut, "/jobs/backup/labels", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("PUT expected 204, got %d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/jobs/backup/labels", nil)
	rec2 := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec2, req2)

	var got map[string]string
	_ = json.NewDecoder(rec2.Body).Decode(&got)
	if got["env"] != "prod" || got["team"] != "ops" {
		t.Errorf("unexpected labels: %v", got)
	}
}

func TestJobLabels_Delete(t *testing.T) {
	sqlDB := tempDB(t)
	srv := api.New(sqlDB)

	_ = db.SetJobLabels(sqlDB, "backup", map[string]string{"env": "prod"})

	req := httptest.NewRequest(http.MethodDelete, "/jobs/backup/labels", nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE expected 204, got %d", rec.Code)
	}

	labels, _ := db.ListJobLabels(sqlDB, "backup")
	if len(labels) != 0 {
		t.Errorf("expected empty labels after delete, got %v", labels)
	}
}

func TestJobLabels_MethodNotAllowed(t *testing.T) {
	sqlDB := tempDB(t)
	srv := api.New(sqlDB)

	req := httptest.NewRequest(http.MethodPost, "/jobs/backup/labels", nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}
