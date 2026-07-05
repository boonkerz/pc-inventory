-- +goose Up
-- Auslastungs-Historie (leichte Samples je Checkin) für Verlaufscharts.
CREATE TABLE metrics_samples (
    device_id TEXT   NOT NULL,
    ts        BIGINT NOT NULL, -- Unix-Millisekunden
    cpu       REAL   NOT NULL,
    mem       REAL   NOT NULL,
    disk      REAL   NOT NULL
);
CREATE INDEX idx_metrics_device_ts ON metrics_samples(device_id, ts);

-- +goose Down
DROP TABLE metrics_samples;
