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

func seedOutputRun(t *testing.T, database *db.DB) int64 {
	t.Helper()
	id, err := database.InsertJobRun("output-job", true, 0, time.Now().UTC())
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	return id
}

func TestJobRunOutput_PutAndGet(t *testing.T) {
	database := tempDB(t)
	runID := seedOutputRun(t, database)
	router := api.New(database.DB)

	body := `{"stdout":"ok output","stderr":""}`
	req := httptest.NewRequest(http.MethodPut, "/runs/"+itoa(runID)+"/output", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("PUT status = %d, want 204", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/runs/"+itoa(runID)+"/output", nil)
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want 200", rec2.Code)
	}

	var out db.JobRunOutput
	if err := json.NewDecoder(rec2.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Stdout != "ok output" {
		t.Errorf("stdout = %q, want %q", out.Stdout, "ok output")
	}
}

func TestJobRunOutput_GetNotFound(t *testing.T) {
	database := tempDB(t)
	router := api.New(database.DB)

	req := httptest.NewRequest(http.MethodGet, "/runs/9999/output", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestJobRunOutput_Delete(t *testing.T) {
	database := tempDB(t)
	runID := seedOutputRun(t, database)
	router := api.New(database.DB)

	_ = db.UpsertJobRunOutput(database.DB, runID, "data", "")

	req := httptest.NewRequest(http.MethodDelete, "/runs/"+itoa(runID)+"/output", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Errorf("DELETE status = %d, want 204", rec.Code)
	}
}

func TestJobRunOutput_MethodNotAllowed(t *testing.T) {
	database := tempDB(t)
	router := api.New(database.DB)

	req := httptest.NewRequest(http.MethodPost, "/runs/1/output", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}
