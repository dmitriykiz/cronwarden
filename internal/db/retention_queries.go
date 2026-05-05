package db

import (
	"fmt"
	"time"
)

// DeleteRunsBefore removes all job_runs rows with a started_at earlier than cutoff.
// Returns the number of rows deleted.
func (d *DB) DeleteRunsBefore(cutoff time.Time) (int64, error) {
	res, err := d.conn.Exec(
		`DELETE FROM job_runs WHERE started_at < ?`,
		cutoff.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return 0, fmt.Errorf("db: DeleteRunsBefore: %w", err)
	}
	return res.RowsAffected()
}

// TrimRunsToLimit deletes the oldest rows so that at most maxRows rows remain
// in job_runs. Returns the number of rows deleted.
func (d *DB) TrimRunsToLimit(maxRows int) (int64, error) {
	var count int
	err := d.conn.QueryRow(`SELECT COUNT(*) FROM job_runs`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("db: TrimRunsToLimit count: %w", err)
	}

	excess := count - maxRows
	if excess <= 0 {
		return 0, nil
	}

	res, err := d.conn.Exec(`
		DELETE FROM job_runs
		WHERE id IN (
			SELECT id FROM job_runs ORDER BY started_at ASC LIMIT ?
		)`, excess)
	if err != nil {
		return 0, fmt.Errorf("db: TrimRunsToLimit delete: %w", err)
	}
	return res.RowsAffected()
}
