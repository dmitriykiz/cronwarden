package db_test

import (
	"testing"

	"github.com/cronwarden/cronwarden/internal/db"
)

func TestAddAndListJobDependencies(t *testing.T) {
	sqlDB := tempDB(t)

	if err := db.AddJobDependency(sqlDB, "backup", "cleanup", 3600); err != nil {
		t.Fatalf("AddJobDependency: %v", err)
	}
	if err := db.AddJobDependency(sqlDB, "backup", "sync", 7200); err != nil {
		t.Fatalf("AddJobDependency: %v", err)
	}

	deps, err := db.ListJobDependencies(sqlDB, "backup")
	if err != nil {
		t.Fatalf("ListJobDependencies: %v", err)
	}
	if len(deps) != 2 {
		t.Fatalf("expected 2 deps, got %d", len(deps))
	}
	if deps[0].DependsOn != "cleanup" || deps[1].DependsOn != "sync" {
		t.Errorf("unexpected deps: %+v", deps)
	}
}

func TestAddJobDependency_Replaces(t *testing.T) {
	sqlDB := tempDB(t)

	if err := db.AddJobDependency(sqlDB, "report", "fetch", 1800); err != nil {
		t.Fatalf("initial insert: %v", err)
	}
	// Upsert with new max_age_secs
	if err := db.AddJobDependency(sqlDB, "report", "fetch", 9000); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	deps, err := db.ListJobDependencies(sqlDB, "report")
	if err != nil {
		t.Fatalf("ListJobDependencies: %v", err)
	}
	if len(deps) != 1 {
		t.Fatalf("expected 1 dep after upsert, got %d", len(deps))
	}
	if deps[0].MaxAgeSecs != 9000 {
		t.Errorf("expected max_age_secs=9000, got %d", deps[0].MaxAgeSecs)
	}
}

func TestListJobDependencies_Empty(t *testing.T) {
	sqlDB := tempDB(t)

	deps, err := db.ListJobDependencies(sqlDB, "nonexistent")
	if err != nil {
		t.Fatalf("ListJobDependencies: %v", err)
	}
	if len(deps) != 0 {
		t.Errorf("expected 0 deps, got %d", len(deps))
	}
}

func TestDeleteJobDependency_Exists(t *testing.T) {
	sqlDB := tempDB(t)

	if err := db.AddJobDependency(sqlDB, "deploy", "test", 600); err != nil {
		t.Fatalf("AddJobDependency: %v", err)
	}
	deps, _ := db.ListJobDependencies(sqlDB, "deploy")
	if len(deps) == 0 {
		t.Fatal("expected at least one dep")
	}

	if err := db.DeleteJobDependency(sqlDB, deps[0].ID); err != nil {
		t.Fatalf("DeleteJobDependency: %v", err)
	}

	deps, _ = db.ListJobDependencies(sqlDB, "deploy")
	if len(deps) != 0 {
		t.Errorf("expected 0 deps after delete, got %d", len(deps))
	}
}

func TestDeleteJobDependency_NotExists(t *testing.T) {
	sqlDB := tempDB(t)

	err := db.DeleteJobDependency(sqlDB, 99999)
	if err == nil {
		t.Fatal("expected error for missing dep, got nil")
	}
}

func TestAddJobDependency_SelfReference(t *testing.T) {
	sqlDB := tempDB(t)

	err := db.AddJobDependency(sqlDB, "loop", "loop", 0)
	if err == nil {
		t.Fatal("expected error for self-referencing dependency")
	}
}
