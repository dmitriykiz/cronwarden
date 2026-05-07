-- Migration 008: job run annotations
-- Stores arbitrary key/value metadata attached to individual job runs.

CREATE TABLE IF NOT EXISTS job_run_annotations (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id     INTEGER NOT NULL,
    key        TEXT    NOT NULL,
    value      TEXT    NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL,
    UNIQUE(run_id, key),
    FOREIGN KEY (run_id) REFERENCES job_runs(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_job_run_annotations_run_id ON job_run_annotations(run_id);
