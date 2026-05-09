package db_test

import (
	"testing"
	"time"

	"github.com/cronwarden/cronwarden/internal/db"
)

func seedHeartbeatAPIRun(t *testing.T, database *db.DB, jobName string) int64 {
	t.Helper()
	id, err := database.InsertJobRun(jobName, "success", 1*time.Second, time.Now())
	if err != nil {
		t.Fatalf("InsertJobRun: %v", err)
	}
	return id
}

func TestListActiveHeartbeats_Empty(t *testing.T) {
	database := tempDB(t)

	results, err := db.ListActiveHeartbeats(database.DB)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestListActiveHeartbeats_ReturnsOnlyActive(t *testing.T) {
	database := tempDB(t)

	run1 := seedHeartbeatAPIRun(t, database, "active-job")
	run2 := seedHeartbeatAPIRun(t, database, "expired-job")

	now := time.Now().UTC()

	// active heartbeat — expires in the future
	_, err := db.UpsertJobRunHeartbeat(database.DB, run1, "running", now, now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("UpsertJobRunHeartbeat active: %v", err)
	}

	// expired heartbeat — expires in the past
	_, err = db.UpsertJobRunHeartbeat(database.DB, run2, "running", now.Add(-2*time.Hour), now.Add(-1*time.Hour))
	if err != nil {
		t.Fatalf("UpsertJobRunHeartbeat expired: %v", err)
	}

	results, err := db.ListActiveHeartbeats(database.DB)
	if err != nil {
		t.Fatalf("ListActiveHeartbeats: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 active heartbeat, got %d", len(results))
	}
	if results[0].RunID != run1 {
		t.Errorf("expected run_id %d, got %d", run1, results[0].RunID)
	}
	if results[0].JobName != "active-job" {
		t.Errorf("expected job_name 'active-job', got %q", results[0].JobName)
	}
	if results[0].Status != "running" {
		t.Errorf("expected status 'running', got %q", results[0].Status)
	}
}
