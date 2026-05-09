package db

import (
	"database/sql"
	"errors"
	"time"
)

// JobRunOutput holds captured stdout/stderr for a single job run.
type JobRunOutput struct {
	ID        int64
	RunID     int64
	Stdout    string
	Stderr    string
	CapturedAt time.Time
}

// UpsertJobRunOutput inserts or replaces the output record for a given run.
func UpsertJobRunOutput(db *sql.DB, runID int64, stdout, stderr string) error {
	_, err := db.Exec(`
		INSERT INTO job_run_outputs (run_id, stdout, stderr, captured_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(run_id) DO UPDATE SET
			stdout = excluded.stdout,
			stderr = excluded.stderr,
			captured_at = excluded.captured_at
	`, runID, stdout, stderr, time.Now().UTC())
	return err
}

// GetJobRunOutput retrieves the output record for a given run.
func GetJobRunOutput(db *sql.DB, runID int64) (*JobRunOutput, error) {
	row := db.QueryRow(`
		SELECT id, run_id, stdout, stderr, captured_at
		FROM job_run_outputs
		WHERE run_id = ?
	`, runID)

	var o JobRunOutput
	var capturedAt string
	err := row.Scan(&o.ID, &o.RunID, &o.Stdout, &o.Stderr, &capturedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	o.CapturedAt, _ = time.Parse(time.RFC3339, capturedAt)
	return &o, nil
}

// DeleteJobRunOutput removes the output record for a given run.
func DeleteJobRunOutput(db *sql.DB, runID int64) error {
	_, err := db.Exec(`DELETE FROM job_run_outputs WHERE run_id = ?`, runID)
	return err
}
