package db

import (
	"database/sql"
	"time"
)

// JobRunWindow holds aggregated stats for a job within a time window.
type JobRunWindow struct {
	JobName      string
	WindowStart  time.Time
	WindowEnd    time.Time
	TotalRuns    int
	SuccessCount int
	FailureCount int
	AvgDurationS float64
	MaxDurationS float64
}

// GetJobRunWindow returns aggregated run stats for a named job within [from, to].
func GetJobRunWindow(db *sql.DB, jobName string, from, to time.Time) (*JobRunWindow, error) {
	const q = `
		SELECT
			COUNT(*) AS total,
			SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) AS successes,
			SUM(CASE WHEN status != 'success' THEN 1 ELSE 0 END) AS failures,
			AVG(duration_seconds) AS avg_dur,
			MAX(duration_seconds) AS max_dur
		FROM job_runs
		WHERE job_name = ?
		  AND started_at >= ?
		  AND started_at <= ?
	`

	row := db.QueryRow(q, jobName, from.UTC().Format(time.RFC3339), to.UTC().Format(time.RFC3339))

	var (
		total, successes, failures int
		avgDur, maxDur             sql.NullFloat64
	)
	if err := row.Scan(&total, &successes, &failures, &avgDur, &maxDur); err != nil {
		return nil, err
	}

	return &JobRunWindow{
		JobName:      jobName,
		WindowStart:  from,
		WindowEnd:    to,
		TotalRuns:    total,
		SuccessCount: successes,
		FailureCount: failures,
		AvgDurationS: avgDur.Float64,
		MaxDurationS: maxDur.Float64,
	}, nil
}
