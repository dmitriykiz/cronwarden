package db

import (
	"database/sql"
	"errors"
	"time"
)

// JobScheduleOverride represents a temporary schedule override for a job.
type JobScheduleOverride struct {
	ID        int64     `json:"id"`
	JobName   string    `json:"job_name"`
	CronExpr  string    `json:"cron_expr"`
	Reason    string    `json:"reason"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// UpsertJobScheduleOverride inserts or replaces a schedule override for a job.
func UpsertJobScheduleOverride(db *sql.DB, o JobScheduleOverride) error {
	_, err := db.Exec(`
		INSERT INTO job_schedule_overrides (job_name, cron_expr, reason, expires_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(job_name) DO UPDATE SET
			cron_expr  = excluded.cron_expr,
			reason     = excluded.reason,
			expires_at = excluded.expires_at
	`, o.JobName, o.CronExpr, o.Reason, o.ExpiresAt.UTC())
	return err
}

// GetJobScheduleOverride returns the active override for a job, if any.
// Returns sql.ErrNoRows if none exists.
func GetJobScheduleOverride(db *sql.DB, jobName string) (JobScheduleOverride, error) {
	row := db.QueryRow(`
		SELECT id, job_name, cron_expr, reason, expires_at, created_at
		FROM job_schedule_overrides
		WHERE job_name = ?
	`, jobName)

	var o JobScheduleOverride
	var expiresAt, createdAt string
	err := row.Scan(&o.ID, &o.JobName, &o.CronExpr, &o.Reason, &expiresAt, &createdAt)
	if err != nil {
		return JobScheduleOverride{}, err
	}
	o.ExpiresAt, _ = time.Parse(time.RFC3339, expiresAt)
	o.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return o, nil
}

// DeleteJobScheduleOverride removes a schedule override for a job.
// Returns sql.ErrNoRows if no override existed.
func DeleteJobScheduleOverride(db *sql.DB, jobName string) error {
	res, err := db.Exec(`DELETE FROM job_schedule_overrides WHERE job_name = ?`, jobName)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("sql: no rows in result set")
	}
	return nil
}

// ListActiveJobScheduleOverrides returns all overrides that have not yet expired.
func ListActiveJobScheduleOverrides(db *sql.DB) ([]JobScheduleOverride, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	rows, err := db.Query(`
		SELECT id, job_name, cron_expr, reason, expires_at, created_at
		FROM job_schedule_overrides
		WHERE expires_at > ?
		ORDER BY job_name
	`, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []JobScheduleOverride
	for rows.Next() {
		var o JobScheduleOverride
		var expiresAt, createdAt string
		if err := rows.Scan(&o.ID, &o.JobName, &o.CronExpr, &o.Reason, &expiresAt, &createdAt); err != nil {
			return nil, err
		}
		o.ExpiresAt, _ = time.Parse(time.RFC3339, expiresAt)
		o.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		out = append(out, o)
	}
	return out, rows.Err()
}
