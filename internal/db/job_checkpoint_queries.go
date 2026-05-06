package db

import (
	"database/sql"
	"errors"
	"time"
)

// JobCheckpoint records the last known progress marker for a cron job.
type JobCheckpoint struct {
	JobName   string
	Marker    string
	UpdatedAt time.Time
}

// UpsertJobCheckpoint inserts or replaces the checkpoint marker for a job.
func UpsertJobCheckpoint(db *sql.DB, jobName, marker string) error {
	_, err := db.Exec(`
		INSERT INTO job_checkpoints (job_name, marker, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(job_name) DO UPDATE SET
			marker     = excluded.marker,
			updated_at = excluded.updated_at
	`, jobName, marker, time.Now().UTC())
	return err
}

// GetJobCheckpoint retrieves the checkpoint for a job.
// Returns sql.ErrNoRows if none exists.
func GetJobCheckpoint(db *sql.DB, jobName string) (*JobCheckpoint, error) {
	row := db.QueryRow(`
		SELECT job_name, marker, updated_at
		FROM job_checkpoints
		WHERE job_name = ?
	`, jobName)

	var cp JobCheckpoint
	if err := row.Scan(&cp.JobName, &cp.Marker, &cp.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &cp, nil
}

// DeleteJobCheckpoint removes the checkpoint for a job.
func DeleteJobCheckpoint(db *sql.DB, jobName string) error {
	_, err := db.Exec(`DELETE FROM job_checkpoints WHERE job_name = ?`, jobName)
	return err
}
