-- Migration 009: job run artifacts
-- Stores file paths or URLs produced by a job run.

CREATE TABLE IF NOT EXISTS job_run_artifacts (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id     INTEGER NOT NULL REFERENCES job_runs(id) ON DELETE CASCADE,
    name       TEXT    NOT NULL,
    path       TEXT    NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_job_run_artifacts_run_id ON job_run_artifacts(run_id);
