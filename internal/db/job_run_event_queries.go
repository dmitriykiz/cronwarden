package db

import (
	"database/sql"
	"time"
)

// JobRunEvent represents a structured event emitted during a job run,
// such as a retry attempt, timeout warning, or custom milestone.
type JobRunEvent struct {
	ID        int64     `json:"id"`
	RunID     int64     `json:"run_id"`
	EventType string    `json:"event_type"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// AddJobRunEvent inserts a new event record associated with a job run.
func AddJobRunEvent(db *sql.DB, runID int64, eventType, message string) (int64, error) {
	const q = `
		INSERT INTO job_run_events (run_id, event_type, message, created_at)
		VALUES (?, ?, ?, ?)
	`
	now := time.Now().UTC()
	res, err := db.Exec(q, runID, eventType, message, now)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ListJobRunEvents returns all events for a given run, ordered by creation time.
func ListJobRunEvents(db *sql.DB, runID int64) ([]JobRunEvent, error) {
	const q = `
		SELECT id, run_id, event_type, message, created_at
		FROM job_run_events
		WHERE run_id = ?
		ORDER BY created_at ASC
	`
	rows, err := db.Query(q, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []JobRunEvent
	for rows.Next() {
		var e JobRunEvent
		if err := rows.Scan(&e.ID, &e.RunID, &e.EventType, &e.Message, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// DeleteJobRunEvent removes a single event by its ID.
// Returns (true, nil) if the row was deleted, (false, nil) if it did not exist.
func DeleteJobRunEvent(db *sql.DB, id int64) (bool, error) {
	const q = `DELETE FROM job_run_events WHERE id = ?`
	res, err := db.Exec(q, id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// DeleteAllJobRunEvents removes every event associated with a job run.
// Useful when purging a run record entirely.
func DeleteAllJobRunEvents(db *sql.DB, runID int64) error {
	const q = `DELETE FROM job_run_events WHERE run_id = ?`
	_, err := db.Exec(q, runID)
	return err
}
