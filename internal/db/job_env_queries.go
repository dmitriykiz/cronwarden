package db

import (
	"database/sql"
	"fmt"
)

// UpsertJobEnvVar sets or replaces an environment variable for a job.
func UpsertJobEnvVar(db *sql.DB, jobName, key, value string) error {
	_, err := db.Exec(`
		INSERT INTO job_env_vars (job_name, key, value)
		VALUES (?, ?, ?)
		ON CONFLICT(job_name, key) DO UPDATE SET value = excluded.value
	`, jobName, key, value)
	if err != nil {
		return fmt.Errorf("upsert job env var: %w", err)
	}
	return nil
}

// ListJobEnvVars returns all env vars for a given job as a map.
func ListJobEnvVars(db *sql.DB, jobName string) (map[string]string, error) {
	rows, err := db.Query(`
		SELECT key, value FROM job_env_vars WHERE job_name = ? ORDER BY key
	`, jobName)
	if err != nil {
		return nil, fmt.Errorf("list job env vars: %w", err)
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, fmt.Errorf("scan job env var: %w", err)
		}
		result[k] = v
	}
	return result, rows.Err()
}

// DeleteJobEnvVar removes a single env var for a job.
func DeleteJobEnvVar(db *sql.DB, jobName, key string) error {
	_, err := db.Exec(`DELETE FROM job_env_vars WHERE job_name = ? AND key = ?`, jobName, key)
	if err != nil {
		return fmt.Errorf("delete job env var: %w", err)
	}
	return nil
}

// DeleteAllJobEnvVars removes all env vars for a job.
func DeleteAllJobEnvVars(db *sql.DB, jobName string) error {
	_, err := db.Exec(`DELETE FROM job_env_vars WHERE job_name = ?`, jobName)
	if err != nil {
		return fmt.Errorf("delete all job env vars: %w", err)
	}
	return nil
}
