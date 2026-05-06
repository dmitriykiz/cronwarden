package db

import (
	"database/sql"
	"fmt"
)

// SetJobLabels replaces all labels for a given job name with the provided map.
func SetJobLabels(db *sql.DB, jobName string, labels map[string]string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	_, err = tx.Exec(`DELETE FROM job_labels WHERE job_name = ?`, jobName)
	if err != nil {
		return fmt.Errorf("delete labels: %w", err)
	}

	for k, v := range labels {
		_, err = tx.Exec(
			`INSERT INTO job_labels (job_name, key, value) VALUES (?, ?, ?)`,
			jobName, k, v,
		)
		if err != nil {
			return fmt.Errorf("insert label %q: %w", k, err)
		}
	}

	return tx.Commit()
}

// ListJobLabels returns all labels for the given job name as a key/value map.
func ListJobLabels(db *sql.DB, jobName string) (map[string]string, error) {
	rows, err := db.Query(
		`SELECT key, value FROM job_labels WHERE job_name = ? ORDER BY key`,
		jobName,
	)
	if err != nil {
		return nil, fmt.Errorf("query labels: %w", err)
	}
	defer rows.Close()

	labels := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		labels[k] = v
	}
	return labels, rows.Err()
}

// DeleteJobLabels removes all labels for the given job name.
func DeleteJobLabels(db *sql.DB, jobName string) error {
	_, err := db.Exec(`DELETE FROM job_labels WHERE job_name = ?`, jobName)
	if err != nil {
		return fmt.Errorf("delete labels: %w", err)
	}
	return nil
}
