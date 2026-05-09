package db

import (
	"database/sql"
	"time"
)

// JobRunMetric represents a key/value metric attached to a specific job run.
type JobRunMetric struct {
	ID        int64
	RunID     int64
	Key       string
	Value     string
	CreatedAt time.Time
}

// AddJobRunMetric inserts a metric for a given run.
func AddJobRunMetric(db *sql.DB, runID int64, key, value string) (int64, error) {
	res, err := db.Exec(
		`INSERT INTO job_run_metrics (run_id, key, value, created_at) VALUES (?, ?, ?, ?)`,
		runID, key, value, time.Now().UTC(),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ListJobRunMetrics returns all metrics for a given run, ordered by id.
func ListJobRunMetrics(db *sql.DB, runID int64) ([]JobRunMetric, error) {
	rows, err := db.Query(
		`SELECT id, run_id, key, value, created_at FROM job_run_metrics WHERE run_id = ? ORDER BY id ASC`,
		runID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []JobRunMetric
	for rows.Next() {
		var m JobRunMetric
		if err := rows.Scan(&m.ID, &m.RunID, &m.Key, &m.Value, &m.CreatedAt); err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}
	return metrics, rows.Err()
}

// DeleteJobRunMetric removes a single metric by its id.
func DeleteJobRunMetric(db *sql.DB, id int64) (bool, error) {
	res, err := db.Exec(`DELETE FROM job_run_metrics WHERE id = ?`, id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}
