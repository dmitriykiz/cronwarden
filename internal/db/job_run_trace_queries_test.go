package db_test

import (
	"context"
	"testing"

	"github.com/cronwarden/cronwarden/internal/db"
)

func seedTraceRun(t *testing.T, d *db.DB) int64 {
	t.Helper()
	err := db.InsertJobRun(context.Background(), d.DB, db.JobRun{
		JobName:  "trace-job",
		Status:   "success",
		Duration: 1.0,
	})
	if err != nil {
		t.Fatalf("seed run: %v", err)
	}
	runs, err := db.ListJobRuns(context.Background(), d.DB, 1)
	if err != nil {
		t.Fatalf("list runs: %v", err)
	}
	return runs[0].ID
}

func TestUpsertAndGetJobRunTrace(t *testing.T) {
	d := tempDB(t)
	runID := seedTraceRun(t, d)

	if err := db.UpsertJobRunTrace(context.Background(), d.DB, runID, "trace-abc", "span-001"); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	got, err := db.GetJobRunTrace(context.Background(), d.DB, runID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got == nil {
		t.Fatal("expected trace, got nil")
	}
	if got.TraceID != "trace-abc" || got.SpanID != "span-001" {
		t.Errorf("unexpected trace: %+v", got)
	}
}

func TestUpsertJobRunTrace_Replaces(t *testing.T) {
	d := tempDB(t)
	runID := seedTraceRun(t, d)

	_ = db.UpsertJobRunTrace(context.Background(), d.DB, runID, "old-trace", "old-span")
	_ = db.UpsertJobRunTrace(context.Background(), d.DB, runID, "new-trace", "new-span")

	got, _ := db.GetJobRunTrace(context.Background(), d.DB, runID)
	if got.TraceID != "new-trace" || got.SpanID != "new-span" {
		t.Errorf("expected replacement, got: %+v", got)
	}
}

func TestGetJobRunTrace_NotFound(t *testing.T) {
	d := tempDB(t)
	got, err := db.GetJobRunTrace(context.Background(), d.DB, 9999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got: %+v", got)
	}
}

func TestDeleteJobRunTrace(t *testing.T) {
	d := tempDB(t)
	runID := seedTraceRun(t, d)

	_ = db.UpsertJobRunTrace(context.Background(), d.DB, runID, "t", "s")
	if err := db.DeleteJobRunTrace(context.Background(), d.DB, runID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	got, _ := db.GetJobRunTrace(context.Background(), d.DB, runID)
	if got != nil {
		t.Error("expected nil after delete")
	}
}
