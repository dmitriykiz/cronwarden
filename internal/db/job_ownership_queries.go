package db

import (
	"database/sql"
	"errors"
	"time"
)

// JobOwnership holds the owner metadata for a named job.
type JobOwnership struct {
	JobName   string    `json:"job_name"`
	Owner     string    `json:"owner"`
	Team      string    `json:"team"`
	SlackHandle string  `json:"slack_handle"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpsertJobOwnership inserts or replaces the ownership record for a job.
func UpsertJobOwnership(db *sql.DB, o JobOwnership) error {
	_, err := db.Exec(`
		INSERT INTO job_ownership (job_name, owner, team, slack_handle, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(job_name) DO UPDATE SET
			owner        = excluded.owner,
			team         = excluded.team,
			slack_handle = excluded.slack_handle,
			updated_at   = excluded.updated_at`,
		o.JobName, o.Owner, o.Team, o.SlackHandle, time.Now().UTC(),
	)
	return err
}

// GetJobOwnership returns the ownership record for the given job name.
// Returns sql.ErrNoRows if not found.
func GetJobOwnership(db *sql.DB, jobName string) (JobOwnership, error) {
	var o JobOwnership
	err := db.QueryRow(`
		SELECT job_name, owner, team, slack_handle, updated_at
		FROM job_ownership WHERE job_name = ?`, jobName).
		Scan(&o.JobName, &o.Owner, &o.Team, &o.SlackHandle, &o.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return JobOwnership{}, sql.ErrNoRows
	}
	return o, err
}

// DeleteJobOwnership removes the ownership record for the given job.
func DeleteJobOwnership(db *sql.DB, jobName string) error {
	_, err := db.Exec(`DELETE FROM job_ownership WHERE job_name = ?`, jobName)
	return err
}

// ListJobOwnerships returns all ownership records ordered by job name.
func ListJobOwnerships(db *sql.DB) ([]JobOwnership, error) {
	rows, err := db.Query(`
		SELECT job_name, owner, team, slack_handle, updated_at
		FROM job_ownership ORDER BY job_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []JobOwnership
	for rows.Next() {
		var o JobOwnership
		if err := rows.Scan(&o.JobName, &o.Owner, &o.Team, &o.SlackHandle, &o.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}
