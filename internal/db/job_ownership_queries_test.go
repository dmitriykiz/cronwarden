package db_test

import (
	"database/sql"
	"testing"

	"github.com/cronwarden/cronwarden/internal/db"
)

func TestUpsertAndGetJobOwnership(t *testing.T) {
	conn := tempDB(t)

	o := db.JobOwnership{
		JobName:     "backup",
		Owner:       "alice",
		Team:        "platform",
		SlackHandle: "@alice",
	}
	if err := db.UpsertJobOwnership(conn, o); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	got, err := db.GetJobOwnership(conn, "backup")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Owner != "alice" || got.Team != "platform" || got.SlackHandle != "@alice" {
		t.Errorf("unexpected record: %+v", got)
	}
}

func TestUpsertJobOwnership_Replaces(t *testing.T) {
	conn := tempDB(t)

	db.UpsertJobOwnership(conn, db.JobOwnership{JobName: "sync", Owner: "bob", Team: "data", SlackHandle: "@bob"})
	db.UpsertJobOwnership(conn, db.JobOwnership{JobName: "sync", Owner: "carol", Team: "infra", SlackHandle: "@carol"})

	got, err := db.GetJobOwnership(conn, "sync")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Owner != "carol" || got.Team != "infra" {
		t.Errorf("expected updated record, got %+v", got)
	}
}

func TestGetJobOwnership_NotFound(t *testing.T) {
	conn := tempDB(t)
	_, err := db.GetJobOwnership(conn, "nonexistent")
	if err != sql.ErrNoRows {
		t.Fatalf("expected ErrNoRows, got %v", err)
	}
}

func TestDeleteJobOwnership(t *testing.T) {
	conn := tempDB(t)

	db.UpsertJobOwnership(conn, db.JobOwnership{JobName: "clean", Owner: "dave", Team: "ops", SlackHandle: "@dave"})
	if err := db.DeleteJobOwnership(conn, "clean"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, err := db.GetJobOwnership(conn, "clean")
	if err != sql.ErrNoRows {
		t.Fatalf("expected ErrNoRows after delete, got %v", err)
	}
}

func TestListJobOwnerships(t *testing.T) {
	conn := tempDB(t)

	db.UpsertJobOwnership(conn, db.JobOwnership{JobName: "alpha", Owner: "x", Team: "a", SlackHandle: "@x"})
	db.UpsertJobOwnership(conn, db.JobOwnership{JobName: "beta", Owner: "y", Team: "b", SlackHandle: "@y"})

	list, err := db.ListJobOwnerships(conn)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 records, got %d", len(list))
	}
	if list[0].JobName != "alpha" || list[1].JobName != "beta" {
		t.Errorf("unexpected order: %v %v", list[0].JobName, list[1].JobName)
	}
}
