-- +goose Up
-- Protokoll der Software-Änderungen je Gerät (Diff zwischen zwei Checkins).
CREATE TABLE software_events (
    id          TEXT PRIMARY KEY,
    device_id   TEXT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    change      TEXT NOT NULL,              -- added | removed | updated
    name        TEXT NOT NULL,
    version     TEXT NOT NULL DEFAULT '',
    old_version TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMP NOT NULL
);
CREATE INDEX idx_software_events_device ON software_events(device_id, created_at);

-- +goose Down
DROP TABLE software_events;
