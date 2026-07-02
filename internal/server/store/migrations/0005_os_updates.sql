-- +goose Up
-- NULL = noch nicht geprüft (Status unbekannt).
ALTER TABLE devices ADD COLUMN updates_count INTEGER;
ALTER TABLE devices ADD COLUMN updates_checked_at TIMESTAMP;

CREATE TABLE os_updates (
    device_id TEXT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    package   TEXT NOT NULL DEFAULT ''
);
CREATE INDEX idx_os_updates_device ON os_updates(device_id);

-- +goose Down
DROP TABLE os_updates;
ALTER TABLE devices DROP COLUMN updates_checked_at;
ALTER TABLE devices DROP COLUMN updates_count;
