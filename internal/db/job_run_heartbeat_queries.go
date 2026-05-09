package db

import (
	"database/sql"
	"errors"
	"time"
)

// JobRunHeartbeat records the last heartbeat time for a running job.
type JobRunHeartbeat struct {
	RunID     int64
	JobName   string
	BeatAt    time.Time
	CreatedAt time.Time
}

// UpsertJobRunHeartbeat inserts or updates the heartbeat for a given run ID.
func UpsertJobRunHeartbeat(db *sql.DB, runID int64, jobName string, beatAt time.Time) error {
	_, err := db.Exec(`
		INSERT INTO job_run_heartbeats (run_id, job_name, beat_at, created_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(run_id) DO UPDATE SET
			beat_at = excluded.beat_at
	`, runID, jobName, beatAt.UTC(), time.Now().UTC())
	return err
}

// GetJobRunHeartbeat retrieves the latest heartbeat for a given run ID.
func GetJobRunHeartbeat(db *sql.DB, runID int64) (*JobRunHeartbeat, error) {
	row := db.QueryRow(`
		SELECT run_id, job_name, beat_at, created_at
		FROM job_run_heartbeats
		WHERE run_id = ?
	`, runID)

	var h JobRunHeartbeat
	var beatAt, createdAt string
	if err := row.Scan(&h.RunID, &h.JobName, &beatAt, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	var err error
	h.BeatAt, err = time.Parse(time.RFC3339Nano, beatAt)
	if err != nil {
		return nil, err
	}
	h.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

// DeleteJobRunHeartbeat removes the heartbeat record for a given run ID.
func DeleteJobRunHeartbeat(db *sql.DB, runID int64) error {
	_, err := db.Exec(`DELETE FROM job_run_heartbeats WHERE run_id = ?`, runID)
	return err
}

// ListStaleHeartbeats returns heartbeats older than the given threshold.
func ListStaleHeartbeats(db *sql.DB, olderThan time.Time) ([]JobRunHeartbeat, error) {
	rows, err := db.Query(`
		SELECT run_id, job_name, beat_at, created_at
		FROM job_run_heartbeats
		WHERE beat_at < ?
		ORDER BY beat_at ASC
	`, olderThan.UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []JobRunHeartbeat
	for rows.Next() {
		var h JobRunHeartbeat
		var beatAt, createdAt string
		if err := rows.Scan(&h.RunID, &h.JobName, &beatAt, &createdAt); err != nil {
			return nil, err
		}
		h.BeatAt, _ = time.Parse(time.RFC3339Nano, beatAt)
		h.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		results = append(results, h)
	}
	return results, rows.Err()
}
