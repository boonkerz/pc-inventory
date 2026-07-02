-- +goose Up
-- Escrow der BitLocker-Wiederherstellungsschlüssel je Gerät/Volume.
CREATE TABLE bitlocker_keys (
    device_id   TEXT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    mount_point TEXT NOT NULL,
    recovery_id TEXT NOT NULL DEFAULT '',
    recovery_key TEXT NOT NULL DEFAULT '',
    updated_at  TIMESTAMP NOT NULL,
    PRIMARY KEY (device_id, mount_point)
);

-- +goose Down
DROP TABLE bitlocker_keys;
