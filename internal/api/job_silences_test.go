package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/cronwarden/internal/api"
	"github.com/user/cronwarden/internal/db"
)

func TestJobSilences_GetEmpty(t *testing.T) {
	database := tempDB(t)
	srv := api.New(database, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/jobs/backup/silences", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result []db.JobSilence
	json.NewDecoder(w.Body).Decode(&result)
	if len(result) != 0 {
		t.Errorf("expected empty list")
	}
}

func TestJobSilences_PostAndGet(t *testing.T) {
	database := tempDB(t)
	srv := api.New(database, nil)

	now := time.Now().UTC().Truncate(time.Second)
	body, _ := json.Marshal(map[string]interface{}{
		"starts_at": now,
		"ends_at":   now.Add(2 * time.Hour),
		"reason":    "planned maintenance",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/jobs/backup/silences", bytes.NewReader(body))
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/jobs/backup/silences", nil)
	w2 := httptest.NewRecorder()
	srv.ServeHTTP(w2, req2)

	var silences []db.JobSilence
	json.NewDecoder(w2.Body).Decode(&silences)
	if len(silences) != 1 {
		t.Fatalf("expected 1 silence, got %d", len(silences))
	}
	if silences[0].Reason != "planned maintenance" {
		t.Errorf("unexpected reason: %q", silences[0].Reason)
	}
}

func TestJobSilences_Delete(t *testing.T) {
	database := tempDB(t)
	srv := api.New(database, nil)

	now := time.Now().UTC()
	id, _ := db.UpsertJobSilence(database, db.JobSilence{
		JobName:  "sync",
		StartsAt: now,
		EndsAt:   now.Add(time.Hour),
		Reason:   "test",
	})

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/jobs/sync/silences/%d", id), nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}

	req2 := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/jobs/sync/silences/%d", id), nil)
	w2 := httptest.NewRecorder()
	srv.ServeHTTP(w2, req2)
	if w2.Code != http.StatusNotFound {
		t.Errorf("expected 404 on second delete, got %d", w2.Code)
	}
}

func TestJobSilences_MethodNotAllowed(t *testing.T) {
	database := tempDB(t)
	srv := api.New(database, nil)

	req := httptest.NewRequest(http.MethodPut, "/api/jobs/backup/silences", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}
