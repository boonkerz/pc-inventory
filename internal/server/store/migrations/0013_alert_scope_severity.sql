-- +goose Up
-- Schweregrad je Check (warning | critical); Standard kritisch.
ALTER TABLE policy_checks ADD COLUMN severity TEXT NOT NULL DEFAULT 'critical';

-- Mindest-Schweregrad je Kanal (warning = alles, critical = nur kritische).
-- Standard 'warning' = bisheriges Verhalten (jeder Fehler benachrichtigt).
ALTER TABLE alert_channels ADD COLUMN min_severity TEXT NOT NULL DEFAULT 'warning';

-- Geltungsbereich je Kanal (Client/Site/Gerät). Keine Einträge = global.
CREATE TABLE alert_channel_assignments (
    id          TEXT PRIMARY KEY,
    channel_id  TEXT NOT NULL REFERENCES alert_channels(id) ON DELETE CASCADE,
    target_type TEXT NOT NULL,  -- client | site | device
    target_id   TEXT NOT NULL
);
CREATE INDEX idx_alert_assign_channel ON alert_channel_assignments(channel_id);
CREATE INDEX idx_alert_assign_target ON alert_channel_assignments(target_type, target_id);

-- +goose Down
DROP TABLE alert_channel_assignments;
ALTER TABLE alert_channels DROP COLUMN min_severity;
ALTER TABLE policy_checks DROP COLUMN severity;
