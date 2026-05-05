package scheduler_test

import (
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"cronwarden/internal/config"
	"cronwarden/internal/db"
	"cronwarden/internal/runner"
	"cronwarden/internal/scheduler"
)

func tempDB(t *testing.T) *sql.DB {
	t.Helper()
	f, err := os.CreateTemp("", "cronwarden-sched-*.db")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	t.Cleanup(func() { os.Remove(f.Name()) })
	f.Close()

	conn, err := db.Open(f.Name())
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

func TestScheduler_Register_ValidJob(t *testing.T) {
	conn := tempDB(t)
	r := runner.New(conn, nil)
	s := scheduler.New(r)

	jobs := []config.Job{
		{Name: "echo-job", Schedule: "@every 1m", Command: "echo hello"},
	}

	if err := s.Register(jobs); err != nil {
		t.Fatalf("Register() unexpected error: %v", err)
	}
}

func TestScheduler_Register_InvalidSchedule(t *testing.T) {
	conn := tempDB(t)
	r := runner.New(conn, nil)
	s := scheduler.New(r)

	jobs := []config.Job{
		{Name: "bad-job", Schedule: "not-a-cron", Command: "echo hi"},
	}

	if err := s.Register(jobs); err == nil {
		t.Fatal("Register() expected error for invalid schedule, got nil")
	}
}

func TestScheduler_NextRun_KnownJob(t *testing.T) {
	conn := tempDB(t)
	r := runner.New(conn, nil)
	s := scheduler.New(r)

	jobs := []config.Job{
		{Name: "tick", Schedule: "@every 5m", Command: "true"},
	}
	if err := s.Register(jobs); err != nil {
		t.Fatalf("Register(): %v", err)
	}
	s.Start()
	defer s.Stop()

	next := s.NextRun("tick")
	if next.IsZero() {
		t.Fatal("NextRun() returned zero time for registered job")
	}
	if next.Before(time.Now()) {
		t.Errorf("NextRun() = %v, want a future time", next)
	}
}

func TestScheduler_NextRun_UnknownJob(t *testing.T) {
	conn := tempDB(t)
	r := runner.New(conn, nil)
	s := scheduler.New(r)

	next := s.NextRun("nonexistent")
	if !next.IsZero() {
		t.Errorf("NextRun() = %v, want zero time for unknown job", next)
	}
}
