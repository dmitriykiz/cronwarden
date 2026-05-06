package db_test

import (
	"testing"
	"time"

	"github.com/user/cronwarden/internal/db"
)

func TestGetJobRun_Found(t *testing.T) {
	sqldb := tempDB(t)

	err := db.InsertJobRun(sqldb, db.JobRun{
		JobName:   "backup",
		StartedAt: time.Now().UTC().Truncate(time.Second),
		Duration:  123.4,
		Success:   true,
		Output:    "ok",
	})
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	runs, err := db.ListJobRuns(sqldb, 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(runs) == 0 {
		t.Fatal("expected at least one run")
	}

	got, err := db.GetJobRun(sqldb, runs[0].ID)
	if err != nil {
		t.Fatalf("GetJobRun: %v", err)
	}
	if got == nil {
		t.Fatal("expected run, got nil")
	}
	if got.JobName != "backup" {
		t.Errorf("job name: want backup, got %s", got.JobName)
	}
	if !got.Success {
		t.Error("expected success=true")
	}
}

func TestGetJobRun_NotFound(t *testing.T) {
	sqldb := tempDB(t)

	got, err := db.GetJobRun(sqldb, 9999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestDeleteJobRun_Exists(t *testing.T) {
	sqldb := tempDB(t)

	_ = db.InsertJobRun(sqldb, db.JobRun{
		JobName:   "cleanup",
		StartedAt: time.Now().UTC(),
		Duration:  50,
		Success:   false,
		Output:    "err",
	})

	runs, _ := db.ListJobRuns(sqldb, 10)
	if len(runs) == 0 {
		t.Fatal("expected a run")
	}
	id := runs[0].ID

	deleted, err := db.DeleteJobRun(sqldb, id)
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if !deleted {
		t.Error("expected deleted=true")
	}

	got, _ := db.GetJobRun(sqldb, id)
	if got != nil {
		t.Error("expected run to be gone")
	}
}

func TestDeleteJobRun_NotExists(t *testing.T) {
	sqldb := tempDB(t)

	deleted, err := db.DeleteJobRun(sqldb, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted {
		t.Error("expected deleted=false for missing row")
	}
}
