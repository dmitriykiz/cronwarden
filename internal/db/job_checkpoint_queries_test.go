package db_test

import (
	"database/sql"
	"testing"

	"github.com/cronwarden/cronwarden/internal/db"
)

func TestUpsertAndGetJobCheckpoint(t *testing.T) {
	conn := tempDB(t)

	if err := db.UpsertJobCheckpoint(conn, "backup", "row-42"); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	cp, err := db.GetJobCheckpoint(conn, "backup")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if cp.JobName != "backup" {
		t.Errorf("job_name: want backup, got %s", cp.JobName)
	}
	if cp.Marker != "row-42" {
		t.Errorf("marker: want row-42, got %s", cp.Marker)
	}
}

func TestUpsertJobCheckpoint_Replaces(t *testing.T) {
	conn := tempDB(t)

	_ = db.UpsertJobCheckpoint(conn, "sync", "v1")
	_ = db.UpsertJobCheckpoint(conn, "sync", "v2")

	cp, err := db.GetJobCheckpoint(conn, "sync")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if cp.Marker != "v2" {
		t.Errorf("expected marker v2, got %s", cp.Marker)
	}
}

func TestGetJobCheckpoint_NotFound(t *testing.T) {
	conn := tempDB(t)

	_, err := db.GetJobCheckpoint(conn, "nonexistent")
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestDeleteJobCheckpoint(t *testing.T) {
	conn := tempDB(t)

	_ = db.UpsertJobCheckpoint(conn, "cleanup", "done")
	if err := db.DeleteJobCheckpoint(conn, "cleanup"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, err := db.GetJobCheckpoint(conn, "cleanup")
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows after delete, got %v", err)
	}
}
