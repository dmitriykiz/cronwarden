package db_test

import (
	"testing"
	"time"

	"github.com/cronwarden/cronwarden/internal/db"
)

func seedHeartbeatAlertRun(t *testing.T, database *db.DB, jobName string) int64 {
	t.Helper()
	id, err := database.InsertJobRun(jobName, "success", 0, time.Now())
	if err != nil {
		t.Fatalf("InsertJobRun: %v", err)
	}
	return id
}

func TestListExpiredHeartbeatAlerts_Empty(t *testing.T) {
	database := tempDB(t)
	alerts, err := db.ListExpiredHeartbeatAlerts(database.DB, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 0 {
		t.Fatalf("expected 0 alerts, got %d", len(alerts))
	}
}

func TestListExpiredHeartbeatAlerts_ReturnsExpired(t *testing.T) {
	database := tempDB(t)
	runID := seedHeartbeatAlertRun(t, database, "backup")

	now := time.Now().UTC()
	expiredAt := now.Add(-1 * time.Minute)

	err := db.UpsertJobRunHeartbeat(database.DB, runID, now.Add(-5*time.Minute), expiredAt)
	if err != nil {
		t.Fatalf("UpsertJobRunHeartbeat: %v", err)
	}

	alerts, err := db.ListExpiredHeartbeatAlerts(database.DB, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].JobName != "backup" {
		t.Errorf("expected job 'backup', got %q", alerts[0].JobName)
	}
	if alerts[0].RunID != runID {
		t.Errorf("expected run id %d, got %d", runID, alerts[0].RunID)
	}
}

func TestListExpiredHeartbeatAlerts_SkipsFuture(t *testing.T) {
	database := tempDB(t)
	runID := seedHeartbeatAlertRun(t, database, "cleanup")

	now := time.Now().UTC()
	futureTimeout := now.Add(10 * time.Minute)

	err := db.UpsertJobRunHeartbeat(database.DB, runID, now, futureTimeout)
	if err != nil {
		t.Fatalf("UpsertJobRunHeartbeat: %v", err)
	}

	alerts, err := db.ListExpiredHeartbeatAlerts(database.DB, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 0 {
		t.Fatalf("expected 0 alerts, got %d", len(alerts))
	}
}
