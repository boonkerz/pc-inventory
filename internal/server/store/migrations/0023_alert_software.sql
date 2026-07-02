-- +goose Up
ALTER TABLE alert_config ADD COLUMN alert_software BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE alert_config DROP COLUMN alert_software;
