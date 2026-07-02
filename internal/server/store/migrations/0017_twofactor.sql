-- +goose Up
ALTER TABLE users ADD COLUMN totp_secret  TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN totp_enabled BOOLEAN NOT NULL DEFAULT FALSE;

-- Einmal-Wiederherstellungscodes (nur Hashes gespeichert).
CREATE TABLE recovery_codes (
    user_id   TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash TEXT NOT NULL,
    used      BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (user_id, code_hash)
);

-- Kurzlebige Login-Challenges für den zweiten Faktor (zwischen Passwort und Code).
CREATE TABLE login_challenges (
    token_hash TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE login_challenges;
DROP TABLE recovery_codes;
ALTER TABLE users DROP COLUMN totp_enabled;
ALTER TABLE users DROP COLUMN totp_secret;
