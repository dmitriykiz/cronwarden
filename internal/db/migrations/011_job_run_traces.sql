-- +migrate Up
CREATE TABLE IF NOT EXISTS job_run_traces (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id     INTEGER NOT NULL UNIQUE REFERENCES job_runs(id) ON DELETE CASCADE,
    trace_id   TEXT    NOT NULL,
    span_id    TEXT    NOT NULL,
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_job_run_traces_run_id ON job_run_traces(run_id);
