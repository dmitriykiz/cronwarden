package db_test

import (
	"testing"

	"github.com/cronwarden/cronwarden/internal/db"
)

func TestUpsertAndListJobEnvVars(t *testing.T) {
	conn := tempDB(t)

	if err := db.UpsertJobEnvVar(conn, "backup", "S3_BUCKET", "my-bucket"); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if err := db.UpsertJobEnvVar(conn, "backup", "AWS_REGION", "us-east-1"); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	vars, err := db.ListJobEnvVars(conn, "backup")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(vars) != 2 {
		t.Fatalf("expected 2 vars, got %d", len(vars))
	}
	if vars["S3_BUCKET"] != "my-bucket" {
		t.Errorf("expected my-bucket, got %s", vars["S3_BUCKET"])
	}
	if vars["AWS_REGION"] != "us-east-1" {
		t.Errorf("expected us-east-1, got %s", vars["AWS_REGION"])
	}
}

func TestUpsertJobEnvVar_Replaces(t *testing.T) {
	conn := tempDB(t)

	db.UpsertJobEnvVar(conn, "myjob", "KEY", "old")
	db.UpsertJobEnvVar(conn, "myjob", "KEY", "new")

	vars, err := db.ListJobEnvVars(conn, "myjob")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if vars["KEY"] != "new" {
		t.Errorf("expected new, got %s", vars["KEY"])
	}
}

func TestListJobEnvVars_Empty(t *testing.T) {
	conn := tempDB(t)
	vars, err := db.ListJobEnvVars(conn, "nonexistent")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(vars) != 0 {
		t.Errorf("expected empty map, got %d entries", len(vars))
	}
}

func TestDeleteJobEnvVar(t *testing.T) {
	conn := tempDB(t)

	db.UpsertJobEnvVar(conn, "myjob", "A", "1")
	db.UpsertJobEnvVar(conn, "myjob", "B", "2")

	if err := db.DeleteJobEnvVar(conn, "myjob", "A"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	vars, _ := db.ListJobEnvVars(conn, "myjob")
	if _, ok := vars["A"]; ok {
		t.Error("expected A to be deleted")
	}
	if vars["B"] != "2" {
		t.Error("expected B to remain")
	}
}

func TestDeleteAllJobEnvVars(t *testing.T) {
	conn := tempDB(t)

	db.UpsertJobEnvVar(conn, "myjob", "A", "1")
	db.UpsertJobEnvVar(conn, "myjob", "B", "2")

	if err := db.DeleteAllJobEnvVars(conn, "myjob"); err != nil {
		t.Fatalf("delete all: %v", err)
	}

	vars, _ := db.ListJobEnvVars(conn, "myjob")
	if len(vars) != 0 {
		t.Errorf("expected empty, got %d", len(vars))
	}
}
