package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cronwarden/internal/api"
	"github.com/cronwarden/internal/db"
)

func tempDB(t *testing.T) *db.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	database, err := db.Open(path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() {
		database.Close()
		os.Remove(path)
	})
	return database
}

func TestHealth(t *testing.T) {
	h := api.New(tempDB(t))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestListRuns_Empty(t *testing.T) {
	h := api.New(tempDB(t))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/runs", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var runs []interface{}
	if err := json.NewDecoder(rr.Body).Decode(&runs); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(runs) != 0 {
		t.Fatalf("expected empty slice, got %d items", len(runs))
	}
}

func TestListRuns_WithData(t *testing.T) {
	database := tempDB(t)
	h := api.New(database)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	_ = database.InsertJobRun("backup", 0, time.Second, "")
	_ = database.InsertJobRun("backup", 1, time.Second, "exit 1")

	req := httptest.NewRequest(http.MethodGet, "/api/runs?job=backup&limit=10", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var runs []interface{}
	if err := json.NewDecoder(rr.Body).Decode(&runs); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(runs) != 2 {
		t.Fatalf("expected 2 runs, got %d", len(runs))
	}
}

func TestListRuns_MethodNotAllowed(t *testing.T) {
	h := api.New(tempDB(t))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/runs", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}
