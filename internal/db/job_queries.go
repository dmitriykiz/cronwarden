package db

import (
	"database/sql"
	"time"
)

// JobRun mirrors the schema row returned from the job_runs table.
type JobRunRow struct {
	ID        int64
	JobName   string
	StartedAt time.Time
	Duration  float64 // seconds
	ExitCode  int
	Output    string
}

// ListJobRunsByName returns up to limit recent runs for the given job name,
// ordered newest-first.
func ListJobRunsByName(db *sql.DB, name string, limit int) ([]JobRunRow, error) {
	const q = `
		SELECT id, job_name, started_at, duration_seconds, exit_code, output
		FROM job_runs
		WHERE job_name = ?
		ORDER BY started_at DESC
		LIMIT ?`

	rows, err := db.Query(q, name, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []JobRunRow
	for rows.Next() {
		var row JobRunRow
		var startedAtStr string
		if err := rows.Scan(
			&row.ID, &row.JobName, &startedAtStr,
			&row.Duration, &row.ExitCode, &row.Output,
		); err != nil {
			return nil, err
		}
		row.StartedAt, _ = time.Parse(time.RFC3339, startedAtStr)
		result = append(result, row)
	}
	return result, rows.Err()
}

// JobStats holds aggregate metrics computed from a slice of runs.
type JobStats struct {
	JobName     string  `json:"job_name"`
	TotalRuns   int     `json:"total_runs"`
	SuccessRuns int     `json:"success_runs"`
	FailureRuns int     `json:"failure_runs"`
	AvgDuration float64 `json:"avg_duration_seconds"`
}

// computeStats derives a JobStats summary from a slice of JobRunRows.
func computeStats(name string, runs []JobRunRow) JobStats {
	stats := JobStats{JobName: name, TotalRuns: len(runs)}
	if len(runs) == 0 {
		return stats
	}
	var totalDur float64
	for _, r := range runs {
		if r.ExitCode == 0 {
			stats.SuccessRuns++
		} else {
			stats.FailureRuns++
		}
		totalDur += r.Duration
	}
	stats.AvgDuration = totalDur / float64(len(runs))
	return stats
}
