package db

import (
	"database/sql"
	"errors"
	"time"
)

// JobRunAnnotation holds a key/value annotation attached to a specific job run.
type JobRunAnnotation struct {
	ID        int64     `json:"id"`
	RunID     int64     `json:"run_id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
}

// UpsertJobRunAnnotation inserts or replaces an annotation key for a given run.
func UpsertJobRunAnnotation(db *sql.DB, runID int64, key, value string) error {
	_, err := db.Exec(`
		INSERT INTO job_run_annotations (run_id, key, value, created_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(run_id, key) DO UPDATE SET value = excluded.value, created_at = excluded.created_at`,
		runID, key, value, time.Now().UTC())
	return err
}

// ListJobRunAnnotations returns all annotations for a given run.
func ListJobRunAnnotations(db *sql.DB, runID int64) ([]JobRunAnnotation, error) {
	rows, err := db.Query(`
		SELECT id, run_id, key, value, created_at
		FROM job_run_annotations
		WHERE run_id = ?
		ORDER BY created_at ASC`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []JobRunAnnotation
	for rows.Next() {
		var a JobRunAnnotation
		if err := rows.Scan(&a.ID, &a.RunID, &a.Key, &a.Value, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// DeleteJobRunAnnotation removes a single annotation by its ID.
func DeleteJobRunAnnotation(db *sql.DB, annotationID int64) error {
	res, err := db.Exec(`DELETE FROM job_run_annotations WHERE id = ?`, annotationID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("annotation not found")
	}
	return nil
}
