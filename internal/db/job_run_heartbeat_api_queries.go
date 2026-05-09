package db

import (
	"database/sql"
	"time"
)

// JobRunHeartbeatSummary holds a lightweight view of a heartbeat for API listing.
type JobRunHeartbeatSummary struct {
	RunID     int64     `json:"run_id"`
	JobName   string    `json:"job_name"`
	BeatAt    time.Time `json:"beat_at"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ListActiveHeartbeats returns all heartbeats whose expires_at is after now.
func ListActiveHeartbeats(db *sql.DB) ([]JobRunHeartbeatSummary, error) {
	const q = `
		SELECT h.run_id, r.job_name, h.beat_at, h.status, h.expires_at
		FROM job_run_heartbeats h
		JOIN job_runs r ON r.id = h.run_id
		WHERE h.expires_at > ?
		ORDER BY h.beat_at DESC`

	rows, err := db.Query(q, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []JobRunHeartbeatSummary
	for rows.Next() {
		var s JobRunHeartbeatSummary
		if err := rows.Scan(&s.RunID, &s.JobName, &s.BeatAt, &s.Status, &s.ExpiresAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
