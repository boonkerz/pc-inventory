-- +goose Up
CREATE TABLE clients (
    id   TEXT PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE sites (
    id        TEXT PRIMARY KEY,
    client_id TEXT NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    name      TEXT NOT NULL
);
CREATE INDEX idx_sites_client ON sites(client_id);

-- Geräte/Tokens verweisen auf eine Site (nullable). FK-Aufräumen erfolgt in Code,
-- da ALTER TABLE ADD COLUMN keine FK-Aktionen zuverlässig nachrüstet.
ALTER TABLE devices ADD COLUMN site_id TEXT;
ALTER TABLE enrollment_tokens ADD COLUMN site_id TEXT;
CREATE INDEX idx_devices_site ON devices(site_id);

-- +goose Down
DROP INDEX idx_devices_site;
ALTER TABLE enrollment_tokens DROP COLUMN site_id;
ALTER TABLE devices DROP COLUMN site_id;
DROP TABLE sites;
DROP TABLE clients;
