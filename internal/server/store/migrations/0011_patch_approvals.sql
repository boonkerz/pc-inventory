-- +goose Up
-- Genehmigungen pro Gerät+Patch. Eigene Tabelle, da os_updates bei jedem Checkin
-- ersetzt wird – die Genehmigung soll das Re-Scannen überleben.
CREATE TABLE patch_approvals (
    device_id TEXT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    name      TEXT NOT NULL,
    PRIMARY KEY (device_id, name)
);

-- +goose Down
DROP TABLE patch_approvals;
