package retention_test

import (
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cronwarden/internal/db"
	"github.com/cronwarden/internal/retention"
)

func tempDB(t *testing.T) *db.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	database, err := db.Open(path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return database
}

func silentLogger() *log.Logger {
	return log.New(os.Discard, "", 0)
}

func TestCleaner_Run_RemovesOldRows(t *testing.T) {
	database := tempDB(t)

	old := time.Now().UTC().Add(-48 * time.Hour)
	recent := time.Now().UTC()

	if err := database.InsertJobRun("job1", "success", 100, old); err != nil {
		t.Fatalf("insert old run: %v", err)
	}
	if err := database.InsertJobRun("job1", "success", 100, recent); err != nil {
		t.Fatalf("insert recent run: %v", err)
	}

	cleaner := retention.New(database, retention.Policy{MaxAgeDays: 1}, silentLogger())
	n, err := cleaner.Run()
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1 deleted row, got %d", n)
	}

	runs, err := database.ListJobRuns("", 100)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(runs) != 1 {
		t.Errorf("expected 1 remaining run, got %d", len(runs))
	}
}

func TestCleaner_Run_TrimsToMaxRows(t *testing.T) {
	database := tempDB(t)

	for i := 0; i < 5; i++ {
		ts := time.Now().UTC().Add(time.Duration(i) * time.Second)
		if err := database.InsertJobRun("job2", "success", 50, ts); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}

	cleaner := retention.New(database, retention.Policy{MaxRows: 3}, silentLogger())
	n, err := cleaner.Run()
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 trimmed rows, got %d", n)
	}

	runs, err := database.ListJobRuns("", 100)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(runs) != 3 {
		t.Errorf("expected 3 remaining runs, got %d", len(runs))
	}
}

func TestCleaner_Run_NoPolicyNoOp(t *testing.T) {
	database := tempDB(t)

	if err := database.InsertJobRun("job3", "success", 10, time.Now().UTC()); err != nil {
		t.Fatalf("insert: %v", err)
	}

	cleaner := retention.New(database, retention.Policy{}, silentLogger())
	n, err := cleaner.Run()
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 deleted rows, got %d", n)
	}
}
