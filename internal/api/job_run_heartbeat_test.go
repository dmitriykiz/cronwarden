package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cronwarden/cronwarden/internal/api"
	"github.com/cronwarden/cronwarden/internal/db"
)

func seedHeartbeatAPIRun(t *testing.T, database *db.DB, jobName string) int64 {
	t.Helper()
	id, err := database.InsertJobRun(jobName, "success", 0, time.Now())
	if err != nil {
		t.Fatalf("InsertJobRun: %v", err)
	}
	return id
}

func TestJobRunHeartbeat_PutAndGet(t *testing.T) {
	database := tempDB(t)
	runID := seedHeartbeatAPIRun(t, database, "nightly")
	h := api.NewRouter(database.DB)

	body := `{"last_seen_at":"2024-01-01T00:00:00Z","timeout_at":"2024-01-01T01:00:00Z"}`
	req := httptest.NewRequest(http.MethodPut, "/api/runs/"+itoa(int(runID))+"/heartbeat", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("PUT expected 204, got %d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/runs/"+itoa(int(runID))+"/heartbeat", nil)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("GET expected 200, got %d: %s", rec2.Code, rec2.Body.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(rec2.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
}

func TestJobRunHeartbeat_GetNotFound(t *testing.T) {
	database := tempDB(t)
	h := api.NewRouter(database.DB)

	req := httptest.NewRequest(http.MethodGet, "/api/runs/9999/heartbeat", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestJobRunHeartbeat_Delete(t *testing.T) {
	database := tempDB(t)
	runID := seedHeartbeatAPIRun(t, database, "weekly")
	h := api.NewRouter(database.DB)

	body := `{"timeout_at":"2099-01-01T00:00:00Z"}`
	req := httptest.NewRequest(http.MethodPut, "/api/runs/"+itoa(int(runID))+"/heartbeat", strings.NewReader(body))
	httptest.NewRecorder()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	req2 := httptest.NewRequest(http.MethodDelete, "/api/runs/"+itoa(int(runID))+"/heartbeat", nil)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusNoContent {
		t.Fatalf("DELETE expected 204, got %d", rec2.Code)
	}
}

func TestJobRunHeartbeat_MethodNotAllowed(t *testing.T) {
	database := tempDB(t)
	h := api.NewRouter(database.DB)

	req := httptest.NewRequest(http.MethodPost, "/api/runs/1/heartbeat", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestListExpiredHeartbeats_Empty(t *testing.T) {
	database := tempDB(t)
	h := api.NewRouter(database.DB)

	req := httptest.NewRequest(http.MethodGet, "/api/heartbeats/expired", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var result []interface{}
	json.NewDecoder(rec.Body).Decode(&result)
	if len(result) != 0 {
		t.Fatalf("expected empty list, got %d items", len(result))
	}
}
