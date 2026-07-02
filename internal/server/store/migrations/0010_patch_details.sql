-- +goose Up
ALTER TABLE os_updates ADD COLUMN severity TEXT NOT NULL DEFAULT 'Other';
ALTER TABLE os_updates ADD COLUMN url      TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE os_updates DROP COLUMN url;
ALTER TABLE os_updates DROP COLUMN severity;
