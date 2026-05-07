package db_test

import (
	"database/sql"
	"testing"

	"github.com/your-org/cronwarden/internal/db"
)

func TestUpsertAndGetJobTimeoutPolicy(t *testing.T) {
	sqlDB := tempDB(t)

	policy := db.JobTimeoutPolicy{
		JobName:       "backup-job",
		TimeoutSec:    120,
		KillOnTimeout: true,
	}

	if err := db.UpsertJobTimeoutPolicy(sqlDB, policy); err != nil {
		t.Fatalf("UpsertJobTimeoutPolicy: %v", err)
	}

	got, err := db.GetJobTimeoutPolicy(sqlDB, "backup-job")
	if err != nil {
		t.Fatalf("GetJobTimeoutPolicy: %v", err)
	}

	if got.JobName != "backup-job" {
		t.Errorf("JobName: got %q, want %q", got.JobName, "backup-job")
	}
	if got.TimeoutSec != 120 {
		t.Errorf("TimeoutSec: got %d, want 120", got.TimeoutSec)
	}
	if !got.KillOnTimeout {
		t.Error("KillOnTimeout: expected true")
	}
}

func TestUpsertJobTimeoutPolicy_Replaces(t *testing.T) {
	sqlDB := tempDB(t)

	first := db.JobTimeoutPolicy{JobName: "sync-job", TimeoutSec: 60, KillOnTimeout: false}
	if err := db.UpsertJobTimeoutPolicy(sqlDB, first); err != nil {
		t.Fatalf("first upsert: %v", err)
	}

	updated := db.JobTimeoutPolicy{JobName: "sync-job", TimeoutSec: 300, KillOnTimeout: true}
	if err := db.UpsertJobTimeoutPolicy(sqlDB, updated); err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	got, err := db.GetJobTimeoutPolicy(sqlDB, "sync-job")
	if err != nil {
		t.Fatalf("GetJobTimeoutPolicy: %v", err)
	}
	if got.TimeoutSec != 300 {
		t.Errorf("TimeoutSec: got %d, want 300", got.TimeoutSec)
	}
	if !got.KillOnTimeout {
		t.Error("KillOnTimeout: expected true after update")
	}
}

func TestGetJobTimeoutPolicy_NotFound(t *testing.T) {
	sqlDB := tempDB(t)

	_, err := db.GetJobTimeoutPolicy(sqlDB, "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing policy, got nil")
	}
	if !isNoRowsTimeout(err) {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestDeleteJobTimeoutPolicy(t *testing.T) {
	sqlDB := tempDB(t)

	policy := db.JobTimeoutPolicy{JobName: "cleanup-job", TimeoutSec: 45, KillOnTimeout: false}
	if err := db.UpsertJobTimeoutPolicy(sqlDB, policy); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	if err := db.DeleteJobTimeoutPolicy(sqlDB, "cleanup-job"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, err := db.GetJobTimeoutPolicy(sqlDB, "cleanup-job")
	if !isNoRowsTimeout(err) {
		t.Errorf("expected ErrNoRows after delete, got %v", err)
	}
}

func TestListJobTimeoutPolicies(t *testing.T) {
	sqlDB := tempDB(t)

	empty, err := db.ListJobTimeoutPolicies(sqlDB)
	if err != nil {
		t.Fatalf("list empty: %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("expected 0 policies, got %d", len(empty))
	}

	for _, name := range []string{"job-a", "job-b", "job-c"} {
		p := db.JobTimeoutPolicy{JobName: name, TimeoutSec: 30, KillOnTimeout: false}
		if err := db.UpsertJobTimeoutPolicy(sqlDB, p); err != nil {
			t.Fatalf("upsert %s: %v", name, err)
		}
	}

	list, err := db.ListJobTimeoutPolicies(sqlDB)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("expected 3 policies, got %d", len(list))
	}
}

func isNoRowsTimeout(err error) bool {
	return err != nil && err.Error() == sql.ErrNoRows.Error()
}
