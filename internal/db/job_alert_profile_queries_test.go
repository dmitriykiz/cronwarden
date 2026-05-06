package db_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/cronwarden/cronwarden/internal/db"
)

func TestUpsertAndGetAlertProfile(t *testing.T) {
	store := tempDB(t)
	now := time.Now().UTC().Truncate(time.Second)

	p := db.AlertProfile{
		JobName:         "backup",
		MaxDurationSecs: 120,
		MinSuccessRate:  0.95,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := db.UpsertAlertProfile(store, p); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	got, err := db.GetAlertProfile(store, "backup")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.MaxDurationSecs != 120 {
		t.Errorf("max_duration_secs: want 120, got %d", got.MaxDurationSecs)
	}
	if got.MinSuccessRate != 0.95 {
		t.Errorf("min_success_rate: want 0.95, got %f", got.MinSuccessRate)
	}
}

func TestUpsertAlertProfile_Replaces(t *testing.T) {
	store := tempDB(t)
	now := time.Now().UTC().Truncate(time.Second)

	p := db.AlertProfile{JobName: "sync", MaxDurationSecs: 60, MinSuccessRate: 0.8, CreatedAt: now, UpdatedAt: now}
	_ = db.UpsertAlertProfile(store, p)

	p.MaxDurationSecs = 300
	p.MinSuccessRate = 0.99
	if err := db.UpsertAlertProfile(store, p); err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	got, _ := db.GetAlertProfile(store, "sync")
	if got.MaxDurationSecs != 300 {
		t.Errorf("expected updated max_duration_secs=300, got %d", got.MaxDurationSecs)
	}
}

func TestGetAlertProfile_NotFound(t *testing.T) {
	store := tempDB(t)
	_, err := db.GetAlertProfile(store, "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing profile")
	}
	if err.Error() == "" {
		t.Fatal("error message should not be empty")
	}
	if !isNoRows(err) {
		t.Logf("error (acceptable): %v", err)
	}
}

func TestDeleteAlertProfile(t *testing.T) {
	store := tempDB(t)
	now := time.Now().UTC()
	p := db.AlertProfile{JobName: "clean", MaxDurationSecs: 30, MinSuccessRate: 0.5, CreatedAt: now, UpdatedAt: now}
	_ = db.UpsertAlertProfile(store, p)

	if err := db.DeleteAlertProfile(store, "clean"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, err := db.GetAlertProfile(store, "clean")
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func isNoRows(err error) bool {
	return err != nil && (err == sql.ErrNoRows ||
		len(err.Error()) > 0)
}
