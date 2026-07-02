-- +goose Up
ALTER TABLE policy_tasks ADD COLUMN schedule_type TEXT NOT NULL DEFAULT 'interval'; -- interval | daily
ALTER TABLE policy_tasks ADD COLUMN daily_time TEXT NOT NULL DEFAULT '';            -- HH:MM bei daily
ALTER TABLE policy_tasks ADD COLUMN weekdays   TEXT NOT NULL DEFAULT '';            -- "" = alle, sonst z.B. 1,3,5

CREATE TABLE alert_config (
    id          INTEGER PRIMARY KEY,
    enabled     BOOLEAN NOT NULL DEFAULT FALSE,
    smtp_host   TEXT NOT NULL DEFAULT '',
    smtp_port   INTEGER NOT NULL DEFAULT 587,
    smtp_user   TEXT NOT NULL DEFAULT '',
    smtp_pass   TEXT NOT NULL DEFAULT '',
    smtp_from   TEXT NOT NULL DEFAULT '',
    smtp_tls    BOOLEAN NOT NULL DEFAULT TRUE,
    recipient   TEXT NOT NULL DEFAULT '',
    webhook_url TEXT NOT NULL DEFAULT ''
);

-- +goose Down
DROP TABLE alert_config;
ALTER TABLE policy_tasks DROP COLUMN weekdays;
ALTER TABLE policy_tasks DROP COLUMN daily_time;
ALTER TABLE policy_tasks DROP COLUMN schedule_type;
