package db

import (
	"database/sql"
	"time"
)

// JobStatus represents the latest run status for a named job.
type JobStatus struct {
	Name      string    `json:"name"`
	LastRunAt time.Time `json:"last_run_at"`
	LastExit  int       `json:"last_exit_code"`
	LastError string    `json:"last_error,omitempty"`
	RunCount  int       `json:"run_count"`
}

// ListJobStatuses returns the most recent run summary for every distinct job
// name recorded in the database, ordered alphabetically.
func ListJobStatuses(db *sql.DB) ([]JobStatus, error) {
	const q = `
		SELECT
			name,
			started_at,
			exit_code,
			COALESCE(error, ''),
			run_count
		FROM (
			SELECT
				name,
				started_at,
				exit_code,
				error,
				COUNT(*) OVER (PARTITION BY name) AS run_count,
				ROW_NUMBER()   OVER (PARTITION BY name ORDER BY started_at DESC) AS rn
			FROM job_runs
		) sub
		WHERE rn = 1
		ORDER BY name ASC
	`

	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statuses []JobStatus
	for rows.Next() {
		var s JobStatus
		if err := rows.Scan(&s.Name, &s.LastRunAt, &s.LastExit, &s.LastError, &s.RunCount); err != nil {
			return nil, err
		}
		statuses = append(statuses, s)
	}
	return statuses, rows.Err()
}
