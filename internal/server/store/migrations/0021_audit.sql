-- +goose Up
-- Audit-Log: wer hat wann welche ändernde Aktion ausgelöst.
CREATE TABLE audit_log (
    id       TEXT PRIMARY KEY,
    ts       TIMESTAMP NOT NULL,
    user_id  TEXT,
    username TEXT NOT NULL DEFAULT '',
    action   TEXT NOT NULL DEFAULT '',
    method   TEXT NOT NULL DEFAULT '',
    path     TEXT NOT NULL DEFAULT '',
    status   INTEGER NOT NULL DEFAULT 0,
    ip       TEXT NOT NULL DEFAULT ''
);
CREATE INDEX idx_audit_ts ON audit_log(ts);

-- +goose Down
DROP TABLE audit_log;
