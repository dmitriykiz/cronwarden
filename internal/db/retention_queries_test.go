package db_test

import (
	"testing"
	"time"
)

func TestDeleteRunsBefore(t *testing.T) {
	database := tempDB(t)

	now := time.Now().UTC()
	past := now.Add(-24 * time.Hour)

	_ = database.InsertJobRun("j", "ok", 1, past)
	_ = database.InsertJobRun("j", "ok", 2, past.Add(-time.Hour))
	_ = database.InsertJobRun("j", "ok", 3, now)

	cutoff := now.Add(-time.Hour)
	n, err := database.DeleteRunsBefore(cutoff)
	if err != nil {
		t.Fatalf("DeleteRunsBefore: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 deleted, got %d", n)
	}

	runs, _ := database.ListJobRuns("", 100)
	if len(runs) != 1 {
		t.Errorf("expected 1 remaining, got %d", len(runs))
	}
}

func TestTrimRunsToLimit(t *testing.T) {
	database := tempDB(t)

	base := time.Now().UTC()
	for i := 0; i < 6; i++ {
		_ = database.InsertJobRun("j", "ok", int64(i), base.Add(time.Duration(i)*time.Second))
	}

	n, err := database.TrimRunsToLimit(4)
	if err != nil {
		t.Fatalf("TrimRunsToLimit: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 trimmed, got %d", n)
	}

	runs, _ := database.ListJobRuns("", 100)
	if len(runs) != 4 {
		t.Errorf("expected 4 remaining, got %d", len(runs))
	}
}

func TestTrimRunsToLimit_NoExcess(t *testing.T) {
	database := tempDB(t)

	_ = database.InsertJobRun("j", "ok", 1, time.Now().UTC())

	n, err := database.TrimRunsToLimit(10)
	if err != nil {
		t.Fatalf("TrimRunsToLimit: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 trimmed, got %d", n)
	}
}
