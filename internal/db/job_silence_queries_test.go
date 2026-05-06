package db_test

import (
	"testing"
	"time"

	"github.com/user/cronwarden/internal/db"
)

func TestUpsertAndListJobSilences(t *testing.T) {
	database := tempDB(t)

	now := time.Now().UTC().Truncate(time.Second)
	id, err := db.UpsertJobSilence(database, db.JobSilence{
		JobName:  "backup",
		StartsAt: now,
		EndsAt:   now.Add(2 * time.Hour),
		Reason:   "maintenance",
	})
	if err != nil {
		t.Fatalf("UpsertJobSilence: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero id")
	}

	silences, err := db.ListJobSilences(database, "backup")
	if err != nil {
		t.Fatalf("ListJobSilences: %v", err)
	}
	if len(silences) != 1 {
		t.Fatalf("expected 1 silence, got %d", len(silences))
	}
	if silences[0].Reason != "maintenance" {
		t.Errorf("expected reason 'maintenance', got %q", silences[0].Reason)
	}
}

func TestListJobSilences_Empty(t *testing.T) {
	database := tempDB(t)
	silences, err := db.ListJobSilences(database, "nojob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(silences) != 0 {
		t.Errorf("expected 0 silences, got %d", len(silences))
	}
}

func TestDeleteJobSilence(t *testing.T) {
	database := tempDB(t)
	now := time.Now().UTC()
	id, _ := db.UpsertJobSilence(database, db.JobSilence{
		JobName:  "cleanup",
		StartsAt: now,
		EndsAt:   now.Add(time.Hour),
		Reason:   "deploy",
	})

	ok, err := db.DeleteJobSilence(database, id)
	if err != nil || !ok {
		t.Fatalf("DeleteJobSilence: ok=%v err=%v", ok, err)
	}

	silences, _ := db.ListJobSilences(database, "cleanup")
	if len(silences) != 0 {
		t.Errorf("expected 0 silences after delete, got %d", len(silences))
	}
}

func TestIsJobSilenced(t *testing.T) {
	database := tempDB(t)
	now := time.Now().UTC()
	db.UpsertJobSilence(database, db.JobSilence{
		JobName:  "sync",
		StartsAt: now.Add(-time.Hour),
		EndsAt:   now.Add(time.Hour),
		Reason:   "window",
	})

	silenced, err := db.IsJobSilenced(database, "sync", now)
	if err != nil {
		t.Fatalf("IsJobSilenced: %v", err)
	}
	if !silenced {
		t.Error("expected job to be silenced")
	}

	notSilenced, _ := db.IsJobSilenced(database, "sync", now.Add(2*time.Hour))
	if notSilenced {
		t.Error("expected job NOT to be silenced after window")
	}
}
