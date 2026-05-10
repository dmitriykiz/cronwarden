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

func TestJobRunLock_PostAndGet(t *testing.T) {
	conn := tempDB(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/jobs/{job}/lock", api.ExportJobRunLockHandler(conn))

	body, _ := json.Marshal(map[string]any{"locked_by": "node-1", "ttl_seconds": 60})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/jobs/backup/lock", bytes.NewReader(body))
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}

	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/jobs/backup/lock", nil)
	mux.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr2.Code)
	}
	var l db.JobRunLock
	json.NewDecoder(rr2.Body).Decode(&l)
	if l.LockedBy != "node-1" {
		t.Fatalf("expected node-1, got %s", l.LockedBy)
	}
}

func TestJobRunLock_Conflict(t *testing.T) {
	conn := tempDB(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/jobs/{job}/lock", api.ExportJobRunLockHandler(conn))

	body, _ := json.Marshal(map[string]any{"locked_by": "node-1", "ttl_seconds": 60})
	req := httptest.NewRequest(http.MethodPost, "/jobs/backup/lock", bytes.NewReader(body))
	httptest.NewRecorder() // discard first
	mux.ServeHTTP(httptest.NewRecorder(), req)

	body2, _ := json.Marshal(map[string]any{"locked_by": "node-2", "ttl_seconds": 60})
	rr := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/jobs/backup/lock", bytes.NewReader(body2))
	mux.ServeHTTP(rr, req2)
	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}
}

func TestJobRunLock_GetNotFound(t *testing.T) {
	conn := tempDB(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/jobs/{job}/lock", api.ExportJobRunLockHandler(conn))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/jobs/ghost/lock", nil)
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestJobRunLock_Delete(t *testing.T) {
	conn := tempDB(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/jobs/{job}/lock", api.ExportJobRunLockHandler(conn))

	body, _ := json.Marshal(map[string]any{"locked_by": "node-1", "ttl_seconds": 60})
	mux.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/jobs/backup/lock", bytes.NewReader(body)))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/jobs/backup/lock?locked_by=node-1", nil)
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
}

func TestJobRunLock_MethodNotAllowed(t *testing.T) {
	conn := tempDB(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/jobs/{job}/lock", api.ExportJobRunLockHandler(conn))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/jobs/backup/lock", nil)
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}
