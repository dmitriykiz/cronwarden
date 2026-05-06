package db

import (
	"database/sql"
	"time"
)

// JobDurationAlert represents a threshold alert for job duration.
type JobDurationAlert struct {
	ID              int64
	JobName         string
	ThresholdSeconds float64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// UpsertJobDurationAlert inserts or replaces a duration alert threshold for a job.
func UpsertJobDurationAlert(db *sql.DB, jobName string, thresholdSeconds float64) error {
	_, err := db.Exec(`
		INSERT INTO job_duration_alerts (job_name, threshold_seconds, created_at, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(job_name) DO UPDATE SET
			threshold_seconds = excluded.threshold_seconds,
			updated_at = CURRENT_TIMESTAMP
	`, jobName, thresholdSeconds)
	return err
}

// GetJobDurationAlert retrieves the duration alert for a given job.
// Returns sql.ErrNoRows if not found.
func GetJobDurationAlert(db *sql.DB, jobName string) (*JobDurationAlert, error) {
	row := db.QueryRow(`
		SELECT id, job_name, threshold_seconds, created_at, updated_at
		FROM job_duration_alerts
		WHERE job_name = ?
	`, jobName)

	var a JobDurationAlert
	err := row.Scan(&a.ID, &a.JobName, &a.ThresholdSeconds, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// DeleteJobDurationAlert removes the duration alert for a given job.
func DeleteJobDurationAlert(db *sql.DB, jobName string) error {
	_, err := db.Exec(`DELETE FROM job_duration_alerts WHERE job_name = ?`, jobName)
	return err
}

// ListJobDurationAlerts returns all configured duration alerts.
func ListJobDurationAlerts(db *sql.DB) ([]JobDurationAlert, error) {
	rows, err := db.Query(`
		SELECT id, job_name, threshold_seconds, created_at, updated_at
		FROM job_duration_alerts
		ORDER BY job_name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []JobDurationAlert
	for rows.Next() {
		var a JobDurationAlert
		if err := rows.Scan(&a.ID, &a.JobName, &a.ThresholdSeconds, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}
