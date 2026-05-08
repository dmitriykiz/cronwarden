package db

import (
	"database/sql"
	"errors"
	"time"
)

// JobRunArtifact represents a named artifact (file path or URL) attached to a job run.
type JobRunArtifact struct {
	ID        int64     `json:"id"`
	RunID     int64     `json:"run_id"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
}

// AddJobRunArtifact inserts a new artifact record for the given run.
func AddJobRunArtifact(db *sql.DB, runID int64, name, path string) (int64, error) {
	res, err := db.Exec(
		`INSERT INTO job_run_artifacts (run_id, name, path, created_at) VALUES (?, ?, ?, ?)`,
		runID, name, path, time.Now().UTC(),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ListJobRunArtifacts returns all artifacts for the given run ID.
func ListJobRunArtifacts(db *sql.DB, runID int64) ([]JobRunArtifact, error) {
	rows, err := db.Query(
		`SELECT id, run_id, name, path, created_at FROM job_run_artifacts WHERE run_id = ? ORDER BY id ASC`,
		runID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var artifacts []JobRunArtifact
	for rows.Next() {
		var a JobRunArtifact
		if err := rows.Scan(&a.ID, &a.RunID, &a.Name, &a.Path, &a.CreatedAt); err != nil {
			return nil, err
		}
		artifacts = append(artifacts, a)
	}
	return artifacts, rows.Err()
}

// DeleteJobRunArtifact removes a single artifact by its ID.
// Returns sql.ErrNoRows if the artifact does not exist.
func DeleteJobRunArtifact(db *sql.DB, id int64) error {
	res, err := db.Exec(`DELETE FROM job_run_artifacts WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("artifact not found")
	}
	return nil
}
