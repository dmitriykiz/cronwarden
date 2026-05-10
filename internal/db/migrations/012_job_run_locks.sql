-- +migrate Up
CREATE TABLE IF NOT EXISTS job_run_locks (
    job_name   TEXT PRIMARY KEY,
    locked_by  TEXT NOT NULL,
    locked_at  TEXT NOT NULL,
    expires_at TEXT NOT NULL
);

-- +migrate Down
DROP TABLE IF EXISTS job_run_locks;
