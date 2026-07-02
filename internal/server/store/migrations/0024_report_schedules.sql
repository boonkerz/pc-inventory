-- +goose Up
-- Geplante Berichte: rendern regelmäßig einen Report und versenden ihn per Alarm-Kanal.
CREATE TABLE report_schedules (
    id         TEXT PRIMARY KEY,
    title      TEXT NOT NULL DEFAULT '',
    frequency  TEXT NOT NULL,                -- daily | weekly | monthly
    channel_id TEXT NOT NULL,                -- Alarm-Kanal (i.d.R. E-Mail)
    last_run   TIMESTAMP,
    created_at TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE report_schedules;
