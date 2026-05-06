package db

import (
	"database/sql"
	"time"
)

// JobSilence represents a silence window that suppresses alerts for a job.
type JobSilence struct {
	ID        int64     `json:"id"`
	JobName   string    `json:"job_name"`
	StartsAt  time.Time `json:"starts_at"`
	EndsAt    time.Time `json:"ends_at"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}

// UpsertJobSilence inserts or replaces a silence window for a job.
func UpsertJobSilence(db *sql.DB, s JobSilence) (int64, error) {
	res, err := db.Exec(`
		INSERT INTO job_silences (job_name, starts_at, ends_at, reason)
		VALUES (?, ?, ?, ?)`,
		s.JobName, s.StartsAt.UTC(), s.EndsAt.UTC(), s.Reason,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ListJobSilences returns all silence windows for the given job name.
func ListJobSilences(db *sql.DB, jobName string) ([]JobSilence, error) {
	rows, err := db.Query(`
		SELECT id, job_name, starts_at, ends_at, reason, created_at
		FROM job_silences
		WHERE job_name = ?
		ORDER BY starts_at ASC`, jobName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []JobSilence
	for rows.Next() {
		var s JobSilence
		if err := rows.Scan(&s.ID, &s.JobName, &s.StartsAt, &s.EndsAt, &s.Reason, &s.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// DeleteJobSilence removes a silence window by ID.
func DeleteJobSilence(db *sql.DB, id int64) (bool, error) {
	res, err := db.Exec(`DELETE FROM job_silences WHERE id = ?`, id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

// IsJobSilenced returns true if the job has an active silence at the given time.
func IsJobSilenced(db *sql.DB, jobName string, at time.Time) (bool, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM job_silences
		WHERE job_name = ? AND starts_at <= ? AND ends_at >= ?`,
		jobName, at.UTC(), at.UTC(),
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
