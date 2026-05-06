package db_test

import (
	"testing"

	"github.com/cronwarden/cronwarden/internal/db"
)

func TestSetAndListJobLabels(t *testing.T) {
	sqlDB := tempDB(t)

	labels := map[string]string{"env": "prod", "team": "platform"}
	if err := db.SetJobLabels(sqlDB, "backup", labels); err != nil {
		t.Fatalf("SetJobLabels: %v", err)
	}

	got, err := db.ListJobLabels(sqlDB, "backup")
	if err != nil {
		t.Fatalf("ListJobLabels: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 labels, got %d", len(got))
	}
	if got["env"] != "prod" || got["team"] != "platform" {
		t.Errorf("unexpected labels: %v", got)
	}
}

func TestSetJobLabels_Replaces(t *testing.T) {
	sqlDB := tempDB(t)

	_ = db.SetJobLabels(sqlDB, "backup", map[string]string{"env": "staging"})
	_ = db.SetJobLabels(sqlDB, "backup", map[string]string{"env": "prod", "owner": "alice"})

	got, err := db.ListJobLabels(sqlDB, "backup")
	if err != nil {
		t.Fatalf("ListJobLabels: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 labels after replace, got %d", len(got))
	}
	if got["env"] != "prod" {
		t.Errorf("expected env=prod, got %q", got["env"])
	}
}

func TestListJobLabels_Empty(t *testing.T) {
	sqlDB := tempDB(t)

	got, err := db.ListJobLabels(sqlDB, "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty map, got %v", got)
	}
}

func TestDeleteJobLabels(t *testing.T) {
	sqlDB := tempDB(t)

	_ = db.SetJobLabels(sqlDB, "backup", map[string]string{"env": "prod"})
	if err := db.DeleteJobLabels(sqlDB, "backup"); err != nil {
		t.Fatalf("DeleteJobLabels: %v", err)
	}

	got, _ := db.ListJobLabels(sqlDB, "backup")
	if len(got) != 0 {
		t.Errorf("expected empty after delete, got %v", got)
	}
}
