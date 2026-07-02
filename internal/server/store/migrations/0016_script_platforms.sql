-- +goose Up
-- Plattformen, auf denen ein Skript laufen darf (JSON-Array: windows|linux|darwin).
-- '[]' = keine Einschränkung über die Shell hinaus.
ALTER TABLE scripts ADD COLUMN platforms TEXT NOT NULL DEFAULT '[]';

-- +goose Down
ALTER TABLE scripts DROP COLUMN platforms;
