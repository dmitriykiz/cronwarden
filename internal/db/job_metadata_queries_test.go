package db_test

import (
	"testing"

	"github.com/cronwarden/cronwarden/internal/db"
)

func TestUpsertAndListJobMetadata(t *testing.T) {
	sqlDB := tempDB(t)

	if err := db.UpsertJobMetadata(sqlDB, "backup", "region", "us-east-1"); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if err := db.UpsertJobMetadata(sqlDB, "backup", "owner", "ops-team"); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	results, err := db.ListJobMetadata(sqlDB, "backup")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(results))
	}
	// Ordered by key ASC: owner < region
	if results[0].Key != "owner" || results[0].Value != "ops-team" {
		t.Errorf("unexpected first entry: %+v", results[0])
	}
	if results[1].Key != "region" || results[1].Value != "us-east-1" {
		t.Errorf("unexpected second entry: %+v", results[1])
	}
}

func TestUpsertJobMetadata_Replaces(t *testing.T) {
	sqlDB := tempDB(t)

	if err := db.UpsertJobMetadata(sqlDB, "sync", "env", "staging"); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if err := db.UpsertJobMetadata(sqlDB, "sync", "env", "production"); err != nil {
		t.Fatalf("upsert replace: %v", err)
	}

	results, err := db.ListJobMetadata(sqlDB, "sync")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 entry after replace, got %d", len(results))
	}
	if results[0].Value != "production" {
		t.Errorf("expected 'production', got %q", results[0].Value)
	}
}

func TestListJobMetadata_Empty(t *testing.T) {
	sqlDB := tempDB(t)

	results, err := db.ListJobMetadata(sqlDB, "nonexistent")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty list, got %d entries", len(results))
	}
}

func TestDeleteJobMetadata_Exists(t *testing.T) {
	sqlDB := tempDB(t)

	_ = db.UpsertJobMetadata(sqlDB, "report", "format", "csv")

	if err := db.DeleteJobMetadata(sqlDB, "report", "format"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	results, _ := db.ListJobMetadata(sqlDB, "report")
	if len(results) != 0 {
		t.Errorf("expected empty after delete, got %d", len(results))
	}
}

func TestDeleteJobMetadata_NotExists(t *testing.T) {
	sqlDB := tempDB(t)

	err := db.DeleteJobMetadata(sqlDB, "ghost", "missing-key")
	if err == nil {
		t.Fatal("expected error for missing key, got nil")
	}
}

func TestDeleteAllJobMetadata(t *testing.T) {
	sqlDB := tempDB(t)

	_ = db.UpsertJobMetadata(sqlDB, "cleanup", "a", "1")
	_ = db.UpsertJobMetadata(sqlDB, "cleanup", "b", "2")

	if err := db.DeleteAllJobMetadata(sqlDB, "cleanup"); err != nil {
		t.Fatalf("delete all: %v", err)
	}

	results, _ := db.ListJobMetadata(sqlDB, "cleanup")
	if len(results) != 0 {
		t.Errorf("expected empty after delete all, got %d", len(results))
	}
}
