-- +goose Up
-- Portables SQL für SQLite und PostgreSQL: nur TEXT/INTEGER/BOOLEAN/TIMESTAMP,
-- IDs sind in Go erzeugte UUID-Strings, Zeiten werden als UTC gespeichert.

CREATE TABLE users (
    id            TEXT PRIMARY KEY,
    username      TEXT NOT NULL UNIQUE,
    email         TEXT NOT NULL DEFAULT '',
    password_hash TEXT NOT NULL DEFAULT '',
    role          TEXT NOT NULL DEFAULT 'viewer',
    auth_source   TEXT NOT NULL DEFAULT 'local',
    created_at    TIMESTAMP NOT NULL,
    last_login    TIMESTAMP
);

CREATE TABLE sessions (
    token_hash TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL
);
CREATE INDEX idx_sessions_user ON sessions(user_id);

CREATE TABLE ldap_config (
    id          INTEGER PRIMARY KEY,
    enabled     BOOLEAN NOT NULL DEFAULT FALSE,
    host        TEXT NOT NULL DEFAULT '',
    port        INTEGER NOT NULL DEFAULT 389,
    use_tls     BOOLEAN NOT NULL DEFAULT TRUE,
    base_dn     TEXT NOT NULL DEFAULT '',
    bind_dn     TEXT NOT NULL DEFAULT '',
    bind_pw     TEXT NOT NULL DEFAULT '',
    user_filter TEXT NOT NULL DEFAULT '(sAMAccountName=%s)'
);

CREATE TABLE groups (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    parent_id   TEXT REFERENCES groups(id) ON DELETE SET NULL
);

CREATE TABLE devices (
    id              TEXT PRIMARY KEY,
    hostname        TEXT NOT NULL DEFAULT '',
    os              TEXT NOT NULL DEFAULT '',
    os_version      TEXT NOT NULL DEFAULT '',
    vendor          TEXT NOT NULL DEFAULT '',
    model           TEXT NOT NULL DEFAULT '',
    serial          TEXT NOT NULL DEFAULT '',
    cpu_model       TEXT NOT NULL DEFAULT '',
    cpu_cores       INTEGER NOT NULL DEFAULT 0,
    memory_bytes    INTEGER NOT NULL DEFAULT 0,
    first_seen      TIMESTAMP NOT NULL,
    last_seen       TIMESTAMP,
    enrolled_at     TIMESTAMP NOT NULL,
    agent_token_hash TEXT NOT NULL DEFAULT '',
    revoked         BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE INDEX idx_devices_token ON devices(agent_token_hash);

CREATE TABLE interfaces (
    device_id TEXT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    name      TEXT NOT NULL DEFAULT '',
    mac       TEXT NOT NULL DEFAULT '',
    ipv4      TEXT NOT NULL DEFAULT '',
    ipv6      TEXT NOT NULL DEFAULT ''
);
CREATE INDEX idx_interfaces_device ON interfaces(device_id);

CREATE TABLE device_groups (
    device_id TEXT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    group_id  TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    PRIMARY KEY (device_id, group_id)
);

CREATE TABLE inventory (
    id           TEXT PRIMARY KEY,
    device_id    TEXT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    collected_at TIMESTAMP NOT NULL,
    data         TEXT NOT NULL DEFAULT '{}'
);
CREATE INDEX idx_inventory_device ON inventory(device_id, collected_at);

CREATE TABLE enrollment_tokens (
    id         TEXT PRIMARY KEY,
    label      TEXT NOT NULL DEFAULT '',
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMP,
    max_uses   INTEGER NOT NULL DEFAULT 0,
    used_count INTEGER NOT NULL DEFAULT 0,
    created_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL
);
CREATE INDEX idx_enroll_token ON enrollment_tokens(token_hash);

CREATE TABLE commands (
    id         TEXT PRIMARY KEY,
    device_id  TEXT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    type       TEXT NOT NULL,
    payload    TEXT NOT NULL DEFAULT '{}',
    status     TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL,
    sent_at    TIMESTAMP,
    result     TEXT NOT NULL DEFAULT '{}'
);
CREATE INDEX idx_commands_device ON commands(device_id, status);

-- +goose Down
DROP TABLE commands;
DROP TABLE enrollment_tokens;
DROP TABLE inventory;
DROP TABLE device_groups;
DROP TABLE interfaces;
DROP TABLE devices;
DROP TABLE groups;
DROP TABLE ldap_config;
DROP TABLE sessions;
DROP TABLE users;
