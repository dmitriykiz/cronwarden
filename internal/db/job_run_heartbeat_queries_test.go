package db_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/cronwarden/cronwarden/internal/db"
)

func seedHeartbeatRun(t *testing.T, database *sql.DB, name string) int64 {
	t.Helper()
	err := db.InsertJobRun(database, name, "success", 1*time.Second)
	if err != nil {
		t.Fatalf("InsertJobRun: %v", err)
	}
	runs, err := db.ListJobRuns(database, 1)
	if err != nil {
		t.Fatalf("ListJobRuns: %v", err)
	}
	if len(runs) == 0 {
		t.Fatal("expected at least one run")
	}
	return runs[0].ID
}

func TestUpsertAndGetJobRunHeartbeat(t *testing.T) {
	database := tempDB(t)
	runID := seedHeartbeatRun(t, database, "heartbeat-job")

	now := time.Now().UTC().Truncate(time.Second)
	if err := db.UpsertJobRunHeartbeat(database, runID, "heartbeat-job", now); err != nil {
		t.Fatalf("UpsertJobRunHeartbeat: %v", err)
	}

	h, err := db.GetJobRunHeartbeat(database, runID)
	if err != nil {
		t.Fatalf("GetJobRunHeartbeat: %v", err)
	}
	if h.RunID != runID {
		t.Errorf("expected run_id %d, got %d", runID, h.RunID)
	}
	if h.JobName != "heartbeat-job" {
		t.Errorf("expected job_name heartbeat-job, got %s", h.JobName)
	}
	if !h.BeatAt.Equal(now) {
		t.Errorf("expected beat_at %v, got %v", now, h.BeatAt)
	}
}

func TestUpsertJobRunHeartbeat_Replaces(t *testing.T) {
	database := tempDB(t)
	runID := seedHeartbeatRun(t, database, "hb-replace-job")

	first := time.Now().UTC().Add(-10 * time.Minute).Truncate(time.Second)
	second := time.Now().UTC().Truncate(time.Second)

	_ = db.UpsertJobRunHeartbeat(database, runID, "hb-replace-job", first)
	_ = db.UpsertJobRunHeartbeat(database, runID, "hb-replace-job", second)

	h, err := db.GetJobRunHeartbeat(database, runID)
	if err != nil {
		t.Fatalf("GetJobRunHeartbeat: %v", err)
	}
	if !h.BeatAt.Equal(second) {
		t.Errorf("expected updated beat_at %v, got %v", second, h.BeatAt)
	}
}

func TestGetJobRunHeartbeat_NotFound(t *testing.T) {
	database := tempDB(t)
	_, err := db.GetJobRunHeartbeat(database, 9999)
	if err == nil {
		t.Fatal("expected error for missing heartbeat, got nil")
	}
}

func TestDeleteJobRunHeartbeat(t *testing.T) {
	database := tempDB(t)
	runID := seedHeartbeatRun(t, database, "hb-delete-job")

	_ = db.UpsertJobRunHeartbeat(database, runID, "hb-delete-job", time.Now().UTC())
	if err := db.DeleteJobRunHeartbeat(database, runID); err != nil {
		t.Fatalf("DeleteJobRunHeartbeat: %v", err)
	}

	_, err := db.GetJobRunHeartbeat(database, runID)
	if err == nil {
		t.Fatal("expected error after deletion, got nil")
	}
}

func TestListStaleHeartbeats(t *testing.T) {
	database := tempDB(t)

	old := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Second)
	recent := time.Now().UTC().Truncate(time.Second)

	run1 := seedHeartbeatRun(t, database, "stale-job-1")
	run2 := seedHeartbeatRun(t, database, "fresh-job-1")

	_ = db.UpsertJobRunHeartbeat(database, run1, "stale-job-1", old)
	_ = db.UpsertJobRunHeartbeat(database, run2, "fresh-job-1", recent)

	threshold := time.Now().UTC().Add(-1 * time.Hour)
	stale, err := db.ListStaleHeartbeats(database, threshold)
	if err != nil {
		t.Fatalf("ListStaleHeartbeats: %v", err)
	}
	if len(stale) != 1 {
		t.Fatalf("expected 1 stale heartbeat, got %d", len(stale))
	}
	if stale[0].JobName != "stale-job-1" {
		t.Errorf("expected stale-job-1, got %s", stale[0].JobName)
	}
}
