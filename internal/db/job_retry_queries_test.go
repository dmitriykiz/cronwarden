package db_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/cronwarden/cronwarden/internal/db"
)

func TestUpsertAndGetJobRetryPolicy(t *testing.T) {
	store := tempDB(t)

	policy := db.JobRetryPolicy{
		JobName:    "backup",
		MaxRetries: 3,
		RetryDelay: 30 * time.Second,
		UpdatedAt:  time.Now().UTC().Truncate(time.Second),
	}

	if err := db.UpsertJobRetryPolicy(store, policy); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	got, err := db.GetJobRetryPolicy(store, "backup")
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if got.JobName != policy.JobName {
		t.Errorf("job name: got %q, want %q", got.JobName, policy.JobName)
	}
	if got.MaxRetries != policy.MaxRetries {
		t.Errorf("max retries: got %d, want %d", got.MaxRetries, policy.MaxRetries)
	}
	if got.RetryDelay != policy.RetryDelay {
		t.Errorf("retry delay: got %v, want %v", got.RetryDelay, policy.RetryDelay)
	}
}

func TestUpsertJobRetryPolicy_Replaces(t *testing.T) {
	store := tempDB(t)

	original := db.JobRetryPolicy{JobName: "sync", MaxRetries: 2, RetryDelay: 10 * time.Second, UpdatedAt: time.Now().UTC()}
	if err := db.UpsertJobRetryPolicy(store, original); err != nil {
		t.Fatalf("first upsert: %v", err)
	}

	updated := db.JobRetryPolicy{JobName: "sync", MaxRetries: 5, RetryDelay: 60 * time.Second, UpdatedAt: time.Now().UTC()}
	if err := db.UpsertJobRetryPolicy(store, updated); err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	got, err := db.GetJobRetryPolicy(store, "sync")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.MaxRetries != 5 {
		t.Errorf("expected max_retries=5, got %d", got.MaxRetries)
	}
	if got.RetryDelay != 60*time.Second {
		t.Errorf("expected delay=60s, got %v", got.RetryDelay)
	}
}

func TestGetJobRetryPolicy_NotFound(t *testing.T) {
	store := tempDB(t)

	_, err := db.GetJobRetryPolicy(store, "nonexistent")
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestDeleteJobRetryPolicy(t *testing.T) {
	store := tempDB(t)

	policy := db.JobRetryPolicy{JobName: "cleanup", MaxRetries: 1, RetryDelay: 5 * time.Second, UpdatedAt: time.Now().UTC()}
	if err := db.UpsertJobRetryPolicy(store, policy); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if err := db.DeleteJobRetryPolicy(store, "cleanup"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, err := db.GetJobRetryPolicy(store, "cleanup")
	if err != sql.ErrNoRows {
		t.Errorf("expected ErrNoRows after delete, got %v", err)
	}
}

func TestListJobRetryPolicies(t *testing.T) {
	store := tempDB(t)

	names := []string{"alpha", "beta", "gamma"}
	for i, name := range names {
		p := db.JobRetryPolicy{JobName: name, MaxRetries: i + 1, RetryDelay: time.Duration(i+1) * 10 * time.Second, UpdatedAt: time.Now().UTC()}
		if err := db.UpsertJobRetryPolicy(store, p); err != nil {
			t.Fatalf("upsert %s: %v", name, err)
		}
	}

	policies, err := db.ListJobRetryPolicies(store)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(policies) != 3 {
		t.Errorf("expected 3 policies, got %d", len(policies))
	}
}
