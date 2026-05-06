package db

import (
	"database/sql"
	"time"
)

// JobNote holds an operator-supplied annotation attached to a specific job run.
type JobNote struct {
	ID        int64     `json:"id"`
	RunID     int64     `json:"run_id"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

// AddJobNote inserts a note for the given run_id and returns the new row id.
func AddJobNote(db *sql.DB, runID int64, note string) (int64, error) {
	res, err := db.Exec(
		`INSERT INTO job_notes (run_id, note, created_at) VALUES (?, ?, ?)`,
		runID, note, time.Now().UTC(),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ListJobNotes returns all notes for the given run_id ordered by creation time.
func ListJobNotes(db *sql.DB, runID int64) ([]JobNote, error) {
	rows, err := db.Query(
		`SELECT id, run_id, note, created_at FROM job_notes WHERE run_id = ? ORDER BY created_at ASC`,
		runID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []JobNote
	for rows.Next() {
		var n JobNote
		if err := rows.Scan(&n.ID, &n.RunID, &n.Note, &n.CreatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

// DeleteJobNote removes a single note by its id. Returns (true, nil) when found.
func DeleteJobNote(db *sql.DB, noteID int64) (bool, error) {
	res, err := db.Exec(`DELETE FROM job_notes WHERE id = ?`, noteID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}
