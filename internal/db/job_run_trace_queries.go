package db

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// JobRunTrace holds a distributed trace identifier linked to a job run.
type JobRunTrace struct {
	ID        int64     `json:"id"`
	RunID     int64     `json:"run_id"`
	TraceID   string    `json:"trace_id"`
	SpanID    string    `json:"span_id"`
	CreatedAt time.Time `json:"created_at"`
}

// UpsertJobRunTrace inserts or replaces the trace for a given run.
func UpsertJobRunTrace(ctx context.Context, db *sql.DB, runID int64, traceID, spanID string) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO job_run_traces (run_id, trace_id, span_id)
		 VALUES (?, ?, ?)
		 ON CONFLICT(run_id) DO UPDATE SET trace_id=excluded.trace_id, span_id=excluded.span_id`,
		runID, traceID, spanID,
	)
	return err
}

// GetJobRunTrace retrieves the trace for a given run.
func GetJobRunTrace(ctx context.Context, db *sql.DB, runID int64) (*JobRunTrace, error) {
	row := db.QueryRowContext(ctx,
		`SELECT id, run_id, trace_id, span_id, created_at FROM job_run_traces WHERE run_id = ?`,
		runID,
	)
	var t JobRunTrace
	if err := row.Scan(&t.ID, &t.RunID, &t.TraceID, &t.SpanID, &t.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

// DeleteJobRunTrace removes the trace record for a given run.
func DeleteJobRunTrace(ctx context.Context, db *sql.DB, runID int64) error {
	_, err := db.ExecContext(ctx,
		`DELETE FROM job_run_traces WHERE run_id = ?`, runID,
	)
	return err
}
