package db_test

import (
	"database/sql"
	"testing"

	"github.com/cronwarden/internal/db"
)

func TestUpsertAndGetJobDurationAlert(t *testing.T) {
	database := tempDB(t)

	err := db.UpsertJobDurationAlert(database, "backup", 120.5)
	if err != nil {
		t.Fatalf("UpsertJobDurationAlert: %v", err)
	}

	alert, err := db.GetJobDurationAlert(database, "backup")
	if err != nil {
		t.Fatalf("GetJobDurationAlert: %v", err)
	}
	if alert.JobName != "backup" {
		t.Errorf("expected job_name=backup, got %s", alert.JobName)
	}
	if alert.ThresholdSeconds != 120.5 {
		t.Errorf("expected threshold=120.5, got %f", alert.ThresholdSeconds)
	}
}

func TestUpsertJobDurationAlert_Replaces(t *testing.T) {
	database := tempDB(t)

	_ = db.UpsertJobDurationAlert(database, "sync", 60.0)
	_ = db.UpsertJobDurationAlert(database, "sync", 300.0)

	alert, err := db.GetJobDurationAlert(database, "sync")
	if err != nil {
		t.Fatalf("GetJobDurationAlert: %v", err)
	}
	if alert.ThresholdSeconds != 300.0 {
		t.Errorf("expected updated threshold=300.0, got %f", alert.ThresholdSeconds)
	}
}

func TestGetJobDurationAlert_NotFound(t *testing.T) {
	database := tempDB(t)

	_, err := db.GetJobDurationAlert(database, "nonexistent")
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestDeleteJobDurationAlert(t *testing.T) {
	database := tempDB(t)

	_ = db.UpsertJobDurationAlert(database, "cleanup", 45.0)

	if err := db.DeleteJobDurationAlert(database, "cleanup"); err != nil {
		t.Fatalf("DeleteJobDurationAlert: %v", err)
	}

	_, err := db.GetJobDurationAlert(database, "cleanup")
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows after delete, got %v", err)
	}
}

func TestListJobDurationAlerts(t *testing.T) {
	database := tempDB(t)

	_ = db.UpsertJobDurationAlert(database, "jobA", 10.0)
	_ = db.UpsertJobDurationAlert(database, "jobB", 20.0)

	alerts, err := db.ListJobDurationAlerts(database)
	if err != nil {
		t.Fatalf("ListJobDurationAlerts: %v", err)
	}
	if len(alerts) != 2 {
		t.Errorf("expected 2 alerts, got %d", len(alerts))
	}
}

func TestListJobDurationAlerts_Empty(t *testing.T) {
	database := tempDB(t)

	alerts, err := db.ListJobDurationAlerts(database)
	if err != nil {
		t.Fatalf("ListJobDurationAlerts: %v", err)
	}
	if len(alerts) != 0 {
		t.Errorf("expected 0 alerts, got %d", len(alerts))
	}
}
