package db_test

import (
	"testing"
	"time"

	"github.com/cronwarden/cronwarden/internal/db"
)

func TestComputeGlobalStats_Empty(t *testing.T) {
	conn := tempDB(t)
	s, err := db.ComputeGlobalStats(conn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.TotalRuns != 0 || s.UniqueJobs != 0 {
		t.Errorf("expected zero stats, got %+v", s)
	}
	if s.LastRunAt != nil {
		t.Errorf("expected nil LastRunAt, got %v", s.LastRunAt)
	}
}

func TestComputeGlobalStats_Mixed(t *testing.T) {
	conn := tempDB(t)
	now := time.Now().UTC().Truncate(time.Second)

	runs := []db.JobRun{
		{JobName: "backup", StartedAt: now.Add(-10 * time.Minute), ExitCode: 0, Output: "ok"},
		{JobName: "backup", StartedAt: now.Add(-5 * time.Minute), ExitCode: 1, Output: "fail"},
		{JobName: "cleanup", StartedAt: now, ExitCode: 0, Output: "done"},
	}
	for _, r := range runs {
		if err := db.InsertJobRun(conn, r); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}

	s, err := db.ComputeGlobalStats(conn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.TotalRuns != 3 {
		t.Errorf("TotalRuns: want 3, got %d", s.TotalRuns)
	}
	if s.SuccessRuns != 2 {
		t.Errorf("SuccessRuns: want 2, got %d", s.SuccessRuns)
	}
	if s.FailureRuns != 1 {
		t.Errorf("FailureRuns: want 1, got %d", s.FailureRuns)
	}
	if s.UniqueJobs != 2 {
		t.Errorf("UniqueJobs: want 2, got %d", s.UniqueJobs)
	}
	if s.LastRunAt == nil {
		t.Fatal("expected non-nil LastRunAt")
	}
}
