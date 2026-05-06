package db

import (
	"database/sql"
	"errors"
	"time"
)

// JobRetryPolicy holds retry configuration for a named job.
type JobRetryPolicy struct {
	JobName    string
	MaxRetries int
	RetryDelay time.Duration // stored as seconds in DB
	UpdatedAt  time.Time
}

// UpsertJobRetryPolicy inserts or replaces the retry policy for a job.
func UpsertJobRetryPolicy(db *sql.DB, policy JobRetryPolicy) error {
	_, err := db.Exec(`
		INSERT INTO job_retry_policies (job_name, max_retries, retry_delay_seconds, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(job_name) DO UPDATE SET
			max_retries = excluded.max_retries,
			retry_delay_seconds = excluded.retry_delay_seconds,
			updated_at = excluded.updated_at`,
		policy.JobName,
		policy.MaxRetries,
		int64(policy.RetryDelay.Seconds()),
		policy.UpdatedAt.UTC(),
	)
	return err
}

// GetJobRetryPolicy retrieves the retry policy for the given job name.
// Returns sql.ErrNoRows if no policy exists.
func GetJobRetryPolicy(db *sql.DB, jobName string) (JobRetryPolicy, error) {
	row := db.QueryRow(`
		SELECT job_name, max_retries, retry_delay_seconds, updated_at
		FROM job_retry_policies
		WHERE job_name = ?`, jobName)

	var p JobRetryPolicy
	var delaySecs int64
	var updatedAt string
	if err := row.Scan(&p.JobName, &p.MaxRetries, &delaySecs, &updatedAt); err != nil {
		return JobRetryPolicy{}, err
	}
	p.RetryDelay = time.Duration(delaySecs) * time.Second
	t, err := time.Parse(time.RFC3339, updatedAt)
	if err == nil {
		p.UpdatedAt = t
	}
	return p, nil
}

// DeleteJobRetryPolicy removes the retry policy for the given job name.
// Returns nil even if no row existed.
func DeleteJobRetryPolicy(db *sql.DB, jobName string) error {
	_, err := db.Exec(`DELETE FROM job_retry_policies WHERE job_name = ?`, jobName)
	return err
}

// ListJobRetryPolicies returns all stored retry policies.
func ListJobRetryPolicies(db *sql.DB) ([]JobRetryPolicy, error) {
	rows, err := db.Query(`
		SELECT job_name, max_retries, retry_delay_seconds, updated_at
		FROM job_retry_policies
		ORDER BY job_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []JobRetryPolicy
	for rows.Next() {
		var p JobRetryPolicy
		var delaySecs int64
		var updatedAt string
		if err := rows.Scan(&p.JobName, &p.MaxRetries, &delaySecs, &updatedAt); err != nil {
			return nil, err
		}
		p.RetryDelay = time.Duration(delaySecs) * time.Second
		if t, err := time.Parse(time.RFC3339, updatedAt); err == nil {
			p.UpdatedAt = t
		}
		policies = append(policies, p)
	}
	if err := rows.Err(); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return policies, nil
}
