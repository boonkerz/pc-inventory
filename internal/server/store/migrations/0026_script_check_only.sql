-- +goose Up
-- Markiert Skripte, die nur als Check dienen (nicht in Ausführen/Sammelaktion zeigen).
ALTER TABLE scripts ADD COLUMN check_only BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE scripts DROP COLUMN check_only;
