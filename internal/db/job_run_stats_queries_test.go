package db_test

import (
	"testing"
	"time"

	"github.com/cronwarden/cronwarden/internal/db"
)

func seedWindowRuns(t *testing.T, database *db.DB) {
	t.Helper()
	now := time.Now().UTC()
	runs := []struct {
		status   string
		dur      float64
		offset   time.Duration
	}{
		{"success", 1.2, -1 * time.Hour},
		{"success", 2.4, -2 * time.Hour},
		{"failure", 0.5, -3 * time.Hour},
		{"success", 3.0, -50 * time.Hour}, // outside window
	}
	for _, r := range runs {
		_, err := database.Exec(
			`INSERT INTO job_runs (job_name, status, started_at, duration_seconds) VALUES (?, ?, ?, ?)`,
			"window-job",
			r.status,
			now.Add(r.offset).Format(time.RFC3339),
			r.dur,
		)
		if err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
}

func TestGetJobRunWindow_Basic(t *testing.T) {
	database := tempDB(t)
	seedWindowRuns(t, database)

	from := time.Now().UTC().Add(-24 * time.Hour)
	to := time.Now().UTC()

	w, err := db.GetJobRunWindow(database.DB, "window-job", from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.TotalRuns != 3 {
		t.Errorf("expected 3 runs in window, got %d", w.TotalRuns)
	}
	if w.SuccessCount != 2 {
		t.Errorf("expected 2 successes, got %d", w.SuccessCount)
	}
	if w.FailureCount != 1 {
		t.Errorf("expected 1 failure, got %d", w.FailureCount)
	}
	if w.MaxDurationS != 2.4 {
		t.Errorf("expected max duration 2.4, got %f", w.MaxDurationS)
	}
}

func TestGetJobRunWindow_NoRuns(t *testing.T) {
	database := tempDB(t)

	from := time.Now().UTC().Add(-1 * time.Hour)
	to := time.Now().UTC()

	w, err := db.GetJobRunWindow(database.DB, "ghost-job", from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.TotalRuns != 0 {
		t.Errorf("expected 0 runs, got %d", w.TotalRuns)
	}
	if w.AvgDurationS != 0 {
		t.Errorf("expected avg 0, got %f", w.AvgDurationS)
	}
}
