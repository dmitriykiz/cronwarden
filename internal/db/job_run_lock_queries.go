package db

import (
	"database/sql"
	"errors"
	"time"
)

// JobRunLock represents a distributed lock on a named job.
type JobRunLock struct {
	JobName   string    `json:"job_name"`
	LockedBy  string    `json:"locked_by"`
	LockedAt  time.Time `json:"locked_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// AcquireJobRunLock attempts to insert a lock row for the given job.
// Returns (true, nil) if the lock was acquired, (false, nil) if already held.
func AcquireJobRunLock(db *sql.DB, jobName, lockedBy string, ttl time.Duration) (bool, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(ttl)

	_, err := db.Exec(`
		INSERT INTO job_run_locks (job_name, locked_by, locked_at, expires_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(job_name) DO UPDATE
		  SET locked_by = excluded.locked_by,
		      locked_at = excluded.locked_at,
		      expires_at = excluded.expires_at
		WHERE job_run_locks.expires_at < ?`,
		jobName, lockedBy, now, expiresAt, now,
	)
	if err != nil {
		return false, err
	}

	var owner string
	err = db.QueryRow(`SELECT locked_by FROM job_run_locks WHERE job_name = ?`, jobName).Scan(&owner)
	if err != nil {
		return false, err
	}
	return owner == lockedBy, nil
}

// GetJobRunLock returns the current lock for a job, if any.
func GetJobRunLock(db *sql.DB, jobName string) (*JobRunLock, error) {
	row := db.QueryRow(`
		SELECT job_name, locked_by, locked_at, expires_at
		FROM job_run_locks WHERE job_name = ?`, jobName)

	var l JobRunLock
	var lockedAt, expiresAt string
	if err := row.Scan(&l.JobName, &l.LockedBy, &lockedAt, &expiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	l.LockedAt, _ = time.Parse(time.RFC3339Nano, lockedAt)
	l.ExpiresAt, _ = time.Parse(time.RFC3339Nano, expiresAt)
	return &l, nil
}

// ReleaseJobRunLock removes the lock for a job if held by the given owner.
func ReleaseJobRunLock(db *sql.DB, jobName, lockedBy string) error {
	_, err := db.Exec(`
		DELETE FROM job_run_locks WHERE job_name = ? AND locked_by = ?`,
		jobName, lockedBy)
	return err
}

// ListJobRunLocks returns all current lock rows.
func ListJobRunLocks(db *sql.DB) ([]JobRunLock, error) {
	rows, err := db.Query(`
		SELECT job_name, locked_by, locked_at, expires_at
		FROM job_run_locks ORDER BY locked_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locks []JobRunLock
	for rows.Next() {
		var l JobRunLock
		var lockedAt, expiresAt string
		if err := rows.Scan(&l.JobName, &l.LockedBy, &lockedAt, &expiresAt); err != nil {
			return nil, err
		}
		l.LockedAt, _ = time.Parse(time.RFC3339Nano, lockedAt)
		l.ExpiresAt, _ = time.Parse(time.RFC3339Nano, expiresAt)
		locks = append(locks, l)
	}
	return locks, rows.Err()
}
