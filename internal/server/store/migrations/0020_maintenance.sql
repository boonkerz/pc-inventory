-- +goose Up
-- Wartungsfenster: unterdrücken Alarme für ein Ziel (Client/Site/Device) im Zeitraum.
CREATE TABLE maintenance_windows (
    id          TEXT PRIMARY KEY,
    target_type TEXT NOT NULL,              -- client | site | device
    target_id   TEXT NOT NULL,
    note        TEXT NOT NULL DEFAULT '',
    starts_at   TIMESTAMP NOT NULL,
    ends_at     TIMESTAMP NOT NULL,
    created_at  TIMESTAMP NOT NULL
);
CREATE INDEX idx_maintenance_window ON maintenance_windows(ends_at);

-- +goose Down
DROP TABLE maintenance_windows;
