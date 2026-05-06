package db_test

import (
	"testing"
	"time"

	"github.com/cronwarden/cronwarden/internal/db"
)

func seedNoteRun(t *testing.T, database *db.DB) int64 {
	t.Helper()
	id, err := db.InsertJobRun(database.SQL(), db.JobRun{
		JobName:   "note-job",
		Status:    "success",
		StartedAt: time.Now().UTC(),
		Duration:  1.0,
	})
	if err != nil {
		t.Fatalf("seed run: %v", err)
	}
	return id
}

func TestAddAndListJobNotes(t *testing.T) {
	database := tempDB(t)
	runID := seedNoteRun(t, database)

	_, err := db.AddJobNote(database.SQL(), runID, "first note")
	if err != nil {
		t.Fatalf("AddJobNote: %v", err)
	}
	_, err = db.AddJobNote(database.SQL(), runID, "second note")
	if err != nil {
		t.Fatalf("AddJobNote: %v", err)
	}

	notes, err := db.ListJobNotes(database.SQL(), runID)
	if err != nil {
		t.Fatalf("ListJobNotes: %v", err)
	}
	if len(notes) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(notes))
	}
	if notes[0].Note != "first note" {
		t.Errorf("expected 'first note', got %q", notes[0].Note)
	}
}

func TestListJobNotes_Empty(t *testing.T) {
	database := tempDB(t)
	notes, err := db.ListJobNotes(database.SQL(), 9999)
	if err != nil {
		t.Fatalf("ListJobNotes: %v", err)
	}
	if len(notes) != 0 {
		t.Errorf("expected empty slice, got %d", len(notes))
	}
}

func TestDeleteJobNote_Exists(t *testing.T) {
	database := tempDB(t)
	runID := seedNoteRun(t, database)

	noteID, err := db.AddJobNote(database.SQL(), runID, "to delete")
	if err != nil {
		t.Fatalf("AddJobNote: %v", err)
	}

	found, err := db.DeleteJobNote(database.SQL(), noteID)
	if err != nil {
		t.Fatalf("DeleteJobNote: %v", err)
	}
	if !found {
		t.Error("expected found=true")
	}

	notes, _ := db.ListJobNotes(database.SQL(), runID)
	if len(notes) != 0 {
		t.Errorf("expected 0 notes after delete, got %d", len(notes))
	}
}

func TestDeleteJobNote_NotExists(t *testing.T) {
	database := tempDB(t)
	found, err := db.DeleteJobNote(database.SQL(), 9999)
	if err != nil {
		t.Fatalf("DeleteJobNote: %v", err)
	}
	if found {
		t.Error("expected found=false for missing note")
	}
}
