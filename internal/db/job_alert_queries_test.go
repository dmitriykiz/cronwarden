package db_test

import (
	"testing"

	"github.com/cronwarden/cronwarden/internal/db"
)

func TestUpsertAndListJobAlerts(t *testing.T) {
	sqldb := tempDB(t)

	if err := db.UpsertJobAlert(sqldb, "backup", "failure", 0); err != nil {
		t.Fatalf("UpsertJobAlert: %v", err)
	}
	if err := db.UpsertJobAlert(sqldb, "backup", "duration", 300); err != nil {
		t.Fatalf("UpsertJobAlert: %v", err)
	}

	alerts, err := db.ListJobAlerts(sqldb, "backup")
	if err != nil {
		t.Fatalf("ListJobAlerts: %v", err)
	}
	if len(alerts) != 2 {
		t.Fatalf("expected 2 alerts, got %d", len(alerts))
	}
}

func TestUpsertJobAlerts_Replaces(t *testing.T) {
	sqldb := tempDB(t)

	if err := db.UpsertJobAlert(sqldb, "sync", "duration", 60); err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	if err := db.UpsertJobAlert(sqldb, "sync", "duration", 120); err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	alerts, err := db.ListJobAlerts(sqldb, "sync")
	if err != nil {
		t.Fatalf("ListJobAlerts: %v", err)
	}
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert after upsert, got %d", len(alerts))
	}
	if alerts[0].Threshold != 120 {
		t.Errorf("expected threshold 120, got %v", alerts[0].Threshold)
	}
}

func TestListJobAlerts_Empty(t *testing.T) {
	sqldb := tempDB(t)

	alerts, err := db.ListJobAlerts(sqldb, "nonexistent")
	if err != nil {
		t.Fatalf("ListJobAlerts: %v", err)
	}
	if len(alerts) != 0 {
		t.Errorf("expected 0 alerts, got %d", len(alerts))
	}
}

func TestDeleteJobAlert_Exists(t *testing.T) {
	sqldb := tempDB(t)

	if err := db.UpsertJobAlert(sqldb, "report", "failure", 0); err != nil {
		t.Fatalf("UpsertJobAlert: %v", err)
	}
	alerts, _ := db.ListJobAlerts(sqldb, "report")
	if len(alerts) == 0 {
		t.Fatal("expected at least one alert")
	}

	ok, err := db.DeleteJobAlert(sqldb, alerts[0].ID)
	if err != nil {
		t.Fatalf("DeleteJobAlert: %v", err)
	}
	if !ok {
		t.Error("expected deleted=true")
	}

	alerts, _ = db.ListJobAlerts(sqldb, "report")
	if len(alerts) != 0 {
		t.Errorf("expected 0 alerts after delete, got %d", len(alerts))
	}
}

func TestDeleteJobAlert_NotExists(t *testing.T) {
	sqldb := tempDB(t)

	ok, err := db.DeleteJobAlert(sqldb, 9999)
	if err != nil {
		t.Fatalf("DeleteJobAlert: %v", err)
	}
	if ok {
		t.Error("expected deleted=false for non-existent id")
	}
}
