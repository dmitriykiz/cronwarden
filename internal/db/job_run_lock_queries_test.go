package db_test

import (
	"testing"
	"time"

	"github.com/cronwarden/cronwarden/internal/db"
)

func TestAcquireJobRunLock_Acquired(t *testing.T) {
	conn := tempDB(t)

	ok, err := db.AcquireJobRunLock(conn, "backup", "node-1", 5*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected lock to be acquired")
	}
}

func TestAcquireJobRunLock_AlreadyHeld(t *testing.T) {
	conn := tempDB(t)

	ok, err := db.AcquireJobRunLock(conn, "backup", "node-1", 5*time.Minute)
	if err != nil || !ok {
		t.Fatalf("first acquire failed: err=%v ok=%v", err, ok)
	}

	ok2, err := db.AcquireJobRunLock(conn, "backup", "node-2", 5*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok2 {
		t.Fatal("expected lock to NOT be acquired by node-2")
	}
}

func TestAcquireJobRunLock_ExpiredIsReplaced(t *testing.T) {
	conn := tempDB(t)

	ok, err := db.AcquireJobRunLock(conn, "backup", "node-1", -1*time.Second)
	if err != nil || !ok {
		t.Fatalf("first acquire failed: %v", err)
	}

	ok2, err := db.AcquireJobRunLock(conn, "backup", "node-2", 5*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok2 {
		t.Fatal("expected node-2 to acquire expired lock")
	}
}

func TestGetJobRunLock_NotFound(t *testing.T) {
	conn := tempDB(t)
	l, err := db.GetJobRunLock(conn, "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l != nil {
		t.Fatal("expected nil lock")
	}
}

func TestReleaseJobRunLock(t *testing.T) {
	conn := tempDB(t)

	db.AcquireJobRunLock(conn, "backup", "node-1", 5*time.Minute)
	if err := db.ReleaseJobRunLock(conn, "backup", "node-1"); err != nil {
		t.Fatalf("release failed: %v", err)
	}
	l, _ := db.GetJobRunLock(conn, "backup")
	if l != nil {
		t.Fatal("expected lock to be gone after release")
	}
}

func TestListJobRunLocks(t *testing.T) {
	conn := tempDB(t)

	db.AcquireJobRunLock(conn, "job-a", "node-1", 5*time.Minute)
	db.AcquireJobRunLock(conn, "job-b", "node-2", 5*time.Minute)

	locks, err := db.ListJobRunLocks(conn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(locks) != 2 {
		t.Fatalf("expected 2 locks, got %d", len(locks))
	}
}
