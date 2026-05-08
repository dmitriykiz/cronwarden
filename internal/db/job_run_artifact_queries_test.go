package db_test

import (
	"testing"

	"github.com/cronwarden/cronwarden/internal/db"
)

func seedArtifactRun(t *testing.T, d *db.DB) int64 {
	t.Helper()
	id, err := db.InsertJobRun(d.DB, "artifact-job", 0, "success", "")
	if err != nil {
		t.Fatalf("seed run: %v", err)
	}
	return id
}

func TestAddAndListJobRunArtifacts(t *testing.T) {
	d := tempDB(t)
	runID := seedArtifactRun(t, d)

	_, err := db.AddJobRunArtifact(d.DB, runID, "output.log", "/var/log/cron/output.log")
	if err != nil {
		t.Fatalf("add artifact: %v", err)
	}
	_, err = db.AddJobRunArtifact(d.DB, runID, "report.json", "/tmp/report.json")
	if err != nil {
		t.Fatalf("add artifact: %v", err)
	}

	artifacts, err := db.ListJobRunArtifacts(d.DB, runID)
	if err != nil {
		t.Fatalf("list artifacts: %v", err)
	}
	if len(artifacts) != 2 {
		t.Fatalf("expected 2 artifacts, got %d", len(artifacts))
	}
	if artifacts[0].Name != "output.log" {
		t.Errorf("expected name output.log, got %s", artifacts[0].Name)
	}
	if artifacts[1].Path != "/tmp/report.json" {
		t.Errorf("expected path /tmp/report.json, got %s", artifacts[1].Path)
	}
}

func TestListJobRunArtifacts_Empty(t *testing.T) {
	d := tempDB(t)
	artifacts, err := db.ListJobRunArtifacts(d.DB, 9999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(artifacts) != 0 {
		t.Errorf("expected 0 artifacts, got %d", len(artifacts))
	}
}

func TestDeleteJobRunArtifact_Exists(t *testing.T) {
	d := tempDB(t)
	runID := seedArtifactRun(t, d)

	id, err := db.AddJobRunArtifact(d.DB, runID, "out.txt", "/tmp/out.txt")
	if err != nil {
		t.Fatalf("add artifact: %v", err)
	}
	if err := db.DeleteJobRunArtifact(d.DB, id); err != nil {
		t.Fatalf("delete artifact: %v", err)
	}
	artifacts, _ := db.ListJobRunArtifacts(d.DB, runID)
	if len(artifacts) != 0 {
		t.Errorf("expected 0 artifacts after delete, got %d", len(artifacts))
	}
}

func TestDeleteJobRunArtifact_NotExists(t *testing.T) {
	d := tempDB(t)
	err := db.DeleteJobRunArtifact(d.DB, 99999)
	if err == nil {
		t.Fatal("expected error for missing artifact, got nil")
	}
}
