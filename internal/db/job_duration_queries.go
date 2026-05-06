package db

import (
	"database/sql"
	"time"
)

// DurationStat holds aggregated duration statistics for a job.
type DurationStat struct {
	JobName    string        `json:"job_name"`
	AvgSeconds float64       `json:"avg_seconds"`
	MinSeconds float64       `json:"min_seconds"`
	MaxSeconds float64       `json:"max_seconds"`
	SampleSize int           `json:"sample_size"`
	LastRun    time.Time     `json:"last_run"`
}

// GetJobDurationStats returns duration aggregates for a specific job over the
// last n runs. If limit <= 0 all recorded runs are considered.
func GetJobDurationStats(db *sql.DB, jobName string, limit int) (*DurationStat, error) {
	query := `
		SELECT
			job_name,
			AVG(duration_ms) / 1000.0,
			MIN(duration_ms) / 1000.0,
			MAX(duration_ms) / 1000.0,
			COUNT(*),
			MAX(started_at)
		FROM (
			SELECT job_name, duration_ms, started_at
			FROM job_runs
			WHERE job_name = ? AND duration_ms IS NOT NULL
			ORDER BY started_at DESC
			LIMIT ?
		)
	`
	if limit <= 0 {
		limit = -1
	}

	var stat DurationStat
	var lastRunStr string
	err := db.QueryRow(query, jobName, limit).Scan(
		&stat.JobName,
		&stat.AvgSeconds,
		&stat.MinSeconds,
		&stat.MaxSeconds,
		&stat.SampleSize,
		&lastRunStr,
	)
	if err != nil {
		return nil, err
	}

	stat.LastRun, err = time.Parse(time.RFC3339, lastRunStr)
	if err != nil {
		stat.LastRun, _ = time.Parse("2006-01-02 15:04:05", lastRunStr)
	}
	return &stat, nil
}
