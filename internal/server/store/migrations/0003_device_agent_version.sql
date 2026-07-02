-- +goose Up
ALTER TABLE devices ADD COLUMN agent_version TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE devices DROP COLUMN agent_version;
