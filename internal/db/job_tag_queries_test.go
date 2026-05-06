package db_test

import (
	"testing"

	"github.com/example/cronwarden/internal/db"
)

func TestSetAndListJobTags(t *testing.T) {
	sqlDB := tempDB(t)

	err := db.SetJobTags(sqlDB, "backup", []string{"infra", "nightly"})
	if err != nil {
		t.Fatalf("SetJobTags: %v", err)
	}

	tags, err := db.ListJobTags(sqlDB, "backup")
	if err != nil {
		t.Fatalf("ListJobTags: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}
	if tags[0] != "infra" || tags[1] != "nightly" {
		t.Errorf("unexpected tags: %v", tags)
	}
}

func TestSetJobTags_Replaces(t *testing.T) {
	sqlDB := tempDB(t)

	_ = db.SetJobTags(sqlDB, "backup", []string{"old", "tags"})
	_ = db.SetJobTags(sqlDB, "backup", []string{"new"})

	tags, err := db.ListJobTags(sqlDB, "backup")
	if err != nil {
		t.Fatalf("ListJobTags: %v", err)
	}
	if len(tags) != 1 || tags[0] != "new" {
		t.Errorf("expected [new], got %v", tags)
	}
}

func TestListJobTags_Empty(t *testing.T) {
	sqlDB := tempDB(t)

	tags, err := db.ListJobTags(sqlDB, "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tags) != 0 {
		t.Errorf("expected empty, got %v", tags)
	}
}

func TestDeleteJobTags(t *testing.T) {
	sqlDB := tempDB(t)

	_ = db.SetJobTags(sqlDB, "deploy", []string{"prod", "critical"})
	if err := db.DeleteJobTags(sqlDB, "deploy"); err != nil {
		t.Fatalf("DeleteJobTags: %v", err)
	}

	tags, _ := db.ListJobTags(sqlDB, "deploy")
	if len(tags) != 0 {
		t.Errorf("expected no tags after delete, got %v", tags)
	}
}
