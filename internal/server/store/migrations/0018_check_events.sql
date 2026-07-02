-- +goose Up
-- Verlauf der Check-Statuswechsel (gut <-> Warnung <-> Fehler) inkl. ob/wann
-- darüber benachrichtigt wurde. check_name wird als Schnappschuss gehalten,
-- damit der Verlauf auch nach Umbenennen/Löschen des Checks lesbar bleibt.
CREATE TABLE check_events (
    id          TEXT PRIMARY KEY,
    device_id   TEXT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    check_id    TEXT NOT NULL,
    check_name  TEXT NOT NULL DEFAULT '',
    old_status  TEXT NOT NULL,
    new_status  TEXT NOT NULL,
    output      TEXT NOT NULL DEFAULT '',
    notified    BOOLEAN NOT NULL DEFAULT FALSE,
    notified_at TIMESTAMP,
    created_at  TIMESTAMP NOT NULL
);
CREATE INDEX idx_check_events_device ON check_events(device_id, created_at);

-- +goose Down
DROP TABLE check_events;
