-- +migrate Up
CREATE TABLE IF NOT EXISTS job_run_events (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id     INTEGER NOT NULL REFERENCES job_runs(id) ON DELETE CASCADE,
    event_type TEXT    NOT NULL,
    message    TEXT    NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_job_run_events_run_id ON job_run_events(run_id);

-- +migrate Down
DROP TABLE IF EXISTS job_run_events;
