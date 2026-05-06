-- Migration 005: job_labels table
-- Stores arbitrary key/value metadata labels per job name.
-- Labels are distinct from tags in that they carry a value (key=value pairs)
-- rather than being plain string identifiers.

CREATE TABLE IF NOT EXISTS job_labels (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    job_name  TEXT    NOT NULL,
    key       TEXT    NOT NULL,
    value     TEXT    NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    UNIQUE (job_name, key)
);

CREATE INDEX IF NOT EXISTS idx_job_labels_job_name ON job_labels (job_name);
