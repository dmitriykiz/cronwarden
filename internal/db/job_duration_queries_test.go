package db_test

import (
	"testing"
	"time"

	"github.com/cronwarden/cronwarden/internal/db"
)

func seedDurationRuns(t *testing.T, database *db.DB, jobName string, durations []int64) {
	t.Helper()
	for _, d := range durations {
		err := database.InsertJobRun(db.JobRun{
			JobName:    jobName,
			Status:     "success",
			StartedAt:  time.Now(),
			DurationMS: d,
		})
		if err != nil {
			t.Fatalf("seed insert failed: %v", err)
		}
	}
}

func TestGetJobDurationStats_Basic(t *testing.T) {
	database := tempDB(t)
	seedDurationRuns(t, database, "backup", []int64{1000, 2000, 3000})

	stat, err := db.GetJobDurationStats(database.DB, "backup", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stat.JobName != "backup" {
		t.Errorf("expected job_name=backup, got %s", stat.JobName)
	}
	if stat.SampleSize != 3 {
		t.Errorf("expected sample_size=3, got %d", stat.SampleSize)
	}
	if stat.AvgSeconds != 2.0 {
		t.Errorf("expected avg=2.0s, got %f", stat.AvgSeconds)
	}
	if stat.MinSeconds != 1.0 {
		t.Errorf("expected min=1.0s, got %f", stat.MinSeconds)
	}
	if stat.MaxSeconds != 3.0 {
		t.Errorf("expected max=3.0s, got %f", stat.MaxSeconds)
	}
}

func TestGetJobDurationStats_LimitRespected(t *testing.T) {
	database := tempDB(t)
	seedDurationRuns(t, database, "sync", []int64{500, 1500, 2500, 3500})

	stat, err := db.GetJobDurationStats(database.DB, "sync", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stat.SampleSize != 2 {
		t.Errorf("expected sample_size=2 with limit, got %d", stat.SampleSize)
	}
}

func TestGetJobDurationStats_NoRuns(t *testing.T) {
	database := tempDB(t)

	_, err := db.GetJobDurationStats(database.DB, "nonexistent", 0)
	if err == nil {
		t.Error("expected error for job with no runs, got nil")
	}
}
