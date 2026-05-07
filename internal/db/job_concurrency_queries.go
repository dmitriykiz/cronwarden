package db

import (
	"database/sql"
	"errors"
	"time"
)

// JobConcurrencyPolicy defines how concurrent executions of a job are handled.
type JobConcurrencyPolicy struct {
	JobName   string
	Policy    string // "allow", "skip", "queue"
	MaxQueue  int
	UpdatedAt time.Time
}

// UpsertJobConcurrencyPolicy inserts or replaces the concurrency policy for a job.
func UpsertJobConcurrencyPolicy(db *sql.DB, p JobConcurrencyPolicy) error {
	_, err := db.Exec(`
		INSERT INTO job_concurrency_policies (job_name, policy, max_queue, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(job_name) DO UPDATE SET
			policy     = excluded.policy,
			max_queue  = excluded.max_queue,
			updated_at = excluded.updated_at
	`, p.JobName, p.Policy, p.MaxQueue, p.UpdatedAt)
	return err
}

// GetJobConcurrencyPolicy retrieves the concurrency policy for a job.
func GetJobConcurrencyPolicy(db *sql.DB, jobName string) (*JobConcurrencyPolicy, error) {
	row := db.QueryRow(`
		SELECT job_name, policy, max_queue, updated_at
		FROM job_concurrency_policies
		WHERE job_name = ?
	`, jobName)

	var p JobConcurrencyPolicy
	err := row.Scan(&p.JobName, &p.Policy, &p.MaxQueue, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &p, nil
}

// DeleteJobConcurrencyPolicy removes the concurrency policy for a job.
func DeleteJobConcurrencyPolicy(db *sql.DB, jobName string) error {
	_, err := db.Exec(`DELETE FROM job_concurrency_policies WHERE job_name = ?`, jobName)
	return err
}

// ListJobConcurrencyPolicies returns all concurrency policies.
func ListJobConcurrencyPolicies(db *sql.DB) ([]JobConcurrencyPolicy, error) {
	rows, err := db.Query(`
		SELECT job_name, policy, max_queue, updated_at
		FROM job_concurrency_policies
		ORDER BY job_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []JobConcurrencyPolicy
	for rows.Next() {
		var p JobConcurrencyPolicy
		if err := rows.Scan(&p.JobName, &p.Policy, &p.MaxQueue, &p.UpdatedAt); err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}
	return policies, rows.Err()
}
