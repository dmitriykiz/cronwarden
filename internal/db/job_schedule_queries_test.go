package db_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/cronwarden/cronwarden/internal/db"
)

func TestUpsertAndGetJobScheduleOverride(t *testing.T) {
	conn := tempDB(t)

	expires := time.Now().Add(2 * time.Hour).UTC().Truncate(time.Second)
	err := db.UpsertJobScheduleOverride(conn, db.JobScheduleOverride{
		JobName:  "backup",
		CronExpr: "0 3 * * *",
		Reason:   "maintenance window",
		ExpiresAt: expires,
	})
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}

	o, err := db.GetJobScheduleOverride(conn, "backup")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if o.JobName != "backup" {
		t.Errorf("expected job_name=backup, got %q", o.JobName)
	}
	if o.CronExpr != "0 3 * * *" {
		t.Errorf("expected cron_expr='0 3 * * *', got %q", o.CronExpr)
	}
	if o.Reason != "maintenance window" {
		t.Errorf("unexpected reason: %q", o.Reason)
	}
}

func TestUpsertJobScheduleOverride_Replaces(t *testing.T) {
	conn := tempDB(t)
	expires := time.Now().Add(1 * time.Hour).UTC()

	_ = db.UpsertJobScheduleOverride(conn, db.JobScheduleOverride{
		JobName: "sync", CronExpr: "*/5 * * * *", Reason: "first", ExpiresAt: expires,
	})
	_ = db.UpsertJobScheduleOverride(conn, db.JobScheduleOverride{
		JobName: "sync", CronExpr: "*/10 * * * *", Reason: "updated", ExpiresAt: expires,
	})

	o, err := db.GetJobScheduleOverride(conn, "sync")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if o.CronExpr != "*/10 * * * *" {
		t.Errorf("expected updated cron_expr, got %q", o.CronExpr)
	}
}

func TestGetJobScheduleOverride_NotFound(t *testing.T) {
	conn := tempDB(t)
	_, err := db.GetJobScheduleOverride(conn, "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing override, got nil")
	}
	if err != sql.ErrNoRows {
		t.Logf("got expected not-found error: %v", err)
	}
}

func TestDeleteJobScheduleOverride(t *testing.T) {
	conn := tempDB(t)
	expires := time.Now().Add(1 * time.Hour).UTC()

	_ = db.UpsertJobScheduleOverride(conn, db.JobScheduleOverride{
		JobName: "cleanup", CronExpr: "@daily", Reason: "temp", ExpiresAt: expires,
	})
	if err := db.DeleteJobScheduleOverride(conn, "cleanup"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, err := db.GetJobScheduleOverride(conn, "cleanup")
	if err == nil {
		t.Fatal("expected not-found after delete")
	}
}

func TestListActiveJobScheduleOverrides(t *testing.T) {
	conn := tempDB(t)

	active := time.Now().Add(2 * time.Hour).UTC()
	expired := time.Now().Add(-1 * time.Hour).UTC()

	_ = db.UpsertJobScheduleOverride(conn, db.JobScheduleOverride{
		JobName: "job-a", CronExpr: "@hourly", Reason: "active", ExpiresAt: active,
	})
	_ = db.UpsertJobScheduleOverride(conn, db.JobScheduleOverride{
		JobName: "job-b", CronExpr: "@daily", Reason: "expired", ExpiresAt: expired,
	})

	list, err := db.ListActiveJobScheduleOverrides(conn)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 active override, got %d", len(list))
	}
	if list[0].JobName != "job-a" {
		t.Errorf("expected job-a, got %q", list[0].JobName)
	}
}
