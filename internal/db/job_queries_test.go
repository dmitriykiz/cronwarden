package db_test

import (
	"testing"
	"time"

	"github.com/cronwarden/internal/db"
)

func TestListJobRunsByName_Empty(t *testing.T) {
	sqlDB := tempDB(t)
	runs, err := db.ListJobRunsByName(sqlDB, "nonexistent", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(runs) != 0 {
		t.Fatalf("expected 0 runs, got %d", len(runs))
	}
}

func TestListJobRunsByName_FiltersByName(t *testing.T) {
	sqlDB := tempDB(t)
	now := time.Now().UTC()

	if err := db.InsertJobRun(sqlDB, "job-a", now, 1.5, 0, "ok"); err != nil {
		t.Fatal(err)
	}
	if err := db.InsertJobRun(sqlDB, "job-b", now, 2.0, 1, "fail"); err != nil {
		t.Fatal(err)
	}
	if err := db.InsertJobRun(sqlDB, "job-a", now.Add(time.Second), 0.8, 0, "ok2"); err != nil {
		t.Fatal(err)
	}

	runs, err := db.ListJobRunsByName(sqlDB, "job-a", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(runs) != 2 {
		t.Fatalf("expected 2 runs for job-a, got %d", len(runs))
	}
	for _, r := range runs {
		if r.JobName != "job-a" {
			t.Errorf("unexpected job name %q", r.JobName)
		}
	}
}

func TestComputeStats_Mixed(t *testing.T) {
	runs := []db.JobRunRow{
		{JobName: "myjob", ExitCode: 0, Duration: 2.0},
		{JobName: "myjob", ExitCode: 1, Duration: 4.0},
		{JobName: "myjob", ExitCode: 0, Duration: 3.0},
	}
	stats := db.ComputeStats("myjob", runs)

	if stats.TotalRuns != 3 {
		t.Errorf("expected TotalRuns=3, got %d", stats.TotalRuns)
	}
	if stats.SuccessRuns != 2 {
		t.Errorf("expected SuccessRuns=2, got %d", stats.SuccessRuns)
	}
	if stats.FailureRuns != 1 {
		t.Errorf("expected FailureRuns=1, got %d", stats.FailureRuns)
	}
	const wantAvg = 3.0
	if stats.AvgDuration != wantAvg {
		t.Errorf("expected AvgDuration=%.2f, got %.2f", wantAvg, stats.AvgDuration)
	}
}

func TestComputeStats_Empty(t *testing.T) {
	stats := db.ComputeStats("emptyjob", nil)
	if stats.TotalRuns != 0 || stats.AvgDuration != 0 {
		t.Errorf("expected zero stats for empty input, got %+v", stats)
	}
}
