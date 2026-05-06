package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/cronwarden/internal/api"
	"github.com/example/cronwarden/internal/db"
)

func TestJobStatuses_Empty(t *testing.T) {
	conn := tempDB(t)
	srv := api.New(conn)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/jobs/status", nil)
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result []db.JobStatus
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty array, got %d items", len(result))
	}
}

func TestJobStatuses_WithData(t *testing.T) {
	conn := tempDB(t)

	now := time.Now().UTC()
	_ = db.InsertJobRun(conn, db.JobRun{Name: "backup", StartedAt: now.Add(-1 * time.Hour), ExitCode: 0})
	_ = db.InsertJobRun(conn, db.JobRun{Name: "backup", StartedAt: now, ExitCode: 2, Error: "disk full"})
	_ = db.InsertJobRun(conn, db.JobRun{Name: "sync", StartedAt: now, ExitCode: 0})

	srv := api.New(conn)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/jobs/status", nil)
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result []db.JobStatus
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
	if result[0].Name != "backup" {
		t.Errorf("expected backup first, got %s", result[0].Name)
	}
	if result[0].LastExit != 2 {
		t.Errorf("expected last exit 2, got %d", result[0].LastExit)
	}
	if result[0].RunCount != 2 {
		t.Errorf("expected run_count 2, got %d", result[0].RunCount)
	}
}

func TestJobStatuses_MethodNotAllowed(t *testing.T) {
	conn := tempDB(t)
	srv := api.New(conn)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/jobs/status", nil)
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
