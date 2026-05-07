-- Migration 006: job environment variable storage
CREATE TABLE IF NOT EXISTS job_env_vars (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    job_name  TEXT    NOT NULL,
    key       TEXT    NOT NULL,
    value     TEXT    NOT NULL DEFAULT '',
    UNIQUE(job_name, key)
);

CREATE INDEX IF NOT EXISTS idx_job_env_vars_job_name ON job_env_vars(job_name);
