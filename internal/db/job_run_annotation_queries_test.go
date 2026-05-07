package db_test

import (
	"testing"
	"time"

	"github.com/user/cronwarden/internal/db"
)

func seedAnnotationRun(t *testing.T, database *db.DB) int64 {
	t.Helper()
	err := db.InsertJobRun(database.DB, db.JobRun{
		JobName:   "annotated-job",
		StartedAt: time.Now().UTC(),
		Success:   true,
	})
	if err != nil {
		t.Fatalf("seed run: %v", err)
	}
	runs, err := db.ListJobRuns(database.DB, 1)
	if err != nil {
		t.Fatalf("list runs: %v", err)
	}
	if len(runs) == 0 {
		t.Fatal("expected at least one run")
	}
	return runs[0].ID
}

func TestUpsertAndListJobRunAnnotations(t *testing.T) {
	database := tempDB(t)
	runID := seedAnnotationRun(t, database)

	if err := db.UpsertJobRunAnnotation(database.DB, runID, "env", "production"); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if err := db.UpsertJobRunAnnotation(database.DB, runID, "region", "us-east-1"); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	anns, err := db.ListJobRunAnnotations(database.DB, runID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(anns) != 2 {
		t.Fatalf("expected 2 annotations, got %d", len(anns))
	}
}

func TestUpsertJobRunAnnotation_Replaces(t *testing.T) {
	database := tempDB(t)
	runID := seedAnnotationRun(t, database)

	_ = db.UpsertJobRunAnnotation(database.DB, runID, "env", "staging")
	_ = db.UpsertJobRunAnnotation(database.DB, runID, "env", "production")

	anns, _ := db.ListJobRunAnnotations(database.DB, runID)
	if len(anns) != 1 {
		t.Fatalf("expected 1 annotation after upsert, got %d", len(anns))
	}
	if anns[0].Value != "production" {
		t.Errorf("expected value 'production', got %q", anns[0].Value)
	}
}

func TestDeleteJobRunAnnotation_Exists(t *testing.T) {
	database := tempDB(t)
	runID := seedAnnotationRun(t, database)

	_ = db.UpsertJobRunAnnotation(database.DB, runID, "key", "val")
	anns, _ := db.ListJobRunAnnotations(database.DB, runID)
	if len(anns) == 0 {
		t.Fatal("expected annotation")
	}

	if err := db.DeleteJobRunAnnotation(database.DB, anns[0].ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	anns, _ = db.ListJobRunAnnotations(database.DB, runID)
	if len(anns) != 0 {
		t.Errorf("expected 0 annotations after delete, got %d", len(anns))
	}
}

func TestDeleteJobRunAnnotation_NotExists(t *testing.T) {
	database := tempDB(t)
	err := db.DeleteJobRunAnnotation(database.DB, 99999)
	if err == nil {
		t.Error("expected error for non-existent annotation")
	}
}
