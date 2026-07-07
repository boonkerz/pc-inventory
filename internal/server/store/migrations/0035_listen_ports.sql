-- +goose Up
-- Lauschende Sockets je Gerät (Angriffsfläche / „nach außen offen"). Wird bei jedem
-- Checkin vollständig ersetzt (Momentaufnahme, keine Historie).
CREATE TABLE listen_ports (
    device_id  TEXT    NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    proto      TEXT    NOT NULL,
    address    TEXT    NOT NULL,
    port       INTEGER NOT NULL,
    process    TEXT    NOT NULL DEFAULT '',
    pid        INTEGER NOT NULL DEFAULT 0,
    public     BOOLEAN NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL
);
CREATE INDEX idx_listen_ports_device ON listen_ports(device_id);
CREATE INDEX idx_listen_ports_public ON listen_ports(public, port);

-- +goose Down
DROP TABLE listen_ports;
