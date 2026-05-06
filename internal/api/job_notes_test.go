package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cronwarden/internal/db"
)

func seedNoteRun(t *testing.T, database *sql.DB) int64 {
	t.Helper()
	run := db.JobRun{
		JobName:   "note-job",
		StartedAt: time.Now().Add(-time.Minute),
		FinishedAt: time.Now(),
		ExitCode:  0,
		Output:    "ok",
		Success:   true,
	}
	id, err := db.InsertJobRun(database, run)
	if err != nil {
		t.Fatalf("seedNoteRun: %v", err)
	}
	return id
}

func TestJobNotes_AddAndList(t *testing.T) {
	database := tempDB(t)
	runID := seedNoteRun(t, database)
	h := newJobNotesHandler(database)

	body, _ := json.Marshal(map[string]string{"note": "looks good"})
	req := httptest.NewRequest(http.MethodPost, "/runs/"+itoa(runID)+"/notes", bytes.NewReader(body))
	req.SetPathValue("id", itoa(runID))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/runs/"+itoa(runID)+"/notes", nil)
	req2.SetPathValue("id", itoa(runID))
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w2.Code)
	}
	var notes []map[string]any
	if err := json.NewDecoder(w2.Body).Decode(&notes); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(notes) != 1 {
		t.Fatalf("expected 1 note, got %d", len(notes))
	}
}

func TestJobNotes_DeleteNote(t *testing.T) {
	database := tempDB(t)
	runID := seedNoteRun(t, database)
	noteID, err := db.AddJobNote(database, runID, "to be deleted")
	if err != nil {
		t.Fatalf("AddJobNote: %v", err)
	}
	h := newJobNotesHandler(database)

	req := httptest.NewRequest(http.MethodDelete, "/notes/"+itoa(noteID), nil)
	req.SetPathValue("note_id", itoa(noteID))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestJobNotes_MethodNotAllowed(t *testing.T) {
	database := tempDB(t)
	h := newJobNotesHandler(database)

	req := httptest.NewRequest(http.MethodPut, "/runs/1/notes", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func itoa(n int64) string {
	return fmt.Sprintf("%d", n)
}
