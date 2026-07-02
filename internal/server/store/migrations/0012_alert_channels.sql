-- +goose Up
CREATE TABLE alert_channels (
    id      TEXT PRIMARY KEY,
    type    TEXT NOT NULL,
    name    TEXT NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    config  TEXT NOT NULL DEFAULT '{}'  -- JSON map[string]string
);

-- +goose Down
DROP TABLE alert_channels;
