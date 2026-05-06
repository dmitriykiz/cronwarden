package db

import (
	"database/sql"
	"time"
)

// JobRunDetail holds a single job run record with all fields.
type JobRunDetail struct {
	ID        int64
	JobName   string
	StartedAt time.Time
	Duration  float64
	Success   bool
	Output    string
}

// GetJobRun retrieves a single job run by its ID.
func GetJobRun(db *sql.DB, id int64) (*JobRunDetail, error) {
	row := db.QueryRow(`
		SELECT id, job_name, started_at, duration_ms, success, output
		FROM job_runs
		WHERE id = ?`, id)

	var r JobRunDetail
	var startedAt string
	err := row.Scan(&r.ID, &r.JobName, &startedAt, &r.Duration, &r.Success, &r.Output)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	r.StartedAt, err = time.Parse(time.RFC3339, startedAt)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// DeleteJobRun removes a single job run by ID. Returns true if a row was deleted.
func DeleteJobRun(db *sql.DB, id int64) (bool, error) {
	res, err := db.Exec(`DELETE FROM job_runs WHERE id = ?`, id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
