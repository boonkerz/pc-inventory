-- +goose Up
-- Theme-Präferenz pro Benutzer ('' = nicht gesetzt -> Frontend nutzt System/lokal).
ALTER TABLE users ADD COLUMN theme TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE users DROP COLUMN theme;
