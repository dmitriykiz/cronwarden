-- +migrate Up
CREATE TABLE IF NOT EXISTS job_concurrency_policies (
    job_name   TEXT    NOT NULL PRIMARY KEY,
    policy     TEXT    NOT NULL DEFAULT 'allow', -- 'allow', 'skip', 'queue'
    max_queue  INTEGER NOT NULL DEFAULT 0,
    updated_at DATETIME NOT NULL
);

-- +migrate Down
DROP TABLE IF EXISTS job_concurrency_policies;
