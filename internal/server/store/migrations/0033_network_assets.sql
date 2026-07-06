-- +goose Up
-- Beim Netzwerk-Scan gefundene (i.d.R. nicht verwaltete) Hosts, einer Site zugeordnet.
CREATE TABLE network_assets (
    id         TEXT PRIMARY KEY,
    site_id    TEXT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    ip         TEXT NOT NULL DEFAULT '',
    mac        TEXT NOT NULL DEFAULT '',
    hostname   TEXT NOT NULL DEFAULT '',
    ports      TEXT NOT NULL DEFAULT '',
    note       TEXT NOT NULL DEFAULT '',
    first_seen TIMESTAMP NOT NULL,
    last_seen  TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX idx_netasset_site_ip ON network_assets(site_id, ip);

-- +goose Down
DROP TABLE network_assets;
