package db_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/cronwarden/cronwarden/internal/db"
)

func TestUpsertAndGetJobConcurrencyPolicy(t *testing.T) {
	database := tempDB(t)

	p := db.JobConcurrencyPolicy{
		JobName:   "backup",
		Policy:    "skip",
		MaxQueue:  0,
		UpdatedAt: time.Now().UTC().Truncate(time.Second),
	}
	if err := db.UpsertJobConcurrencyPolicy(database, p); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	got, err := db.GetJobConcurrencyPolicy(database, "backup")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Policy != "skip" {
		t.Errorf("policy: got %q, want %q", got.Policy, "skip")
	}
	if got.MaxQueue != 0 {
		t.Errorf("max_queue: got %d, want 0", got.MaxQueue)
	}
}

func TestUpsertJobConcurrencyPolicy_Replaces(t *testing.T) {
	database := tempDB(t)

	upsert := func(policy string, maxQueue int) {
		err := db.UpsertJobConcurrencyPolicy(database, db.JobConcurrencyPolicy{
			JobName:   "sync",
			Policy:    policy,
			MaxQueue:  maxQueue,
			UpdatedAt: time.Now().UTC(),
		})
		if err != nil {
			t.Fatalf("upsert: %v", err)
		}
	}

	upsert("allow", 0)
	upsert("queue", 5)

	got, err := db.GetJobConcurrencyPolicy(database, "sync")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Policy != "queue" {
		t.Errorf("policy: got %q, want %q", got.Policy, "queue")
	}
	if got.MaxQueue != 5 {
		t.Errorf("max_queue: got %d, want 5", got.MaxQueue)
	}
}

func TestGetJobConcurrencyPolicy_NotFound(t *testing.T) {
	database := tempDB(t)

	_, err := db.GetJobConcurrencyPolicy(database, "nonexistent")
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestDeleteJobConcurrencyPolicy(t *testing.T) {
	database := tempDB(t)

	err := db.UpsertJobConcurrencyPolicy(database, db.JobConcurrencyPolicy{
		JobName:   "cleanup",
		Policy:    "skip",
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}

	if err := db.DeleteJobConcurrencyPolicy(database, "cleanup"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, err = db.GetJobConcurrencyPolicy(database, "cleanup")
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows after delete, got %v", err)
	}
}

func TestListJobConcurrencyPolicies(t *testing.T) {
	database := tempDB(t)

	for _, name := range []string{"alpha", "beta", "gamma"} {
		if err := db.UpsertJobConcurrencyPolicy(database, db.JobConcurrencyPolicy{
			JobName:   name,
			Policy:    "allow",
			UpdatedAt: time.Now().UTC(),
		}); err != nil {
			t.Fatalf("upsert %s: %v", name, err)
		}
	}

	policies, err := db.ListJobConcurrencyPolicies(database)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(policies) != 3 {
		t.Errorf("got %d policies, want 3", len(policies))
	}
}
