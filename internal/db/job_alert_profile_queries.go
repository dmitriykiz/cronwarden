package db

import (
	"database/sql"
	"fmt"
	"time"
)

// AlertProfile holds threshold-based alerting config for a job.
type AlertProfile struct {
	ID              int64
	JobName         string
	MaxDurationSecs int
	MinSuccessRate  float64 // 0.0–1.0
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// UpsertAlertProfile inserts or replaces the alert profile for a job.
func UpsertAlertProfile(db *sql.DB, p AlertProfile) error {
	_, err := db.Exec(`
		INSERT INTO job_alert_profiles (job_name, max_duration_secs, min_success_rate, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(job_name) DO UPDATE SET
			max_duration_secs = excluded.max_duration_secs,
			min_success_rate  = excluded.min_success_rate,
			updated_at        = excluded.updated_at`,
		p.JobName, p.MaxDurationSecs, p.MinSuccessRate,
		p.CreatedAt.UTC(), p.UpdatedAt.UTC(),
	)
	return err
}

// GetAlertProfile returns the alert profile for a job, or sql.ErrNoRows.
func GetAlertProfile(db *sql.DB, jobName string) (AlertProfile, error) {
	var p AlertProfile
	err := db.QueryRow(`
		SELECT id, job_name, max_duration_secs, min_success_rate, created_at, updated_at
		FROM job_alert_profiles WHERE job_name = ?`, jobName).
		Scan(&p.ID, &p.JobName, &p.MaxDurationSecs, &p.MinSuccessRate, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return AlertProfile{}, fmt.Errorf("get alert profile: %w", err)
	}
	return p, nil
}

// DeleteAlertProfile removes the alert profile for a job.
func DeleteAlertProfile(db *sql.DB, jobName string) error {
	_, err := db.Exec(`DELETE FROM job_alert_profiles WHERE job_name = ?`, jobName)
	return err
}
