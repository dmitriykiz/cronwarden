package db

import (
	"database/sql"
	"fmt"
)

// SetJobTags replaces all tags for a given job name.
func SetJobTags(db *sql.DB, jobName string, tags []string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	_, err = tx.Exec(`DELETE FROM job_tags WHERE job_name = ?`, jobName)
	if err != nil {
		return fmt.Errorf("delete tags: %w", err)
	}

	for _, tag := range tags {
		_, err = tx.Exec(`INSERT INTO job_tags (job_name, tag) VALUES (?, ?)`, jobName, tag)
		if err != nil {
			return fmt.Errorf("insert tag %q: %w", tag, err)
		}
	}

	return tx.Commit()
}

// ListJobTags returns all tags associated with a job name.
func ListJobTags(db *sql.DB, jobName string) ([]string, error) {
	rows, err := db.Query(`SELECT tag FROM job_tags WHERE job_name = ? ORDER BY tag ASC`, jobName)
	if err != nil {
		return nil, fmt.Errorf("query tags: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

// DeleteJobTags removes all tags for a job.
func DeleteJobTags(db *sql.DB, jobName string) error {
	_, err := db.Exec(`DELETE FROM job_tags WHERE job_name = ?`, jobName)
	return err
}
