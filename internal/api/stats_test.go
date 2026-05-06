package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cronwarden/cronwarden/internal/api"
	"github.com/cronwarden/cronwarden/internal/db"
)

func TestStats_Empty(t *testing.T) {
	conn := tempDB(t)
	srv := api.New(conn)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["total_runs"].(float64) != 0 {
		t.Errorf("expected 0 total_runs")
	}
}

func TestStats_WithData(t *testing.T) {
	conn := tempDB(t)
	now := time.Now().UTC()

	runs := []db.JobRun{
		{JobName: "alpha", StartedAt: now.Add(-2 * time.Minute), ExitCode: 0, Output: "ok"},
		{JobName: "alpha", StartedAt: now.Add(-1 * time.Minute), ExitCode: 1, Output: "err"},
		{JobName: "beta", StartedAt: now, ExitCode: 0, Output: "ok"},
	}
	for _, r := range runs {
		if err := db.InsertJobRun(conn, r); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}

	srv := api.New(conn)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["total_runs"].(float64) != 3 {
		t.Errorf("total_runs: want 3, got %v", resp["total_runs"])
	}
	if resp["unique_jobs"].(float64) != 2 {
		t.Errorf("unique_jobs: want 2, got %v", resp["unique_jobs"])
	}
	if resp["last_run_at"] == nil {
		t.Errorf("expected last_run_at to be set")
	}
}

func TestStats_MethodNotAllowed(t *testing.T) {
	conn := tempDB(t)
	srv := api.New(conn)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/stats", nil)
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("want 405, got %d", rec.Code)
	}
}
