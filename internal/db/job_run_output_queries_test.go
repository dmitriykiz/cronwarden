package db_test

import (
	"testing"
	"time"

	"github.com/cronwarden/cronwarden/internal/db"
)

func seedOutputRun(t *testing.T, database *db.DB) int64 {
	t.Helper()
	id, err := database.InsertJobRun("output-job", true, 0, time.Now().UTC())
	if err != nil {
		t.Fatalf("seed run: %v", err)
	}
	return id
}

func TestUpsertAndGetJobRunOutput(t *testing.T) {
	database := tempDB(t)
	runID := seedOutputRun(t, database)

	err := db.UpsertJobRunOutput(database.DB, runID, "hello stdout", "hello stderr")
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}

	out, err := db.GetJobRunOutput(database.DB, runID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if out == nil {
		t.Fatal("expected output, got nil")
	}
	if out.Stdout != "hello stdout" {
		t.Errorf("stdout = %q, want %q", out.Stdout, "hello stdout")
	}
	if out.Stderr != "hello stderr" {
		t.Errorf("stderr = %q, want %q", out.Stderr, "hello stderr")
	}
}

func TestUpsertJobRunOutput_Replaces(t *testing.T) {
	database := tempDB(t)
	runID := seedOutputRun(t, database)

	_ = db.UpsertJobRunOutput(database.DB, runID, "first", "")
	_ = db.UpsertJobRunOutput(database.DB, runID, "second", "err2")

	out, err := db.GetJobRunOutput(database.DB, runID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if out.Stdout != "second" {
		t.Errorf("stdout = %q, want %q", out.Stdout, "second")
	}
}

func TestGetJobRunOutput_NotFound(t *testing.T) {
	database := tempDB(t)

	out, err := db.GetJobRunOutput(database.DB, 9999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil, got %+v", out)
	}
}

func TestDeleteJobRunOutput(t *testing.T) {
	database := tempDB(t)
	runID := seedOutputRun(t, database)

	_ = db.UpsertJobRunOutput(database.DB, runID, "data", "")
	if err := db.DeleteJobRunOutput(database.DB, runID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	out, err := db.GetJobRunOutput(database.DB, runID)
	if err != nil {
		t.Fatalf("get after delete: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil after delete, got %+v", out)
	}
}
