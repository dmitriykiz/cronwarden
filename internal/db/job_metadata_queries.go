package db

import (
	"database/sql"
	"errors"
	"time"
)

// JobMetadata holds arbitrary key/value metadata attached to a job.
type JobMetadata struct {
	JobName   string    `json:"job_name"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpsertJobMetadata inserts or replaces a metadata key for a job.
func UpsertJobMetadata(db *sql.DB, jobName, key, value string) error {
	_, err := db.Exec(`
		INSERT INTO job_metadata (job_name, key, value, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(job_name, key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`,
		jobName, key, value, time.Now().UTC())
	return err
}

// ListJobMetadata returns all metadata entries for a given job.
func ListJobMetadata(db *sql.DB, jobName string) ([]JobMetadata, error) {
	rows, err := db.Query(`
		SELECT job_name, key, value, updated_at
		FROM job_metadata
		WHERE job_name = ?
		ORDER BY key ASC`, jobName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []JobMetadata
	for rows.Next() {
		var m JobMetadata
		if err := rows.Scan(&m.JobName, &m.Key, &m.Value, &m.UpdatedAt); err != nil {
			return nil, err
		}
		results = append(results, m)
	}
	return results, rows.Err()
}

// DeleteJobMetadata removes a specific metadata key for a job.
func DeleteJobMetadata(db *sql.DB, jobName, key string) error {
	res, err := db.Exec(`DELETE FROM job_metadata WHERE job_name = ? AND key = ?`, jobName, key)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("metadata key not found")
	}
	return nil
}

// DeleteAllJobMetadata removes all metadata for a job.
func DeleteAllJobMetadata(db *sql.DB, jobName string) error {
	_, err := db.Exec(`DELETE FROM job_metadata WHERE job_name = ?`, jobName)
	return err
}
