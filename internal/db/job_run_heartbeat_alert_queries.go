package db

import (
	"database/sql"
	"time"
)

// HeartbeatAlert represents a stale heartbeat that should trigger an alert.
type HeartbeatAlert struct {
	RunID     int64
	JobName   string
	LastSeen  time.Time
	TimeoutAt time.Time
}

// ListExpiredHeartbeatAlerts returns heartbeats whose timeout has passed and
// have not been acknowledged (deleted). Results are ordered oldest-first.
func ListExpiredHeartbeatAlerts(db *sql.DB, now time.Time) ([]HeartbeatAlert, error) {
	const q = `
		SELECT h.run_id, r.job_name, h.last_seen_at, h.timeout_at
		FROM job_run_heartbeats h
		JOIN job_runs r ON r.id = h.run_id
		WHERE h.timeout_at <= ?
		ORDER BY h.timeout_at ASC`

	rows, err := db.Query(q, now.UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []HeartbeatAlert
	for rows.Next() {
		var a HeartbeatAlert
		if err := rows.Scan(&a.RunID, &a.JobName, &a.LastSeen, &a.TimeoutAt); err != nil {
			return nil, err
		}
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}
