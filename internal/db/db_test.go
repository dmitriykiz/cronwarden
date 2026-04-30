package db

import (
	"os"
	"testing"
	"time"
)

func tempDB(t *testing.T) (*DB, func()) {
	t.Helper()
	f, err := os.CreateTemp("", "cronwarden-test-*.db")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	f.Close()
	d, err := Open(f.Name())
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	return d, func() {
		d.Close()
		os.Remove(f.Name())
	}
}

func TestInsertAndListJobRuns(t *testing.T) {
	d, cleanup := tempDB(t)
	defer cleanup()

	now := time.Now().UTC().Truncate(time.Second)
	finished := now.Add(2 * time.Second)

	run := &JobRun{
		JobName:    "backup",
		StartedAt:  now,
		FinishedAt: &finished,
		ExitCode:   0,
		Output:     "backup complete",
		Success:    true,
	}

	id, err := d.InsertJobRun(run)
	if err != nil {
		t.Fatalf("InsertJobRun: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive id, got %d", id)
	}

	runs, err := d.ListJobRuns("backup", 10)
	if err != nil {
		t.Fatalf("ListJobRuns: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}

	got := runs[0]
	if got.JobName != "backup" {
		t.Errorf("job name: want backup, got %s", got.JobName)
	}
	if !got.Success {
		t.Error("expected success=true")
	}
	if got.ExitCode != 0 {
		t.Errorf("exit code: want 0, got %d", got.ExitCode)
	}
}

func TestListJobRuns_Empty(t *testing.T) {
	d, cleanup := tempDB(t)
	defer cleanup()

	runs, err := d.ListJobRuns("nonexistent", 10)
	if err != nil {
		t.Fatalf("ListJobRuns: %v", err)
	}
	if len(runs) != 0 {
		t.Errorf("expected 0 runs, got %d", len(runs))
	}
}

func TestListJobRuns_LimitRespected(t *testing.T) {
	d, cleanup := tempDB(t)
	defer cleanup()

	for i := 0; i < 5; i++ {
		_, err := d.InsertJobRun(&JobRun{
			JobName:   "cleanup",
			StartedAt: time.Now().UTC(),
			Success:   true,
		})
		if err != nil {
			t.Fatalf("insert: %v", err)
		}
	}

	runs, err := d.ListJobRuns("cleanup", 3)
	if err != nil {
		t.Fatalf("ListJobRuns: %v", err)
	}
	if len(runs) != 3 {
		t.Errorf("expected 3 runs (limit), got %d", len(runs))
	}
}
