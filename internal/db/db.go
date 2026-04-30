package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps a SQLite database connection.
type DB struct {
	conn *sql.DB
}

// JobRun represents a single cron job execution record.
type JobRun struct {
	ID        int64
	JobName   string
	StartedAt time.Time
	FinishedAt *time.Time
	ExitCode  int
	Output    string
	Success   bool
}

// Open opens (or creates) the SQLite database at the given path and runs migrations.
func Open(path string) (*DB, error) {
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("db open: %w", err)
	}
	conn.SetMaxOpenConns(1)
	d := &DB{conn: conn}
	if err := d.migrate(); err != nil {
		return nil, fmt.Errorf("db migrate: %w", err)
	}
	return d, nil
}

// Close closes the database connection.
func (d *DB) Close() error {
	return d.conn.Close()
}

func (d *DB) migrate() error {
	_, err := d.conn.Exec(`
		CREATE TABLE IF NOT EXISTS job_runs (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			job_name    TEXT    NOT NULL,
			started_at  DATETIME NOT NULL,
			finished_at DATETIME,
			exit_code   INTEGER NOT NULL DEFAULT 0,
			output      TEXT    NOT NULL DEFAULT '',
			success     BOOLEAN NOT NULL DEFAULT 0
		);
	`)
	return err
}

// InsertJobRun persists a completed job run and returns its assigned ID.
func (d *DB) InsertJobRun(r *JobRun) (int64, error) {
	res, err := d.conn.Exec(
		`INSERT INTO job_runs (job_name, started_at, finished_at, exit_code, output, success)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		r.JobName, r.StartedAt, r.FinishedAt, r.ExitCode, r.Output, r.Success,
	)
	if err != nil {
		return 0, fmt.Errorf("insert job run: %w", err)
	}
	return res.LastInsertId()
}

// ListJobRuns returns the most recent `limit` runs for the given job name.
func (d *DB) ListJobRuns(jobName string, limit int) ([]JobRun, error) {
	rows, err := d.conn.Query(
		`SELECT id, job_name, started_at, finished_at, exit_code, output, success
		 FROM job_runs WHERE job_name = ? ORDER BY started_at DESC LIMIT ?`,
		jobName, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list job runs: %w", err)
	}
	defer rows.Close()

	var runs []JobRun
	for rows.Next() {
		var r JobRun
		if err := rows.Scan(&r.ID, &r.JobName, &r.StartedAt, &r.FinishedAt,
			&r.ExitCode, &r.Output, &r.Success); err != nil {
			return nil, fmt.Errorf("scan job run: %w", err)
		}
		runs = append(runs, r)
	}
	return runs, rows.Err()
}
