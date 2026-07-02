-- +goose Up
CREATE TABLE scripts (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    shell      TEXT NOT NULL DEFAULT 'shell', -- powershell | shell
    content    TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL
);

CREATE TABLE policies (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT ''
);

CREATE TABLE policy_checks (
    id        TEXT PRIMARY KEY,
    policy_id TEXT NOT NULL REFERENCES policies(id) ON DELETE CASCADE,
    name      TEXT NOT NULL DEFAULT '',
    type      TEXT NOT NULL,                  -- disk | memory | cpu | updates | script
    config    TEXT NOT NULL DEFAULT '{}',     -- JSON
    script_id TEXT REFERENCES scripts(id) ON DELETE SET NULL
);
CREATE INDEX idx_policy_checks_policy ON policy_checks(policy_id);

CREATE TABLE policy_tasks (
    id               TEXT PRIMARY KEY,
    policy_id        TEXT NOT NULL REFERENCES policies(id) ON DELETE CASCADE,
    name             TEXT NOT NULL DEFAULT '',
    script_id        TEXT REFERENCES scripts(id) ON DELETE SET NULL,
    interval_minutes INTEGER NOT NULL DEFAULT 60
);
CREATE INDEX idx_policy_tasks_policy ON policy_tasks(policy_id);

CREATE TABLE policy_assignments (
    id          TEXT PRIMARY KEY,
    policy_id   TEXT NOT NULL REFERENCES policies(id) ON DELETE CASCADE,
    target_type TEXT NOT NULL,                -- client | site | device
    target_id   TEXT NOT NULL
);
CREATE INDEX idx_assign_target ON policy_assignments(target_type, target_id);
CREATE INDEX idx_assign_policy ON policy_assignments(policy_id);

CREATE TABLE check_results (
    device_id  TEXT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    check_id   TEXT NOT NULL,
    status     TEXT NOT NULL DEFAULT 'unknown',
    output     TEXT NOT NULL DEFAULT '',
    value      REAL,
    updated_at TIMESTAMP NOT NULL,
    PRIMARY KEY (device_id, check_id)
);

CREATE TABLE task_results (
    id        TEXT PRIMARY KEY,
    device_id TEXT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    task_id   TEXT NOT NULL,
    exit_code INTEGER NOT NULL DEFAULT 0,
    output    TEXT NOT NULL DEFAULT '',
    ran_at    TIMESTAMP NOT NULL
);
CREATE INDEX idx_task_results_device ON task_results(device_id, ran_at);

-- +goose Down
DROP TABLE task_results;
DROP TABLE check_results;
DROP TABLE policy_assignments;
DROP TABLE policy_tasks;
DROP TABLE policy_checks;
DROP TABLE policies;
DROP TABLE scripts;
