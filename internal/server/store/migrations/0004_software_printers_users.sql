-- +goose Up
CREATE TABLE software (
    device_id TEXT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    name      TEXT NOT NULL DEFAULT '',
    version   TEXT NOT NULL DEFAULT '',
    publisher TEXT NOT NULL DEFAULT ''
);
CREATE INDEX idx_software_device ON software(device_id);

CREATE TABLE printers (
    device_id  TEXT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    name       TEXT NOT NULL DEFAULT '',
    driver     TEXT NOT NULL DEFAULT '',
    port       TEXT NOT NULL DEFAULT '',
    is_default BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE INDEX idx_printers_device ON printers(device_id);

ALTER TABLE devices ADD COLUMN logged_in_users TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE devices DROP COLUMN logged_in_users;
DROP TABLE printers;
DROP TABLE software;
