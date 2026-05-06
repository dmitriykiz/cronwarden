package db

import (
	"database/sql"
	"errors"
	"fmt"
)

// JobDependency represents a dependency relationship between two jobs.
type JobDependency struct {
	ID         int64  `json:"id"`
	JobName    string `json:"job_name"`
	DependsOn  string `json:"depends_on"`
	MaxAgeSecs int    `json:"max_age_secs"`
}

// AddJobDependency inserts a dependency for a job. If the same pair already
// exists, it updates max_age_secs.
func AddJobDependency(db *sql.DB, jobName, dependsOn string, maxAgeSecs int) error {
	if jobName == dependsOn {
		return fmt.Errorf("job cannot depend on itself: %q", jobName)
	}
	_, err := db.Exec(`
		INSERT INTO job_dependencies (job_name, depends_on, max_age_secs)
		VALUES (?, ?, ?)
		ON CONFLICT(job_name, depends_on) DO UPDATE SET max_age_secs = excluded.max_age_secs`,
		jobName, dependsOn, maxAgeSecs)
	return err
}

// ListJobDependencies returns all dependencies registered for the given job.
func ListJobDependencies(db *sql.DB, jobName string) ([]JobDependency, error) {
	rows, err := db.Query(`
		SELECT id, job_name, depends_on, max_age_secs
		FROM job_dependencies
		WHERE job_name = ?
		ORDER BY id ASC`, jobName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deps []JobDependency
	for rows.Next() {
		var d JobDependency
		if err := rows.Scan(&d.ID, &d.JobName, &d.DependsOn, &d.MaxAgeSecs); err != nil {
			return nil, err
		}
		deps = append(deps, d)
	}
	return deps, rows.Err()
}

// DeleteJobDependency removes a specific dependency by ID.
// Returns sql.ErrNoRows if the record does not exist.
func DeleteJobDependency(db *sql.DB, id int64) error {
	res, err := db.Exec(`DELETE FROM job_dependencies WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("sql: no rows in result set")
	}
	return nil
}
