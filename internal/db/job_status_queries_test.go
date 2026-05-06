package db_test

import (
	"testing"
	"time"

	"github.com/example/cronwarden/internal/db"
)

func TestListJobStatuses_Empty(t *testing.T) {
	conn := tempDB(t)
	statuses, err := db.ListJobStatuses(conn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(statuses) != 0 {
		t.Fatalf("expected 0 statuses, got %d", len(statuses))
	}
}

func TestListJobStatuses_ReturnsLatestPerJob(t *testing.T) {
	conn := tempDB(t)

	now := time.Now().UTC()

	runs := []db.JobRun{
		{Name: "alpha", StartedAt: now.Add(-2 * time.Hour), ExitCode: 0},
		{Name: "alpha", StartedAt: now.Add(-1 * time.Hour), ExitCode: 1, Error: "timeout"},
		{Name: "beta", StartedAt: now.Add(-30 * time.Minute), ExitCode: 0},
	}
	for _, r := range runs {
		if err := db.InsertJobRun(conn, r); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}

	statuses, err := db.ListJobStatuses(conn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(statuses) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(statuses))
	}

	// alpha should reflect its most recent (failed) run
	alpha := statuses[0]
	if alpha.Name != "alpha" {
		t.Errorf("expected alpha first, got %s", alpha.Name)
	}
	if alpha.LastExit != 1 {
		t.Errorf("expected exit 1, got %d", alpha.LastExit)
	}
	if alpha.LastError != "timeout" {
		t.Errorf("expected error 'timeout', got %q", alpha.LastError)
	}
	if alpha.RunCount != 2 {
		t.Errorf("expected run_count 2, got %d", alpha.RunCount)
	}

	beta := statuses[1]
	if beta.Name != "beta" {
		t.Errorf("expected beta second, got %s", beta.Name)
	}
	if beta.LastExit != 0 {
		t.Errorf("expected exit 0, got %d", beta.LastExit)
	}
	if beta.RunCount != 1 {
		t.Errorf("expected run_count 1, got %d", beta.RunCount)
	}
}
