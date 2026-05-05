package scheduler_test

import (
	"testing"
	"time"

	"cronwarden/internal/config"
	"cronwarden/internal/db"
	"cronwarden/internal/runner"
	"cronwarden/internal/scheduler"
)

// TestScheduler_JobRuns_RecordedInDB verifies that a job triggered by the
// scheduler actually persists a run record in the database.
func TestScheduler_JobRuns_RecordedInDB(t *testing.T) {
	conn := tempDB(t)
	r := runner.New(conn, nil)
	s := scheduler.New(r)

	// Use a 1-second interval so the job fires quickly during the test.
	jobs := []config.Job{
		{Name: "fast-job", Schedule: "@every 1s", Command: "echo integration"},
	}
	if err := s.Register(jobs); err != nil {
		t.Fatalf("Register(): %v", err)
	}

	s.Start()
	defer s.Stop()

	// Wait long enough for at least one execution.
	time.Sleep(2500 * time.Millisecond)

	runs, err := db.ListJobRuns(conn, "fast-job", 10)
	if err != nil {
		t.Fatalf("ListJobRuns(): %v", err)
	}
	if len(runs) == 0 {
		t.Fatal("expected at least one job run recorded, got none")
	}

	for _, run := range runs {
		if run.JobName != "fast-job" {
			t.Errorf("run.JobName = %q, want %q", run.JobName, "fast-job")
		}
		if run.ExitCode != 0 {
			t.Errorf("run.ExitCode = %d, want 0", run.ExitCode)
		}
	}
}
