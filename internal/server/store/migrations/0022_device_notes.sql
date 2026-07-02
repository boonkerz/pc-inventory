-- +goose Up
ALTER TABLE devices ADD COLUMN notes TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE devices DROP COLUMN notes;
