-- +goose Up
ALTER TABLE devices ADD COLUMN cpu_sockets    INTEGER NOT NULL DEFAULT 0;
ALTER TABLE devices ADD COLUMN cpu_threads    INTEGER NOT NULL DEFAULT 0;
ALTER TABLE devices ADD COLUMN public_ip      TEXT NOT NULL DEFAULT '';
ALTER TABLE devices ADD COLUMN disks          TEXT NOT NULL DEFAULT '';   -- JSON
ALTER TABLE devices ADD COLUMN physical_disks TEXT NOT NULL DEFAULT '';   -- JSON
ALTER TABLE devices ADD COLUMN gpus           TEXT NOT NULL DEFAULT '';   -- JSON

-- +goose Down
ALTER TABLE devices DROP COLUMN gpus;
ALTER TABLE devices DROP COLUMN physical_disks;
ALTER TABLE devices DROP COLUMN disks;
ALTER TABLE devices DROP COLUMN public_ip;
ALTER TABLE devices DROP COLUMN cpu_threads;
ALTER TABLE devices DROP COLUMN cpu_sockets;
