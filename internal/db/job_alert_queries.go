package db

import (
	"database/sql"
	"time"
)

// JobAlert represents a stored alert rule for a specific job.
type JobAlert struct {
	ID          int64     `json:"id"`
	JobName     string    `json:"job_name"`
	AlertType   string    `json:"alert_type"` // "failure", "duration", "missing"
	Threshold   float64   `json:"threshold"`  // seconds for duration; 0 for failure/missing
	CreatedAt   time.Time `json:"created_at"`
}

// UpsertJobAlert inserts or replaces an alert rule for the given job and alert type.
func UpsertJobAlert(db *sql.DB, jobName, alertType string, threshold float64) error {
	_, err := db.Exec(`
		INSERT INTO job_alerts (job_name, alert_type, threshold, created_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(job_name, alert_type) DO UPDATE SET threshold = excluded.threshold`,
		jobName, alertType, threshold, time.Now().UTC(),
	)
	return err
}

// ListJobAlerts returns all alert rules for a given job.
func ListJobAlerts(db *sql.DB, jobName string) ([]JobAlert, error) {
	rows, err := db.Query(`
		SELECT id, job_name, alert_type, threshold, created_at
		FROM job_alerts
		WHERE job_name = ?
		ORDER BY created_at ASC`, jobName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []JobAlert
	for rows.Next() {
		var a JobAlert
		if err := rows.Scan(&a.ID, &a.JobName, &a.AlertType, &a.Threshold, &a.CreatedAt); err != nil {
			return nil, err
		}
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}

// DeleteJobAlert removes a specific alert rule by ID.
func DeleteJobAlert(db *sql.DB, id int64) (bool, error) {
	res, err := db.Exec(`DELETE FROM job_alerts WHERE id = ?`, id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}
