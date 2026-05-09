package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cronwarden/cronwarden/internal/api"
	"github.com/cronwarden/cronwarden/internal/db"
)

func seedMetricAPIRun(t *testing.T, database *db.DB) int64 {
	t.Helper()
	id, err := db.InsertJobRun(database.DB, db.JobRun{
		JobName:   "api-metric-job",
		Status:    "success",
		StartedAt: time.Now().UTC(),
		Duration:  0.5,
	})
	if err != nil {
		t.Fatalf("seed run: %v", err)
	}
	return id
}

func TestJobRunMetrics_ListEmpty(t *testing.T) {
	database := tempDB(t)
	runID := seedMetricAPIRun(t, database)

	srv := api.New(database.DB, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/runs/"+itoa(int(runID))+"/metrics", nil)
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var result []map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&result)
	if len(result) != 0 {
		t.Errorf("expected empty list")
	}
}

func TestJobRunMetrics_PostAndList(t *testing.T) {
	database := tempDB(t)
	runID := seedMetricAPIRun(t, database)

	srv := api.New(database.DB, nil)

	body, _ := json.Marshal(map[string]string{"key": "duration_ms", "value": "320"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/runs/"+itoa(int(runID))+"/metrics", bytes.NewReader(body))
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/runs/"+itoa(int(runID))+"/metrics", nil)
	srv.ServeHTTP(rec2, req2)
	var list []map[string]interface{}
	json.NewDecoder(rec2.Body).Decode(&list)
	if len(list) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(list))
	}
	if list[0]["key"] != "duration_ms" {
		t.Errorf("unexpected key: %v", list[0]["key"])
	}
}

func TestJobRunMetrics_Delete(t *testing.T) {
	database := tempDB(t)
	runID := seedMetricAPIRun(t, database)

	mID, _ := db.AddJobRunMetric(database.DB, runID, "mem_mb", "128")

	srv := api.New(database.DB, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/runs/"+itoa(int(runID))+"/metrics/"+itoa(int(mID)), nil)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestJobRunMetrics_MethodNotAllowed(t *testing.T) {
	database := tempDB(t)
	runID := seedMetricAPIRun(t, database)

	srv := api.New(database.DB, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/runs/"+itoa(int(runID))+"/metrics", nil)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}
