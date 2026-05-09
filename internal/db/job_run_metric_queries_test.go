package db_test

import (
	"testing"
	"time"

	"github.com/cronwarden/cronwarden/internal/db"
)

func seedMetricRun(t *testing.T, database *db.DB) int64 {
	t.Helper()
	id, err := db.InsertJobRun(database.DB, db.JobRun{
		JobName:   "metric-job",
		Status:    "success",
		StartedAt: time.Now().UTC(),
		Duration:  1.0,
	})
	if err != nil {
		t.Fatalf("seed run: %v", err)
	}
	return id
}

func TestAddAndListJobRunMetrics(t *testing.T) {
	database := tempDB(t)
	runID := seedMetricRun(t, database)

	_, err := db.AddJobRunMetric(database.DB, runID, "exit_code", "0")
	if err != nil {
		t.Fatalf("add metric: %v", err)
	}
	_, err = db.AddJobRunMetric(database.DB, runID, "rows_processed", "42")
	if err != nil {
		t.Fatalf("add metric: %v", err)
	}

	metrics, err := db.ListJobRunMetrics(database.DB, runID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(metrics) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(metrics))
	}
	if metrics[0].Key != "exit_code" || metrics[0].Value != "0" {
		t.Errorf("unexpected first metric: %+v", metrics[0])
	}
	if metrics[1].Key != "rows_processed" || metrics[1].Value != "42" {
		t.Errorf("unexpected second metric: %+v", metrics[1])
	}
}

func TestListJobRunMetrics_Empty(t *testing.T) {
	database := tempDB(t)
	runID := seedMetricRun(t, database)

	metrics, err := db.ListJobRunMetrics(database.DB, runID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(metrics) != 0 {
		t.Errorf("expected 0 metrics, got %d", len(metrics))
	}
}

func TestDeleteJobRunMetric_Exists(t *testing.T) {
	database := tempDB(t)
	runID := seedMetricRun(t, database)

	id, err := db.AddJobRunMetric(database.DB, runID, "cpu_pct", "12.5")
	if err != nil {
		t.Fatalf("add: %v", err)
	}

	ok, err := db.DeleteJobRunMetric(database.DB, id)
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if !ok {
		t.Error("expected deleted=true")
	}

	metrics, _ := db.ListJobRunMetrics(database.DB, runID)
	if len(metrics) != 0 {
		t.Errorf("expected 0 after delete, got %d", len(metrics))
	}
}

func TestDeleteJobRunMetric_NotExists(t *testing.T) {
	database := tempDB(t)

	ok, err := db.DeleteJobRunMetric(database.DB, 9999)
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if ok {
		t.Error("expected deleted=false for non-existent id")
	}
}
