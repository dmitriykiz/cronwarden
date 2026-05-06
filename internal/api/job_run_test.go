package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/user/cronwarden/internal/api"
	"github.com/user/cronwarden/internal/db"
)

func insertRun(t *testing.T, database interface{ /* *sql.DB */ }, name string) {
	t.Helper()
}

func TestGetJobRun_API_Found(t *testing.T) {
	sqldb := tempDB(t)
	_ = db.InsertJobRun(sqldb, db.JobRun{
		JobName:   "nightly",
		StartedAt: time.Now().UTC(),
		Duration:  200,
		Success:   true,
		Output:    "done",
	})
	runs, _ := db.ListJobRuns(sqldb, 1)
	if len(runs) == 0 {
		t.Fatal("no runs inserted")
	}
	id := runs[0].ID

	srv := api.New(sqldb, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/runs/"+strconv.FormatInt(id, 10), nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["JobName"] != "nightly" {
		t.Errorf("unexpected job name: %v", body["JobName"])
	}
}

func TestGetJobRun_API_NotFound(t *testing.T) {
	sqldb := tempDB(t)
	srv := api.New(sqldb, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/runs/9999", nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

func TestDeleteJobRun_API_Success(t *testing.T) {
	sqldb := tempDB(t)
	_ = db.InsertJobRun(sqldb, db.JobRun{
		JobName:   "cleanup",
		StartedAt: time.Now().UTC(),
		Duration:  10,
		Success:   false,
		Output:    "fail",
	})
	runs, _ := db.ListJobRuns(sqldb, 1)
	id := runs[0].ID

	srv := api.New(sqldb, nil)
	req := httptest.NewRequest(http.MethodDelete, "/api/runs/"+strconv.FormatInt(id, 10), nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d", rec.Code)
	}
}

func TestDeleteJobRun_API_NotFound(t *testing.T) {
	sqldb := tempDB(t)
	srv := api.New(sqldb, nil)
	req := httptest.NewRequest(http.MethodDelete, "/api/runs/8888", nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

func TestJobRun_API_MethodNotAllowed(t *testing.T) {
	sqldb := tempDB(t)
	srv := api.New(sqldb, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/runs/1", nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("want 405, got %d", rec.Code)
	}
}
