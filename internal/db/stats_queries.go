package db

import (
	"database/sql"
	"time"
)

// GlobalStats holds aggregate metrics across all jobs.
type GlobalStats struct {
	TotalRuns    int
	SuccessRuns  int
	FailureRuns  int
	UniqueJobs   int
	LastRunAt    *time.Time
}

// ComputeGlobalStats returns aggregate run statistics across all jobs.
func ComputeGlobalStats(db *sql.DB) (GlobalStats, error) {
	const q = `
		SELECT
			COUNT(*) AS total,
			SUM(CASE WHEN exit_code = 0 THEN 1 ELSE 0 END) AS success,
			SUM(CASE WHEN exit_code != 0 THEN 1 ELSE 0 END) AS failure,
			COUNT(DISTINCT job_name) AS unique_jobs,
			MAX(started_at) AS last_run
		FROM job_runs
	`
	row := db.QueryRow(q)

	var s GlobalStats
	var lastRun sql.NullString
	if err := row.Scan(&s.TotalRuns, &s.SuccessRuns, &s.FailureRuns, &s.UniqueJobs, &lastRun); err != nil {
		return GlobalStats{}, err
	}
	if lastRun.Valid && lastRun.String != "" {
		t, err := time.Parse(time.RFC3339, lastRun.String)
		if err == nil {
			s.LastRunAt = &t
		}
	}
	return s, nil
}
