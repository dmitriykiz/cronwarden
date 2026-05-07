package db

import (
	"database/sql"
	"errors"
	"time"
)

// JobTimeoutPolicy holds the timeout configuration for a cron job.
type JobTimeoutPolicy struct {
	JobName   string
	TimeoutSec int
	KillOnTimeout bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UpsertJobTimeoutPolicy inserts or replaces the timeout policy for a job.
func UpsertJobTimeoutPolicy(db *sql.DB, p JobTimeoutPolicy) error {
	_, err := db.Exec(`
		INSERT INTO job_timeout_policies (job_name, timeout_sec, kill_on_timeout, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(job_name) DO UPDATE SET
			timeout_sec = excluded.timeout_sec,
			kill_on_timeout = excluded.kill_on_timeout,
			updated_at = excluded.updated_at
	`,
		p.JobName,
		p.TimeoutSec,
		p.KillOnTimeout,
		time.Now().UTC(),
		time.Now().UTC(),
	)
	return err
}

// GetJobTimeoutPolicy retrieves the timeout policy for a job by name.
// Returns sql.ErrNoRows if not found.
func GetJobTimeoutPolicy(db *sql.DB, jobName string) (JobTimeoutPolicy, error) {
	var p JobTimeoutPolicy
	err := db.QueryRow(`
		SELECT job_name, timeout_sec, kill_on_timeout, created_at, updated_at
		FROM job_timeout_policies
		WHERE job_name = ?
	`, jobName).Scan(&p.JobName, &p.TimeoutSec, &p.KillOnTimeout, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return JobTimeoutPolicy{}, err
	}
	return p, nil
}

// DeleteJobTimeoutPolicy removes the timeout policy for a job.
// Returns nil even if no row existed.
func DeleteJobTimeoutPolicy(db *sql.DB, jobName string) error {
	_, err := db.Exec(`DELETE FROM job_timeout_policies WHERE job_name = ?`, jobName)
	return err
}

// ListJobTimeoutPolicies returns all configured timeout policies.
func ListJobTimeoutPolicies(db *sql.DB) ([]JobTimeoutPolicy, error) {
	rows, err := db.Query(`
		SELECT job_name, timeout_sec, kill_on_timeout, created_at, updated_at
		FROM job_timeout_policies
		ORDER BY job_name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []JobTimeoutPolicy
	for rows.Next() {
		var p JobTimeoutPolicy
		if err := rows.Scan(&p.JobName, &p.TimeoutSec, &p.KillOnTimeout, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if policies == nil {
		return []JobTimeoutPolicy{}, nil
	}
	return policies, nil
}

// isNoRowsTimeout is a convenience helper used in tests.
func isNoRowsTimeout(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
