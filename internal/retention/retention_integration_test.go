package retention_test

import (
	"testing"
	"time"

	"github.com/cronwarden/internal/retention"
)

func TestCleaner_Start_RunsPeriodically(t *testing.T) {
	database := tempDB(t)

	old := time.Now().UTC().Add(-72 * time.Hour)
	for i := 0; i < 3; i++ {
		if err := database.InsertJobRun("periodic-job", "success", 20, old); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}

	cleaner := retention.New(database, retention.Policy{MaxAgeDays: 1}, silentLogger())

	done := make(chan struct{})
	go cleaner.Start(50*time.Millisecond, done)

	time.Sleep(120 * time.Millisecond)
	close(done)

	runs, err := database.ListJobRuns("", 100)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(runs) != 0 {
		t.Errorf("expected all old runs removed, got %d remaining", len(runs))
	}
}
